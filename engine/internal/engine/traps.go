package engine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Container traps live in VAL4 (GMSCRIPT.DOC "Item Value Fields"): 1-13 are
// mechanical traps, x001 codes are glyphs/runes. randomChestDrop arms treasure
// chests; checkTrap springs a trap and consumes it. A trap springs when its
// container is opened or its lock is picked, wherever the container is — on the
// floor, carried, or worn. DISARM finds it first (TrapFound) and then removes it.

const (
	trapExplosive       = 8 // "Explosive, destroys objects and contents"
	disarmRoundTime     = 5
	disarmFumbleRoll    = 90 // a failed disarm rolling above this springs the trap
	disarmBaseChance    = 20
	disarmChancePerRank = 10
)

// trapCodeOf returns the trap armed on an item, or 0. Only things with a lid or
// door can hold a trap; LIQCONTAINERs use VAL4 for their potion's appearance
// adjective, and a BREW in progress packs its ingredient codes into VAL4 with the
// step count in VAL5 (see doBrew), so neither is ever read as a trap.
func trapCodeOf(def *gameworld.ItemDef, val4, val5 int) int {
	if def == nil || val4 <= 0 || def.Type == "LIQCONTAINER" {
		return 0
	}
	if !containsFlag(def.Flags, "OPENABLE") && !isPortal(def.Type) {
		return 0
	}
	if val5 > 0 && val5 <= 3 && def.Container == "IN" {
		return 0
	}
	return val4
}

// trapKind names a trap for DISARM's "You find..." message.
func trapKind(code int) string {
	switch {
	case code >= 1000:
		switch (code / 1000) % 10 {
		case 7, 8:
			return "rune"
		case 9:
			return "glyph of death"
		default:
			return "glyph"
		}
	case code == 1 || code == 5 || code == 7 || code == 13:
		return "needle trap"
	case code == 2 || code == 12:
		return "gas trap"
	case code == 3 || code == 9:
		return "acid trap"
	case code == 4:
		return "blade trap"
	case code == trapExplosive:
		return "explosive trap"
	}
	return "trap"
}

// trapComplexity rates how hard a trap is to disarm (1 = simplest). LEGENDS.DOC:
// "Success when disarming is a percentage based on the complexity of the trap
// versus a character's trap lore skill level."
func trapComplexity(code int) int {
	switch {
	case code >= 1000:
		return 4
	case code == 1 || code == 2:
		return 1
	case code == 3 || code == 4 || code == 5:
		return 2
	case code == 13:
		return 4
	default: // 7, 8, 9, 12
		return 3
	}
}

// disarmChance calibrates to the rori0627 capture: rank 1, Perception 62 → 36%
// on simple chest traps. Each complexity step past the first costs 10%.
func disarmChance(player *Player, code int) int {
	chance := disarmBaseChance + player.Skills[12]*disarmChancePerRank + player.Perception/10 -
		(trapComplexity(code)-1)*10 + unemoteSkillBonus(player)
	return max(5, min(95, chance))
}

// springTrap fires the trap held in val4 and clears it. The messages are checkTrap's.
func (e *GameEngine) springTrap(player *Player, val4 *int, found *bool) []string {
	tmp := gameworld.RoomItem{Val4: *val4}
	msgs := e.checkTrap(player, &tmp)
	*val4 = 0
	*found = false
	return msgs
}

