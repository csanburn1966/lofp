package engine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Web (127) and Plant Snare (500) on players, and Wall of Thorns (516). Creeping
// Death (526) and Insect Swarm (522) are in swarm.go.
//
// Wall of Thorns springs up around its target: the target can't leave the room,
// advance or retreat, but can fight whatever has closed in on it, taking damage from
// the thorns with every melee attack. Missiles, spells and psionics cost it nothing.

const (
	webSpellID          = 127
	plantSnareSpellID   = 500
	wallOfThornsSpellID = 516

	snareDuration = 60 * time.Second
)

// spellActive reports whether a spell expiry is set and still to come.
func spellActive(expiry time.Time) bool {
	return !expiry.IsZero() && time.Now().Before(expiry)
}

// heldInPlaceMessage returns why the player can't leave where they're standing
// (psi Immobilize, an Imprisonment Rune, or a snare), or "".
func heldInPlaceMessage(p *Player) string {
	if p.IsImmobilized() {
		return "You are immobilized and cannot move!"
	}
	if len(p.Entangles) > 0 {
		return fmt.Sprintf("You are bound by %s and cannot move!", entangleNoun(p.Entangles[0]))
	}
	return ""
}

// heldInPlace reports whether something holds the monster where it is, so it can't
// wander or flee. The Webbed and Tentacled expiries are checked here because
// monsterCombatTick only clears them for a monster that's fighting.
func (inst *MonsterInstance) heldInPlace() bool {
	now := time.Now()
	return inst.Immobilized || inst.Swarmed || inst.Creeping || inst.thorned() || inst.snared() ||
		(inst.Webbed && now.Before(inst.WebExpiry)) || (inst.Tentacled && now.Before(inst.TentacleExpiry))
}

// thorned reports whether a Wall of Thorns surrounds the monster.
func (inst *MonsterInstance) thorned() bool { return spellActive(inst.ThornExpiry) }

// snared reports whether a Plant Snare holds the monster.
func (inst *MonsterInstance) snared() bool { return spellActive(inst.SnareExpiry) }

// snareHoldsBack reports whether a snare keeps the monster from advancing on anyone,
// as one does a player. A monster that fights from a distance needn't advance.
func (e *GameEngine) snareHoldsBack(inst *MonsterInstance, def *gameworld.MonsterDef) bool {
	return (inst.thorned() || inst.snared() || inst.Creeping) && !e.monsterIsRanged(def)
}

// snareImmunity returns the lines (to the player, to the room) for a snaring spell
// that can't hold them, or "" if it can.
func snareImmunity(p *Player, spell *SpellDef) (selfMsg, roomMsg string) {
	switch {
	case p.MistForm || p.SlimeForm:
		return "It finds nothing solid to hold.", fmt.Sprintf("It finds nothing solid to hold on %s.", p.DisplayName())
	case spellActive(p.FreeActionExpiry):
		return "It slides off you without taking hold.", fmt.Sprintf("It slides off %s without taking hold.", p.DisplayName())
	case spell.ID == webSpellID:
		if spellActive(p.RepelPlantsAndWebsExpiry) {
			return "The webbing recoils from you!", fmt.Sprintf("The webbing recoils from %s!", p.DisplayName())
		}
	case playerImmuneToPlantSnare(p):
		return "The plants recoil from you!", fmt.Sprintf("The plants recoil from %s!", p.DisplayName())
	}
	return "", ""
}

// snarePlayer applies Web, Plant Snare, Wall of Thorns or Creeping Death to a player,
// returning the lines for them and the line for everyone else. The caller has
// checked snareImmunity and saves the player.
func snarePlayer(p *Player, spell *SpellDef) (selfMsgs []string, roomMsg string) {
	switch spell.ID {
	case webSpellID:
		selfMsgs = []string{"You are covered with strands of sticky webbing!"}
		roomMsg = fmt.Sprintf("%s is covered with strands of sticky webbing!", p.DisplayNameCap()) // chan0409.txt
	case wallOfThornsSpellID:
		selfMsgs = []string{"A wall of thorns springs up around you!"}
		roomMsg = fmt.Sprintf("A wall of thorns springs up around %s!", p.DisplayName())
	case creepingDeathSpellID:
		selfMsgs = []string{fmt.Sprintf(creepingDeathSnare.catchMsg, "you")}
		roomMsg = fmt.Sprintf(creepingDeathSnare.catchMsg, p.DisplayName())
	default:
		selfMsgs = []string{"Plant tendrils shoot up from the ground and ensnare you!"} // shirla.cap
		roomMsg = fmt.Sprintf("Plant tendrils shoot up from the ground and ensnare %s!", p.DisplayName())
	}
	// Per MAGIC.TXT, being snared or webbed gets rid of an insect swarm.
	if spell.ID == webSpellID || spell.ID == plantSnareSpellID {
		if dropMsg := removeInsectSwarm(p); dropMsg != "" {
			selfMsgs = append(selfMsgs, dropMsg)
		}
	}
	entanglePlayer(p, spell.ID, spell.Name, snareDuration)
	return selfMsgs, roomMsg
}

