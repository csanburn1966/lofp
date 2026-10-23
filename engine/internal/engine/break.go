package engine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// BREAK <object> WITH <object> (LEGENDS.DOC: "Allows you to destroy an item with
// another item of greater strength. Synonym: DESTROY").
//
// The original formula is lost; this is reconstructed from GMSCRIPT.DOC's SUBSTANCE
// ladder and the captures (csanburn/chandra.txt, bog.txt). In those, a SOFTMETAL medal
// broke a SOFTMETAL chest (item 87) only after ~20 tries — so an equal substance is
// breakable but rare, and the 5-second round time (2 when hasted) is charged per try.
// A broken container is gone; its contents spill onto the floor, its trap fires
// ("You hear gas being released..."), and some of what was inside gets damaged.

const breakRoundTime = 5

// substanceStrength ranks GMSCRIPT.DOC's substance types from weakest to strongest.
// FUR/HIDE aren't in the doc's list but appear in the scripts; they sit with CLOTH,
// HARDSTONE with SOFTMETAL. An undefined substance is INDESTRUCTABLE per the doc.
var substanceStrength = map[string]int{
	"INSUBSTANTIAL": 0,
	"PAPER":         1,
	"CLOTH":         2,
	"FUR":           2,
	"HIDE":          2,
	"BRITTLE":       3,
	"WOOD":          4,
	"STONE":         5,
	"HARDSTONE":     6,
	"SOFTMETAL":     6,
	"HARDMETAL":     7,
}

// breakChanceByMargin is the per-try chance by how many substance ranks the tool
// outclasses the target (0 = same substance, the medal-vs-chest case). Tunable guesses.
var breakChanceByMargin = []int{5, 20, 40, 60}

const (
	breakHeavyToolBonus = 10 // hammers, maces, axes — "using a hammer or axe seemed to help"
	breakGemDamagePct   = 50 // per gem inside a broken container
)

func substanceRank(def *gameworld.ItemDef) (int, bool) {
	rank, ok := substanceStrength[strings.ToUpper(def.Substance)]
	return rank, ok
}

// isHeavyBreakingTool reports whether a tool gets breakHeavyToolBonus: crush weapons
// and anything axe-like (axe, battleaxe, handaxe...).
func (e *GameEngine) isHeavyBreakingTool(def *gameworld.ItemDef) bool {
	if def.Type == "CRUSH_WEAPON" || def.Type == "DRAKIN_CRUSH" {
		return true
	}
	return strings.HasSuffix(e.getItemNounName(def), "axe")
}

// breakChance returns the per-try percent chance of tool breaking target, or 0 if
// it never can (indestructible target, or a tool of weaker substance).
func (e *GameEngine) breakChance(player *Player, targetDef, toolDef *gameworld.ItemDef) int {
	targetRank, ok := substanceRank(targetDef)
	if !ok {
		return 0
	}
	toolRank, ok := substanceRank(toolDef)
	if !ok {
		toolRank = substanceStrength["HARDMETAL"] + 1 // an indestructible tool beats anything breakable
	}
	margin := toolRank - targetRank
	if margin < 0 {
		return 0
	}
	chance := breakChanceByMargin[min(margin, len(breakChanceByMargin)-1)]
	if e.isHeavyBreakingTool(toolDef) {
		chance += breakHeavyToolBonus
	}
	chance += (player.Strength - 50) / 10
	return max(1, min(chance, 95))
}

// destroyedWhenContainerBreaks reports whether an item inside a broken container is
// destroyed outright: potions and their vials, flasks and bottles, and other glass.
func (e *GameEngine) destroyedWhenContainerBreaks(def *gameworld.ItemDef) bool {
	if _, isGem := gemArchetypeSellBase[def.Number]; isGem {
		return false
	}
	if def.Type == "LIQCONTAINER" || def.Type == "LIQUID" || strings.EqualFold(def.Substance, "BRITTLE") {
		return true
	}
	switch e.getItemNounName(def) {
	case "potion", "vial", "flask", "bottle":
		return true
	}
	return false
}

