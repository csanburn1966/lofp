package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Energy Maelstrom (138, Conjuration level 31), as original/csanburn/chan0409.txt shows
// it: cast at no one in particular, "A swirling vortex of black mists and crackling
// energy arises!" and knocks people and monsters in the room over, the caster's own
// group included; then "A black maelstrom of wind and energy rages!" every few seconds.
// It stays in the room when the caster leaves, and they "cannot prepare a spell while
// you maintain your current magic."
//
// MAGIC.TXT: "Flattens, damages". Four of the seven kvantar went down at the cast and
// nobody who got up was knocked over again, so the knocking over is at the cast alone.
// The damage falls on hostile creatures with each rage; there's no PvP. It had stopped
// by the time the caster came back a while later, so it lasts half a minute.

const (
	maelstromDuration       = 30 * time.Second
	maelstromMonsterStands  = 40 // % of monsters that keep their feet, resistance aside
	maintainingMagicMsg     = "You cannot prepare a spell while you maintain your current magic."
	maelstromKnockdownTime  = 12 * time.Second
	maelstromPlayerStandMax = 50
)

type maelstrom struct {
	room   int
	expiry time.Time
}

// maelstroms holds the active ones, by caster.
type maelstromSet struct {
	mu       sync.Mutex
	byCaster map[string]maelstrom
}

func (s *maelstromSet) start(caster string, m maelstrom) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byCaster == nil {
		s.byCaster = map[string]maelstrom{}
	}
	s.byCaster[caster] = m
}

func (s *maelstromSet) end(caster string) (maelstrom, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.byCaster[caster]
	delete(s.byCaster, caster)
	return m, ok
}

func (s *maelstromSet) snapshot() map[string]maelstrom {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]maelstrom, len(s.byCaster))
	for k, v := range s.byCaster {
		out[k] = v
	}
	return out
}

// maintainsMaelstrom reports whether the player has a maelstrom going.
func (e *GameEngine) maintainsMaelstrom(p *Player) bool {
	m, ok := e.maelstroms.snapshot()[p.FirstName]
	return ok && spellActive(m.expiry)
}

// releaseMaelstrom ends the caster's maelstrom, reporting whether there was one.
func (e *GameEngine) releaseMaelstrom(caster string) bool {
	m, ok := e.maelstroms.end(caster)
	if ok && e.localRoomBroadcast != nil {
		e.localRoomBroadcast(m.room, []string{"The black maelstrom dies away."})
	}
	return ok
}

// endMaelstromsIn ends every maelstrom in the room (Disjunction).
func (e *GameEngine) endMaelstromsIn(room int) {
	for caster, m := range e.maelstroms.snapshot() {
		if m.room == room {
			e.releaseMaelstrom(caster)
		}
	}
}