// hostileSpellAtPlayer finds the player a hostile spell (Wither Limb, Wall of Thorns) is
// aimed at and gives them their chance to resist it, the way Truename does: someone who
// has SUBMITted puts up next to no resistance, and nobody resists their own spell. It
// returns the target, or the finished result if there's no such player or they resisted.
func (e *GameEngine) hostileSpellAtPlayer(player *Player, targetName string) (*Player, *CommandResult) {
	switch strings.ToLower(targetName) {
	case "me", "myself", "self":
		return player, nil
	}
	target := e.findPlayerInRoom(player, strings.ToLower(targetName))
	if target == nil {
		return nil, &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", targetName)}, TargetNotFound: true}
	}
	shielded := !target.Submitting && rand.Intn(100) < spellShieldResistBonus(target)
	if player.IsGM || !(shielded || playerMagicResistRoll(player, target)) {
		return target, nil
	}
	line := fmt.Sprintf("%s resists the spell!", target.DisplayNameCap())
	return nil, &CommandResult{
		Messages:      []string{fmt.Sprintf("You gesture at %s.", target.DisplayName()), line},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), target.DisplayName()), line},
		TargetName:    target.FirstName,
		TargetMsg:     []string{fmt.Sprintf("%s gestures at you.", player.DisplayNameCap()), "You resist the spell!"},
		Resisted:      true,
	}
}

// castWallOfThorns handles Wall of Thorns (516) cast at a monster or a player.
func (e *GameEngine) castWallOfThorns(player *Player, spell *SpellDef, args []string, castRoll, castChance int) *CommandResult {
	targetName := strings.Join(args, " ")
	if targetName == "" {
		targetName = e.autoTargetMonsterName(player)
	}
	if targetName == "" {
		return &CommandResult{Messages: []string{"Cast Wall of Thorns at what?"}, TargetNotFound: true}
	}
	inst, def, refusal := e.findHostileMonsterInRoom(player, targetName)
	if refusal != nil {
		return refusal
	}
	if inst == nil {
		target, refusal := e.hostileSpellAtPlayer(player, targetName)
		if refusal != nil {
			return refusal
		}
		return e.castSnareAtPlayer(player, target, spell)
	}

	name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
	article := articleFor(name, def.Unique)
	gestureSelf := fmt.Sprintf("You gesture at %s%s.", article, name)
	gestureRoom := fmt.Sprintf("%s gestures at %s%s.", player.DisplayNameCap(), article, name)
	if def.Discorporate {
		line := fmt.Sprintf("A wall of thorns springs up, but %s%s drifts straight through it!", article, name)
		return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}}
	}
	if magicResisted(player, castRoll, castChance, def.MagicResist) {
		line := fmt.Sprintf("The %s resists the spell!", name)
		return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}, Resisted: true}
	}
	e.updateMonster(inst.ID, func(m *MonsterInstance) {
		m.ThornExpiry = time.Now().Add(snareDuration)
		m.ThornCasterName = player.FirstName
	})
	line := fmt.Sprintf("A wall of thorns springs up around %s%s!", article, name)
	return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}}
}

// castSnareAtPlayer resolves a snaring spell one player has cast at another, or at
// themselves.
func (e *GameEngine) castSnareAtPlayer(player, target *Player, spell *SpellDef) *CommandResult {
	selfMsg, roomMsg := snareImmunity(target, spell)
	selfMsgs := []string{selfMsg}
	if selfMsg == "" {
		selfMsgs, roomMsg = snarePlayer(target, spell)
		e.SavePlayer(context.Background(), target)
	}
	if target == player {
		return &CommandResult{
			Messages:      append([]string{"You gesture."}, selfMsgs...),
			RoomBroadcast: []string{fmt.Sprintf("%s gestures.", player.DisplayNameCap()), roomMsg},
		}
	}
	return &CommandResult{
		Messages:      []string{fmt.Sprintf("You gesture at %s.", target.DisplayName()), roomMsg},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), target.DisplayName()), roomMsg},
		TargetName:    target.FirstName,
		TargetMsg:     append([]string{fmt.Sprintf("%s gestures at you.", player.DisplayNameCap())}, selfMsgs...),
	}
}

// thornDamage rolls what the thorns do to whoever fights their way through them.
func thornDamage() int {
	spell := FindSpellByID(wallOfThornsSpellID)
	return rand.Intn(spell.DmgMax-spell.DmgMin+1) + spell.DmgMin
}

