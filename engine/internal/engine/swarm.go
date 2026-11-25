package engine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Insect Swarm (522, Druidic level 25): "Limited, damaging" entanglement per
// MAGIC.TXT. Insects boil up from the ground and swarm the target, holding it in
// place and biting every few seconds until the swarm dies off or is removed. The
// original session capture (original/csanburn/bog.txt) shows a bog ooze casting
// it on a player: "Insects suddenly boil up from the ground and swarm over you!",
// then "You are being devoured by insects!" roughly every 5 seconds with an I
// (immobilized) prompt code, until a wolfling TRANSFORM ends it with "The insects
// drop off and die." The swarmed player could still cast spells.
//
// On a player the swarm is a PlayerEntangle (SpellID 522), so Freedom, Dispel
// Lesser Magic, Mist/Slime Form and the movement/stand gates all handle it like
// Plant Snare. Per MAGIC.TXT's escape list, a wolfling's TRANSFORM and being
// snared by Plant Snare also drive the insects off. STRUGGLE does not: it isn't in
// that list, and you can't wrestle free of a cloud of insects.
//
// No capture shows the duration or damage. 60 seconds matches Plant Snare; the
// per-tick damage (the SpellDef's DmgMin/DmgMax, 3-8) is kept low, since the
// capture never shows it threatening anyone. Neither is shown to the player.

const (
	insectSwarmSpellID  = 522
	insectSwarmDuration = 60 * time.Second
	insectSwarmTick     = 5 * time.Second
	insectSwarmDropMsg  = "The insects drop off and die."
)

// damagingSnare is a spell that holds its target in place and hurts it every
// insectSwarmTick: Insect Swarm, and Creeping Death (526), MAGIC.TXT's "Damaging
// snare". Creeping Death has no text in any capture, so its lines are our own,
// modelled on Plant Snare's; its damage (the SpellDef's DmgMin/DmgMax) is Insect
// Swarm's scaled by spell level, 22 to 25.
type damagingSnare struct {
	spellID   int
	spellName string
	noun      string // what holds the player, for "You are bound by ..." lines
	killer    string
	catchMsg  string // %s = the target
	noBiteMsg string // discorporate target, %s = the target
	tickMsg   string // to the player
	dropMsg   string // to the player, as it ends
	monTick   string // %s = the monster, capitalized
	monHit    string // "bite to left leg"
	monDrop   string // %s = the monster
	suffix    string // after the monster's name in the room
	monster   func(*MonsterInstance) (*bool, *time.Time, *string)
}

const creepingDeathSpellID = 526

var insectSwarmSnare = &damagingSnare{
	spellID: insectSwarmSpellID, spellName: "Insect Swarm",
	noun: "a swarm of insects", killer: "a swarm of insects",
	catchMsg:  "Insects suddenly boil up from the ground and swarm over %s!",
	noBiteMsg: "Insects boil up from the ground, but find nothing to bite on %s!",
	tickMsg:   "You are being devoured by insects!",
	dropMsg:   insectSwarmDropMsg,
	monTick:   "%s is being devoured by insects!", monHit: "bite",
	monDrop: "The insects swarming over %s drop off and die.",
	suffix:  "swarmed by insects",
	monster: func(m *MonsterInstance) (*bool, *time.Time, *string) {
		return &m.Swarmed, &m.SwarmExpiry, &m.SwarmCasterName
	},
}

var creepingDeathSnare = &damagingSnare{
	spellID: creepingDeathSpellID, spellName: "Creeping Death",
	noun: "strangling vines", killer: "strangling vines",
	catchMsg:  "Thorny black vines creep up from the ground and coil around %s!",
	noBiteMsg: "Thorny black vines creep up from the ground, but close on nothing as they pass through %s!",
	tickMsg:   "The black vines tighten around you, tearing at your flesh!",
	dropMsg:   "The black vines around you rapidly wither and turn to dust.",
	monTick:   "The black vines tighten around %s!", monHit: "gash",
	monDrop: "The black vines around %s rapidly wither and turn to dust.",
	suffix:  "ensnared",
	monster: func(m *MonsterInstance) (*bool, *time.Time, *string) {
		return &m.Creeping, &m.CreepExpiry, &m.CreepCasterName
	},
}