// castEnergyMaelstrom handles Energy Maelstrom (138).
func (e *GameEngine) castEnergyMaelstrom(player *Player, spell *SpellDef, castRoll, castChance int) *CommandResult {
	e.maelstroms.start(player.FirstName, maelstrom{room: player.RoomNumber, expiry: time.Now().Add(maelstromDuration)})

	const arises = "A swirling vortex of black mists and crackling energy arises!"
	gesture := fmt.Sprintf("%s gestures.", player.DisplayNameCap())
	type knocked struct {
		p    *Player
		line string
	}
	var downs []knocked
	var others []*Player
	if e.sessions != nil {
		for _, p := range e.sessions.OnlinePlayers() {
			if p == player || p.RoomNumber != player.RoomNumber {
				continue
			}
			others = append(others, p)
			if p.Dead || p.GMInvis || p.Position != 0 || p.IsFormLocked() || p.IsImprisoned() || p.TooInsubstantialToStop() {
				continue
			}
			if rand.Intn(100) < min(p.Agility/5, maelstromPlayerStandMax) {
				continue
			}
			p.Position = 2
			p.KnockedDown = true
			name := p.DisplayNameCap()
			if p.IsConcealed() {
				name = "Someone"
			}
			downs = append(downs, knocked{p, fmt.Sprintf("%s is knocked over!", name)})
		}
	}
	var monLines []string
	if e.monsterMgr != nil {
		e.monsterMgr.mu.Lock()
		for i := range e.monsterMgr.instances {
			m := &e.monsterMgr.instances[i]
			def := e.monsters[m.DefNumber]
			if !m.Alive || m.RoomNumber != player.RoomNumber || def == nil || m.Imprisoned || m.SummonerName == player.FirstName ||
				def.Discorporate || !monsterCanBeKnockedDown(def) || monsterIsDown(m) {
				continue
			}
			if rand.Intn(100) < maelstromMonsterStands || magicResisted(player, castRoll, castChance, def.MagicResist) {
				continue
			}
			m.KnockedDown = true
			m.KnockedDownExpiry = time.Now().Add(maelstromKnockdownTime)
			name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
			monLines = append(monLines, fmt.Sprintf("%s%s is knocked over!", capArticle(articleFor(name, def.Unique)), name))
		}
		e.monsterMgr.mu.Unlock()
	}

	// Everyone sees the same lines except that those knocked over see their own in the
	// second person, so they're told one by one.
	linesFor := func(viewer *Player) []string {
		lines := []string{arises}
		for _, d := range downs {
			if d.p == viewer {
				lines = append(lines, "You are knocked over!")
			} else {
				lines = append(lines, d.line)
			}
		}
		return append(lines, monLines...)
	}
	for _, p := range others {
		e.sendToPlayerMsgs(p, append([]string{gesture}, linesFor(p)...))
	}
	return &CommandResult{Messages: append([]string{"You gesture."}, linesFor(player)...)}
}

// maelstromTick rages each maelstrom once, or ends it when its time is up or its caster
// is dead or gone.
func (e *GameEngine) maelstromTick() {
	for caster, m := range e.maelstroms.snapshot() {
		var p *Player
		if e.sessions != nil {
			p = e.findOnlinePlayerByName(caster)
		}
		if p == nil || p.Dead || !spellActive(m.expiry) {
			e.releaseMaelstrom(caster)
			continue
		}
		e.maelstromRage(p, m.room)
	}
}

// maelstromRage is one rage of a maelstrom: it damages the hostile creatures in the
// room, along with any others that are fighting there.
func (e *GameEngine) maelstromRage(caster *Player, room int) {
	if e.localRoomBroadcast != nil {
		e.localRoomBroadcast(room, []string{"A black maelstrom of wind and energy rages!"})
	}
	spell := FindSpellByID(138)
	if spell == nil || e.monsterMgr == nil {
		return
	}
	for _, m := range e.monsterMgr.MonstersInRoom(room) {
		def := e.monsters[m.DefNumber]
		if def == nil || m.IsSummoned || m.Sedated || m.Imprisoned || (def.Strategy < 301 && m.Target == "") {
			continue
		}
		dmg := rand.Intn(spell.DmgMax-spell.DmgMin+1) + spell.DmgMin
		if level, ok := def.Immunities[elementalImmunityType(spell.DmgType)]; ok {
			dmg = applyImmunity(dmg, level)
		}
		if dmg <= 0 {
			continue
		}
		e.maelstromStrike(caster, m, def, dmg)
	}
}

func (e *GameEngine) maelstromStrike(caster *Player, m MonsterInstance, def *gameworld.MonsterDef, dmg int) {
	name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
	if e.localRoomBroadcast != nil {
		e.localRoomBroadcast(m.RoomNumber, []string{
			fmt.Sprintf("Energy crackles over %s%s!", articleFor(name, def.Unique), name),
			fmt.Sprintf(" %s %s to %s. [%d Damage]", damageSeverity(dmg, m.MaxHP), spellDmgNoun("electric"), randomBodyPart(def.BodyType), dmg),
		})
	}
	killed, _ := e.damageMonster(m.ID, dmg, caster.FirstName)
	if !killed {
		return
	}
	e.announceMonsterDeath(&m, def)
	if recipients := e.dotKillRecipients(caster.FirstName); len(recipients) > 0 {
		e.handleMonsterDeath(recipients, &m, def)
	} else {
		e.clearCombatForMonster(m.ID)
	}
}
