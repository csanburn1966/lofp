package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Disjunction (230, Enchantment level 21): per MAGIC.TXT it "dispels ALL spells active
// in the area... It is the only way (other than time) to get someone out of an
// Imprisonment bubble." Everyone in the room is caught in it, friend or foe, the caster
// included. The lines are from original/csanburn/chan0409.txt.
//
// What it leaves alone: Phantom Form ("Undispellable invisibility"), Mist, Slime and
// Wolf Form (ended by TRANSFORM), psionics, control over summoned creatures, sigils, and
// withered limbs.

// fadeLines are the lines (to the player, to the room) for a spell ending on someone,
// from the original captures; %s in the room line is who it ended on.
var fadeLines = map[string][2]string{
	"Mystic Armor":           {"The glowing outline of armor fades from around you.", "The glowing outline of armor fades from around %s."},
	"Spectral Shield":        {"The ghostly shield protecting you slowly fades.", "The ghostly shield protecting %s slowly fades."},
	"Globe of Protection":    {"The prismatic globe vanishes from around you.", "The prismatic globe around %s vanishes."},
	"Globe of Protection II": {"The brilliant prismatic globe vanishes from around you.", "The brilliant prismatic globe around %s vanishes."},
	"Haste":                  {"You return to normal speed.", "%s returns to normal speed."},
	"Slow":                   {"You return to normal speed.", "%s returns to normal speed."},
	"Strength":               {"You feel the strength leave you.", ""},
	"Divine Blessing":        {"The divine light fades from you.", "%s looks less confident."},
	"True Aim":               {trueAimFadeMsg, "The gleam leaves %s's eyes."},
	"Imprison":               {"The blue force bubble vanishes from around you.", "The blue force bubble vanishes from around %s."},
	"Invisibility":           {"You suddenly fade into view.", "%s suddenly fades into view."},
	"Plant Snare":            {"The plant tendrils ensnaring you rapidly wither and turn to dust.", "The plant tendrils ensnaring %s rapidly wither and turn to dust."},
	"Light":                  {"The glow surrounding you fades away.", "The glow surrounding %s fades away."},
	"Web":                    {"The sticky webbing holding you falls away.", "The sticky webbing holding %s falls away."},
	"Wall of Thorns":         {"The wall of thorns around you withers and crumbles away.", "The wall of thorns around %s withers and crumbles away."},
	"Creeping Death":         {"The black vines around you rapidly wither and turn to dust.", "The black vines around %s rapidly wither and turn to dust."},
}

// spellFade is one spell that a Disjunction ended on a player.
type spellFade struct {
	player   *Player // nil for a monster
	selfLine string
	roomLine string
}

// disjunctPlayer strips every spell from the player, returning what faded.
func (e *GameEngine) disjunctPlayer(p *Player) []spellFade {
	var fades []spellFade
	add := func(name, selfLine string) {
		f := spellFade{player: p, selfLine: selfLine}
		if lines, ok := fadeLines[name]; ok {
			f.selfLine = lines[0]
			if lines[1] != "" && (!p.IsConcealed() || name == "Invisibility") {
				f.roomLine = capitalize(strings.Replace(lines[1], "%s", p.DisplayName(), 1))
			}
		}
		fades = append(fades, f)
	}
	if p.IsImprisoned() {
		p.ImprisonedExpiry = time.Time{}
		add("Imprison", "")
	}
	// Each removal shifts the lists the other effects index into, so they're taken
	// one at a time off a fresh list.
	lostFlight := false
	for {
		effects := activeDispellableEffects(p)
		if len(effects) == 0 {
			break
		}
		line := effects[0].remove()
		add(effects[0].name, line)
		lostFlight = lostFlight || effects[0].name == "Fly"
	}
	if p.LightSpellActive() {
		p.LightSpellExpiry = time.Time{}
		add("Light", "")
	}
	if p.NightVisionActive() {
		p.NightVisionExpiry = time.Time{}
		add("Night Vision", "Your night vision fades.")
	}
	if p.Invisible {
		p.Invisible = false
		add("Invisibility", "")
	}
	if spellActive(p.CharmExpiry) {
		p.CharmExpiry = time.Time{}
		add("Charm", "You come to your senses.")
	}
	for _, item := range []*InventoryItem{p.Wielded, p.OffHand} {
		if item != nil && spellActive(item.BladeSpellExpiry) {
			item.BladeSpellExpiry = time.Now().Add(-time.Second)
			revertExpiredBladeSpell(item)
			add("Blade", "The elemental energy fades from your weapon.")
		}
	}
	if wake := wakeFromSlumber(p); wake != "" {
		add("Slumber", wake)
	}
	if lostFlight {
		if fallMsgs := e.fallIfFlightLost(context.Background(), p, true); fallMsgs != nil {
			for _, m := range fallMsgs {
				fades = append(fades, spellFade{player: p, selfLine: m})
			}
		} else if p.Position == 4 && !p.IsFlying() {
			p.Position = 0
		}
	}
	if len(fades) > 0 {
		e.SavePlayer(context.Background(), p)
	}
	return fades
}

