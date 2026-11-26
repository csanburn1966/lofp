package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Turn Undead I/II/III (301, 302, 304). The original documents only list the
// spells' levels, so the behavior is a house design: the turn line is crowd
// control, not damage. Undead are split into three tiers (1 lowest, 3 highest):
//
//	Turn Undead I:   tiers 1 and 2 flee in fear.
//	Turn Undead II:  tier 1 has a 50% chance to be destroyed outright, or loses
//	                 half its remaining body points; tiers 2 and 3 flee in fear.
//	Turn Undead III: tier 1 is destroyed; tier 2 has a 50% chance to be
//	                 destroyed, or loses half its remaining body points; tier 3
//	                 flees in fear.
//
// A frightened undead flees for turnFearDuration. If it can't flee (snared, or
// no exits), it cowers without attacking for that long.
const turnFearDuration = 30 * time.Second

func isTurnUndeadSpell(id int) bool {
	return id == 301 || id == 302 || id == 304
}

// undeadTiers assigns each undead monster number to a turning tier, ranked by
// overall strength (BODY, attack, defense, armor, resistances and spells).
var undeadTiers = map[int]int{
	// Tier 1, the lowest.
	34: 1, 181: 1, 180: 1, 286: 1, 187: 1, 53: 1, 76: 1, 104: 1, 326: 1, 17: 1,
	25: 1, 327: 1, 75: 1, 24: 1, 370: 1, 2: 1, 26: 1, 1: 1,
	// Tier 2, the middle.
	210: 2, 151: 2, 100: 2, 146: 2, 55: 2, 432: 2, 65: 2, 374: 2, 377: 2, 292: 2,
	149: 2, 183: 2, 184: 2, 43: 2, 337: 2, 182: 2, 112: 2, 248: 2, 310: 2, 246: 2,
	// Tier 3, the highest.
	192: 3, 268: 3, 133: 3, 87: 3, 244: 3, 140: 3, 236: 3, 376: 3, 185: 3, 16: 3,
	186: 3, 141: 3, 136: 3, 264: 3, 303: 3, 179: 3, 293: 3, 102: 3, 263: 3, 135: 3,
}

// undeadTier returns a monster's turning tier, or 0 if turning doesn't affect it.
// Only the undead are turned: RACE 22, or the script's undead affecting immunity
// (slot 8) set, as on the spriggan and the toxic spirit. The check comes first
// because some numbers in the table belong to archived scripts that aren't
// loaded, and the loaded monster with that number is something else (263 and
// 264 are vith workers, 326 a mutated bear). Undead missing from the table fall
// back to that immunity (3 lesser, 2 greater, 1 most powerful), and to tier 2
// when it isn't set.
func undeadTier(def *gameworld.MonsterDef) int {
	if def.Race != 22 && def.Immunities[8] == 0 {
		return 0
	}
	if t, ok := undeadTiers[def.Number]; ok {
		return t
	}
	switch def.Immunities[8] {
	case 3:
		return 1
	case 1:
		return 3
	}
	return 2
}

// turnOutcome is what one Turn Undead spell does to one tier.
type turnOutcome int

const (
	turnNone turnOutcome = iota
	turnFear
	turnMaybeDestroy // 50% destroyed, otherwise half its remaining body points
	turnDestroy
)

func turnUndeadOutcome(spellID, tier int) turnOutcome {
	switch spellID {
	case 301:
		if tier == 1 || tier == 2 {
			return turnFear
		}
	case 302:
		switch tier {
		case 1:
			return turnMaybeDestroy
		case 2, 3:
			return turnFear
		}
	case 304:
		switch tier {
		case 1:
			return turnDestroy
		case 2:
			return turnMaybeDestroy
		case 3:
			return turnFear
		}
	}
	return turnNone
}

// castTurnUndead resolves a Turn Undead spell against a monster that has already
// passed castDamageSpell's targeting, force-bubble and undead checks.
func (e *GameEngine) castTurnUndead(player *Player, spell *SpellDef, inst *MonsterInstance, def *gameworld.MonsterDef, name string) *CommandResult {
	name = strings.ToLower(name)
	article := articleFor(name, def.Unique)
	who := article + name
	cast := fmt.Sprintf("You raise your hand and cast %s at %s!", spell.Name, who)
	roomCast := fmt.Sprintf("%s raises a hand and casts %s at %s!", player.DisplayNameCap(), spell.Name, who)

	outcome := turnUndeadOutcome(spell.ID, undeadTier(def))
	if outcome == turnMaybeDestroy && rand.Intn(2) == 0 {
		outcome = turnDestroy
	}

	switch outcome {
	case turnFear:
		e.monsterMgr.mu.Lock()
		if i := e.monsterMgr.indexOfID(inst.ID); i >= 0 {
			m := &e.monsterMgr.instances[i]
			m.Feared = true
			m.FearExpiry = time.Now().Add(turnFearDuration)
			m.LastAttacker = player.FirstName
		}
		e.monsterMgr.mu.Unlock()
		return &CommandResult{
			Messages:      []string{cast, fmt.Sprintf("The %s recoils from the turning magic and cowers in fear!", name)},
			RoomBroadcast: []string{roomCast, fmt.Sprintf("The %s recoils and cowers in fear!", name)},
		}

	case turnDestroy:
		killed, _ := e.damageMonster(inst.ID, max(1, inst.CurrentHP), player.FirstName)
		msgs := []string{cast}
		roomMsgs := []string{roomCast}
		if killed {
			line := fmt.Sprintf("The %s shrieks and crumbles to dust!", name)
			msgs = append(msgs, line)
			roomMsgs = append(roomMsgs, line)
			e.handleMonsterDeath(e.sharedXPRecipients(player), inst, def)
		}
		return &CommandResult{Messages: msgs, RoomBroadcast: roomMsgs}

	case turnMaybeDestroy:
		// The 50% destroy roll failed: it loses half its remaining body points.
		dmg := inst.CurrentHP / 2
		msgs := []string{cast}
		roomMsgs := []string{roomCast}
		if dmg > 0 {
			killed, _ := e.damageMonster(inst.ID, dmg, player.FirstName)
			msgs = append(msgs, fmt.Sprintf("The %s withers under the turning magic, but holds together! [%d Damage]", name, dmg))
			roomMsgs = append(roomMsgs, fmt.Sprintf("The %s withers under the turning magic!", name))
			if killed {
				e.handleMonsterDeath(e.sharedXPRecipients(player), inst, def)
			}
		} else {
			msgs = append(msgs, fmt.Sprintf("The %s shudders, but holds together.", name))
		}
		return &CommandResult{Messages: msgs, RoomBroadcast: roomMsgs}
	}

	return &CommandResult{
		Messages:      []string{cast, fmt.Sprintf("The %s is too powerful to be turned by this spell.", name)},
		RoomBroadcast: []string{roomCast, fmt.Sprintf("The %s ignores it.", name)},
	}
}
