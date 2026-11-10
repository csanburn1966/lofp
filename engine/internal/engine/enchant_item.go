package engine

import (
	"fmt"
	"strings"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Enchant an Item (244) creates a new magic item in three steps. No original
// description of the spell survives (MAGIC.TXT, spells.html and LEGENDS.DOC give
// only its name and level), so the ritual is new design built on the item
// conventions the generic magic items already use (Val3 = spell ID, Val2 =
// charges — see doRoutine in scripts.go):
//
//  1. Rite of Preparation (412) cast at a blank spell-holding item (Val3 == 0)
//     primes it: Val3 = itemPrimedVal3.
//  2. Any spell cast at the primed item is bound into it instead of taking
//     effect: Val3 = that spell's ID, Val2 = 0. The spell can come from memory,
//     a chanted scroll, or another magic item — anything that sets PreparedSpell.
//  3. Enchant an Item, prepared with a diamond, cast at the item outdoors during
//     a storm, drains the caster's remaining mana into charges the same way
//     Charge Wand (243) does, and leaves the caster with no mana or fatigue.
const (
	enchantItemSpellID     = 244
	ritePreparationSpellID = 412
	// itemPrimedVal3 marks an item that Rite of Preparation has readied to receive
	// a spell. Negative so every existing "Val3 > 0 means it holds a spell" check
	// keeps treating it as empty.
	itemPrimedVal3 = -1
)

// Generic magic-item archetypes (ITEM1.SCR/ITEM2.SCR INUMBER 149-189: amulets,
// idols, medallions, necklaces, orbs, pendants, rings, rods, wands, talismans).
// Unique items elsewhere also use ROUTINE, but their spells are part of their
// own scripted design, so only this catalog can be enchanted.
const (
	minSpellHolderArch = 149
	maxSpellHolderArch = 189
)

// isSpellHolderItem reports whether an item archetype is one of the generic magic
// items built to hold a spell: in the catalog range and carrying a ROUTINE 1/2
// script that casts or prepares its Val3 spell. The range alone isn't enough —
// the scroll (168) sits inside it.
func isSpellHolderItem(def *gameworld.ItemDef) bool {
	if def == nil || def.Number < minSpellHolderArch || def.Number > maxSpellHolderArch {
		return false
	}
	for i := range def.Scripts {
		if scriptBlockHasRoutine(&def.Scripts[i]) {
			return true
		}
	}
	return false
}

func scriptBlockHasRoutine(b *gameworld.ScriptBlock) bool {
	for _, a := range b.Actions {
		if strings.EqualFold(a.Command, "ROUTINE") {
			return true
		}
	}
	for _, steps := range [][]gameworld.ScriptStep{b.Body, b.ElseBody} {
		for _, s := range steps {
			if s.Action != nil && strings.EqualFold(s.Action.Command, "ROUTINE") {
				return true
			}
			if s.Block != nil && scriptBlockHasRoutine(s.Block) {
				return true
			}
		}
	}
	return false
}

// spellCanBeImprinted reports whether a spell can be bound into a primed item.
// The three spells of the enchanting ritual itself can't.
func spellCanBeImprinted(spellID int) bool {
	switch spellID {
	case ritePreparationSpellID, 243, enchantItemSpellID:
		return false
	}
	return true
}

// castRiteOfPreparationOnItem handles step 1: CAST <item> with Rite of
// Preparation prepared. Returns nil when the args don't name a carried
// spell-holding item, so the caller keeps its generic behavior.
func (e *GameEngine) castRiteOfPreparationOnItem(player *Player, args []string) *CommandResult {
	if len(args) == 0 {
		return nil
	}
	item, def := e.findMagicItemTarget(player, strings.Join(args, " "), false)
	if item == nil || !isSpellHolderItem(def) {
		return nil
	}
	itemName := e.formatItemName(def, item.Adj1, item.Adj2, item.Adj3, item.Tail)
	switch {
	case item.Val3 == itemPrimedVal3:
		return &CommandResult{Messages: []string{fmt.Sprintf("%s has already been prepared to receive a spell.", capitalize(itemName))}, TargetNotFound: true}
	case item.Val3 != 0:
		return &CommandResult{Messages: []string{fmt.Sprintf("%s already holds the magic of a spell.", capitalize(itemName))}, TargetNotFound: true}
	}
	item.Val3 = itemPrimedVal3
	item.Val2 = 0
	return &CommandResult{
		Messages: []string{
			fmt.Sprintf("You gesture at %s.", itemName),
			fmt.Sprintf("A pale shimmer washes over %s and fades, leaving it hollow and waiting for a spell.", itemName),
		},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s, and a pale shimmer washes over it.", player.DisplayNameCap(), itemName)},
	}
}

// tryImprintSpellIntoItem handles step 2: a spell cast at a carried primed item
// is bound into it instead of taking its normal effect. Returns nil when the
// args don't name a carried primed item.
func (e *GameEngine) tryImprintSpellIntoItem(player *Player, spell *SpellDef, args []string) *CommandResult {
	if len(args) == 0 || !spellCanBeImprinted(spell.ID) {
		return nil
	}
	item, def := e.findMagicItemTarget(player, strings.Join(args, " "), false)
	if item == nil || item.Val3 != itemPrimedVal3 || !isSpellHolderItem(def) {
		return nil
	}
	itemName := e.formatItemName(def, item.Adj1, item.Adj2, item.Adj3, item.Tail)

	msgs := []string{fmt.Sprintf("You gesture at %s.", itemName)}
	roomMsgs := []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), itemName)}

	// A reagent the spell would have consumed on its own at cast time (spider eye,
	// imp toe, spriggan finger) is consumed by the binding instead.
	if reqArch := player.PreparedSpellReagentArch; reqArch != 0 {
		player.PreparedSpellReagentArch = 0
		if !spellConsumesReagentAtPrepare(spell.ID) {
			for i, ii := range player.Inventory {
				if ii.Archetype == reqArch {
					player.Inventory = append(player.Inventory[:i], player.Inventory[i+1:]...)
					msgs = append(msgs, spellReagentConsumeMessage(spell.ID))
					roomMsgs = append(roomMsgs, spellReagentConsumeMessageRoom(spell.ID, player.DisplayName()))
					break
				}
			}
		}
	}

	item.Val3 = spell.ID
	item.Val2 = 0
	msgs = append(msgs, fmt.Sprintf("The magic of %s swirls into %s and settles deep within it, waiting to be bound.", spell.Name, itemName))
	roomMsgs = append(roomMsgs, fmt.Sprintf("Swirling magic sinks into %s and vanishes.", itemName))
	return &CommandResult{Messages: msgs, RoomBroadcast: roomMsgs}
}