// disjunctMonster strips every spell from a monster, returning the room's lines. Called
// with monsterMgr.mu held.
func disjunctMonster(m *MonsterInstance, name string) []string {
	var lines []string
	fade := func(key string) {
		if l, ok := fadeLines[key]; ok && l[1] != "" {
			lines = append(lines, capitalize(strings.Replace(l[1], "%s", name, 1)))
		}
	}
	if m.Imprisoned {
		m.Imprisoned, m.ImprisonExpiry = false, time.Time{}
		fade("Imprison")
	}
	for _, b := range m.TimedDefenseBuffs {
		if spellActive(b.Expiry) {
			fade(b.SpellName)
		}
	}
	m.TimedDefenseBuffs = nil
	if m.MysticArmorBonus > 0 && spellActive(m.MysticArmorExpiry) {
		fade("Mystic Armor")
	}
	m.MysticArmorBonus, m.MysticArmorExpiry = 0, time.Time{}
	if spellActive(m.HasteExpiry) || spellActive(m.SlowExpiry) {
		fade("Haste")
	}
	m.HasteExpiry, m.SlowExpiry = time.Time{}, time.Time{}
	if m.DivineBlessingBonus > 0 && spellActive(m.DivineBlessingExpiry) {
		fade("Divine Blessing")
	}
	m.DivineBlessingBonus, m.DivineBlessingExpiry = 0, time.Time{}
	m.StrengthBuffID, m.StrengthBuffBonus, m.StrengthBuffExpiry = 0, 0, time.Time{}
	m.AgilityBuffID, m.AgilityBuffBonus, m.AgilityBuffExpiry = 0, 0, time.Time{}
	if m.Webbed {
		m.Webbed = false
		fade("Web")
	}
	if m.Tentacled {
		m.Tentacled = false
		lines = append(lines, fmt.Sprintf("The tentacles release %s and sink back into the ground.", name))
	}
	for _, ds := range damagingSnares {
		if on, _, _ := ds.monster(m); *on {
			*on = false
			lines = append(lines, fmt.Sprintf(ds.monDrop, name))
		}
	}
	if m.snared() {
		m.SnareExpiry = time.Time{}
		fade("Plant Snare")
	}
	if m.thorned() {
		m.ThornExpiry = time.Time{}
		fade("Wall of Thorns")
	}
	if m.Sleeping {
		m.Sleeping, m.SleepExpiry, m.SleepStand = false, time.Time{}, true
		lines = append(lines, capitalize(name)+" stirs and wakes up.")
	}
	m.Feared, m.Charmed, m.CharmTarget, m.Silenced = false, false, "", false
	return lines
}

