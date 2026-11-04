package engine

import (
	"context"
	"fmt"
	"strings"
)

// NOSELL lets a player protect items they mean to keep from being sold by
// accident. "NOSELL <item>" toggles InventoryItem.NoSell on a held, worn or carried
// item; a bare "NOSELL" lists everything marked, including items packed inside
// containers. Marked items are refused by SELL and left behind by SELL ALL
// (doSell, doSellAll), and INVENTORY tags them (noSellTag).

// noSellTag is the INVENTORY suffix for a marked item.
func noSellTag(ii InventoryItem) string {
	if ii.NoSell {
		return " (nosell)"
	}
	return ""
}

// noSellRefusal is SELL's answer for a marked item.
func noSellRefusal(displayName string) *CommandResult {
	return &CommandResult{Messages: []string{
		fmt.Sprintf("You've marked %s NOSELL, so you keep it. (NOSELL it again to allow selling it.)", displayName),
	}}
}

func (e *GameEngine) doNoSell(ctx context.Context, player *Player, args []string) *CommandResult {
	if len(args) == 0 {
		return e.listNoSell(player)
	}
	target := strings.ToLower(strings.Join(args, " "))
	target, _ = stripMyPrefix(target)
	target, skip := parseOrdinal(target)

	// Hands first, then worn, then carried: the same order as runVerbScriptsForTarget.
	candidates := []*InventoryItem{player.Wielded, player.OffHand}
	for i := range player.Worn {
		candidates = append(candidates, &player.Worn[i])
	}
	for i := range player.Inventory {
		candidates = append(candidates, &player.Inventory[i])
	}
	for _, ii := range candidates {
		if ii == nil {
			continue
		}
		def := e.items[ii.Archetype]
		if def == nil || !matchesTarget(e.getItemNounName(def), target, e.getAdjName(ii.Adj1), e.getAdjName(ii.Adj2), e.getAdjName(ii.Adj3)) {
			continue
		}
		if skip > 0 {
			skip--
			continue
		}
		ii.NoSell = !ii.NoSell
		e.SavePlayer(ctx, player)
		name := e.formatItemName(def, ii.Adj1, ii.Adj2, ii.Adj3, ii.Tail)
		if ii.NoSell {
			return &CommandResult{Messages: []string{fmt.Sprintf("You mark %s NOSELL. Merchants won't buy it from you until you NOSELL it again.", name)}}
		}
		return &CommandResult{Messages: []string{fmt.Sprintf("You clear the NOSELL mark from %s. It can be sold again.", name)}}
	}
	return &CommandResult{Messages: []string{"You don't have that."}}
}

// listNoSell lists every marked item the player has on them, with where it is.
func (e *GameEngine) listNoSell(player *Player) *CommandResult {
	var lines []string
	var walk func(ii *InventoryItem, where string)
	walk = func(ii *InventoryItem, where string) {
		if ii == nil {
			return
		}
		def := e.items[ii.Archetype]
		if def == nil {
			return
		}
		name := e.formatItemName(def, ii.Adj1, ii.Adj2, ii.Adj3, ii.Tail)
		if ii.NoSell {
			lines = append(lines, fmt.Sprintf("  %s%s", name, where))
		}
		for i := range ii.Contents {
			walk(&ii.Contents[i], fmt.Sprintf(" (in %s)", name))
		}
	}
	walk(player.Wielded, " (wielded)")
	walk(player.OffHand, " (off-hand)")
	for i := range player.Worn {
		walk(&player.Worn[i], " (worn)")
	}
	for i := range player.Inventory {
		walk(&player.Inventory[i], "")
	}
	if len(lines) == 0 {
		return &CommandResult{Messages: []string{"You haven't marked anything NOSELL. Type NOSELL <item> to protect an item from being sold."}}
	}
	return &CommandResult{Messages: append([]string{"Marked NOSELL:"}, lines...)}
}
