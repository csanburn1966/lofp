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

// StartInsectSwarmLoop runs the 5-second swarm damage/expiry tick for players
// and monsters.
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
	e.insectSwarmPlayerTick()
	e.insectSwarmMonsterTick()
}

// entangleReleaseMsg is the line a player sees when an entangle ends, however
// it ends (expiry, Freedom, Dispel Lesser Magic, a form change).
func entangleReleaseMsg(ent PlayerEntangle) string {
	if ent.SpellID == insectSwarmSpellID {
		return insectSwarmDropMsg
	}
	return fmt.Sprintf("The %s releases its hold on you.", ent.SpellName)
}

// isInsectSwarmed reports whether an insect swarm is on the player.
func isInsectSwarmed(p *Player) bool {
	for _, ent := range p.Entangles {
		if ent.SpellID == insectSwarmSpellID {
			return true
		}
	}
	return false
}

// swarmPlayer puts an insect swarm on the player, or refreshes the duration of
// one already there.
func swarmPlayer(p *Player) {
	expiry := time.Now().Add(insectSwarmDuration)
	for i := range p.Entangles {
		if p.Entangles[i].SpellID == insectSwarmSpellID {
			p.Entangles[i].Expiry = expiry
			return
		}
	}
	p.Entangles = append(p.Entangles, PlayerEntangle{SpellID: insectSwarmSpellID, SpellName: "Insect Swarm", Expiry: expiry})
}

// removeInsectSwarm strips any insect swarm from the player, returning
// insectSwarmDropMsg if there was one and "" otherwise.
func removeInsectSwarm(p *Player) string {
	kept := p.Entangles[:0]
	removed := false
	for _, ent := range p.Entangles {
		if ent.SpellID == insectSwarmSpellID {
			removed = true
			continue
		}
		kept = append(kept, ent)
	}
	if !removed {
		return ""
	}
	if len(kept) == 0 {
		p.Entangles = nil
	} else {
		p.Entangles = kept
	}
	return insectSwarmDropMsg
}