var damagingSnares = []*damagingSnare{insectSwarmSnare, creepingDeathSnare}

// damagingSnareForSpell returns the damaging snare a spell makes, or nil.
func damagingSnareForSpell(spellID int) *damagingSnare {
	for _, ds := range damagingSnares {
		if ds.spellID == spellID {
			return ds
		}
	}
	return nil
}

// StartInsectSwarmLoop runs the 5-second tick for damaging snares on players and
// monsters, and for the short spell effects that can't wait for regenTick's minute
// (spell_ticks.go).
func (e *GameEngine) StartInsectSwarmLoop() {
	go func() {
		ticker := time.NewTicker(insectSwarmTick)
		defer ticker.Stop()
		for range ticker.C {
			e.insectSwarmStep()
		}
	}()
}

func (e *GameEngine) insectSwarmStep() {
	defer recoverBackground("insect swarm tick")
	for _, ds := range damagingSnares {
		e.damagingSnarePlayerTick(ds)
		e.damagingSnareMonsterTick(ds)
	}
	e.snareMonsterTick()
	e.spellEffectTick()
	e.maelstromTick()
}

// entangleReleaseMsg is the line a player sees when an entangle ends, however
// it ends (expiry, Freedom, Dispel Lesser Magic, a form change).
func entangleReleaseMsg(ent PlayerEntangle) string {
	if ds := damagingSnareForSpell(ent.SpellID); ds != nil {
		return ds.dropMsg
	}
	switch ent.SpellID {
	case plantSnareSpellID:
		return "The plant tendrils ensnaring you rapidly wither and turn to dust." // original/csanburn
	case webSpellID:
		return "The sticky webbing holding you falls away."
	case wallOfThornsSpellID:
		return "The wall of thorns around you withers and crumbles away."
	}
	return fmt.Sprintf("The %s releases its hold on you.", ent.SpellName)
}

// hasEntangle reports whether the player is held by the given spell.
func hasEntangle(p *Player, spellID int) bool {
	for _, ent := range p.Entangles {
		if ent.SpellID == spellID {
			return true
		}
	}
	return false
}

// isInsectSwarmed reports whether an insect swarm is on the player.
func isInsectSwarmed(p *Player) bool { return hasEntangle(p, insectSwarmSpellID) }

// entanglePlayer holds the player with the given spell for d, or refreshes the
// duration if it already does.
func entanglePlayer(p *Player, spellID int, spellName string, d time.Duration) {
	expiry := time.Now().Add(d)
	for i := range p.Entangles {
		if p.Entangles[i].SpellID == spellID {
			p.Entangles[i].Expiry = expiry
			return
		}
	}
	p.Entangles = append(p.Entangles, PlayerEntangle{SpellID: spellID, SpellName: spellName, Expiry: expiry})
}

// swarmPlayer puts an insect swarm on the player, or refreshes the duration of
// one already there.
func swarmPlayer(p *Player) {
	entanglePlayer(p, insectSwarmSpellID, "Insect Swarm", insectSwarmDuration)
}

// removeInsectSwarm strips any insect swarm from the player, returning
// insectSwarmDropMsg if there was one and "" otherwise.
func removeInsectSwarm(p *Player) string {
	if !removeEntangle(p, insectSwarmSpellID) {
		return ""
	}
	return insectSwarmDropMsg
}

// removeEntangle strips the given spell's hold from the player, reporting whether
// there was one.
func removeEntangle(p *Player, spellID int) bool {
	var kept []PlayerEntangle
	for _, ent := range p.Entangles {
		if ent.SpellID != spellID {
			kept = append(kept, ent)
		}
	}
	removed := len(kept) != len(p.Entangles)
	p.Entangles = kept
	return removed
}