// thornsTearPlayer hurts a player making a melee attack from inside a Wall of Thorns.
// Like the damaging snares' ticks it knocks out but doesn't kill.
func (e *GameEngine) thornsTearPlayer(p *Player) []string {
	dmg := formDamageReduction(p, thornDamage())
	p.BodyPoints -= dmg
	msgs := []string{fmt.Sprintf("The thorns tear at you as you fight your way through them! [%d Damage]", dmg)}
	if p.BodyPoints <= 0 {
		p.BodyPoints = 0
		p.Unconscious = true
		p.Position = 2
		msgs = append(msgs, "You collapse, unconscious!")
	}
	return msgs
}

// thornsTearMonster hurts a monster about to make a melee attack from inside a Wall
// of Thorns, reporting whether that killed it. Called without monsterMgr.mu held.
func (e *GameEngine) thornsTearMonster(inst *MonsterInstance, def *gameworld.MonsterDef) bool {
	if !inst.thorned() || e.monsterIsRanged(def) {
		return false
	}
	caster := inst.ThornCasterName
	dmg := thornDamage()
	name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
	if e.localRoomBroadcast != nil {
		e.localRoomBroadcast(inst.RoomNumber, []string{fmt.Sprintf("The thorns tear at %s%s as it fights its way through them! [%d Damage]", articleFor(name, def.Unique), name, dmg)})
	}
	killed, _ := e.damageMonster(inst.ID, dmg, caster)
	if !killed {
		return false
	}
	e.announceMonsterDeath(inst, def)
	if recipients := e.dotKillRecipients(caster); len(recipients) > 0 {
		e.handleMonsterDeath(recipients, inst, def)
	} else {
		e.clearCombatForMonster(inst.ID)
	}
	return true
}

// snareMonsterTick ends the Plant Snares and Walls of Thorns on monsters that have run
// their course.
func (e *GameEngine) snareMonsterTick() {
	if e.monsterMgr == nil {
		return
	}
	e.monsterMgr.mu.Lock()
	defer e.monsterMgr.mu.Unlock()
	for i := range e.monsterMgr.instances {
		m := &e.monsterMgr.instances[i]
		for _, s := range []struct {
			expiry *time.Time
			spell  string
		}{{&m.ThornExpiry, "Wall of Thorns"}, {&m.SnareExpiry, "Plant Snare"}} {
			if s.expiry.IsZero() || spellActive(*s.expiry) {
				continue
			}
			*s.expiry = time.Time{}
			if def := e.monsters[m.DefNumber]; def != nil && m.Alive && e.localRoomBroadcast != nil {
				name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
				e.localRoomBroadcast(m.RoomNumber, []string{fmt.Sprintf(fadeLines[s.spell][1], articleFor(name, def.Unique)+name)})
			}
		}
	}
}

// castPlantSnareSpell handles Plant Snare (500): outdoors only, a monster or player is
// held where they stand for a minute. They can still fight. The lines are the captures'
// ("Plant tendrils shoot up from the ground and ensnare Eldaran!", chandra.txt), and
// chan0327.txt shows snared monsters: "a laash (ensnared)".
func (e *GameEngine) castPlantSnareSpell(player *Player, spell *SpellDef, args []string, castRoll, castChance int) *CommandResult {
	room := e.rooms[player.RoomNumber]
	if room == nil || !isOutdoorTerrain(room.Terrain) {
		return &CommandResult{Messages: []string{"Plant Snare only works outdoors."}}
	}
	targetName := strings.Join(args, " ")
	if targetName == "" {
		targetName = e.autoTargetMonsterName(player)
	}
	if targetName == "" {
		return &CommandResult{Messages: []string{"Snare whom?"}, TargetNotFound: true}
	}
	inst, def, refusal := e.findHostileMonsterInRoom(player, targetName)
	if refusal != nil {
		return refusal
	}
	if inst == nil {
		target := e.findPlayerInRoom(player, strings.ToLower(targetName))
		if target == nil {
			return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", targetName)}, TargetNotFound: true}
		}
		return e.castSnareAtPlayer(player, target, spell)
	}

	name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
	full := articleFor(name, def.Unique) + name
	gestureSelf := fmt.Sprintf("You gesture at %s.", full)
	gestureRoom := fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), full)
	if def.Discorporate {
		line := fmt.Sprintf("Plant tendrils shoot up from the ground, but find nothing solid to hold on %s!", full)
		return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}}
	}
	if magicResisted(player, castRoll, castChance, def.MagicResist) {
		line := fmt.Sprintf("The %s resists the spell!", name)
		return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}, Resisted: true}
	}
	lines := []string{fmt.Sprintf("Plant tendrils shoot up from the ground and ensnare %s!", full)}
	e.updateMonster(inst.ID, func(m *MonsterInstance) {
		m.SnareExpiry = time.Now().Add(snareDuration)
		if m.Swarmed { // being snared gets rid of an insect swarm (MAGIC.TXT)
			m.Swarmed = false
			lines = append(lines, fmt.Sprintf(insectSwarmSnare.monDrop, full))
		}
	})
	return &CommandResult{
		Messages:      append([]string{gestureSelf}, lines...),
		RoomBroadcast: append([]string{gestureRoom}, lines...),
	}
}
