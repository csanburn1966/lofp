package engine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Monsters' spells that aren't plain damage. Every monster spell without its own handler
// used to go off as a bolt of energy for 1 point, so a giant spider's Web, a werewolf's
// Disjunction and Fear, a shaman's Plant Snare and a priest's Mystic Armor all did
// nothing else.
//
// Not handled: Invisibility (225; the phase spider and three others), since monsters
// can't be invisible. They don't try to cast it.

// monsterSelfBuffs are the spells a monster casts on itself rather than at anyone.
var monsterSelfBuffs = map[int]bool{
	102: true, // Mystic Armor
	105: true, // Globe of Protection
	133: true, // Globe of Protection II
	210: true, // Haste
	326: true, // Spectral Shield
	511: true, // Carapace
}

// monsterUtilitySpells are the other non-damage spells resolveMonsterUtilitySpell gives
// an effect.
var monsterUtilitySpells = map[int]bool{
	webSpellID: true, plantSnareSpellID: true, wallOfThornsSpellID: true, creepingDeathSpellID: true,
	200: true,                       // Fear
	201: true,                       // Charm
	211: true,                       // Slow
	216: true, 217: true, 218: true, // Slumber
	219: true, // Silence
	226: true, // Paranoia
	230: true, // Disjunction
	231: true, // Imprison
	351: true, // Wither Limb
	519: true, // Sunray
}

// monsterCanUseSpell reports whether it's worth a monster's casting the spell: it has an
// effect, works here, and isn't one the monster already has on it.
func (e *GameEngine) monsterCanUseSpell(inst *MonsterInstance, sp *SpellDef) bool {
	if monsterSelfBuffs[sp.ID] {
		switch sp.ID {
		case 102:
			return !(inst.MysticArmorBonus > 0 && spellActive(inst.MysticArmorExpiry))
		case 210:
			return !spellActive(inst.HasteExpiry)
		}
		for _, b := range inst.TimedDefenseBuffs {
			if b.SpellID == sp.ID && spellActive(b.Expiry) {
				return false
			}
		}
		return true
	}
	if sp.ID == plantSnareSpellID {
		room := e.rooms[inst.RoomNumber]
		return room != nil && isOutdoorTerrain(room.Terrain)
	}
	_, summons := sorcerousSummonsDemon[sp.ID]
	return sp.Effect == "damage" || summons || sp.ID == insectSwarmSpellID ||
		afflictionForSpell(sp.ID) != nil || monsterUtilitySpells[sp.ID]
}

// resolveMonsterSelfBuff resolves a spell a monster casts on itself, returning the room's
// lines. The appearing lines are the captures' ("A brilliant prismatic globe encircles a
// zharkier."). Called without monsterMgr.mu held.
func (e *GameEngine) resolveMonsterSelfBuff(inst *MonsterInstance, def *gameworld.MonsterDef, sp *SpellDef) []string {
	name := FormatMonsterName(def, e.monAdjs)
	full := articleFor(name, def.Unique) + name
	gesture := capitalize(full) + " gestures."
	if texl := def.TextOverrides["TEXL"]; texl != "" {
		gesture = texl + "."
	}
	if rand.Intn(100) >= monsterCastChance(def) {
		texq := def.TextOverrides["TEXQ"]
		if texq == "" {
			texq = fmt.Sprintf("%s's spell fizzles.", capitalize(full))
		}
		return []string{gesture, texq}
	}

	duration := scaleSpellDuration(20*time.Minute, spellDurationMult(def.CastLevel, 0))
	expiry := time.Now().Add(duration)
	var line string
	e.updateMonster(inst.ID, func(m *MonsterInstance) {
		switch sp.ID {
		case 102:
			m.MysticArmorBonus, m.MysticArmorExpiry = sp.DefBonus, expiry
			line = fmt.Sprintf("The glowing outline of armor appears around %s.", full)
		case 210:
			m.SlowExpiry = time.Time{}
			m.HasteExpiry = expiry
			line = capitalize(full) + " moves with blurring speed."
		default:
			m.TimedDefenseBuffs = append(m.TimedDefenseBuffs, TimedDefenseBuff{SpellID: sp.ID, SpellName: sp.Name, Bonus: sp.DefBonus, Expiry: expiry})
			switch sp.ID {
			case 105:
				line = fmt.Sprintf("A prismatic globe encircles %s.", full)
			case 133:
				line = fmt.Sprintf("A brilliant prismatic globe encircles %s.", full)
			case 326:
				line = fmt.Sprintf("A ghostly shield hovers before %s.", full)
			case 511:
				line = fmt.Sprintf("%s's skin hardens into a tough, chitinous carapace.", capitalize(full))
			}
		}
	})
	return []string{gesture, line}
}