// damagingSnarePlayerTick hurts every player the snare holds and ends those that
// have run their course. Damage follows the poison tick's rules (regenTick): a
// tick that takes a player to 0 knocks them out, and only a later tick while
// they're still out at 0 kills them.
func (e *GameEngine) damagingSnarePlayerTick(ds *damagingSnare) {
	if e.sessions == nil {
		return
	}
	spell := FindSpellByID(ds.spellID)
	if spell == nil {
		return
	}
	now := time.Now()
	for _, p := range e.sessions.OnlinePlayers() {
		if p == nil || !hasEntangle(p, ds.spellID) {
			continue
		}
		if p.Dead {
			removeEntangle(p, ds.spellID)
			continue
		}
		var expired bool
		for _, ent := range p.Entangles {
			if ent.SpellID == ds.spellID && now.After(ent.Expiry) {
				expired = true
			}
		}
		if expired {
			removeEntangle(p, ds.spellID)
			e.SavePlayer(context.Background(), p)
			e.sendToPlayerMsgs(p, []string{ds.dropMsg})
			continue
		}
		if p.IsFormLocked() {
			continue
		}

		msgs := []string{ds.tickMsg}
		if p.Unconscious && p.BodyPoints <= 0 {
			removeEntangle(p, ds.spellID)
			msgs = append(msgs, e.handlePlayerDeath(p, ds.killer)...)
		} else {
			p.BodyPoints -= rand.Intn(spell.DmgMax-spell.DmgMin+1) + spell.DmgMin
			if p.BodyPoints <= 0 {
				p.BodyPoints = 0
				p.Unconscious = true
				p.Position = 2
				msgs = append(msgs, "You collapse, unconscious!")
			}
		}
		e.SavePlayer(context.Background(), p)
		e.sendToPlayerMsgs(p, msgs)
	}
}

func (e *GameEngine) sendToPlayerMsgs(p *Player, msgs []string) {
	if e.sendToPlayer != nil {
		e.sendToPlayer(p.FirstName, msgs)
	}
}

// castDamagingSnareSpell handles a player casting Insect Swarm (522) or Creeping
// Death (526) at a monster: the monster is held in place (no wandering or fleeing,
// though it can still fight back) and hurt every 5 seconds by
// damagingSnareMonsterTick.
func (e *GameEngine) castDamagingSnareSpell(player *Player, ds *damagingSnare, args []string, castRoll, castChance int) *CommandResult {
	targetName := strings.Join(args, " ")
	if targetName == "" {
		targetName = e.autoTargetMonsterName(player)
	}
	if targetName == "" {
		return &CommandResult{Messages: []string{fmt.Sprintf("Cast %s at what?", ds.spellName)}, TargetNotFound: true}
	}

	inst, def, refusal := e.findHostileMonsterInRoom(player, targetName)
	if refusal != nil {
		return refusal
	}
	if inst == nil {
		return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", targetName)}, TargetNotFound: true}
	}

	name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
	article := articleFor(name, def.Unique)
	gestureSelf := fmt.Sprintf("You gesture at %s%s.", article, name)
	gestureRoom := fmt.Sprintf("%s gestures at %s%s.", player.DisplayNameCap(), article, name)

	if def.Discorporate {
		line := fmt.Sprintf(ds.noBiteMsg, article+name)
		return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}}
	}
	if magicResisted(player, castRoll, castChance, def.MagicResist) {
		return &CommandResult{
			Messages:      []string{gestureSelf, fmt.Sprintf("The %s resists the spell!", name)},
			RoomBroadcast: []string{gestureRoom, fmt.Sprintf("The %s resists the spell!", name)},
			Resisted:      true,
		}
	}

	e.updateMonster(inst.ID, func(m *MonsterInstance) {
		on, expiry, caster := ds.monster(m)
		*on, *expiry, *caster = true, time.Now().Add(insectSwarmDuration), player.FirstName
	})

	line := fmt.Sprintf(ds.catchMsg, article+name)
	return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}}
}

