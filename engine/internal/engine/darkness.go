package engine

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Darkness — a room scripted with DARKNESS lighting (GMSCRIPT.DOC) can't be seen
// in unless something lights it or the viewer can see in the dark. Matches the
// original's two messages (original/csanburn/*.txt, legends/*.cap):
//
//	> look                       > look sw  /  look in chest  /  skin ...
//	[Mine, Entrance]             It is too dark to do that.
//	It is too dark to see anything.
//	Obvious exits: sw, out.
//
// Only DARKNESS rooms are ever dark — day/night and the other lighting types
// don't affect sight (a deliberate scope decision, not an oversight).
//
// What lights a dark room (roomIsUnlitDark): any LIT item lying in it or carried,
// wielded or worn by anyone there (torch, lantern, flashlight...), or anyone there
// with an active Light spell (113). What lets one viewer see anyway
// (seesInDark): Night Vision (520), being a Murg, Highlander or Ephemeral
// (LEGENDS.DOC: they "see well in darkness"), a Beta or better Bionic Eye
// (cyberware.go), or being a GM.
const (
	tooDarkToSeeMsg = "It is too dark to see anything."
	tooDarkToDoMsg  = "It is too dark to do that."

	lightSpellDuration   = 20 * time.Minute // per cast; recasting extends it
	nightVisionDuration  = 20 * time.Minute
	darkSpellMaxDuration = 4 * time.Hour // cap on stacked recasts, like Camouflage
)

// LightSpellActive reports whether the player is carrying a Light spell (113) glow.
func (p *Player) LightSpellActive() bool {
	return !p.LightSpellExpiry.IsZero() && time.Now().Before(p.LightSpellExpiry)
}

// NightVisionActive reports whether the player's Night Vision spell (520) is up.
func (p *Player) NightVisionActive() bool {
	return !p.NightVisionExpiry.IsZero() && time.Now().Before(p.NightVisionExpiry)
}

// seesInDark reports whether p can see in an unlit dark room.
func (p *Player) seesInDark() bool {
	switch p.Race {
	case RaceMurg, RaceHighlander, RaceEphemeral:
		return true
	}
	return p.NightVisionActive() || p.cyberNightVision() || p.IsGM
}

// tooDarkToSee reports whether the player's current room is too dark for them to see in.
func (e *GameEngine) tooDarkToSee(p *Player) bool {
	return !p.seesInDark() && e.roomIsUnlitDark(p.RoomNumber)
}

// tooDarkResult is the refusal for a sight-based command in the dark.
func tooDarkResult() *CommandResult {
	return &CommandResult{Messages: []string{tooDarkToDoMsg}}
}

// extendTimedSpell starts a timed spell at duration, or adds duration to one
// already running (capped at darkSpellMaxDuration from now). Returns the new
// expiry, the minutes left, and whether it was freshly started.
func extendTimedSpell(expiry time.Time, duration time.Duration) (time.Time, int, bool) {
	now := time.Now()
	if !expiry.IsZero() && now.Before(expiry) {
		next := expiry.Add(duration)
		if limit := now.Add(darkSpellMaxDuration); next.After(limit) {
			next = limit
		}
		return next, int(time.Until(next).Minutes()) + 1, false
	}
	return now.Add(duration), int(duration.Minutes()), true
}

// castLightSpell handles Light (113): a soft glow surrounds the caster, or a named
// player in the room, lighting any dark room they're in until it fades.
func (e *GameEngine) castLightSpell(player *Player, spell *SpellDef, args []string) *CommandResult {
	target, isSelf, refusal := e.darkSpellTarget(player, args)
	if refusal != nil {
		return refusal
	}
	var mins int
	var fresh bool
	target.LightSpellExpiry, mins, fresh = extendTimedSpell(target.LightSpellExpiry, lightSpellDuration)
	if !fresh {
		return e.darkSpellResult(player, target, isSelf, spell,
			fmt.Sprintf("The glow around you brightens. (%d minutes remaining)", mins),
			fmt.Sprintf("The glow around %s brightens.", target.DisplayName()))
	}
	return e.darkSpellResult(player, target, isSelf, spell,
		fmt.Sprintf("A soft glow surrounds you, lighting the area. (%d minutes)", mins),
		fmt.Sprintf("A soft glow surrounds %s.", target.DisplayName()))
}

// castNightVisionSpell handles Night Vision (520): the caster, or a named player
// in the room, can see in dark rooms until it fades.
func (e *GameEngine) castNightVisionSpell(player *Player, spell *SpellDef, args []string) *CommandResult {
	target, isSelf, refusal := e.darkSpellTarget(player, args)
	if refusal != nil {
		return refusal
	}
	var mins int
	var fresh bool
	target.NightVisionExpiry, mins, fresh = extendTimedSpell(target.NightVisionExpiry, nightVisionDuration)
	if !fresh {
		return e.darkSpellResult(player, target, isSelf, spell,
			fmt.Sprintf("Your night vision is renewed. (%d minutes remaining)", mins), "")
	}
	return e.darkSpellResult(player, target, isSelf, spell,
		fmt.Sprintf("The darkness retreats from your eyes. (%d minutes)", mins),
		fmt.Sprintf("%s's eyes take on a faint silvery sheen.", target.DisplayNameCap()))
}

// darkSpellTarget resolves the optional player target of Light / Night Vision.
func (e *GameEngine) darkSpellTarget(player *Player, args []string) (target *Player, isSelf bool, refusal *CommandResult) {
	if len(args) > 0 {
		t := strings.ToLower(strings.Join(args, " "))
		if t != "me" && t != "myself" && t != "self" {
			found := e.findPlayerInRoom(player, t)
			if found == nil {
				return nil, false, &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", strings.Join(args, " "))}, TargetNotFound: true}
			}
			return found, false, nil
		}
	}
	return player, true, nil
}

// darkSpellResult builds the cast echoes. effectMsg is phrased to the recipient
// ("you"); roomEffect (optional) is what onlookers see happen to them.
func (e *GameEngine) darkSpellResult(caster, target *Player, isSelf bool, spell *SpellDef, effectMsg, roomEffect string) *CommandResult {
	if isSelf {
		res := &CommandResult{
			Messages:      []string{fmt.Sprintf("You gesture and cast %s.", spell.Name), effectMsg},
			RoomBroadcast: []string{fmt.Sprintf("%s gestures and casts %s.", caster.DisplayNameCap(), spell.Name)},
		}
		if roomEffect != "" {
			res.RoomBroadcast = append(res.RoomBroadcast, roomEffect)
		}
		return res
	}
	e.SavePlayer(context.Background(), target)
	res := &CommandResult{
		Messages:      []string{fmt.Sprintf("You gesture and cast %s on %s.", spell.Name, target.DisplayName())},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures and casts %s on %s.", caster.DisplayNameCap(), spell.Name, target.DisplayName())},
		TargetName:    target.FirstName,
		TargetMsg:     []string{fmt.Sprintf("%s casts %s on you.", caster.DisplayNameCap(), spell.Name), effectMsg},
	}
	if roomEffect != "" {
		res.RoomBroadcast = append(res.RoomBroadcast, roomEffect)
	}
	return res
}