// cloakedMind reports whether Cloak Mind (235) wards the player: per MAGIC.TXT it "also
// protects against sleep, charm, true name and other mind-affecting spells."
func cloakedMind(p *Player) bool { return cloakMindResistBonus(p) > 0 }

// slumberBodyLimit is the most body points each Slumber can put to sleep.
var slumberBodyLimit = map[int]int{216: 80, 217: 200, 218: 450}

// wakeFromSlumber wakes a player put to sleep by a Slumber spell (it lasts "until next
// struck or until the spell wears off", MAGIC.TXT), leaving them lying down. Returns the
// line for them, or "" if they weren't asleep from one.
func wakeFromSlumber(p *Player) string {
	if p.SleepExpiry.IsZero() {
		return ""
	}
	p.SleepExpiry = time.Time{}
	if !p.Unconscious || p.BodyPoints <= 0 {
		return ""
	}
	p.Unconscious = false
	return "You wake up!"
}

// charmedBy reports whether the player is charmed by that monster (or, with no
// CharmMonsterID, by a script) and won't attack it.
func charmedBy(p *Player, monsterID int) bool {
	return (p.CharmMonsterID == monsterID || p.CharmMonsterID == 0) && spellActive(p.CharmExpiry)
}

// resolveMonsterUtilitySpell gives a monster's non-damage spell its effect on the player
// it was cast at. The results are resolveMonsterCast's. Called without monsterMgr.mu held.
func (e *GameEngine) resolveMonsterUtilitySpell(inst *MonsterInstance, def *gameworld.MonsterDef, sp *SpellDef, target *Player, gestureMsg string) (playerMsgs, roomMsgs []string, defender *Player) {
	if sp.ID == 230 { // Disjunction takes in the whole room, whoever it was waved at
		e.resolveMonsterDisjunction(inst, def, gestureMsg)
		return nil, nil, nil
	}
	done := func(selfMsgs []string, roomMsg string) ([]string, []string, *Player) {
		e.SavePlayer(context.Background(), target)
		roomMsgs := []string{gestureMsg}
		if roomMsg != "" {
			roomMsgs = append(roomMsgs, roomMsg)
		}
		return append([]string{gestureMsg}, selfMsgs...), roomMsgs, target
	}

	if resist := spellShieldResistBonus(target); resist > 0 && rand.Intn(100) < resist {
		return done([]string{"Your antimagical field flares and deflects the spell!"}, fmt.Sprintf("An antimagical field around %s deflects the spell!", target.DisplayName()))
	}
	self, room := e.hostileSpellEffect(target, sp, inst, def, 1)
	return done(self, room)
}

