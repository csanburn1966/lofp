package engine

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// GMScript represents a GM-uploaded script stored in MongoDB.
type GMScript struct {
	Filename            string           `bson:"_id" json:"filename"`
	Name                string           `bson:"name" json:"name"`
	Content             string           `bson:"content" json:"content"`
	Priority            int              `bson:"priority" json:"priority"` // higher = loads sooner
	// Draft scripts are stored (and readable/editable in the GM panel) but are not
	// applied to the running game and are skipped at startup. Missing in older
	// documents, so the zero value (false) means "live" — existing scripts stay live.
	Draft               bool             `bson:"draft" json:"draft"`
	Size                int              `bson:"size" json:"size"`
	UploadedBy          string           `bson:"uploadedBy" json:"uploadedBy"`
	UploadedByAccountID string           `bson:"uploadedByAccountId" json:"uploadedByAccountId"`
	UploadedAt          time.Time        `bson:"uploadedAt" json:"uploadedAt"`
	ParseStats          ScriptApplyStats `bson:"parseStats" json:"parseStats"`
	History             []GMScriptVersion `bson:"history" json:"history"`
}

// GMScriptVersion is a historical version of a script.
type GMScriptVersion struct {
	Content    string    `bson:"content" json:"content"`
	UploadedBy string    `bson:"uploadedBy" json:"uploadedBy"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploadedAt"`
}

const gmScriptsCollection = "gm_scripts"
const maxScriptHistory = 10
const MaxScriptSize = 268245 // 110% of largest disk script (MONSTERS.SCR at 243,859 bytes)