// chipGem flaws an undamaged gem as chipped or cracked, replacing any
// flawless/perfect/polished quality word and scaling its locked value to match.
// Returns false if the gem was already flawed or has no adjective slot to spare.
func (e *GameEngine) chipGem(def *gameworld.ItemDef, gem *InventoryItem) bool {
	adjID := func(name string, fallback int) int {
		if id := e.adjByName(name); id != 0 {
			return id
		}
		return fallback
	}
	// ADJNOUN.SCR fallbacks, as in doMold.
	chippedID, crackedID, damagedID := adjID("chipped", 53), adjID("cracked", 384), adjID("damaged", 83)
	qualityMult := map[int]float64{adjID("flawless", 0): 2.0, adjID("perfect", 0): 4.0, adjID("polished", 241): 1.0}
	delete(qualityMult, 0)

	slots := [3]*int{&gem.Adj1, &gem.Adj2, &gem.Adj3}
	var slot *int
	oldMult := 1.0
	for _, a := range slots {
		if *a == chippedID || *a == crackedID || *a == damagedID {
			return false
		}
		if m, ok := qualityMult[*a]; ok && slot == nil {
			slot, oldMult = a, m
		}
	}
	if slot == nil {
		for _, a := range slots {
			if *a == 0 {
				slot = a
				break
			}
		}
	}
	if slot == nil {
		return false
	}
	newID, newMult := chippedID, 0.60 // gemQualities in containers.go
	if rand.Intn(2) == 0 {
		newID, newMult = crackedID, 0.25
	}
	value := e.currentGemValue(def, gem)
	*slot = newID
	gem.Val1 = max(1, int(float64(value)*newMult/oldMult))
	if gem.Val2 > 0 {
		gem.Val2 = max(1, int(float64(gem.Val2)*newMult/oldMult))
	}
	return true
}

// spillBrokenContents puts a broken container's contents on the floor of the
// room, destroying fragile items and chipping gems along the way.
// Reports whether anything was damaged.
func (e *GameEngine) spillBrokenContents(room *gameworld.Room, contents []InventoryItem) bool {
	damaged := false
	for _, ci := range contents {
		def := e.items[ci.Archetype]
		if def == nil {
			continue
		}
		if ci.State != "MONEY" {
			if e.destroyedWhenContainerBreaks(def) {
				damaged = true
				continue
			}
			if _, isGem := gemArchetypeSellBase[def.Number]; isGem && rand.Intn(100) < breakGemDamagePct && e.chipGem(def, &ci) {
				damaged = true
			}
		}
		ri := roomItemFromInventory(&ci)
		ri.Ref = nextRoomItemRef(room)
		if isContainerDef(def) {
			e.roomContainerSet(room.Number, ri.Ref, ci.Contents)
		}
		room.Items = append(room.Items, ri)
		e.notifyRoomChange(RoomChange{RoomNumber: room.Number, Type: "item_add", Item: &ri})
	}
	return damaged
}

