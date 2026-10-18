package engine

import "fmt"

// Mechanoid race traits — LEGENDS.DOC "Mechanoid:", EMOTE and UNEMOTE:
//
//   - UNEMOTE "causes mechanoids to switch over to the analytical half of their
//     brains. While in this state they can perform skills with greater precision.
//     However they cannot experience emotion in this mode." EMOTE switches back:
//     "They gain the ability to experience emotion but lose their logistical edge."
//   - Unemoted, every non-combat skill roll gets mechanoidUnemoteSkillBonus added to
//     its success chance (before the roll's own cap) — see unemoteSkillBonus.
//     Combat, spells and psionics are unaffected. No original number survives; the
//     constant is a tunable guess.
//   - Unemoted mechanoids can't use emotional emotes (emotionalEmoteVerbs); neutral
//     gestures (NOD, BOW, WAVE...) still work.
//   - GM Pages updates.html #22: "Mechanoids INTONE when unemoted." Captures show
//     it for statements, questions and exclamations alike: Warsprite intones,
//     "do we have any other kind ?"
//
// Every mechanoid starts in the emotional (normal) mode.
const mechanoidUnemoteSkillBonus = 10

// IsUnemoted reports whether the player is a mechanoid running on the analytical
// half of their brain. Checks race too, so a GM race change away from Mechanoid
// can never leave a stray Unemoted flag in effect.
func (p *Player) IsUnemoted() bool { return p.Race == RaceMechanoid && p.Unemoted }

// unemoteSkillBonus is the success-chance bonus an unemoted mechanoid adds to a
// non-combat skill roll, or 0.
func unemoteSkillBonus(p *Player) int {
	if p.IsUnemoted() {
		return mechanoidUnemoteSkillBonus
	}
	return 0
}

// emotionalEmoteVerbs are the emotes that express a feeling, refused while a
// mechanoid is unemoted.
var emotionalEmoteVerbs = map[string]bool{
	"SMILE": true, "LAUGH": true, "CHUCKLE": true, "GRIN": true, "FROWN": true,
	"SIGH": true, "WINK": true, "CRY": true, "SOB": true, "SNIFFLE": true,
	"DANCE": true, "JIG": true, "HULA": true, "HUG": true, "KISS": true,
	"CUDDLE": true, "SNUGGLE": true, "NUZZLE": true, "CARESS": true, "COMFORT": true,
	"TICKLE": true, "SLAP": true, "FIDGET": true, "SHIVER": true, "SHUDDER": true,
	"GROAN": true, "MOAN": true, "BEAM": true, "SWOON": true, "GLARE": true,
	"SCOWL": true, "BLUSH": true, "CRINGE": true, "GASP": true, "GIGGLE": true,
	"GRIMACE": true, "POUT": true, "SMIRK": true, "SNICKER": true, "WINCE": true,
	"SCREAM": true, "FUME": true, "SNARL": true, "GROWL": true, "CURSE": true,
	"PUCKER": true,
}

// unemotedEmoteBlock returns the refusal for an emotional emote by an unemoted
// mechanoid, or "" if the emote is allowed.
func unemotedEmoteBlock(p *Player, verb string) string {
	if !p.IsUnemoted() || !emotionalEmoteVerbs[verb] {
		return ""
	}
	return "Your emotions are disengaged. You feel nothing."
}

// doEmoteMode handles EMOTE (emotional=true) and UNEMOTE (emotional=false).
func (e *GameEngine) doEmoteMode(player *Player, emotional bool) *CommandResult {
	if player.Race != RaceMechanoid {
		return &CommandResult{Messages: []string{"Only mechanoids can switch their emotions on and off."}}
	}
	if emotional {
		if !player.Unemoted {
			return &CommandResult{Messages: []string{"Your emotions are already engaged."}}
		}
		player.Unemoted = false
		return &CommandResult{
			Messages:      []string{"You engage the emotional half of your brain. Feelings flood back in as your analytical edge fades."},
			RoomBroadcast: []string{fmt.Sprintf("%s's eyes soften as %s emotions return.", player.DisplayNameCap(), player.Possessive())},
		}
	}
	if player.Unemoted {
		return &CommandResult{Messages: []string{"Your emotions are already disengaged."}}
	}
	player.Unemoted = true
	return &CommandResult{
		Messages:      []string{"You switch over to the analytical half of your brain. Your emotions fall silent and your thoughts become precise."},
		RoomBroadcast: []string{fmt.Sprintf("%s's expression goes cold and blank.", player.DisplayNameCap())},
	}
}