// updateMonster changes a live monster instance, found by ID, under the lock.
func (e *GameEngine) updateMonster(id int, change func(*MonsterInstance)) {
	e.monsterMgr.mu.Lock()
	defer e.monsterMgr.mu.Unlock()
	if i := e.monsterMgr.indexOfID(id); i >= 0 {
		change(&e.monsterMgr.instances[i])
	}
}

// damagingSnareMonsterTick hurts every monster the snare holds and ends those that
// have run their course. Kills credit the caster, as with Tentacles.
func (e *GameEngine) damagingSnareMonsterTick(ds *damagingSnare) {
	if e.monsterMgr == nil {
		return
	}
	spell := FindSpellByID(ds.spellID)
	if spell == nil {
		return
	}

	type snareKill struct {
		inst   MonsterInstance
		def    *gameworld.MonsterDef
		caster string
	}
	var kills []snareKill
	now := time.Now()

	e.monsterMgr.mu.Lock()
	for i := range e.monsterMgr.instances {
		inst := &e.monsterMgr.instances[i]
		on, expiry, caster := ds.monster(inst)
		if !*on {
			continue
		}
		def := e.monsters[inst.DefNumber]
		if !inst.Alive || def == nil {
			*on = false
			continue
		}
		name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
		article := articleFor(name, def.Unique)

		if now.After(*expiry) {
			*on = false
			if e.localRoomBroadcast != nil {
				e.localRoomBroadcast(inst.RoomNumber, []string{fmt.Sprintf(ds.monDrop, article+name)})
			}
			continue
		}
		if inst.Imprisoned {
			continue // nothing reaches it inside the bubble
		}

		dmg := rand.Intn(spell.DmgMax-spell.DmgMin+1) + spell.DmgMin
		if e.localRoomBroadcast != nil {
			e.localRoomBroadcast(inst.RoomNumber, []string{
				capitalize(fmt.Sprintf(ds.monTick, article+name)),
				fmt.Sprintf(" %s %s to %s. [%d Damage]", damageSeverity(dmg, inst.MaxHP), ds.monHit, randomBodyPart(def.BodyType), dmg),
			})
		}
		inst.CurrentHP -= dmg
		if inst.CurrentHP <= 0 {
			inst.Alive = false
			inst.CurrentHP = 0
			inst.DeathTime = now
			*on = false
			kills = append(kills, snareKill{inst: *inst, def: def, caster: *caster})
		}
	}
	e.monsterMgr.mu.Unlock()

	for _, k := range kills {
		e.announceMonsterDeath(&k.inst, k.def)
		if recipients := e.dotKillRecipients(k.caster); len(recipients) > 0 {
			e.handleMonsterDeath(recipients, &k.inst, k.def)
		}
	}
}

// announceMonsterDeath tells the room a monster has died of something other than
// a blow that described its own kill.
func (e *GameEngine) announceMonsterDeath(inst *MonsterInstance, def *gameworld.MonsterDef) {
	if e.localRoomBroadcast == nil {
		return
	}
	name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
	capName := capArticle(articleFor(name, def.Unique)) + name
	if deathText := def.TextOverrides["TEXD"]; deathText != "" {
		e.localRoomBroadcast(inst.RoomNumber, []string{fmt.Sprintf("%s %s", capName, deathText)})
	} else {
		e.localRoomBroadcast(inst.RoomNumber, []string{fmt.Sprintf("%s collapses, dead!", capName)})
	}
}

// entangleNoun names what is holding the player, for "You are bound by ..." lines.
func entangleNoun(ent PlayerEntangle) string {
	if ds := damagingSnareForSpell(ent.SpellID); ds != nil {
		return ds.noun
	}
	switch ent.SpellID {
	case plantSnareSpellID:
		return "plant tendrils"
	case webSpellID:
		return "sticky webbing"
	case wallOfThornsSpellID:
		return "a wall of thorns"
	}
	return ent.SpellName
}