// ListGMScripts returns all GM scripts sorted by priority (descending) then name.
func (e *GameEngine) ListGMScripts(ctx context.Context) ([]GMScript, error) {
	if e.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	opts := options.Find().SetSort(bson.D{{Key: "priority", Value: -1}, {Key: "name", Value: 1}})
	// Don't include content or history in list — just metadata
	opts.SetProjection(bson.M{"content": 0, "history": 0})
	cursor, err := e.db.Collection(gmScriptsCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	var scripts []GMScript
	if err := cursor.All(ctx, &scripts); err != nil {
		return nil, err
	}
	return scripts, nil
}

// GetGMScript returns a single GM script by filename.
func (e *GameEngine) GetGMScript(ctx context.Context, filename string) (*GMScript, error) {
	if e.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	var script GMScript
	err := e.db.Collection(gmScriptsCollection).FindOne(ctx, bson.M{"_id": filename}).Decode(&script)
	if err != nil {
		return nil, err
	}
	return &script, nil
}

// SaveGMScript upserts a GM script, pushing the previous version to history.
func (e *GameEngine) SaveGMScript(ctx context.Context, script *GMScript) error {
	if e.db == nil {
		return fmt.Errorf("database not available")
	}
	// Check if exists, push old content to history
	existing, err := e.GetGMScript(ctx, script.Filename)
	if err == nil && existing.Content != "" {
		script.History = existing.History
		// Push previous version to front of history
		prev := GMScriptVersion{
			Content:    existing.Content,
			UploadedBy: existing.UploadedBy,
			UploadedAt: existing.UploadedAt,
		}
		script.History = append([]GMScriptVersion{prev}, script.History...)
		// Trim to max history
		if len(script.History) > maxScriptHistory {
			script.History = script.History[:maxScriptHistory]
		}
	}

	opts := options.Replace().SetUpsert(true)
	_, err = e.db.Collection(gmScriptsCollection).ReplaceOne(ctx, bson.M{"_id": script.Filename}, script, opts)
	return err
}

// SetGMScriptDraft flips a stored script between draft and live without touching
// its content or version history. The caller applies/withdraws it in the engine.
func (e *GameEngine) SetGMScriptDraft(ctx context.Context, filename string, draft bool) error {
	if e.db == nil {
		return fmt.Errorf("database not available")
	}
	res, err := e.db.Collection(gmScriptsCollection).UpdateOne(ctx, bson.M{"_id": filename}, bson.M{"$set": bson.M{"draft": draft}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("script %q not found", filename)
	}
	return nil
}

// DeleteGMScript removes a GM script from MongoDB.
func (e *GameEngine) DeleteGMScript(ctx context.Context, filename string) error {
	if e.db == nil {
		return fmt.Errorf("database not available")
	}
	_, err := e.db.Collection(gmScriptsCollection).DeleteOne(ctx, bson.M{"_id": filename})
	return err
}

// LoadGMScripts loads all GM scripts from MongoDB and applies them to the engine.
// Called at startup after disk scripts are loaded.
func (e *GameEngine) LoadGMScripts() {
	if e.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Sort by priority desc (higher loads first), then by uploadedAt asc
	opts := options.Find().SetSort(bson.D{{Key: "priority", Value: -1}, {Key: "uploadedAt", Value: 1}})
	cursor, err := e.db.Collection(gmScriptsCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		log.Printf("Warning: could not load GM scripts: %v", err)
		return
	}
	var scripts []GMScript
	if err := cursor.All(ctx, &scripts); err != nil {
		log.Printf("Warning: could not decode GM scripts: %v", err)
		return
	}
	if len(scripts) == 0 {
		return
	}
	log.Printf("Loading %d GM scripts from MongoDB...", len(scripts))
	for _, s := range scripts {
		if s.Draft {
			log.Printf("  %s: draft, not loaded", s.Filename)
			continue
		}
		parsed, err := e.parseAndApplyScript(s.Content, s.Filename)
		if err != nil {
			log.Printf("Warning: GM script %q failed to parse: %v", s.Filename, err)
			continue
		}
		log.Printf("  %s (priority %d): %d rooms, %d items, %d monsters",
			s.Filename, s.Priority, parsed.Rooms, parsed.Items, parsed.Monsters)
	}
}

// parseAndApplyScript parses script content and applies it to the engine.
// Returns stats of what was applied. Imported here to avoid circular deps —
// caller must pass content, this calls out to scriptparser via a callback.
var ScriptParser func(content, filename string) (*gameworld.ParsedData, error)

func (e *GameEngine) parseAndApplyScript(content, filename string) (ScriptApplyStats, error) {
	if ScriptParser == nil {
		return ScriptApplyStats{}, fmt.Errorf("script parser not initialized")
	}
	parsed, err := ScriptParser(content, filename)
	if err != nil {
		return ScriptApplyStats{}, err
	}
	stats := e.ApplyParsedData(parsed)
	return stats, nil
}

// ScriptApplyStats summarizes what a hot-loaded script changed.
type ScriptApplyStats struct {
	Rooms     int `json:"rooms"`
	Items     int `json:"items"`
	Monsters  int `json:"monsters"`
	Nouns     int `json:"nouns"`
	Variables int `json:"variables"`
}

// ApplyParsedData hot-loads parsed script data into the running engine.
// Entities are added or overwritten by number. This is additive — it never removes entities.
func (e *GameEngine) ApplyParsedData(parsed *gameworld.ParsedData) ScriptApplyStats {
	var stats ScriptApplyStats

	for i := range parsed.Rooms {
		e.rooms[parsed.Rooms[i].Number] = &parsed.Rooms[i]
		stats.Rooms++
	}
	for i := range parsed.Items {
		e.items[parsed.Items[i].Number] = &parsed.Items[i]
		stats.Items++
	}
	for i := range parsed.Monsters {
		e.monsters[parsed.Monsters[i].Number] = &parsed.Monsters[i]
		stats.Monsters++
	}
	for i := range parsed.Nouns {
		e.nouns[parsed.Nouns[i].ID] = parsed.Nouns[i].Name
		stats.Nouns++
	}
	for i := range parsed.Adjectives {
		e.adjectives[parsed.Adjectives[i].ID] = parsed.Adjectives[i].Name
	}
	for i := range parsed.MonsterAdjs {
		e.monAdjs[parsed.MonsterAdjs[i].ID] = parsed.MonsterAdjs[i].Name
	}
	for i := range parsed.BreakMods {
		e.breakMods[parsed.BreakMods[i].AdjID] = parsed.BreakMods[i].Modifier
	}
	for i := range parsed.MoneyDefs {
		if md := parsed.MoneyDefs[i]; md.ID >= 1 && md.ID <= 9 && md.ExchangeRate > 0 {
			if e.moneyDefs == nil {
				e.moneyDefs = map[int]*gameworld.MoneyDef{}
			}
			e.moneyDefs[md.ID] = &md
		}
	}
	for _, v := range parsed.Variables {
		name := strings.ToUpper(v.Name)
		if !e.namedVarNames[name] {
			e.namedVarNames[name] = true
			e.NamedVars[name] = 0
			stats.Variables++
		}
	}
	if len(parsed.ForageDefs) > 0 {
		e.forageDefs = append(e.forageDefs, parsed.ForageDefs...)
	}
	// A CEVENT with the same ID and room as one already loaded replaces it rather than
	// running alongside it — otherwise re-uploading an edited GM script (e.g. tweaking
	// FALLFEST.SCR mid-event) would double every ambient echo and spawner it defines.
	for _, ce := range parsed.CEvents {
		replaced := false
		for i := range e.cevents {
			if e.cevents[i].ID == ce.ID && e.cevents[i].Room == ce.Room {
				e.cevents[i] = ce
				replaced = true
				break
			}
		}
		if !replaced {
			e.cevents = append(e.cevents, ce)
		}
	}
	for i := range parsed.Macros {
		e.macros[parsed.Macros[i].ID] = parsed.Macros[i].Scripts
	}
	for i := range parsed.OrgDefs {
		def := &parsed.OrgDefs[i]
		e.orgDefs[def.Number] = def
	}

	if stats.Rooms > 0 || stats.Items > 0 || stats.Monsters > 0 {
		log.Printf("Hot-loaded script: %d rooms, %d items, %d monsters, %d nouns, %d vars",
			stats.Rooms, stats.Items, stats.Monsters, stats.Nouns, stats.Variables)
	}

	return stats
}

// loadPVals loads persistent global values from MongoDB.
func (e *GameEngine) loadPVals() {
	if e.db == nil {
		return
	}
	ctx := context.Background()
	var doc struct {
		Vals map[string]int `bson:"vals"`
	}
	err := e.db.Collection("pvals").FindOne(ctx, bson.M{"_id": "pvals"}).Decode(&doc)
	if err != nil {
		return // no saved pvals yet
	}
	for k, v := range doc.Vals {
		idx, err := strconv.Atoi(k)
		if err == nil {
			e.PVals[idx] = v
		}
	}
}

// savePVals saves persistent global values to MongoDB.
func (e *GameEngine) savePVals() {
	if e.db == nil {
		return
	}
	ctx := context.Background()
	vals := make(map[string]int)
	for k, v := range e.PVals {
		vals[strconv.Itoa(k)] = v
	}
	_, _ = e.db.Collection("pvals").ReplaceOne(ctx, bson.M{"_id": "pvals"}, bson.M{"_id": "pvals", "vals": vals}, options.Replace().SetUpsert(true))
}

// StartCEventLoop starts the background goroutine that fires cyclic events.
func (e *GameEngine) StartCEventLoop() {
	if len(e.cevents) == 0 {
		return
	}
	go func() {
		tick := 0
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			tick++
			for _, ce := range e.cevents {
				// Cycles<=0 (e.g. "CEVENT 91 0 350") means fire every tick, not never —
				// several original scripts (KREETEK's subway timer, EVILFAYD's ambient
				// echoes) rely on a 0 interval and roll their own internal randomization
				// or counters for pacing.
				if ce.Cycles <= 0 || tick%ce.Cycles == 0 {
					e.runCEvent(ce)
				}
			}
		}
	}()
}

// runCEvent fires one cyclic event. Split out of StartCEventLoop so a panic in one
// event's script is logged (recoverBackground) and skipped instead of killing the
// whole server.
func (e *GameEngine) runCEvent(ce gameworld.CEvent) {
	defer recoverBackground(fmt.Sprintf("CEVENT %d (room %d)", ce.ID, ce.Room))
	room := e.rooms[ce.Room]
	if room == nil {
		return
	}
	sc := &ScriptContext{Room: room, Engine: e, Player: &Player{}}
	for _, block := range ce.Scripts {
		if block.Type == "ACTION" {
			for _, action := range block.Actions {
				sc.execAction(action)
			}
		} else {
			sc.execBlock(block)
		}
	}
	// Deliver ECHO messages from CEVENT scripts to players in the room
	if e.roomBroadcast != nil && len(sc.RoomMsgs) > 0 {
		e.roomBroadcast(ce.Room, sc.RoomMsgs)
	}
	e.Events.Publish("cevent", fmt.Sprintf("CEVENT %d fired in room %d (%s)", ce.ID, ce.Room, room.Name))
}

// ── Drafts: withdrawing a live GM script from the running game ───────────────
//
// Applying a GM script is additive (ApplyParsedData replaces entities by number).
// Withdrawing one — disabling it to a draft, deleting it, or re-saving it without
// some of its old entities — puts back what the game would have without it:
// each room/item/monster/macro/CEVENT it defined reverts to the disk-script
// definition captured at startup (or disappears if the disk scripts never had
// it), and then any other live GM script that also defines that entity is laid
// back on top, in normal load order. Vocabulary, named variables, monster
// adjectives, break mods, money/org/forage definitions are left alone: they're
// harmless leftovers and clear on the next restart.

// snapshotBaseDefinitions records the disk-script definitions, before any GM
// script is applied. Called once from NewGameEngine.
func (e *GameEngine) snapshotBaseDefinitions() {
	e.baseRooms = make(map[int]*gameworld.Room, len(e.rooms))
	for n, r := range e.rooms {
		e.baseRooms[n] = r
	}
	e.baseItems = make(map[int]*gameworld.ItemDef, len(e.items))
	for n, it := range e.items {
		e.baseItems[n] = it
	}
	e.baseMonsters = make(map[int]*gameworld.MonsterDef, len(e.monsters))
	for n, m := range e.monsters {
		e.baseMonsters[n] = m
	}
	e.baseMacros = make(map[int][]gameworld.ScriptBlock, len(e.macros))
	for n, m := range e.macros {
		e.baseMacros[n] = m
	}
	e.baseCEvents = append([]gameworld.CEvent(nil), e.cevents...)
}

type ceventKey struct{ id, room int }

// scriptEntitySet is the set of entity numbers a parsed script defines.
type scriptEntitySet struct {
	rooms, items, monsters, macros map[int]bool
	cevents                        map[ceventKey]bool
}

func entitySetOf(p *gameworld.ParsedData) scriptEntitySet {
	s := scriptEntitySet{map[int]bool{}, map[int]bool{}, map[int]bool{}, map[int]bool{}, map[ceventKey]bool{}}
	if p == nil {
		return s
	}
	for i := range p.Rooms {
		s.rooms[p.Rooms[i].Number] = true
	}
	for i := range p.Items {
		s.items[p.Items[i].Number] = true
	}
	for i := range p.Monsters {
		s.monsters[p.Monsters[i].Number] = true
	}
	for i := range p.Macros {
		s.macros[p.Macros[i].ID] = true
	}
	for _, ce := range p.CEvents {
		s.cevents[ceventKey{ce.ID, ce.Room}] = true
	}
	return s
}

// minus removes everything o defines from s (entities a re-saved script still
// defines don't need reverting — the new version is applied over them).
func (s scriptEntitySet) minus(o scriptEntitySet) scriptEntitySet {
	for n := range o.rooms {
		delete(s.rooms, n)
	}
	for n := range o.items {
		delete(s.items, n)
	}
	for n := range o.monsters {
		delete(s.monsters, n)
	}
	for n := range o.macros {
		delete(s.macros, n)
	}
	for k := range o.cevents {
		delete(s.cevents, k)
	}
	return s
}

// WithdrawStats summarizes what withdrawing a GM script put back.
type WithdrawStats struct {
	Rooms     int `json:"rooms"`
	Items     int `json:"items"`
	Monsters  int `json:"monsters"`
	Macros    int `json:"macros"`
	CEvents   int `json:"cevents"`
	Despawned int `json:"despawned"` // creatures removed because their definition went away
}

// ParsedStats counts what a parsed script defines, without applying it (shown
// for drafts, where nothing is applied).
func ParsedStats(p *gameworld.ParsedData) ScriptApplyStats {
	return ScriptApplyStats{Rooms: len(p.Rooms), Items: len(p.Items), Monsters: len(p.Monsters),
		Nouns: len(p.Nouns), Variables: len(p.Variables)}
}

// WithdrawGMScript removes a live GM script's effects from the running game.
// oldContent is the version currently applied; keep (optional) is a new version
// about to be applied in its place — entities it still defines are left for it to
// overwrite. Other live GM scripts are read from MongoDB so their definitions win
// back any entity they share with the withdrawn one.
func (e *GameEngine) WithdrawGMScript(ctx context.Context, filename, oldContent string, keep *gameworld.ParsedData) (WithdrawStats, error) {
	if ScriptParser == nil {
		return WithdrawStats{}, fmt.Errorf("script parser not initialized")
	}
	old, err := ScriptParser(oldContent, filename)
	if err != nil {
		return WithdrawStats{}, fmt.Errorf("parse previous version: %w", err)
	}
	set := entitySetOf(old)
	if keep != nil {
		set = set.minus(entitySetOf(keep))
	}

	// Other live scripts, in the same order LoadGMScripts applies them.
	var others []*gameworld.ParsedData
	if e.db != nil {
		opts := options.Find().SetSort(bson.D{{Key: "priority", Value: -1}, {Key: "uploadedAt", Value: 1}})
		cursor, err := e.db.Collection(gmScriptsCollection).Find(ctx, bson.M{"_id": bson.M{"$ne": filename}, "draft": bson.M{"$ne": true}}, opts)
		if err != nil {
			return WithdrawStats{}, err
		}
		var scripts []GMScript
		if err := cursor.All(ctx, &scripts); err != nil {
			return WithdrawStats{}, err
		}
		for _, sc := range scripts {
			p, err := ScriptParser(sc.Content, sc.Filename)
			if err != nil {
				log.Printf("Warning: withdraw %s: skipping unparseable GM script %s: %v", filename, sc.Filename, err)
				continue
			}
			others = append(others, p)
		}
	}
	stats := e.revertScriptEntities(set, others)
	log.Printf("Withdrew GM script %s: %d rooms, %d items, %d monsters, %d macros, %d cevents reverted; %d creatures despawned",
		filename, stats.Rooms, stats.Items, stats.Monsters, stats.Macros, stats.CEvents, stats.Despawned)
	return stats, nil
}

// revertScriptEntities restores every entity in set to its disk definition (or
// removes it), then re-applies matching definitions from others in order.
func (e *GameEngine) revertScriptEntities(set scriptEntitySet, others []*gameworld.ParsedData) WithdrawStats {
	var stats WithdrawStats
	seasonal := e.seasonalRooms[e.currentSeason]

	for n := range set.rooms {
		stats.Rooms++
		base := e.baseRooms[n]
		var override *gameworld.Room
		for i := range seasonal {
			if seasonal[i].Number == n {
				override = &seasonal[i]
				break
			}
		}
		switch {
		case base == nil && override == nil:
			delete(e.rooms, n)
			continue
		case base == nil:
			// A room only a seasonal script defines comes back as that.
			base = override
		case override != nil && override != base:
			// The season may have changed while the GM version was in place.
			applySeasonalRoomOverride(base, override)
		}
		e.rooms[n] = base
	}
	for n := range set.items {
		stats.Items++
		if base := e.baseItems[n]; base != nil {
			e.items[n] = base
		} else {
			delete(e.items, n)
		}
	}
	for n := range set.macros {
		stats.Macros++
		if base, ok := e.baseMacros[n]; ok {
			e.macros[n] = base
		} else {
			delete(e.macros, n)
		}
	}
	for n := range set.monsters {
		stats.Monsters++
		if base := e.baseMonsters[n]; base != nil {
			e.monsters[n] = base
		} else {
			delete(e.monsters, n)
		}
	}
	if len(set.cevents) > 0 {
		var kept []gameworld.CEvent
		for _, ce := range e.cevents {
			if set.cevents[ceventKey{ce.ID, ce.Room}] {
				stats.CEvents++
				continue
			}
			kept = append(kept, ce)
		}
		for _, ce := range e.baseCEvents {
			if set.cevents[ceventKey{ce.ID, ce.Room}] {
				kept = append(kept, ce)
			}
		}
		e.cevents = kept
	}

	// Lay other live scripts back over anything they share with the withdrawn one.
	for _, p := range others {
		for i := range p.Rooms {
			if set.rooms[p.Rooms[i].Number] {
				e.rooms[p.Rooms[i].Number] = &p.Rooms[i]
			}
		}
		for i := range p.Items {
			if set.items[p.Items[i].Number] {
				e.items[p.Items[i].Number] = &p.Items[i]
			}
		}
		for i := range p.Monsters {
			if set.monsters[p.Monsters[i].Number] {
				e.monsters[p.Monsters[i].Number] = &p.Monsters[i]
			}
		}
		for i := range p.Macros {
			if set.macros[p.Macros[i].ID] {
				e.macros[p.Macros[i].ID] = p.Macros[i].Scripts
			}
		}
		for _, ce := range p.CEvents {
			if !set.cevents[ceventKey{ce.ID, ce.Room}] {
				continue
			}
			replaced := false
			for i := range e.cevents {
				if e.cevents[i].ID == ce.ID && e.cevents[i].Room == ce.Room {
					e.cevents[i] = ce
					replaced = true
					break
				}
			}
			if !replaced {
				e.cevents = append(e.cevents, ce)
			}
		}
	}

	// Creatures whose definition no longer exists anywhere are removed.
	if e.monsterMgr != nil {
		for n := range set.monsters {
			if e.monsters[n] == nil {
				stats.Despawned += e.monsterMgr.RemoveByDef(n)
			}
		}
	}
	return stats
}
