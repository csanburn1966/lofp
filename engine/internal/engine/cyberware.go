package engine

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Cyberware: implants sold by CyberTech (ADDAMS.SCR room 352; the older
// FAYDINDR.SCR copy of the room is overridden). Buying one grafts it onto the
// body instead of handing over an item. The item defs (205-212) are bare
// INTANGIBLE names with no scripts, so everything here was rebuilt from the
// surviving evidence:
//
//   - legends/TALIESIN.TXT and legends/talibak.cap: INFO lists a separate
//     "Cyberware:" section (e.g. "Omega Bionic Arm", "Omega Cyberplate Armor",
//     "Alpha Telepathor", "Omega Reflexor"). Between those two captures the
//     character added an Omega Arm, Leg and Reflexor and gained +24 Strength,
//     +29 Quickness and +35 Agility while Empathy fell from 53 to 0 and Mana
//     from 104 to 0. The omega column below reproduces that exactly.
//   - GM Pages/char.html: racial attribute 3, "Telepathy: 1=Telepathor,
//     2=Tel&Contact" (a telepathor gives telepathy, not CONTACT).
//   - MANUAL.DOC: Strength, Quickness, Perception and Empathy are "affected by
//     cybernetics"; legends/bugs.txt: spells going inactive after buying
//     cyberware (the Empathy drain).
//   - ADDAMS.SCR rooms 788/789: the surgery the buyer wakes up in.
//
// The alpha-gamma tiers, the eye/ear split, night vision and the armor numbers
// are design decisions (2026-10-18), not recovered data.

// CyberImplant is one installed implant. The stat fields hold what was actually
// applied at install (Empathy can't go below 0, so its drain may be smaller than
// the table's), so an upgrade or removal undoes exactly that.
type CyberImplant struct {
	Kind     string `bson:"kind" json:"kind"`
	Tier     int    `bson:"tier" json:"tier"` // 1 alpha, 2 beta, 3 gamma, 4 omega
	Str      int    `bson:"str,omitempty" json:"str,omitempty"`
	Agi      int    `bson:"agi,omitempty" json:"agi,omitempty"`
	Qui      int    `bson:"qui,omitempty" json:"qui,omitempty"`
	Per      int    `bson:"per,omitempty" json:"per,omitempty"`
	Emp      int    `bson:"emp,omitempty" json:"emp,omitempty"`           // negative: Empathy removed
	ManaLost int    `bson:"manaLost,omitempty" json:"manaLost,omitempty"` // MaxMana removed with that Empathy
}

const (
	cyberArm        = "arm"
	cyberLeg        = "leg"
	cyberEar        = "ear"
	cyberArmor      = "armor"
	cyberTelepathor = "telepathor"
	cyberReflexor   = "reflexor"
	cyberEye        = "eye"
)

var cyberTierNames = [5]string{"", "Alpha", "Beta", "Gamma", "Omega"}

// cyberTierAdjs maps the STOREITEM adjectives (ADJNOUN.SCR 369-372) to tiers.
var cyberTierAdjs = map[int]int{369: 1, 370: 2, 371: 3, 372: 4}

// cyberKindDef is one implant type's per-tier effects, indexed by tier (0 unused).
type cyberKindDef struct {
	kind      string
	archetype int
	name      string // as INFO shows it after the tier: "Omega Bionic Arm"
	str, agi, qui, per [5]int
	empCost   [5]int // Empathy drained
	armorDR   [5]int // damage reduction % on body, arms, back and legs
	hearing   [5]int // added to Perception when detecting hidden or sneaking people
	nightVisionTier int // lowest tier that sees in the dark (0 = never)
}

// cyberKinds, in item-number order (also INFO's listing order).
//
//	              Alpha  Beta  Gamma  Omega
//	Arm   STR       +5   +10   +15    +24
//	Leg   QUI       +4    +7   +10    +15
//	Reflexor AGI    +8   +15   +22    +35   QUI +3/+6/+9/+14
//	Eye   PER       +5   +10   +15    +25   night vision from Beta
//	Ear   hearing  +10   +20   +30    +40   (finding hidden/sneaking people only)
//	Armor DR %      70    75    80     85   body, arms, back, legs
//	Empathy cost     3     6    10     18   (armor 5/10/15/25, telepathor 5)
var cyberKinds = []*cyberKindDef{
	{kind: cyberArm, archetype: 206, name: "Bionic Arm",
		str: [5]int{0, 5, 10, 15, 24}, empCost: [5]int{0, 3, 6, 10, 18}},
	{kind: cyberLeg, archetype: 207, name: "Bionic Leg",
		qui: [5]int{0, 4, 7, 10, 15}, empCost: [5]int{0, 3, 6, 10, 18}},
	{kind: cyberEar, archetype: 208, name: "Bionic Ear",
		hearing: [5]int{0, 10, 20, 30, 40}, empCost: [5]int{0, 3, 6, 10, 18}},
	{kind: cyberArmor, archetype: 209, name: "Cyberplate Armor",
		armorDR: [5]int{0, 70, 75, 80, 85}, empCost: [5]int{0, 5, 10, 15, 25}},
	{kind: cyberTelepathor, archetype: 210, name: "Telepathor",
		empCost: [5]int{0, 5, 5, 5, 5}},
	{kind: cyberReflexor, archetype: 211, name: "Reflexor",
		agi: [5]int{0, 8, 15, 22, 35}, qui: [5]int{0, 3, 6, 9, 14}, empCost: [5]int{0, 3, 6, 10, 18}},
	{kind: cyberEye, archetype: 212, name: "Bionic Eye",
		per: [5]int{0, 5, 10, 15, 25}, empCost: [5]int{0, 3, 6, 10, 18}, nightVisionTier: 2},
}

// cyberSurgeryRoom is where a buyer wakes up (ADDAMS.SCR 789, "CyberTech, surgery").
const cyberSurgeryRoom = 789

func cyberKindByName(kind string) *cyberKindDef {
	for _, k := range cyberKinds {
		if k.kind == kind {
			return k
		}
	}
	return nil
}

// cyberKindForItem returns the implant type an item def represents. Only the
// INTANGIBLE CyberTech items count: INTANGIBLE also marks natural weapons.
// Item 205 (cyber claws) has no entry; CyberTech stopped selling them
// ("unavailable until they can be fixed").
func cyberKindForItem(def *gameworld.ItemDef) *cyberKindDef {
	if def == nil || !containsFlag(def.Flags, "INTANGIBLE") {
		return nil
	}
	for _, k := range cyberKinds {
		if k.archetype == def.Number {
			return k
		}
	}
	return nil
}

func (k *cyberKindDef) displayName(tier int) string {
	return cyberTierNames[tier] + " " + k.name
}

// cyberImplant returns p's installed implant of the given kind, or nil.
func (p *Player) cyberImplant(kind string) *CyberImplant {
	for i := range p.Cyberware {
		if p.Cyberware[i].Kind == kind {
			return &p.Cyberware[i]
		}
	}
	return nil
}

func (p *Player) cyberTier(kind string) int {
	if imp := p.cyberImplant(kind); imp != nil {
		return imp.Tier
	}
	return 0
}

// cyberArmorDR is the damage reduction from Cyberplate Armor (0 without it).
func (p *Player) cyberArmorDR() int {
	t := p.cyberTier(cyberArmor)
	return cyberKindByName(cyberArmor).armorDR[t]
}

// cyberHearingBonus is a Bionic Ear's bonus to noticing hidden or sneaking people.
func (p *Player) cyberHearingBonus() int {
	return cyberKindByName(cyberEar).hearing[p.cyberTier(cyberEar)]
}

// cyberNightVision reports whether a Bionic Eye lets p see in the dark.
func (p *Player) cyberNightVision() bool {
	k := cyberKindByName(cyberEye)
	t := p.cyberTier(cyberEye)
	return t > 0 && t >= k.nightVisionTier
}

func (p *Player) hasTelepathor() bool { return p.cyberTier(cyberTelepathor) > 0 }

// hasInnateTelepathy: telepathy that is part of the character rather than a
// timed effect (Ephemerals, a Telepathor). Mindlink's expiry and Dispel leave it alone.
func (p *Player) hasInnateTelepathy() bool { return p.IsEphemeral() || p.hasTelepathor() }

// cyberNoMagicMsg is why an implanted character can't cast (legends/bugs.txt:
// "spells being inactive upon purchase of cyberware") or grow Empathy back with
// Coriam Seeds. Removing every implant (@cyberware <player> remove all) lifts both.
const cyberNoMagicMsg = "The cybernetics in your body leave you deaf to the flow of mana."

// cyberArmorBlocks reports whether Cyberplate Armor rules out wearing def. The
// plating covers chest, back, shoulders, torso, arms and legs; head, hands and
// feet stay free for helms, gloves, gauntlets and boots. Body armor and bracers
// are always blocked; on other body slots only items that actually give
// protection (PARAMETER1 > 0) are, so cloaks, capes and kilts still fit.
func cyberArmorBlocks(def *gameworld.ItemDef) bool {
	if def == nil || def.Type != "ARMOR" {
		return false
	}
	switch def.WornSlot {
	case "WORN_ARMOR", "WORN_WRIST":
		return true
	case "WORN_BACK", "WORN_TRUNK1", "WORN_TRUNK2", "WORN_TORSO1", "WORN_TORSO2",
		"WORN_TORSO3", "WORN_BODY", "WORN_DON":
		return def.Parameter1 > 0
	}
	return false
}

// applyCyberStats installs k at tier on p's stats and returns the record.
func applyCyberStats(p *Player, k *cyberKindDef, tier int) CyberImplant {
	imp := CyberImplant{Kind: k.kind, Tier: tier,
		Str: k.str[tier], Agi: k.agi[tier], Qui: k.qui[tier], Per: k.per[tier]}
	p.Strength += imp.Str
	p.Agility += imp.Agi
	p.Quickness += imp.Qui
	p.Perception += imp.Per

	drain := min(k.empCost[tier], max(p.Empathy, 0))
	if drain > 0 {
		oldEmp := p.Empathy
		p.Empathy -= drain
		imp.Emp = -drain
		// Mana comes from Empathy, so it shrinks in proportion; no Empathy, no mana.
		newMax := p.MaxMana * p.Empathy / oldEmp
		imp.ManaLost = p.MaxMana - newMax
		p.MaxMana = newMax
		p.Mana = min(p.Mana, p.MaxMana)
	}
	return imp
}

// undoCyberStats reverses applyCyberStats.
func undoCyberStats(p *Player, imp CyberImplant) {
	p.Strength -= imp.Str
	p.Agility -= imp.Agi
	p.Quickness -= imp.Qui
	p.Perception -= imp.Per
	p.Empathy -= imp.Emp
	p.MaxMana += imp.ManaLost
}

// installCyberware grafts k at tier onto p, replacing (and undoing) any lower
// implant of the same kind. It returns notes for the player: armor stripped
// off by new Cyberplate Armor, and telepathy coming online.
func (e *GameEngine) installCyberware(p *Player, k *cyberKindDef, tier int) []string {
	var notes []string
	for i := range p.Cyberware {
		if p.Cyberware[i].Kind == k.kind {
			undoCyberStats(p, p.Cyberware[i])
			p.Cyberware = append(p.Cyberware[:i], p.Cyberware[i+1:]...)
			break
		}
	}
	p.Cyberware = append(p.Cyberware, applyCyberStats(p, k, tier))
	p.PreparedSpell = 0
	sort.Slice(p.Cyberware, func(a, b int) bool {
		return cyberKindByName(p.Cyberware[a].Kind).archetype < cyberKindByName(p.Cyberware[b].Kind).archetype
	})

	switch k.kind {
	case cyberArmor:
		var removed []string
		kept := p.Worn[:0]
		for _, w := range p.Worn {
			def := e.items[w.Archetype]
			if cyberArmorBlocks(def) {
				w.WornSlot = ""
				p.Inventory = append(p.Inventory, w)
				removed = append(removed, e.formatItemName(def, w.Adj1, w.Adj2, w.Adj3, w.Tail))
				continue
			}
			kept = append(kept, w)
		}
		p.Worn = kept
		if len(removed) > 0 {
			notes = append(notes, fmt.Sprintf("The technicians had to remove %s to fit the plating; it's in your pack.", joinWithAnd(removed)))
		}
	case cyberTelepathor:
		p.TelepathyActive = true
		p.TelepathyExpiry = time.Time{}
		notes = append(notes, "A faint hum of distant thoughts fills the back of your mind. (TELEPATHY turns it on or off; THINK sends.)")
	}
	return notes
}

// removeCyberware takes out p's implant of the given kind (GM use).
func (e *GameEngine) removeCyberware(p *Player, kind string) bool {
	for i := range p.Cyberware {
		if p.Cyberware[i].Kind != kind {
			continue
		}
		undoCyberStats(p, p.Cyberware[i])
		p.Cyberware = append(p.Cyberware[:i], p.Cyberware[i+1:]...)
		if kind == cyberTelepathor && !p.hasInnateTelepathy() && p.TelepathyExpiry.IsZero() {
			p.TelepathyActive = false
		}
		return true
	}
	return false
}

// cyberwareLines is INFO/STATUS's "Cyberware:" section (none if no implants).
func (p *Player) cyberwareLines() []string {
	if len(p.Cyberware) == 0 {
		return nil
	}
	lines := []string{"", "Cyberware:"}
	for _, imp := range p.Cyberware {
		if k := cyberKindByName(imp.Kind); k != nil {
			lines = append(lines, k.displayName(imp.Tier))
		}
	}
	return lines
}

// buyCyberware handles BUY of a CyberTech implant: check, charge, install, then
// wake the buyer on the operating table in the surgery (ADDAMS.SCR 789).
func (e *GameEngine) buyCyberware(ctx context.Context, player *Player, k *cyberKindDef, adj, price int) *CommandResult {
	tier := cyberTierAdjs[adj]
	if tier == 0 {
		tier = 1
	}
	if player.IsEphemeral() {
		return &CommandResult{Messages: []string{"The robot receptionist scans you and says, \"We regret that there is nothing physical to graft our devices onto.\""}}
	}
	if have := player.cyberTier(k.kind); have >= tier {
		return &CommandResult{Messages: []string{fmt.Sprintf("Your %s is already installed.", k.displayName(have))}}
	}
	md := e.roomCurrency(player.RoomNumber)
	if !e.spendFunds(player, md, price) {
		return &CommandResult{Messages: []string{fmt.Sprintf("You can't afford that. %s costs %s.", k.displayName(tier), formatCurrency(md, price))}}
	}

	upgrade := player.cyberTier(k.kind) > 0
	notes := e.installCyberware(player, k, tier)
	name := k.displayName(tier)

	msgs := []string{
		"You hand over your money. The robot receptionist ushers you through the oak door, and a masked physician helps you onto a cold metal table.",
		"A needle pricks your arm, and the world slides away...",
		"",
	}
	if upgrade {
		msgs = append(msgs, fmt.Sprintf("You wake on the operating table. Your old implant is gone, and your new %s hums in its place.", name))
	} else {
		msgs = append(msgs, fmt.Sprintf("You wake on the operating table, sore and woozy. Your new %s has been grafted into you.", name))
	}
	msgs = append(msgs, notes...)
	result := &CommandResult{RoomBroadcast: nil}

	oldRoom := player.RoomNumber
	if e.rooms[cyberSurgeryRoom] != nil {
		e.breakAllMeleeEngagements(player)
		player.RoomNumber = cyberSurgeryRoom
		player.Position = 2 // laying on the table; the room's STAND script handles getting up
		look := e.doLook(player)
		msgs = append(msgs, "")
		msgs = append(msgs, look.Messages...)
		result.RoomName, result.RoomDesc, result.Exits, result.Items = look.RoomName, look.RoomDesc, look.Exits, look.Items
		result.OldRoom = oldRoom
		result.OldRoomMsg = []string{fmt.Sprintf("The robot receptionist leads %s through the oak door to surgery.", player.DisplayName())}
		result.RoomBroadcast = []string{fmt.Sprintf("Orderlies wheel %s in on a gurney and lift %s onto the operating table.", player.DisplayName(), player.Objective())}
	} else {
		result.RoomBroadcast = []string{fmt.Sprintf("%s is led away to surgery and returns some time later, moving a little stiffly.", player.DisplayNameCap())}
	}
	result.Messages = msgs
	result.PlayerState = player
	e.SavePlayer(ctx, player)
	return result
}

// migrateCyberwareItems converts CyberTech purchases made before installation
// existed (inventory objects that did nothing) into installed implants. The
// highest tier of each kind wins; the objects are removed.
func (e *GameEngine) migrateCyberwareItems(p *Player) bool {
	best := map[*cyberKindDef]int{}
	kept := p.Inventory[:0]
	for _, ii := range p.Inventory {
		k := cyberKindForItem(e.items[ii.Archetype])
		if k == nil {
			kept = append(kept, ii)
			continue
		}
		tier := 1
		for _, adj := range []int{ii.Adj1, ii.Adj2, ii.Adj3} {
			if t := cyberTierAdjs[adj]; t > 0 {
				tier = t
			}
		}
		best[k] = max(best[k], tier)
	}
	if len(best) == 0 {
		return false
	}
	p.Inventory = kept
	for k, tier := range best {
		if tier > p.cyberTier(k.kind) {
			e.installCyberware(p, k, tier)
		}
	}
	p.LoginNotices = append(p.LoginNotices, "CyberTech has finished installing the cyberware you bought. Type INFO to see it.")
	return true
}

// gmCyberware handles @cyberware:
//
//	@cyberware <player>                          list implants
//	@cyberware <player> add <kind> <tier>        install (tier: alpha..omega or 1-4)
//	@cyberware <player> remove <kind>            take one out, undoing its stats
func (e *GameEngine) gmCyberware(ctx context.Context, args []string) *CommandResult {
	usage := "Usage: @cyberware <player> [add <arm|leg|ear|armor|telepathor|reflexor|eye> <alpha|beta|gamma|omega> | remove <kind|all>]"
	if len(args) < 1 {
		return &CommandResult{Messages: []string{usage}}
	}
	target, err := e.resolveOnlinePlayer(args[0])
	if err != nil {
		return &CommandResult{Messages: []string{err.Error()}}
	}
	if len(args) == 1 {
		lines := target.cyberwareLines()
		if lines == nil {
			return &CommandResult{Messages: []string{fmt.Sprintf("%s has no cyberware.", target.FullName())}}
		}
		return &CommandResult{Messages: append([]string{target.FullName() + ":"}, lines[1:]...)}
	}
	if len(args) < 3 {
		return &CommandResult{Messages: []string{usage}}
	}
	if verb := strings.ToLower(args[1]); (verb == "remove" || verb == "rem" || verb == "del") && strings.EqualFold(args[2], "all") {
		if len(target.Cyberware) == 0 {
			return &CommandResult{Messages: []string{fmt.Sprintf("%s has no cyberware.", target.FullName())}}
		}
		for len(target.Cyberware) > 0 {
			e.removeCyberware(target, target.Cyberware[0].Kind)
		}
		e.SavePlayer(ctx, target)
		if e.sendToPlayer != nil {
			e.sendToPlayer(target.FirstName, []string{"The last of the cold metal is gone from your body, and you feel the flow of mana around you once more."})
		}
		return &CommandResult{Messages: []string{fmt.Sprintf("Removed all of %s's cyberware; Empathy and Mana restored and spellcasting allowed again.", target.FullName())}}
	}
	k := cyberKindByName(strings.ToLower(args[2]))
	if k == nil {
		return &CommandResult{Messages: []string{usage}}
	}
	switch strings.ToLower(args[1]) {
	case "add", "install":
		if len(args) < 4 {
			return &CommandResult{Messages: []string{usage}}
		}
		tier, err := strconv.Atoi(args[3])
		if err != nil {
			for i, n := range cyberTierNames {
				if i > 0 && strings.EqualFold(n, args[3]) {
					tier = i
				}
			}
		}
		if tier < 1 || tier > 4 {
			return &CommandResult{Messages: []string{usage}}
		}
		notes := e.installCyberware(target, k, tier)
		e.SavePlayer(ctx, target)
		if e.sendToPlayer != nil {
			e.sendToPlayer(target.FirstName, append([]string{fmt.Sprintf("You feel your new %s settle into place.", k.displayName(tier))}, notes...))
		}
		return &CommandResult{Messages: []string{fmt.Sprintf("Installed %s in %s.", k.displayName(tier), target.FullName())}}
	case "remove", "rem", "del":
		if !e.removeCyberware(target, k.kind) {
			return &CommandResult{Messages: []string{fmt.Sprintf("%s has no %s.", target.FullName(), k.name)}}
		}
		e.SavePlayer(ctx, target)
		return &CommandResult{Messages: []string{fmt.Sprintf("Removed %s's %s; its stat changes were undone.", target.FullName(), k.name)}}
	}
	return &CommandResult{Messages: []string{usage}}
}
