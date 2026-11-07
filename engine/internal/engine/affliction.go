package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Poison and disease follow the original rule from GM Pages/mongentr.html (the
// POISON field; its DISEASE field says "See POISON above for a description of how
// Disease works", and GMSCRIPT.DOC calls DISEASE "basically the same as POISON"):
// a level L affliction costs L body points the first minute, L-1 the second, and
// so on until it wears off after L minutes. original/csanburn/lysis.txt shows
// poison on HEALTH (level 9, three "Poison burns in your veins." ticks, then
// level 6); the captures end them with "The poison leaves you." and "The disease
// leaves you." A new affliction only replaces the current one of the same kind if
// it's a higher level.
//
// Undead and creatures immune to life-affecting magic (GMSCRIPT.DOC immunity #7)
// are unaffected.

// affliction describes one damage-over-time condition: poison or disease.
type affliction struct {
	spellID   int // the Necromancy spell that inflicts it
	spellName string
	noun      string // "poison", "disease"
	adj       string // "poisoned", "diseased"
	tickMsg   string // the player's once-a-minute line
	monTick   string // the onlookers' line for a monster, %s = its name
	// itemLevel is the level when an item or room script casts the spell, which
	// has no spellcraft of its own: the spell's rank in GM Pages/expl.html.
	itemLevel int
	player    func(*Player) (*bool, *int)
	monster   func(*MonsterInstance) (*bool, *int, *string)
}

const maxAfflictionLevel = 50 // same cap as ANOINT and weapon VAL4 poison (51-100)

var poisonAffliction = &affliction{
	spellID: 321, spellName: "Poison", noun: "poison", adj: "poisoned",
	tickMsg:   "Poison burns in your veins.",
	monTick:   "%s shudders as poison burns in its veins.",
	itemLevel: 20,
	player:    func(p *Player) (*bool, *int) { return &p.Poisoned, &p.PoisonLevel },
	monster: func(m *MonsterInstance) (*bool, *int, *string) {
		return &m.Poisoned, &m.PoisonLevel, &m.PoisonCasterName
	},
}

// No capture shows disease's once-a-minute line, so tickMsg is our own.
var diseaseAffliction = &affliction{
	spellID: 320, spellName: "Contagion", noun: "disease", adj: "diseased",
	tickMsg:   "The sickness ravages your body!",
	monTick:   "%s shudders as the sickness ravages its body.",
	itemLevel: 23,
	player:    func(p *Player) (*bool, *int) { return &p.Diseased, &p.DiseaseLevel },
	monster: func(m *MonsterInstance) (*bool, *int, *string) {
		return &m.Diseased, &m.DiseaseLevel, &m.DiseaseCasterName
	},
}

var afflictions = []*affliction{poisonAffliction, diseaseAffliction}

// afflictionForSpell returns the affliction a spell inflicts, or nil.
func afflictionForSpell(spellID int) *affliction {
	for _, a := range afflictions {
		if a.spellID == spellID {
			return a
		}
	}
	return nil
}

func (a *affliction) endMsg() string { return fmt.Sprintf("The %s leaves you.", a.noun) }

// afflictPlayer gives the player the affliction at the given level unless they're
// undead or already suffering one at least that strong. Returns whether it took hold.
func afflictPlayer(p *Player, a *affliction, level int) bool {
	if p.IsUndead() || level < 1 {
		return false
	}
	on, lvl := a.player(p)
	if *on && *lvl >= level {
		return false
	}
	*on, *lvl = true, level
	return true
}

func poisonPlayer(p *Player, level int) bool  { return afflictPlayer(p, poisonAffliction, level) }
func diseasePlayer(p *Player, level int) bool { return afflictPlayer(p, diseaseAffliction, level) }

// playerAfflictionSpellLevel is the level of a player's Poison or Contagion:
// their Spellcraft rank, capped at maxAfflictionLevel.
func playerAfflictionSpellLevel(p *Player) int {
	return max(1, min(p.Skills[23], maxAfflictionLevel))
}

// monsterAfflictionSpellLevel is the level of a monster's Poison or Contagion. A
// monster's SPELLSKILL runs on a larger scale than a player's Spellcraft rank
// (65-200 among the monsters that know 321), so a quarter of it is used.
func monsterAfflictionSpellLevel(def *gameworld.MonsterDef) int {
	return max(1, min(def.SpellSkill/4, maxAfflictionLevel))
}

// monsterAfflictionLevel adjusts a poison or disease level for the monster it's
// being put on: 0 if the monster can't be afflicted, otherwise scaled by its
// life-affecting immunity (#7) the way weapon and spell damage are scaled by theirs.
func monsterAfflictionLevel(def *gameworld.MonsterDef, level int) int {
	if def.Race == 22 || def.Discorporate {
		return 0
	}
	if imm, ok := def.Immunities[7]; ok {
		level = applyImmunity(level, imm)
	}
	return level
}

// afflictMonster gives the monster instance the affliction at the given (already
// adjusted) level unless it's carrying one at least that strong. casterName is
// credited if it kills the monster. Returns whether it took hold.
func (e *GameEngine) afflictMonster(instID int, a *affliction, level int, casterName string) bool {
	if level < 1 || e.monsterMgr == nil {
		return false
	}
	e.monsterMgr.mu.Lock()
	defer e.monsterMgr.mu.Unlock()
	for i := range e.monsterMgr.instances {
		inst := &e.monsterMgr.instances[i]
		if inst.ID != instID || !inst.Alive {
			continue
		}
		on, lvl, caster := a.monster(inst)
		if *on && *lvl >= level {
			return false
		}
		*on, *lvl, *caster = true, level, casterName
		return true
	}
	return false
}

func (e *GameEngine) poisonMonster(instID, level int, casterName string) bool {
	return e.afflictMonster(instID, poisonAffliction, level, casterName)
}

// playerAfflictionTick runs one minute of the affliction on a player, from
// regenTick. A tick that takes them to 0 knocks them out; a tick while they were
// already out at 0 when regenTick started (wasUnconsciousAtZero) kills them.
// Mist Form / Slime Form take no damage from anything, so the affliction waits.
func (e *GameEngine) playerAfflictionTick(p *Player, a *affliction, wasUnconsciousAtZero bool) (died, changed bool) {
	on, lvlp := a.player(p)
	if !*on || p.IsFormLocked() {
		return false, false
	}
	if wasUnconsciousAtZero {
		if e.sendToPlayer != nil {
			e.sendToPlayer(p.FirstName, e.handlePlayerDeath(p, a.noun))
		}
		return true, true
	}
	lvl := max(*lvlp, 1)
	p.BodyPoints = max(p.BodyPoints-lvl, 0)
	msgs := []string{fmt.Sprintf("%s [-%d BP, %d/%d]", a.tickMsg, lvl, p.BodyPoints, p.MaxBodyPoints)}
	if p.BodyPoints == 0 {
		p.Unconscious = true
		p.Position = 2
		msgs = append(msgs, "You collapse, unconscious!")
	}
	*lvlp = lvl - 1
	if *lvlp <= 0 {
		*on, *lvlp = false, 0
		msgs = append(msgs, a.endMsg())
	}
	if e.sendToPlayer != nil {
		e.sendToPlayer(p.FirstName, msgs)
	}
	return false, true
}

// monsterAfflictionTick is the once-a-minute poison and disease tick for
// monsters, run from StartRegenCycle alongside the players' regenTick.
func (e *GameEngine) monsterAfflictionTick() {
	defer recoverBackground("monster affliction tick")
	if e.monsterMgr == nil {
		return
	}

	type afflictionKill struct {
		inst   MonsterInstance
		def    *gameworld.MonsterDef
		caster string
	}
	var kills []afflictionKill
	type roomMsg struct {
		room int
		msgs []string
	}
	var out []roomMsg

	e.monsterMgr.mu.Lock()
	for i := range e.monsterMgr.instances {
		inst := &e.monsterMgr.instances[i]
		def := e.monsters[inst.DefNumber]
		for _, a := range afflictions {
			on, lvl, caster := a.monster(inst)
			if !*on {
				continue
			}
			if !inst.Alive || def == nil {
				*on, *lvl, *caster = false, 0, ""
				continue
			}
			name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
			article := articleFor(name, def.Unique)

			dmg := max(*lvl, 1)
			inst.CurrentHP -= dmg
			msgs := []string{fmt.Sprintf(a.monTick+" [%d Damage]", capArticle(article)+name, dmg)}
			*lvl = dmg - 1
			if inst.CurrentHP <= 0 {
				inst.Alive = false
				inst.CurrentHP = 0
				inst.DeathTime = time.Now()
				kills = append(kills, afflictionKill{inst: *inst, def: def, caster: *caster})
				for _, b := range afflictions {
					bOn, bLvl, bCaster := b.monster(inst)
					*bOn, *bLvl, *bCaster = false, 0, ""
				}
			} else if *lvl <= 0 {
				*on, *lvl, *caster = false, 0, ""
				msgs = append(msgs, fmt.Sprintf("The %s leaves %s%s.", a.noun, article, name))
			}
			out = append(out, roomMsg{room: inst.RoomNumber, msgs: msgs})
		}
	}
	e.monsterMgr.mu.Unlock()

	if e.localRoomBroadcast != nil {
		for _, m := range out {
			e.localRoomBroadcast(m.room, m.msgs)
		}
	}
	for _, k := range kills {
		name := strings.ToLower(FormatMonsterName(k.def, e.monAdjs))
		capName := capArticle(articleFor(name, k.def.Unique)) + name
		if e.localRoomBroadcast != nil {
			if deathText := k.def.TextOverrides["TEXD"]; deathText != "" {
				e.localRoomBroadcast(k.inst.RoomNumber, []string{fmt.Sprintf("%s %s", capName, deathText)})
			} else {
				e.localRoomBroadcast(k.inst.RoomNumber, []string{fmt.Sprintf("%s collapses, dead!", capName)})
			}
		}
		if recipients := e.dotKillRecipients(k.caster); len(recipients) > 0 {
			e.handleMonsterDeath(recipients, &k.inst, k.def)
		}
	}
}

// castAfflictionSpell handles a player casting Poison (321) or Contagion (320)
// at a monster. There's no PvP, so a player target is spared.
func (e *GameEngine) castAfflictionSpell(player *Player, a *affliction, args []string, castRoll, castChance int) *CommandResult {
	targetName := strings.Join(args, " ")
	if targetName == "" {
		targetName = e.autoTargetMonsterName(player)
	}
	if targetName == "" {
		return &CommandResult{Messages: []string{fmt.Sprintf("Cast %s at what?", a.spellName)}}
	}

	inst, def, refusal := e.findHostileMonsterInRoom(player, targetName)
	if refusal != nil {
		return refusal
	}
	if inst == nil {
		if e.sessions != nil {
			for _, p := range e.sessions.OnlinePlayers() {
				if p != player && p.RoomNumber == player.RoomNumber && !p.IsConcealed() && p.NameMatches(strings.ToLower(targetName)) {
					return &CommandResult{
						Messages:      []string{fmt.Sprintf("You gesture at %s.", p.DisplayName()), fmt.Sprintf("Your spell passes harmlessly by %s.", p.DisplayName())},
						RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), p.DisplayName())},
					}
				}
			}
		}
		return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", targetName)}, TargetNotFound: true}
	}

	name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
	article := articleFor(name, def.Unique)
	gestureSelf := fmt.Sprintf("You gesture at %s%s.", article, name)
	gestureRoom := fmt.Sprintf("%s gestures at %s%s.", player.DisplayNameCap(), article, name)

	if magicResisted(player, castRoll, castChance, def.MagicResist) {
		return &CommandResult{
			Messages:      []string{gestureSelf, fmt.Sprintf("The %s resists the spell!", name)},
			RoomBroadcast: []string{gestureRoom, fmt.Sprintf("The %s resists the spell!", name)},
			Resisted:      true,
		}
	}

	level := monsterAfflictionLevel(def, playerAfflictionSpellLevel(player))
	if level < 1 {
		line := fmt.Sprintf("%s%s seems unaffected.", capArticle(article), name)
		return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}}
	}
	if !e.afflictMonster(inst.ID, a, level, player.FirstName) {
		return &CommandResult{
			Messages:      []string{gestureSelf, fmt.Sprintf("The %s already in %s%s is stronger than yours.", a.noun, article, name)},
			RoomBroadcast: []string{gestureRoom},
		}
	}
	line := fmt.Sprintf("%s%s is %s!", capArticle(article), name, a.adj)
	return &CommandResult{
		Messages:      []string{gestureSelf, line},
		RoomBroadcast: []string{gestureRoom, line},
	}
}

// resolveMonsterAfflictionCast applies a monster's successful Poison or Contagion
// cast to its target player, as resolveMonsterCast does for Insect Swarm.
func (e *GameEngine) resolveMonsterAfflictionCast(def *gameworld.MonsterDef, a *affliction, target *Player, gestureMsg string) (playerMsgs, roomMsgs []string) {
	if target.IsUndead() {
		return []string{gestureMsg, "You feel a brief chill, but nothing more."}, nil
	}
	if !afflictPlayer(target, a, monsterAfflictionSpellLevel(def)) {
		return []string{gestureMsg, fmt.Sprintf("You feel a brief wave of nausea, but the %s already in you is stronger.", a.noun)}, nil
	}
	e.SavePlayer(context.Background(), target)
	return []string{gestureMsg, fmt.Sprintf("You are %s!", a.adj)}, []string{fmt.Sprintf("%s is %s!", target.DisplayName(), a.adj)}
}