// insectSwarmPlayerTick bites every swarmed player once and ends swarms that
// have run their course. Damage follows the poison tick's rules (regenTick): a
// tick that takes a player to 0 knocks them out, and only a later tick while
// they're still out at 0 kills them.
func (e *GameEngine) insectSwarmPlayerTick() {
	if e.sessions == nil {
		return
	}
	spell := FindSpellByID(insectSwarmSpellID)
	if spell == nil {
		return
	}
	now := time.Now()
	for _, p := range e.sessions.OnlinePlayers() {
		if p == nil || !isInsectSwarmed(p) {
			continue
		}
		if p.Dead {
			removeInsectSwarm(p)
			continue
		}
		var expired bool
		for _, ent := range p.Entangles {
			if ent.SpellID == insectSwarmSpellID && now.After(ent.Expiry) {
				expired = true
			}
		}
		if expired {
			removeInsectSwarm(p)
			e.SavePlayer(context.Background(), p)
			e.sendToPlayerMsgs(p, []string{insectSwarmDropMsg})
			continue
		}
		if p.IsFormLocked() {
			continue
		}

		msgs := []string{"You are being devoured by insects!"}
		if p.Unconscious && p.BodyPoints <= 0 {
			removeInsectSwarm(p)
			msgs = append(msgs, e.handlePlayerDeath(p, "a swarm of insects")...)
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

// castInsectSwarmSpell handles a player casting Insect Swarm (522) at a
// monster: the monster is held in place (no wandering or fleeing, though it can
// still fight back) and bitten every 5 seconds by insectSwarmMonsterTick.
func (e *GameEngine) castInsectSwarmSpell(player *Player, spell *SpellDef, args []string, castRoll, castChance int) *CommandResult {
	targetName := strings.Join(args, " ")
	if targetName == "" {
		targetName = e.autoTargetMonsterName(player)
	}
	if targetName == "" {
		return &CommandResult{Messages: []string{"Cast Insect Swarm at what?"}}
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
		return &CommandResult{
			Messages:      []string{gestureSelf, fmt.Sprintf("Insects boil up from the ground, but find nothing to bite on %s%s!", article, name)},
			RoomBroadcast: []string{gestureRoom, fmt.Sprintf("Insects boil up from the ground, but find nothing to bite on %s%s!", article, name)},
		}
	}
	if magicResisted(player, castRoll, castChance, def.MagicResist) {
		return &CommandResult{
			Messages:      []string{gestureSelf, fmt.Sprintf("The %s resists the spell!", name)},
			RoomBroadcast: []string{gestureRoom, fmt.Sprintf("The %s resists the spell!", name)},
			Resisted:      true,
		}
	}

	e.monsterMgr.mu.Lock()
	for i := range e.monsterMgr.instances {
		if e.monsterMgr.instances[i].ID == inst.ID {
			e.monsterMgr.instances[i].Swarmed = true
			e.monsterMgr.instances[i].SwarmExpiry = time.Now().Add(insectSwarmDuration)
			e.monsterMgr.instances[i].SwarmCasterName = player.FirstName
			break
		}
	}
	e.monsterMgr.mu.Unlock()

	swarmLine := fmt.Sprintf("Insects suddenly boil up from the ground and swarm over %s%s!", article, name)
	return &CommandResult{
		Messages:      []string{gestureSelf, swarmLine},
		RoomBroadcast: []string{gestureRoom, swarmLine},
	}
}

// insectSwarmMonsterTick bites every swarmed monster once and ends swarms that
// have run their course. Kills credit the caster, as with Tentacles.
func (e *GameEngine) insectSwarmMonsterTick() {
	if e.monsterMgr == nil {
		return
	}
	spell := FindSpellByID(insectSwarmSpellID)
	if spell == nil {
		return
	}

	type swarmKill struct {
		inst MonsterInstance
		def  *gameworld.MonsterDef
	}
	var kills []swarmKill
	now := time.Now()

	e.monsterMgr.mu.Lock()
	for i := range e.monsterMgr.instances {
		inst := &e.monsterMgr.instances[i]
		if !inst.Swarmed {
			continue
		}
		def := e.monsters[inst.DefNumber]
		if !inst.Alive || def == nil {
			inst.Swarmed = false
			continue
		}
		name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
		article := articleFor(name, def.Unique)

		if now.After(inst.SwarmExpiry) {
			inst.Swarmed = false
			if e.localRoomBroadcast != nil {
				e.localRoomBroadcast(inst.RoomNumber, []string{fmt.Sprintf("The insects swarming over %s%s drop off and die.", article, name)})
			}
			continue
		}

		dmg := rand.Intn(spell.DmgMax-spell.DmgMin+1) + spell.DmgMin
		if e.localRoomBroadcast != nil {
			e.localRoomBroadcast(inst.RoomNumber, []string{
				fmt.Sprintf("%s%s is being devoured by insects!", capArticle(article), name),
				fmt.Sprintf(" %s bite to %s. [%d Damage]", damageSeverity(dmg, inst.MaxHP), randomBodyPart(def.BodyType), dmg),
			})
		}
		inst.CurrentHP -= dmg
		if inst.CurrentHP <= 0 {
			inst.Alive = false
			inst.CurrentHP = 0
			inst.DeathTime = now
			inst.Swarmed = false
			kills = append(kills, swarmKill{inst: *inst, def: def})
		}
	}
	e.monsterMgr.mu.Unlock()

	for _, k := range kills {
		name := strings.ToLower(FormatMonsterName(k.def, e.monAdjs))
		if e.localRoomBroadcast != nil {
			if deathText := k.def.TextOverrides["TEXD"]; deathText != "" {
				e.localRoomBroadcast(k.inst.RoomNumber, []string{fmt.Sprintf("A %s %s", name, deathText)})
			} else {
				e.localRoomBroadcast(k.inst.RoomNumber, []string{fmt.Sprintf("A %s collapses, dead!", name)})
			}
		}
		if recipients := e.dotKillRecipients(k.inst.SwarmCasterName); len(recipients) > 0 {
			e.handleMonsterDeath(recipients, &k.inst, k.def)
		}
	}
}

// entangleNoun names what is holding the player, for "You are bound by ..." lines.
func entangleNoun(ent PlayerEntangle) string {
	if ent.SpellID == insectSwarmSpellID {
		return "a swarm of insects"
	}
	return ent.SpellName
}