// takeRoomItemWithContents removes room.Items[idx] and everything inside it
// (both the roomContainerContents map and script-placed IsPut items) from the
// room, notifying other machines, and returns them.
func (e *GameEngine) takeRoomItemWithContents(room *gameworld.Room, idx int) (InventoryItem, []InventoryItem) {
	ri := room.Items[idx]
	item := inventoryFromRoomItem(&ri)
	contents := e.roomContainerFullContents(room.Number, ri.Ref)
	e.roomContainerDelete(room.Number, ri.Ref)
	kept := room.Items[:0]
	for _, r2 := range room.Items {
		switch {
		case r2.Ref == ri.Ref:
		case r2.IsPut && r2.PutIn == ri.Ref: // script-placed contents, already in contents
			e.notifyRoomChange(RoomChange{RoomNumber: room.Number, Type: "item_remove", ItemRef: r2.Ref})
		default:
			kept = append(kept, r2)
		}
	}
	room.Items = kept
	e.notifyRoomChange(RoomChange{RoomNumber: room.Number, Type: "item_remove", ItemRef: ri.Ref})
	return item, contents
}

// explodedMessage is appended when an explosive trap destroys its container.
func explodedMessage(name string) string {
	return fmt.Sprintf("%s and everything in it are destroyed!", capitalize(name))
}

// springRoomTrap springs the trap on room.Items[idx]. An explosive trap destroys
// the item and its contents (destroyed reports this, and idx is then invalid).
func (e *GameEngine) springRoomTrap(player *Player, room *gameworld.Room, idx int, name string) (msgs []string, destroyed bool) {
	code := room.Items[idx].Val4
	msgs = e.springTrap(player, &room.Items[idx].Val4, &room.Items[idx].TrapFound)
	if code == trapExplosive {
		e.takeRoomItemWithContents(room, idx)
		return append(msgs, explodedMessage(name)), true
	}
	itemCopy := room.Items[idx]
	e.notifyRoomChange(RoomChange{RoomNumber: room.Number, Type: "item_update", ItemRef: itemCopy.Ref, Item: &itemCopy})
	return msgs, false
}

// springCarriedTrap springs the trap on (*items)[idx] — the player's Inventory
// or Worn slice. An explosive trap destroys the item and its contents.
func (e *GameEngine) springCarriedTrap(player *Player, items *[]InventoryItem, idx int, name string) (msgs []string, destroyed bool) {
	code := (*items)[idx].Val4
	msgs = e.springTrap(player, &(*items)[idx].Val4, &(*items)[idx].TrapFound)
	if code == trapExplosive {
		*items = append((*items)[:idx], (*items)[idx+1:]...)
		return append(msgs, explodedMessage(name)), true
	}
	return msgs, false
}

