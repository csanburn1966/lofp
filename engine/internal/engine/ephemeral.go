package engine

import (
	"fmt"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Ephemeral race traits — LEGENDS.DOC "Ephemeral:":
//
//   - "non-magical weapons do not harm them except when they themselves go on the
//     offense with weapons, magic or psionics. This draws them onto the physical
//     plane and causes them to become vulnerable to mundane attacks."
//     Preparing a spell, selecting a psionic discipline, casting/projecting,
//     advancing on or attacking a target makes an Ephemeral substantial for
//     ephemeralPhysicalDuration, refreshed by each further offensive act;
//     ephemeralFadeTick returns them to their ethereal form once it lapses. While
//     ethereal, monsters' mundane physical attacks (weapons, claws, bites, physical
//     special attacks from a monster with no WEAPONPLUS bonus) pass through them.
//     Spells, psionics, elemental special attacks and magical monsters still harm.
//   - "they do not bleed" — Player.CanBleed.
//   - "they cannot manipulate objects that weigh more than a couple of pounds" —
//     ephemeralTooHeavy gates GET, WIELD, WEAR, being handed an item, and BUY.
//     Gear a character already had when this was added is left alone.
//   - "Ephemerals can see well in the dark" — darkness gives hiders no bonus
//     against an Ephemeral's SEARCH (see hideConcealmentBonus).
//   - "natural telepaths" — innate TELEPATHY, handled in ProcessCommand/NewPlayer.
//
// The original numbers don't survive; the constants below are tunable.
const (
	ephemeralPhysicalDuration = 30 * time.Second
	ephemeralMaxHandleWeight  = 2 // pounds ("a couple of pounds")
)

// IsEphemeral reports whether the player is of the Ephemeral race.
func (p *Player) IsEphemeral() bool { return p.Race == RaceEphemeral }

// EphemeralEthereal reports whether the player is an Ephemeral who hasn't been
// drawn onto the physical plane by recent offensive action, so mundane physical
// attacks pass through them.
func (p *Player) EphemeralEthereal() bool {
	return p.IsEphemeral() && !time.Now().Before(p.EphemeralPhysicalUntil)
}

// CanBleed reports whether wounds dealt to the player bleed: not undead (no active
// life force — GMSCRIPT.DOC immunity #7) and not an Ephemeral (LEGENDS.DOC).
func (p *Player) CanBleed() bool { return !p.IsUndead() && !p.IsEphemeral() }

// ephemeralGoOnOffense draws an Ephemeral onto the physical plane for an offensive
// act that has just gone through: starts or refreshes the substantial window and,
// on the transition from ethereal, adds the "becomes more substantial" echoes to
// res. A no-op for every other race. Call it only once the action is committed —
// a refused command (unknown spell, no target, round time) shouldn't count.
func ephemeralGoOnOffense(p *Player, res *CommandResult) *CommandResult {
	if res == nil || !p.IsEphemeral() {
		return res
	}
	p.EphemeralPhysicalUntil = time.Now().Add(ephemeralPhysicalDuration)
	if p.EphemeralSubstantial {
		return res
	}
	p.EphemeralSubstantial = true
	res.Messages = append(res.Messages, "You become more substantial.")
	res.RoomBroadcast = append(res.RoomBroadcast, fmt.Sprintf("%s becomes more substantial.", p.DisplayNameCap()))
	return res
}

// ephemeralFadeTick returns Ephemerals whose substantial window has lapsed to
// their ethereal form. Runs on the 3-second monster loop.
func (e *GameEngine) ephemeralFadeTick() {
	if e.sessions == nil {
		return
	}
	now := time.Now()
	for _, p := range e.sessions.OnlinePlayers() {
		if !p.EphemeralSubstantial || now.Before(p.EphemeralPhysicalUntil) {
			continue
		}
		p.EphemeralSubstantial = false
		if e.sendToPlayer != nil {
			e.sendToPlayer(p.FirstName, []string{"You become more ethereal."})
		}
		if !p.IsConcealed() && e.roomBroadcastExclude != nil {
			e.roomBroadcastExclude(p.RoomNumber, p.FirstName, []string{fmt.Sprintf("%s becomes more ethereal.", p.DisplayNameCap())})
		}
	}
}

// ephemeralPassesThrough reports whether a monster's mundane physical attack
// passes through player: an ethereal Ephemeral, and a monster with no magic
// bonus (WEAPONPLUS — the same field that decides whether a monster can harm a
// MAGICWEAPON-protected guard, see the elemental guard intercept in combat.go).
func ephemeralPassesThrough(player *Player, def *gameworld.MonsterDef) bool {
	return player.EphemeralEthereal() && def.WeaponPlus <= 0
}

// ephemeralTooHeavy returns the refusal message if player is an Ephemeral and an
// item of the given total weight (including anything inside it) is more than
// they can manipulate, or "" if they can handle it.
func ephemeralTooHeavy(player *Player, weight int) string {
	if !player.IsEphemeral() || weight <= ephemeralMaxHandleWeight {
		return ""
	}
	return "Your insubstantial form cannot grasp anything that heavy."
}

// inventoryItemWeight is an item's own weight plus, recursively, its contents'.
func (e *GameEngine) inventoryItemWeight(ii *InventoryItem) int {
	total := 0
	if def := e.items[ii.Archetype]; def != nil {
		total = def.Weight
	}
	for i := range ii.Contents {
		total += e.inventoryItemWeight(&ii.Contents[i])
	}
	return total
}

// roomItemWeight is a floor item's own weight plus, for a container, everything in it.
func (e *GameEngine) roomItemWeight(room *gameworld.Room, ri *gameworld.RoomItem) int {
	def := e.items[ri.Archetype]
	if def == nil {
		return 0
	}
	total := def.Weight
	if isContainerDef(def) {
		total += e.roomContainerContentsWeight(room, ri.Ref)
	}
	return total
}