// doBreakWith runs the BREAK <target> WITH <tool> mechanic for an object no
// IFPREVERB BREAK script handled (see doBreak).
func (e *GameEngine) doBreakWith(ctx context.Context, player *Player, targetArgs, toolArgs []string) *CommandResult {
	if len(toolArgs) == 0 {
		return &CommandResult{Messages: []string{"What do you want to break it with?"}}
	}
	if player.RoundTimeExpiry.After(time.Now()) {
		remaining := int(time.Until(player.RoundTimeExpiry).Seconds()) + 1
		return &CommandResult{Messages: []string{fmt.Sprintf("[Wait %d seconds...]", remaining)}}
	}
	room := e.rooms[player.RoomNumber]
	if room == nil {
		return &CommandResult{Messages: []string{"You can't do that here."}}
	}

	// The tool: wielded, off-hand, or carried.
	toolTarget, toolSkip := parseOrdinal(strings.ToLower(strings.Join(toolArgs, " ")))
	var tool *InventoryItem
	held := []*InventoryItem{player.Wielded, player.OffHand}
	for i := range player.Inventory {
		held = append(held, &player.Inventory[i])
	}
	for _, ii := range held {
		if ii == nil {
			continue
		}
		if d := e.items[ii.Archetype]; d != nil && matchesTargetOrdinal(e.getItemNounName(d), toolTarget, &toolSkip, e.getAdjName(ii.Adj1), e.getAdjName(ii.Adj2), e.getAdjName(ii.Adj3)) {
			tool = ii
			break
		}
	}
	if tool == nil {
		return &CommandResult{Messages: []string{"You aren't holding that."}}
	}
	toolDef := e.items[tool.Archetype]

	// The target: on the floor (not inside another container), else in inventory.
	target, mine := stripMyPrefix(strings.ToLower(strings.Join(targetArgs, " ")))
	target, skip := parseOrdinal(target)
	roomIdx, invIdx := -1, -1
	if !mine {
		for i, ri := range room.Items {
			if ri.IsPut {
				continue
			}
			if d := e.items[ri.Archetype]; d != nil && matchesTargetOrdinal(e.getItemNounName(d), target, &skip, e.getAdjName(ri.Adj1), e.getAdjName(ri.Adj2), e.getAdjName(ri.Adj3)) {
				roomIdx = i
				break
			}
		}
	}
	if roomIdx < 0 {
		for i := range player.Inventory {
			ii := &player.Inventory[i]
			if d := e.items[ii.Archetype]; d != nil && matchesTargetOrdinal(e.getItemNounName(d), target, &skip, e.getAdjName(ii.Adj1), e.getAdjName(ii.Adj2), e.getAdjName(ii.Adj3)) {
				invIdx = i
				break
			}
		}
	}

	var targetDef *gameworld.ItemDef
	var targetName string
	switch {
	case roomIdx >= 0:
		ri := room.Items[roomIdx]
		targetDef = e.items[ri.Archetype]
		targetName = e.formatItemName(targetDef, ri.Adj1, ri.Adj2, ri.Adj3, ri.Extend)
		if targetDef.Weight >= 1000 || isPortal(targetDef.Type) || containsFlag(targetDef.Flags, "FIXED") ||
			e.isPersistentContainerRef(room.Number, ri.Ref) {
			return &CommandResult{Messages: []string{"You can't break that."}}
		}
	case invIdx >= 0:
		ii := player.Inventory[invIdx]
		if &player.Inventory[invIdx] == tool {
			return &CommandResult{Messages: []string{"You can't break something with itself."}}
		}
		targetDef = e.items[ii.Archetype]
		targetName = e.formatItemName(targetDef, ii.Adj1, ii.Adj2, ii.Adj3, ii.Tail)
	default:
		return &CommandResult{Messages: []string{"You don't see that here."}}
	}

	rt := applyRoundTime(player, breakRoundTime)
	player.RoundTime = rt
	player.RoundTimeExpiry = time.Now().Add(time.Duration(rt) * time.Second)
	rtMsg := fmt.Sprintf(" [Round: %d sec]", rt)
	toolName := e.formatItemName(toolDef, tool.Adj1, tool.Adj2, tool.Adj3, tool.Tail)
	attemptBroadcast := fmt.Sprintf("%s strikes %s with %s.", player.DisplayNameCap(), targetName, toolName)

	// Striking it counts as touching it — an armed sigil goes off (see triggerSigilIfArmed).
	var sigilMsgs []string
	if roomIdx >= 0 {
		sigilMsgs = e.triggerSigilIfArmed(ctx, player, &room.Items[roomIdx].SigilSpellID, &room.Items[roomIdx].SigilOwner, targetName)
		if sigilMsgs != nil {
			itemCopy := room.Items[roomIdx]
			e.notifyRoomChange(RoomChange{RoomNumber: room.Number, Type: "item_update", ItemRef: itemCopy.Ref, Item: &itemCopy})
		}
	} else {
		sigilMsgs = e.triggerSigilIfArmed(ctx, player, &player.Inventory[invIdx].SigilSpellID, &player.Inventory[invIdx].SigilOwner, targetName)
	}
	if sigilMsgs != nil {
		e.SavePlayer(ctx, player)
		return &CommandResult{Messages: append(sigilMsgs, rtMsg), RoomBroadcast: []string{attemptBroadcast}, PlayerState: player}
	}

	if rand.Intn(100) >= e.breakChance(player, targetDef, toolDef) {
		e.SavePlayer(ctx, player)
		return &CommandResult{
			Messages:      []string{fmt.Sprintf("%s doesn't break.", capitalize(targetName)), rtMsg},
			RoomBroadcast: []string{attemptBroadcast},
			PlayerState:   player,
		}
	}

	// It breaks. Pull the item and its contents out of the world.
	var broken InventoryItem
	var contents []InventoryItem
	if roomIdx >= 0 {
		broken, contents = e.takeRoomItemWithContents(room, roomIdx)
	} else {
		broken = player.Inventory[invIdx]
		contents = broken.Contents
		player.Inventory = append(player.Inventory[:invIdx], player.Inventory[invIdx+1:]...)
	}

	var msgs []string
	if isContainerDef(targetDef) {
		msgs = append(msgs, fmt.Sprintf("%s busts wide open!", capitalize(targetName)))
	} else if strings.EqualFold(targetDef.Substance, "BRITTLE") {
		msgs = append(msgs, fmt.Sprintf("%s shatters!", capitalize(targetName)))
	} else {
		msgs = append(msgs, fmt.Sprintf("%s breaks apart!", capitalize(targetName)))
	}
	// A trapped container springs its trap (see traps.go); an explosive one takes
	// the contents with it instead of spilling them.
	if code := trapCodeOf(targetDef, broken.Val4, broken.Val5); code != 0 {
		msgs = append(msgs, e.springTrap(player, &broken.Val4, &broken.TrapFound)...)
		if code == trapExplosive {
			msgs = append(msgs, "Everything inside is destroyed!")
			contents = nil
		}
	}
	if e.spillBrokenContents(room, contents) {
		msgs = append(msgs, "Some things inside were damaged.")
	}
	msgs = append(msgs, rtMsg)

	e.SavePlayer(ctx, player)
	return &CommandResult{
		Messages:      msgs,
		RoomBroadcast: []string{fmt.Sprintf("%s breaks %s with %s!", player.DisplayNameCap(), targetName, toolName)},
		PlayerState:   player,
	}
}