// hostileSpellEffect gives a player the effect of one of monsterUtilitySpells (Disjunction
// aside), returning the lines for them and for the room (nil and "" if nobody notices).
// inst and def are the monster casting it, nil for a script; mult stretches the long
// durations. The caller saves the player.
func (e *GameEngine) hostileSpellEffect(target *Player, sp *SpellDef, inst *MonsterInstance, def *gameworld.MonsterDef, mult float64) (selfMsgs []string, roomMsg string) {
	who, whoCap := target.DisplayName(), target.DisplayNameCap()
	mindAffecting := sp.ID == 200 || sp.ID == 201 || sp.ID == 226 || slumberBodyLimit[sp.ID] > 0
	if mindAffecting && cloakedMind(target) {
		return []string{"You feel something slide off your cloaked mind."}, ""
	}
	interrupt := func() {
		if msg := interruptPreparedSpell(target); msg != "" {
			selfMsgs = append(selfMsgs, msg)
		}
	}

	switch sp.ID {
	case webSpellID, plantSnareSpellID, wallOfThornsSpellID, creepingDeathSpellID:
		if selfMsg, roomMsg := snareImmunity(target, sp); selfMsg != "" {
			return []string{selfMsg}, roomMsg
		}
		return snarePlayer(target, sp)
	case 200: // Fear
		target.FearExpiry = time.Now().Add(fearDuration)
		return []string{"You are filled with terror!"}, fmt.Sprintf("%s is filled with terror!", whoCap) // badge.txt
	case 201: // Charm: a script's charm is by no one in particular, so it's of everyone
		target.CharmExpiry = time.Now().Add(60 * time.Second)
		target.CharmMonsterID = 0
		full := "anyone"
		roomMsg = fmt.Sprintf("%s gazes about fondly.", whoCap)
		if inst != nil {
			target.CharmMonsterID = inst.ID
			name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
			full = articleFor(name, def.Unique) + name
			roomMsg = fmt.Sprintf("%s gazes fondly at %s.", whoCap, full)
		}
		return []string{fmt.Sprintf("You are charmed! You can't imagine wanting to harm %s.", full)}, roomMsg
	case 211: // Slow
		if isHasted(target) {
			target.HasteExpiry = time.Time{}
			return []string{"You return to normal speed."}, fmt.Sprintf("%s returns to normal speed.", whoCap)
		}
		target.SlowExpiry = time.Now().Add(scaleSpellDuration(20*time.Minute, mult))
		return []string{"You slow down considerably."}, fmt.Sprintf("%s slows down considerably.", whoCap)
	case 216, 217, 218: // Slumber
		if target.IsUndead() || target.MaxBodyPoints > slumberBodyLimit[sp.ID] || target.Unconscious {
			return []string{"You feel drowsy for a moment."}, fmt.Sprintf("%s seems drowsy for a moment.", whoCap)
		}
		target.Unconscious = true
		target.Position = 2
		target.SleepExpiry = time.Now().Add(scaleSpellDuration(time.Duration(5+rand.Intn(6))*time.Second, mult))
		selfMsgs = []string{"You fall asleep."}
		interrupt()
		return selfMsgs, fmt.Sprintf("%s falls asleep.", whoCap)
	case 219: // Silence
		target.SilencedExpiry = time.Now().Add(60 * time.Second)
		return []string{"Your voice catches in your throat -- you have been silenced!"}, fmt.Sprintf("%s becomes absolutely silent.", whoCap)
	case 226: // Paranoia: the victim never knows
		target.ParanoiaExpiry = time.Now().Add(scaleSpellDuration(20*time.Minute, mult))
		return nil, ""
	case 231: // Imprison
		target.ImprisonedExpiry = time.Now().Add(5 * time.Minute)
		return []string{"A blue force bubble envelops you! You cannot attack or cast spells until it fades."}, fmt.Sprintf("A blue force bubble envelops %s!", who)
	case 351: // Wither Limb
		return e.witherMessages(target)
	case 519: // Sunray
		target.Stunned = true
		target.StunExpiry = time.Now().Add(time.Duration(4+rand.Intn(3)) * time.Second)
		selfMsgs = []string{"A searing ray of sunlight blinds you, and you reel, stunned!"}
		interrupt()
		return selfMsgs, fmt.Sprintf("%s is blinded by a searing ray of sunlight and reels, stunned!", whoCap)
	}
	// monsterCanUseSpell keeps monsters from choosing any other, so this is only one
	// begun before its definition changed.
	return []string{"Nothing happens."}, "Nothing happens."
}