// enchantItemSkyBlocked returns why the caster can't call down Enchant an Item's
// lightning where they stand, or "" if they can: they must be outdoors while the
// region's weather is a Thunderstorm, Gale or Hurricane (WeatherNames 6-8).
func (e *GameEngine) enchantItemSkyBlocked(player *Player) string {
	room := e.rooms[player.RoomNumber]
	if room == nil || !isOutdoorTerrain(room.Terrain) {
		return "You must be outdoors, beneath the open sky, to call down the lightning that Enchant an Item requires."
	}
	if wea := e.RegionWeather[room.Region]; wea < 6 || wea > 8 {
		return "Enchant an Item needs the lightning of a raging storm, and the sky here is not stormy enough."
	}
	return ""
}

// castEnchantItemSpell handles step 3: Enchant an Item (244) cast at a carried
// item holding a bound spell but no charges. Converts the caster's remaining
// mana (after this spell's own cost) into charges with Charge Wand's formula,
// then empties the caster's mana and fatigue.
func (e *GameEngine) castEnchantItemSpell(player *Player, spell *SpellDef, args []string) *CommandResult {
	if len(args) == 0 {
		return &CommandResult{Messages: []string{"Enchant what? Specify the item you are holding."}, TargetNotFound: true}
	}
	if msg := e.enchantItemSkyBlocked(player); msg != "" {
		return &CommandResult{Messages: []string{msg}, TargetNotFound: true}
	}
	item, def := e.findMagicItemTarget(player, strings.Join(args, " "), false)
	if item == nil {
		return &CommandResult{Messages: []string{"You aren't holding or wearing that."}, TargetNotFound: true}
	}
	itemName := e.formatItemName(def, item.Adj1, item.Adj2, item.Adj3, item.Tail)
	if !isSpellHolderItem(def) {
		return &CommandResult{Messages: []string{fmt.Sprintf("%s was not made to hold a spell.", capitalize(itemName))}, TargetNotFound: true}
	}
	switch {
	case item.Val3 == 0:
		return &CommandResult{Messages: []string{fmt.Sprintf("%s must first be readied with the Rite of Preparation.", capitalize(itemName))}, TargetNotFound: true}
	case item.Val3 == itemPrimedVal3:
		return &CommandResult{Messages: []string{fmt.Sprintf("No spell has been cast into %s yet.", itemName)}, TargetNotFound: true}
	case item.Val2 > 0:
		return &CommandResult{Messages: []string{fmt.Sprintf("%s is already enchanted.", capitalize(itemName))}, TargetNotFound: true}
	}
	storedSpell := FindSpellByID(item.Val3)
	if storedSpell == nil || storedSpell.ManaCost <= 0 {
		return &CommandResult{Messages: []string{fmt.Sprintf("%s's magic is beyond your comprehension.", capitalize(itemName))}}
	}

	// Consume the diamond verified at PREPARE time (mirrors castChargeWandSpell).
	if reqArch := player.PreparedSpellReagentArch; reqArch != 0 {
		player.PreparedSpellReagentArch = 0
		consumed := false
		for i, ii := range player.Inventory {
			if ii.Archetype == reqArch {
				player.Inventory = append(player.Inventory[:i], player.Inventory[i+1:]...)
				consumed = true
				break
			}
		}
		if !consumed {
			return &CommandResult{Messages: []string{fmt.Sprintf("You no longer have the required reagent (%s). The spell fizzles.", spellReagentName(spell.ID))}}
		}
	}

	charges := (player.Mana + storedSpell.ManaCost - 1) / storedSpell.ManaCost // round up, as Charge Wand does
	if charges < 1 {
		charges = 1
	}
	player.Mana = 0
	player.Fatigue = 0
	item.Val2 = charges

	return &CommandResult{
		Messages: []string{
			fmt.Sprintf("You raise %s before you and reach skyward with your free hand, calling forth a mighty bolt of lightning. Its blinding energy tears through your outstretched hand and floods your body as you channel every last ounce of your magical strength into %s!", itemName, itemName),
			spellReagentConsumeMessage(spell.ID),
			fmt.Sprintf("%s thrums with the bound power of %s.", capitalize(itemName), storedSpell.Name),
		},
		RoomBroadcast: []string{
			fmt.Sprintf("%s raises %s before %s and reaches skyward with %s free hand. A mighty bolt of lightning descends from the heavens, striking %s outstretched hand and flooding %s body with blinding energy as %s channels every last ounce of %s magical strength into %s!",
				player.DisplayNameCap(), itemName, player.Objective(), player.Possessive(), player.Possessive(), player.Possessive(), player.Pronoun(), player.Possessive(), itemName),
			spellReagentConsumeMessageRoom(spell.ID, player.DisplayName()),
		},
	}
}
