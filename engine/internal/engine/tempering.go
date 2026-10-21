package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Tempering (spell 223) — new, not part of the original 1990s data. Level 28
// Enchantment that adds temperingHardnessBonus to a weapon's or shield's Weapon
// Clash break-resistance (see temperingHardness in combat.go). Two versions:
//
//   - PREPARE 223: temporary, lasts temperingDuration. Recasting resets the timer
//     back to a full hour; it never adds time.
//   - PREPARE 223 WITH <diamond>: permanent. The diamond must be large or huge AND
//     polished (only a Highlander's MOLD adds "polished" — see doMold). It's only
//     checked at PREPARE and consumed on a successful cast, never a failed one.
//
// The two never stack: a permanently tempered item refuses both versions, and the
// permanent cast clears any running temporary timer.
const (
	temperingSpellID     = 223
	temperingDiamondArch = 107 // diamond (also Summon Gargoyle's reagent, any diamond)
	temperingDuration    = time.Hour
)

// isTemperingDiamond reports whether ii is a diamond fit for permanent Tempering:
// large or huge, and polished.
func (e *GameEngine) isTemperingDiamond(ii *InventoryItem) bool {
	if ii.Archetype != temperingDiamondArch {
		return false
	}
	polishedID := e.adjByName("polished")
	if polishedID == 0 {
		polishedID = 241 // ADJNOUN.SCR fallback, same as doMold
	}
	largeID, hugeID := e.adjByName("large"), e.adjByName("huge")
	polished, big := false, false
	for _, a := range []int{ii.Adj1, ii.Adj2, ii.Adj3} {
		if a == 0 {
			continue
		}
		if a == polishedID {
			polished = true
		}
		if a == largeID || a == hugeID {
			big = true
		}
	}
	return polished && big
}

// findTemperingDiamond returns the inventory index of the first qualifying diamond
// matching target (any qualifying diamond when target is ""), or -1. matched reports
// whether target named any diamond at all, so PREPARE can tell "that diamond isn't
// good enough" apart from "that isn't a diamond" (e.g. a moonstone).
func (e *GameEngine) findTemperingDiamond(player *Player, target string) (idx int, matched bool) {
	for i := range player.Inventory {
		ii := &player.Inventory[i]
		def := e.items[ii.Archetype]
		if def == nil || ii.Archetype != temperingDiamondArch {
			continue
		}
		if target != "" && !matchesTarget(e.getItemNounName(def), target, e.getAdjName(ii.Adj1), e.getAdjName(ii.Adj2), e.getAdjName(ii.Adj3)) {
			continue
		}
		matched = true
		if e.isTemperingDiamond(ii) {
			return i, true
		}
	}
	return -1, matched
}

// castTemperingSpell handles Tempering (223) on a carried, wielded, held or worn
// weapon or shield. withDiamond is whether it was prepared with a qualifying
// diamond (PreparedTemperingDiamond, captured by doCastSpell).
func (e *GameEngine) castTemperingSpell(player *Player, args []string, withDiamond bool) *CommandResult {
	if len(args) == 0 {
		return &CommandResult{Messages: []string{"Temper what? Specify a weapon or shield in your possession."}, TargetNotFound: true}
	}
	target, skip := parseOrdinal(strings.ToLower(strings.Join(args, " ")))

	type candidate struct {
		item *InventoryItem
		def  *gameworld.ItemDef
	}
	temperable := func(def *gameworld.ItemDef) bool {
		return def != nil && (isWeaponItemType(def.Type) || isWeapon(def.Type) || def.Type == "SHIELD")
	}
	var candidates []candidate
	if player.Wielded != nil {
		if def := e.items[player.Wielded.Archetype]; temperable(def) {
			candidates = append(candidates, candidate{player.Wielded, def})
		}
	}
	if player.OffHand != nil {
		if def := e.items[player.OffHand.Archetype]; temperable(def) {
			candidates = append(candidates, candidate{player.OffHand, def})
		}
	}
	for i := range player.Worn {
		if def := e.items[player.Worn[i].Archetype]; temperable(def) {
			candidates = append(candidates, candidate{&player.Worn[i], def})
		}
	}
	for i := range player.Inventory {
		if def := e.items[player.Inventory[i].Archetype]; temperable(def) {
			candidates = append(candidates, candidate{&player.Inventory[i], def})
		}
	}

	for _, c := range candidates {
		if !matchesTargetOrdinal(e.getItemNounName(c.def), target, &skip, e.getAdjName(c.item.Adj1), e.getAdjName(c.item.Adj2), e.getAdjName(c.item.Adj3)) {
			continue
		}
		itemName := "your " + e.formatItemNameNoArticle(c.def, c.item.Adj1, c.item.Adj2, c.item.Adj3, c.item.Tail)
		if c.item.Tempered {
			return &CommandResult{Messages: []string{fmt.Sprintf("%s has already been permanently tempered. The spell would have no effect.", capitalize(itemName))}, TargetNotFound: true}
		}

		if !withDiamond {
			renewed := c.item.TemperedUntil > time.Now().Unix()
			c.item.TemperedUntil = time.Now().Add(temperingDuration).Unix()
			msg := fmt.Sprintf("A faint silvery sheen settles over %s as the spell draws it tight and strong.", itemName)
			if renewed {
				msg = fmt.Sprintf("The silvery sheen over %s brightens as its tempering is renewed for another hour.", itemName)
			}
			return &CommandResult{
				Messages:      []string{msg},
				RoomBroadcast: []string{fmt.Sprintf("A faint silvery sheen settles over an item %s is holding.", player.DisplayName())},
			}
		}

		// Permanent: the diamond was only verified at PREPARE, so it may have been
		// dropped or sold since. Consumed only now, on a successful cast.
		di, _ := e.findTemperingDiamond(player, "")
		if di < 0 {
			return &CommandResult{Messages: []string{"You no longer have a suitable diamond. The spell fizzles."}}
		}
		diamond := player.Inventory[di]
		diamondName := e.formatItemName(e.items[diamond.Archetype], diamond.Adj1, diamond.Adj2, diamond.Adj3, diamond.Tail)
		// Temper before removing the diamond: c.item may point into player.Inventory,
		// and the removal shifts everything after the diamond down a slot.
		c.item.Tempered = true
		c.item.TemperedUntil = 0
		player.Inventory = append(player.Inventory[:di], player.Inventory[di+1:]...)
		return &CommandResult{
			Messages: []string{
				fmt.Sprintf("%s turns to glittering dust that sinks into %s, tempering it permanently!", capitalize(diamondName), itemName),
			},
			RoomBroadcast: []string{fmt.Sprintf("%s turns to glittering dust that sinks into an item %s is holding.", capitalize(diamondName), player.DisplayName())},
		}
	}
	return &CommandResult{Messages: []string{"You don't have a weapon or shield matching that."}, TargetNotFound: true}
}