// doDisarm implements DISARM <object>. LEGENDS.DOC: "The first use of this command
// on an object attempts to locate a trap on it. If there is a trap, it can then be
// disarmed by using this command again." Finding needs Trap & Poison Lore (skill
// 12); an untrapped object, or a trap you can't recognize, gets the capture's "You
// don't see a trap." with no round time.
func (e *GameEngine) doDisarm(ctx context.Context, player *Player, args []string) *CommandResult {
	if len(args) == 0 {
		return &CommandResult{Messages: []string{"Disarm what?"}}
	}
	if player.RoundTimeExpiry.After(time.Now()) {
		remaining := int(time.Until(player.RoundTimeExpiry).Seconds()) + 1
		return &CommandResult{Messages: []string{fmt.Sprintf("[Wait %d seconds...]", remaining)}}
	}
	target := strings.ToLower(strings.Join(args, " "))
	target, mine := stripMyPrefix(target)
	target, skip := parseOrdinal(target)
	room := e.rooms[player.RoomNumber]
	if room == nil {
		return &CommandResult{Messages: []string{"You can't do that here."}}
	}

	// Carried, then worn, then (unless "my") on the floor — same order as doOpen.
	var items *[]InventoryItem
	idx, roomIdx := -1, -1
	for _, list := range []*[]InventoryItem{&player.Inventory, &player.Worn} {
		for i, ii := range *list {
			if d := e.items[ii.Archetype]; d != nil && matchesTargetOrdinal(e.getItemNounName(d), target, &skip, e.getAdjName(ii.Adj1), e.getAdjName(ii.Adj2), e.getAdjName(ii.Adj3)) {
				items, idx = list, i
				break
			}
		}
		if items != nil {
			break
		}
	}
	if items == nil && !mine {
		for i, ri := range room.Items {
			if d := e.items[ri.Archetype]; d != nil && matchesTargetOrdinal(e.getItemNounName(d), target, &skip, e.getAdjName(ri.Adj1), e.getAdjName(ri.Adj2), e.getAdjName(ri.Adj3)) {
				roomIdx = i
				break
			}
		}
	}

	var def *gameworld.ItemDef
	var name string
	var val4, val5 int
	var found *bool
	switch {
	case items != nil:
		ii := &(*items)[idx]
		def = e.items[ii.Archetype]
		name = e.formatItemName(def, ii.Adj1, ii.Adj2, ii.Adj3, ii.Tail)
		val4, val5, found = ii.Val4, ii.Val5, &ii.TrapFound
	case roomIdx >= 0:
		ri := &room.Items[roomIdx]
		def = e.items[ri.Archetype]
		name = e.formatItemName(def, ri.Adj1, ri.Adj2, ri.Adj3, ri.Extend)
		val4, val5, found = ri.Val4, ri.Val5, &ri.TrapFound
	default:
		return &CommandResult{Messages: []string{"You don't see that here."}}
	}

	code := trapCodeOf(def, val4, val5)
	if code == 0 || player.Skills[12] < 1 {
		return &CommandResult{Messages: []string{"You don't see a trap."}}
	}
	notifyRoom := func() {
		if roomIdx >= 0 && roomIdx < len(room.Items) {
			itemCopy := room.Items[roomIdx]
			e.notifyRoomChange(RoomChange{RoomNumber: room.Number, Type: "item_update", ItemRef: itemCopy.Ref, Item: &itemCopy})
		}
	}

	if !*found {
		*found = true
		notifyRoom()
		e.SavePlayer(ctx, player)
		return &CommandResult{Messages: []string{fmt.Sprintf("You find a %s on %s!", trapKind(code), name)}}
	}

	chance := disarmChance(player, code)
	roll := rand.Intn(100) + 1
	rt := applyRoundTime(player, disarmRoundTime)
	player.RoundTime = rt
	player.RoundTimeExpiry = time.Now().Add(time.Duration(rt) * time.Second)
	rtMsg := fmt.Sprintf(" [Round: %d sec]", rt)

	if roll <= chance {
		if items != nil {
			(*items)[idx].Val4, (*items)[idx].TrapFound = 0, false
		} else {
			room.Items[roomIdx].Val4, room.Items[roomIdx].TrapFound = 0, false
			notifyRoom()
		}
		e.SavePlayer(ctx, player)
		return &CommandResult{
			Messages:      []string{fmt.Sprintf("[Success: %d%%, Roll: %d] Success!", chance, roll), fmt.Sprintf("You disarm %s.", name), rtMsg},
			RoomBroadcast: []string{fmt.Sprintf("%s disarms %s.", player.DisplayNameCap(), name)},
			PlayerState:   player,
		}
	}

	msgs := []string{fmt.Sprintf("[Success: %d%%, Roll: %d] Failure.", chance, roll), fmt.Sprintf("You are unable to disarm %s.", name)}
	broadcast := []string{fmt.Sprintf("%s fiddles with %s.", player.DisplayNameCap(), name)}
	if roll > disarmFumbleRoll {
		var trapMsgs []string
		if items != nil {
			trapMsgs, _ = e.springCarriedTrap(player, items, idx, name)
		} else {
			trapMsgs, _ = e.springRoomTrap(player, room, roomIdx, name)
		}
		msgs = append(msgs, trapMsgs...)
		broadcast = append(broadcast, fmt.Sprintf("%s sets off a trap!", player.DisplayNameCap()))
	}
	msgs = append(msgs, rtMsg)
	e.SavePlayer(ctx, player)
	return &CommandResult{Messages: msgs, RoomBroadcast: broadcast, PlayerState: player}
}