// disjunct ends every spell in the room. Everyone there is told, each seeing their own
// spells end in the second person; from's lines are returned instead of sent, for the
// caller's command result (from is nil when a monster cast it). opening lines come first
// for everyone but from. Takes monsterMgr.mu itself.
func (e *GameEngine) disjunct(room int, from *Player, opening []string) []string {
	var players []*Player
	if from != nil {
		players = append(players, from)
	}
	if e.sessions != nil {
		for _, p := range e.sessions.OnlinePlayers() {
			if p != from && p.RoomNumber == room {
				players = append(players, p)
			}
		}
	}
	e.endMaelstromsIn(room)
	var fades []spellFade
	// Someone whose flight fails over open sky may fall out of the room; they still
	// see their own lines.
	for _, p := range players {
		fades = append(fades, e.disjunctPlayer(p)...)
	}
	if e.monsterMgr != nil {
		e.monsterMgr.mu.Lock()
		for i := range e.monsterMgr.instances {
			m := &e.monsterMgr.instances[i]
			def := e.monsters[m.DefNumber]
			if !m.Alive || m.RoomNumber != room || def == nil {
				continue
			}
			name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
			for _, line := range disjunctMonster(m, articleFor(name, def.Unique)+name) {
				fades = append(fades, spellFade{roomLine: line})
			}
		}
		e.monsterMgr.mu.Unlock()
	}

	var fromLines []string
	for _, p := range players {
		var lines []string
		if p != from {
			lines = append(lines, opening...)
		}
		for _, f := range fades {
			if f.player == p {
				lines = append(lines, f.selfLine)
			} else if f.roomLine != "" {
				lines = append(lines, f.roomLine)
			}
		}
		if p == from {
			fromLines = lines
		} else {
			e.sendToPlayerMsgs(p, lines)
		}
	}
	return fromLines
}

// castDisjunction handles a player casting Disjunction (230).
func (e *GameEngine) castDisjunction(player *Player) *CommandResult {
	name := player.DisplayName()
	lines := e.disjunct(player.RoomNumber, player, []string{
		fmt.Sprintf("%s gestures.", player.DisplayNameCap()),
		fmt.Sprintf("A wave of deep red light emanates from %s.", name),
	})
	return &CommandResult{
		Messages: append([]string{"You gesture.", "A wave of deep red light emanates from you."}, lines...),
	}
}

// resolveMonsterDisjunction handles a monster's Disjunction: "A greater werewolf
// gestures with one humanlike hand at Mortif. / A wave of deep red light emanates from
// a werewolf." (badge.txt)
func (e *GameEngine) resolveMonsterDisjunction(inst *MonsterInstance, def *gameworld.MonsterDef, gestureMsg string) {
	name := FormatMonsterName(def, e.monAdjs)
	e.disjunct(inst.RoomNumber, nil, []string{
		gestureMsg,
		fmt.Sprintf("A wave of deep red light emanates from %s%s.", articleFor(name, def.Unique), name),
	})
}

// castDispelInvisibility handles Dispel Invisibility (406): everyone in the room under
// the Invisibility spell, or only the named player, becomes visible. Phantom Form is
// "Undispellable invisibility" (MAGIC.TXT), and hiding isn't magic. The captures show
// nothing at all when there's no one to reveal.
func (e *GameEngine) castDispelInvisibility(player *Player, args []string) *CommandResult {
	result := &CommandResult{
		Messages:      []string{"You gesture."},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures.", player.DisplayNameCap())},
	}
	if e.sessions == nil {
		return result
	}
	only := strings.ToLower(strings.Join(args, " "))
	for _, p := range e.sessions.OnlinePlayers() {
		if p == player || p.RoomNumber != player.RoomNumber || !p.Invisible || p.GMInvis || p.GMHidden {
			continue
		}
		if only != "" && !p.NameMatches(only) {
			continue
		}
		e.revealInvisible(p)
		e.sendToPlayerMsgs(p, []string{"You suddenly fade into view.", "You suddenly become visible."})
		// chandra.txt and moryan.txt show both lines, one after the other.
		lines := []string{
			fmt.Sprintf("%s suddenly fades into view.", p.DisplayNameCap()),
			fmt.Sprintf("%s suddenly becomes visible.", p.DisplayNameCap()),
		}
		result.Messages = append(result.Messages, lines...)
		for _, o := range e.sessions.OnlinePlayers() {
			if o != p && o != player && o.RoomNumber == player.RoomNumber {
				e.sendToPlayerMsgs(o, lines)
			}
		}
	}
	return result
}

func (e *GameEngine) revealInvisible(p *Player) {
	p.Invisible = false
	e.SavePlayer(context.Background(), p)
}
