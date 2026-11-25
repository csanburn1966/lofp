package engine

import (
	"fmt"
	"strings"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Capacitance (Mind over Matter #6). LEGENDS.DOC: "Drains 1 charge from a magical
// charged item and rearranges the energy into psi -- a net gain of 5. The item in
// question must have had an actual magical spell imbued in it." Here it drains every
// charge at once, 5 psi each and more for a stronger spell (capacitancePerCharge),
// and the psi may run past the psionicist's maximum. As in the original the gain is
// net: the discipline's own cost comes back with it.

const capacitanceID = 5 // internal ID; players know it as #6

// capacitancePerCharge is the psi a charge of a spell of the given level yields: 5,
// plus a tenth of the level, rounded up, for a spell of level 5 or more.
func capacitancePerCharge(level int) int {
	if level < 5 {
		return 5
	}
	return 5 + (level+9)/10
}

// projectCapacitance handles PROJECT <item> for Capacitance. The psi has been spent;
// cost is how much, to come back with a success.
func (e *GameEngine) projectCapacitance(player *Player, disc *PsiDiscipline, cost int, args []string) *CommandResult {
	refund := func(msg string) *CommandResult {
		player.Psi += cost
		return &CommandResult{Messages: []string{msg}, TargetNotFound: true}
	}
	if len(args) == 0 {
		return refund("Project Capacitance at what? Name a charged magic item you are carrying.")
	}
	item, def := e.findMagicItemTarget(player, strings.Join(args, " "), false)
	if item == nil {
		return refund("You aren't carrying that.")
	}
	name := e.formatItemName(def, item.Adj1, item.Adj2, item.Adj3, item.Tail)
	spell := FindSpellByID(item.Val3)
	// The spell has to be bound into a magic item that casts it: a weapon's Val3 is
	// its critical type and a potion's sips aren't charges.
	if spell == nil || !itemCastsItsSpell(def) {
		return refund(fmt.Sprintf("You probe %s, but there is no spell imbued in it to draw upon.", name))
	}
	if item.Val2 <= 0 {
		return refund(fmt.Sprintf("You probe %s, but its magic is spent.", name))
	}
	charges := item.Val2
	gain := charges * capacitancePerCharge(spell.Level)
	item.Val2 = 0
	player.Psi += gain + cost
	plural := "s"
	if charges == 1 {
		plural = ""
	}
	return &CommandResult{
		Messages: []string{
			fmt.Sprintf("You draw the magic out of %s, and its %d charge%s of %s rearrange themselves into psionic energy!", name, charges, plural, spell.Name),
			fmt.Sprintf("[+%d Psi, %d/%d]", gain, player.Psi, player.MaxPsi),
			fmt.Sprintf("%s's magic is spent.", capitalize(name)),
		},
		RoomBroadcast: []string{fmt.Sprintf("%s concentrates intently on %s, which flickers and goes dull.", player.DisplayNameCap(), name)},
	}
}

// itemCastsItsSpell reports whether an item holds its spell (Val3) as charges it casts
// (Val2): one with a ROUTINE script, as the generic magic items and enchanted ones have.
func itemCastsItsSpell(def *gameworld.ItemDef) bool {
	if def == nil {
		return false
	}
	for i := range def.Scripts {
		if scriptBlockHasRoutine(&def.Scripts[i]) {
			return true
		}
	}
	return false
}
