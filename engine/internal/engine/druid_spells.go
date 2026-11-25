package engine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Spells that were on the list but did nothing: True Aim (512), Free Action (528),
// Tree Door (531), Ride the Lightning (532), Claws of the Elder Wolf (534), Form Lock
// (535), Wolf Form (536), and from other schools Unstun (338), Analyze Ore (407) and
// Tindarath's Chaotic Summons (144).

const (
	treeDoorSpellID = 531
	wolfFormSpellID = 536

	trueAimDuration  = 10 * time.Second
	trueAimRanks     = 20
	formLockDuration = 5 * time.Minute
	fearDuration     = 60 * time.Second
	// fearAttackPenalty is what Fear (200) costs a player's attack rating. The captures
	// (badge.txt) only show that a terrified player can still fight.
	fearAttackPenalty = 20

	trueAimFadeMsg = "The gleam leaves your eyes."
)

// timedEffect is a spell effect that is nothing more than an expiry on the player.
type timedEffect struct {
	spellID int // 0: not one castTimedEffect handles
	name    string
	field   func(*Player) *time.Time
	gain    string // to the player
	gainOth string // to the room, %s = the player
	extend  string
	fade    string
}

// timedEffects are ended with their fade line by spellEffectTick, and by Dispel Lesser
// Magic and Disjunction (activeDispellableEffects).
var timedEffects = []timedEffect{
	{spellID: 532, name: "Ride the Lightning", field: func(p *Player) *time.Time { return &p.RideLightningExpiry },
		gain:    "Tiny arcs of lightning crackle over your skin. Electricity can no longer harm you!",
		gainOth: "Tiny arcs of lightning crackle over %s.",
		extend:  "The lightning crackling over you grows brighter.",
		fade:    "The arcs of lightning crackling over you sputter out."},
	{spellID: 528, name: "Free Action", field: func(p *Player) *time.Time { return &p.FreeActionExpiry },
		gain:    "You feel as though nothing could hold you!",
		gainOth: "%s moves with sudden, easy freedom.",
		extend:  "Your freedom of movement is renewed.",
		fade:    "Your freedom of movement fades."},
	{spellID: 534, name: "Claws of the Elder Wolf", field: func(p *Player) *time.Time { return &p.ElderWolfExpiry },
		gain:    "The power of the elder wolf flows into your claws!",
		gainOth: "A grey light flickers around %s's hands.",
		extend:  "The power of the elder wolf flows on through your claws.",
		fade:    "The power of the elder wolf leaves your claws."},
	{name: "Form Lock", field: func(p *Player) *time.Time { return &p.FormLockExpiry }, fade: "Your body feels free to change its form again."},
	{name: "True Aim", field: func(p *Player) *time.Time { return &p.TrueAimExpiry }, fade: trueAimFadeMsg},
	{name: "Fear", field: func(p *Player) *time.Time { return &p.FearExpiry }, fade: "The feeling of terror leaves you."}, // badge.txt
}

func timedEffectForSpell(spellID int) *timedEffect {
	for i := range timedEffects {
		if timedEffects[i].spellID == spellID && spellID != 0 {
			return &timedEffects[i]
		}
	}
	return nil
}

// spellEffectTick ends player spell effects as they run out. regenTick's once a minute
// is too slow for a ten-second True Aim, a slumber of a few seconds, or a web.
func (e *GameEngine) spellEffectTick() {
	if e.sessions == nil {
		return
	}
	now := time.Now()
	for _, p := range e.sessions.OnlinePlayers() {
		if p == nil {
			continue
		}
		e.sentenceServed(p)
		var msgs []string
		for _, te := range timedEffects {
			if expiry := te.field(p); !expiry.IsZero() && now.After(*expiry) {
				*expiry = time.Time{}
				msgs = append(msgs, te.fade)
				if te.name == "True Aim" && !p.IsConcealed() && e.roomBroadcastExclude != nil {
					e.roomBroadcastExclude(p.RoomNumber, p.FirstName, []string{fmt.Sprintf("The gleam leaves %s's eyes.", p.DisplayName())})
				}
			}
		}
		if !p.CharmExpiry.IsZero() && now.After(p.CharmExpiry) {
			p.CharmExpiry = time.Time{}
			msgs = append(msgs, "You come to your senses.")
		}
		if !p.SleepExpiry.IsZero() && now.After(p.SleepExpiry) {
			if wake := wakeFromSlumber(p); wake != "" {
				msgs = append(msgs, wake)
			}
		}
		if p.Stunned && !p.StunExpiry.IsZero() && now.After(p.StunExpiry) {
			p.Stunned = false
			p.StunExpiry = time.Time{}
			msgs = append(msgs, "You shake off the stun and can act again.")
		}
		var held []PlayerEntangle
		for _, ent := range p.Entangles {
			if now.After(ent.Expiry) && damagingSnareForSpell(ent.SpellID) == nil {
				msgs = append(msgs, entangleReleaseMsg(ent))
			} else {
				held = append(held, ent)
			}
		}
		if len(held) != len(p.Entangles) {
			p.Entangles = held
		}
		if len(msgs) > 0 {
			e.SavePlayer(context.Background(), p)
			e.sendToPlayerMsgs(p, msgs)
		}
	}
}

// castTimedEffect handles Free Action (528), Ride the Lightning (532) and Claws of the
// Elder Wolf (534): 20 minutes, extended by further casts up to four hours.
func (e *GameEngine) castTimedEffect(player *Player, spell *SpellDef, args []string) *CommandResult {
	te := timedEffectForSpell(spell.ID)
	target, mInst, _, isSelf, notFound := e.resolveBuffTarget(player, args)
	if notFound {
		return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", strings.Join(args, " "))}, TargetNotFound: true}
	}
	if mInst != nil {
		return &CommandResult{Messages: []string{fmt.Sprintf("%s has no effect on creatures.", spell.Name)}, TargetNotFound: true}
	}
	mins, applied := applyElementalShield(te.field(target), player.spellDurationMult(spell))
	line, left := te.gain, fmt.Sprintf(" (%s)", minutesText(mins))
	if !applied {
		line, left = te.extend, fmt.Sprintf(" (%d minutes remaining)", mins)
	}
	roomLine := fmt.Sprintf(te.gainOth, target.DisplayName())
	if isSelf {
		result := &CommandResult{
			Messages:      []string{"You gesture.", line + left},
			RoomBroadcast: []string{fmt.Sprintf("%s gestures.", player.DisplayNameCap())},
		}
		if applied {
			result.RoomBroadcast = append(result.RoomBroadcast, capitalize(roomLine))
		}
		return result
	}
	e.SavePlayer(context.Background(), target)
	result := &CommandResult{
		Messages:      []string{fmt.Sprintf("You gesture at %s.", target.DisplayName())},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), target.DisplayName())},
		TargetName:    target.FirstName,
		TargetMsg:     []string{fmt.Sprintf("%s gestures at you.", player.DisplayNameCap()), line + left},
	}
	if applied {
		result.Messages = append(result.Messages, capitalize(roomLine))
		result.RoomBroadcast = append(result.RoomBroadcast, capitalize(roomLine))
	}
	return result
}

// absorbLightning turns electrical damage into mana for a player under Ride the
// Lightning (532), all of it, even past their maximum. It returns the lines to show the
// player and the room, or "" if the damage should land as usual.
func absorbLightning(p *Player, dmg int, dmgType string) (selfMsg, roomMsg string) {
	if !strings.EqualFold(dmgType, "electric") || dmg <= 0 || !spellActive(p.RideLightningExpiry) {
		return "", ""
	}
	p.Mana += dmg
	return fmt.Sprintf("The lightning courses into you, filling you with power! [+%d Mana, %d/%d]", dmg, p.Mana, p.MaxMana),
		fmt.Sprintf("The lightning courses into %s harmlessly!", p.DisplayName())
}

// stormStrikesRider is regenTick's once-a-minute lightning strike on a player riding the
// lightning outdoors in a storm (the weather Call Lightning and Enchant an Item need):
// a Lightning Bolt's worth (103), all of it turned into mana. Reports whether it struck.
func (e *GameEngine) stormStrikesRider(p *Player) bool {
	if p.Dead || !spellActive(p.RideLightningExpiry) {
		return false
	}
	room := e.rooms[p.RoomNumber]
	if room == nil || !isOutdoorTerrain(room.Terrain) {
		return false
	}
	if wea := e.RegionWeather[room.Region]; wea < 6 || wea > 8 {
		return false
	}
	bolt := FindSpellByID(103)
	selfMsg, _ := absorbLightning(p, rand.Intn(bolt.DmgMax-bolt.DmgMin+1)+bolt.DmgMin, "electric")
	e.sendToPlayerMsgs(p, []string{"A bolt of lightning streaks down from the storm and strikes you!", selfMsg})
	if !p.IsConcealed() && e.roomBroadcastExclude != nil {
		e.roomBroadcastExclude(p.RoomNumber, p.FirstName, []string{fmt.Sprintf("A bolt of lightning streaks down from the storm and strikes %s, who seems to drink it in!", p.DisplayName())})
	}
	return true
}

// usingClaws reports whether the player's current attack is with claws: grown (Claw
// Growth), wielded, or a wolf's.
func usingClaws(p *Player, weaponDef *gameworld.ItemDef) bool {
	if weaponDef == nil {
		return p.WolfForm
	}
	return weaponDef.Type == "CLAW_WEAPON"
}

// trueAimBonus is the ranks of missile skill True Aim (512) adds to a bow or thrown
// weapon, or 0.
func trueAimBonus(p *Player, weaponDef *gameworld.ItemDef) int {
	if weaponDef == nil || !isRangedWeaponType(weaponDef.Type) || !spellActive(p.TrueAimExpiry) {
		return 0
	}
	return trueAimRanks
}

// applyTrueAim starts or resets True Aim. Per MAGIC.TXT it doesn't accumulate: a recast
// "will 'reset' the duration from the beginning."
func applyTrueAim(p *Player, mult float64) {
	p.TrueAimExpiry = time.Now().Add(scaleSpellDuration(trueAimDuration, mult))
}

// castTrueAim handles True Aim (512): "Adds 20 ranks of missile skill for 10 seconds."
// The gleam in the eyes is from quilan.txt, where an archer casts it before every shot.
func (e *GameEngine) castTrueAim(player *Player, spell *SpellDef, args []string) *CommandResult {
	target, mInst, _, isSelf, notFound := e.resolveBuffTarget(player, args)
	if notFound {
		return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", strings.Join(args, " "))}, TargetNotFound: true}
	}
	if mInst != nil {
		return &CommandResult{Messages: []string{"True Aim has no effect on creatures."}, TargetNotFound: true}
	}
	applyTrueAim(target, player.spellDurationMult(spell))
	gleam := fmt.Sprintf("A gleam appears in %s's eyes.", target.DisplayName())
	if isSelf {
		return &CommandResult{
			Messages:      []string{"You gesture.", "A gleam appears in your eyes."},
			RoomBroadcast: []string{fmt.Sprintf("%s gestures.", player.DisplayNameCap()), gleam},
		}
	}
	e.SavePlayer(context.Background(), target)
	return &CommandResult{
		Messages:      []string{fmt.Sprintf("You gesture at %s.", target.DisplayName()), gleam},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), target.DisplayName()), gleam},
		TargetName:    target.FirstName,
		TargetMsg:     []string{fmt.Sprintf("%s gestures at you.", player.DisplayNameCap()), "A gleam appears in your eyes."},
	}
}

// castUnstun handles Unstun (338): "Sun is no longer stunned." (rorinquest.txt)
func (e *GameEngine) castUnstun(player *Player, spell *SpellDef, args []string) *CommandResult {
	target, mInst, mDef, isSelf, notFound := e.resolveBuffTarget(player, args)
	if notFound {
		return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", strings.Join(args, " "))}, TargetNotFound: true}
	}
	if mInst != nil {
		name := strings.ToLower(FormatMonsterName(mDef, e.monAdjs))
		full := articleFor(name, mDef.Unique) + name
		line := capitalize(full) + " isn't stunned."
		if mInst.Stunned {
			e.updateMonster(mInst.ID, func(m *MonsterInstance) { m.Stunned, m.StunExpiry = false, time.Time{} })
			line = capitalize(full) + " is no longer stunned."
		}
		return &CommandResult{
			Messages:      []string{fmt.Sprintf("You gesture at %s.", full), line},
			RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), full), line},
		}
	}
	wasStunned := target.Stunned
	target.Stunned = false
	target.StunExpiry = time.Time{}
	if isSelf {
		line := "You aren't stunned."
		if wasStunned {
			line = "You are no longer stunned."
		}
		return &CommandResult{Messages: []string{"You gesture.", line}, RoomBroadcast: []string{fmt.Sprintf("%s gestures.", player.DisplayNameCap())}}
	}
	result := &CommandResult{
		Messages:      []string{fmt.Sprintf("You gesture at %s.", target.DisplayName())},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), target.DisplayName())},
		TargetName:    target.FirstName,
		TargetMsg:     []string{fmt.Sprintf("%s gestures at you.", player.DisplayNameCap())},
	}
	if !wasStunned {
		result.Messages = append(result.Messages, fmt.Sprintf("%s isn't stunned.", target.DisplayNameCap()))
		return result
	}
	e.SavePlayer(context.Background(), target)
	line := fmt.Sprintf("%s is no longer stunned.", target.DisplayNameCap())
	result.Messages = append(result.Messages, line)
	result.RoomBroadcast = append(result.RoomBroadcast, line)
	result.TargetMsg = append(result.TargetMsg, "You are no longer stunned.")
	return result
}

// castAnalyzeOre handles Analyze Ore (407): "Tells an ore's purity", no Mining needed,
// for ore the caster is carrying or ore lying in the room.
func (e *GameEngine) castAnalyzeOre(player *Player, args []string) *CommandResult {
	if len(args) == 0 {
		return &CommandResult{Messages: []string{"Cast Analyze Ore at what?"}, TargetNotFound: true}
	}
	target := strings.Join(args, " ")
	target, _ = stripMyPrefix(strings.ToLower(target))
	invItem, roomItem, def := e.findMutableItemTarget(player, target)
	if def == nil {
		return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", target)}, TargetNotFound: true}
	}
	var name string
	var purity int
	if roomItem != nil {
		name = e.formatItemName(def, roomItem.Adj1, roomItem.Adj2, roomItem.Adj3, roomItem.Extend)
		purity = roomItem.Val1
	} else {
		name = e.formatItemName(def, invItem.Adj1, invItem.Adj2, invItem.Adj3, invItem.Tail)
		purity = invItem.Val1
	}
	result := &CommandResult{
		Messages:      []string{fmt.Sprintf("You gesture at %s.", name)},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), name)},
	}
	if def.Type != "ORE" {
		result.Messages = append(result.Messages, "The spell tells you nothing; that isn't ore.")
		return result
	}
	result.Messages = append(result.Messages, fmt.Sprintf("You sense that the ore is of %s quality. (Purity: %d%%)", orePurityWord(purity), purity))
	return result
}

// roomHasTree reports whether a Tree Door can open in the room: forest, jungle, or
// anywhere with a tree.
func (e *GameEngine) roomHasTree(room *gameworld.Room) bool {
	if room == nil {
		return false
	}
	if room.Terrain == "FOREST" || room.Terrain == "JUNGLE" {
		return true
	}
	for i := range room.Items {
		if def := e.items[room.Items[i].Archetype]; def != nil {
			switch strings.ToLower(e.nouns[def.NameID]) {
			case "tree", "trees":
				return true
			}
		}
	}
	return false
}

// castTreeDoor handles Tree Door (531): the caster alone steps into a tree here and out
// at a mark that has trees too. MAGIC.TXT calls the original "very picky about which
// tree works"; this deliberately isn't. The lines are
// FULLTEIW.SCR's (a script that isn't loaded).
func (e *GameEngine) castTreeDoor(ctx context.Context, player *Player, args []string) *CommandResult {
	if !e.roomHasTree(e.rooms[player.RoomNumber]) {
		return &CommandResult{Messages: []string{"There is no tree here to step into."}, TargetNotFound: true}
	}
	dest, errMsg := e.resolveBendSpaceMark(player, args)
	if errMsg != "" {
		return &CommandResult{Messages: []string{errMsg}, TargetNotFound: true}
	}
	if !e.roomHasTree(dest) {
		return &CommandResult{Messages: []string{"There is no tree at that mark to step out of."}, TargetNotFound: true}
	}

	originalRoom := player.RoomNumber
	concealed := player.IsConcealed()
	player.RoomNumber = dest.Number
	player.breakGuard()
	e.SavePlayer(ctx, player)

	look := e.doLook(player)
	result := &CommandResult{
		Messages: append([]string{"You gesture.", "The tree accepts you as you become one with it..."}, look.Messages...),
		RoomName: look.RoomName,
		RoomDesc: look.RoomDesc,
		Exits:    look.Exits,
		Items:    look.Items,
		OldRoom:  originalRoom,
	}
	if !player.IsGM && !concealed {
		result.OldRoomMsg = []string{fmt.Sprintf("%s sinks into the tree and is gone.", player.DisplayNameCap())}
		result.RoomBroadcast = []string{fmt.Sprintf("%s seems to appear out of nowhere.", player.DisplayNameCap())}
	}
	e.applyEntryScripts(ctx, player, dest, result)
	return result
}

// formFixedMessage refuses a change of form to a player under Form Lock (535), or
// returns "".
func formFixedMessage(p *Player) string {
	if !spellActive(p.FormLockExpiry) {
		return ""
	}
	return "You strain to change your form, but something holds it fixed!"
}

// lockForm starts or resets Form Lock. It doesn't accumulate.
func lockForm(p *Player) { p.FormLockExpiry = time.Now().Add(formLockDuration) }

// castFormLock handles Form Lock (535): "Locks target into their current form (human,
// slime, mist, wolf.)"
func (e *GameEngine) castFormLock(player *Player, spell *SpellDef, args []string) *CommandResult {
	target, mInst, _, isSelf, notFound := e.resolveBuffTarget(player, args)
	if notFound {
		return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", strings.Join(args, " "))}, TargetNotFound: true}
	}
	if mInst != nil {
		return &CommandResult{Messages: []string{"Form Lock has no effect on creatures."}, TargetNotFound: true}
	}
	lockForm(target)
	const locked = "Your body stiffens as its form becomes fixed!"
	if isSelf {
		return &CommandResult{
			Messages:      []string{"You gesture.", locked},
			RoomBroadcast: []string{fmt.Sprintf("%s gestures.", player.DisplayNameCap())},
		}
	}
	e.SavePlayer(context.Background(), target)
	line := fmt.Sprintf("%s stiffens for a moment.", target.DisplayNameCap())
	return &CommandResult{
		Messages:      []string{fmt.Sprintf("You gesture at %s.", target.DisplayName()), line},
		RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), target.DisplayName()), line},
		TargetName:    target.FirstName,
		TargetMsg:     []string{fmt.Sprintf("%s gestures at you.", player.DisplayNameCap()), locked},
	}
}

// wolfFormRefusal says why the player can't cast Wolf Form (536), or "". It takes a good
// alignment, above neutral.
func wolfFormRefusal(p *Player) string {
	if p.Alignment <= 0 && !p.IsGM {
		return "The spirit of the wolf does not answer you. It comes only to those of good heart."
	}
	return ""
}

// becomeWolf puts the player into wolf form, shaking off whatever held them as a
// wolfling's TRANSFORM does (MAGIC.TXT). Returns the lines for them and the room.
func becomeWolf(p *Player) (selfMsgs []string, roomMsg string) {
	roomMsg = fmt.Sprintf("Without warning, %s howls and collapses to the ground, shaking. Undergoing a terrible transformation, %s changes shape into that of a wolf!", p.DisplayName(), p.Pronoun())
	selfMsgs = []string{"You groan in pain as your body undergoes a metamorphosis and assumes the form of a wolf. [Type TRANSFORM to resume your own shape]"}
	for _, ent := range p.Entangles {
		selfMsgs = append(selfMsgs, entangleReleaseMsg(ent))
	}
	p.Entangles = nil
	p.WolfForm = true
	return selfMsgs, roomMsg
}

// castWolfForm handles Wolf Form (536): the caster alone becomes a wolf, until they
// TRANSFORM back.
func (e *GameEngine) castWolfForm(player *Player) *CommandResult {
	if msg := wolfFormRefusal(player); msg != "" { // prepared from a scroll
		return &CommandResult{Messages: []string{"You gesture.", msg}}
	}
	if player.WolfForm || player.MistForm || player.SlimeForm {
		return &CommandResult{Messages: []string{"You are already in another form."}, TargetNotFound: true}
	}
	if msg := formFixedMessage(player); msg != "" {
		return &CommandResult{Messages: []string{"You gesture.", msg}}
	}
	if player.Wielded != nil || player.OffHand != nil {
		return &CommandResult{Messages: []string{"You must unwield your weapon and shield before you can transform."}, TargetNotFound: true}
	}
	selfMsgs, roomMsg := becomeWolf(player)
	return &CommandResult{Messages: append([]string{"You gesture."}, selfMsgs...), RoomBroadcast: []string{roomMsg}}
}

// castChaoticSummons handles Tindarath's Chaotic Summons (144). MAGIC.TXT lists it among
// the guards: it gets a high level conjurer "something equivalent to, but better looking
// than, an elemental", "Component varies". Which component brought what isn't recorded;
// an ursine tongue and the fire tiger (MNUMBER 114), between the gargoyle of rank 16 and
// the spectral warrior of rank 32 in strength, stand in until it's known.
func (e *GameEngine) castChaoticSummons(player *Player) *CommandResult {
	if player.SummonedCreatureID != 0 {
		return &CommandResult{Messages: []string{"You already have a summoned creature. Use COMMAND BEGONE to dismiss it first."}}
	}
	player.PreparedSpellReagentArch = 0
	return e.spawnSummonedCreature(player, 114, false, "Reality twists and tears, and out of the chaos steps")
}

// applyItemUtilitySpell gives the player the effect of one of these spells when an item
// or a room script casts it on them (SPELL n), reporting whether it was one of them.
func (sc *ScriptContext) applyItemUtilitySpell(spell *SpellDef, mult float64) bool {
	p, e := sc.Player, sc.Engine
	name := p.DisplayName()
	if te := timedEffectForSpell(spell.ID); te != nil {
		mins, applied := applyElementalShield(te.field(p), mult)
		if applied {
			sc.Messages = append(sc.Messages, fmt.Sprintf("%s (%s)", te.gain, minutesText(mins)))
			sc.RoomMsgs = append(sc.RoomMsgs, capitalize(fmt.Sprintf(te.gainOth, name)))
		} else {
			sc.Messages = append(sc.Messages, fmt.Sprintf("%s (%d minutes remaining)", te.extend, mins))
		}
		return true
	}
	switch spell.ID {
	case 230: // Disjunction
		sc.Messages = append(sc.Messages, "A wave of deep red light emanates from you.")
		sc.Messages = append(sc.Messages, e.disjunct(p.RoomNumber, p, []string{fmt.Sprintf("A wave of deep red light emanates from %s.", name)})...)
	case 232, 245: // Mist Form, Slime Form
		if p.MistForm || p.SlimeForm || p.WolfForm {
			sc.Messages = append(sc.Messages, "You are already in another form.")
			return true
		}
		if msg := formFixedMessage(p); msg != "" {
			sc.Messages = append(sc.Messages, msg)
			return true
		}
		selfMsgs, roomMsg := takeForm(p, spell.ID)
		sc.Messages = append(sc.Messages, selfMsgs...)
		sc.RoomMsgs = append(sc.RoomMsgs, roomMsg)
	case 248: // Phantom Form
		p.PhantomForm = true
		sc.Messages = append(sc.Messages, "You feel a tingling sensation.")
		sc.RoomMsgs = append(sc.RoomMsgs, fmt.Sprintf("%s fades from sight.", capitalize(name)))
	case 303, 319: // Cure Poison, Cure Disease: the lights are the captures' (shemri.txt, quilan.txt)
		a, color := poisonAffliction, "blue"
		if spell.ID == 319 {
			a, color = diseaseAffliction, "purple"
		}
		sc.Messages = append(sc.Messages, fmt.Sprintf("A %s light flashes around you.", color))
		sc.RoomMsgs = append(sc.RoomMsgs, fmt.Sprintf("A %s light flashes around %s.", color, name))
		if on, level := a.player(p); *on {
			*on, *level = false, 0
			sc.Messages = append(sc.Messages, a.endMsg())
		}
	case 403: // Mindlink
		mins, applied := applyMindlink(p, mult)
		if applied {
			sc.Messages = append(sc.Messages, "You feel the power of your mind expanding.")
			sc.RoomMsgs = append(sc.RoomMsgs, fmt.Sprintf("%s seems changed.", capitalize(name)))
		} else {
			sc.Messages = append(sc.Messages, fmt.Sprintf("You feel the power of your mind expanding further. (%d minutes remaining)", mins))
		}
	case 518: // Claw Growth
		mins, applied := applyClawGrowth(p, mult)
		if applied {
			sc.Messages = append(sc.Messages, fmt.Sprintf("Wicked claws erupt from your fingertips! (%s)", minutesText(mins)))
			sc.RoomMsgs = append(sc.RoomMsgs, fmt.Sprintf("Claws erupt from %s's fingertips!", name))
		} else {
			sc.Messages = append(sc.Messages, fmt.Sprintf("Your claws grow sharper still! (%d minutes remaining)", mins))
		}
	case 406: // Dispel Invisibility: EVILFAYD.SCR and others cast it at whoever walks in
		if p.Invisible {
			e.revealInvisible(p)
			sc.Messages = append(sc.Messages, "You suddenly fade into view.", "You suddenly become visible.")
			sc.RoomMsgs = append(sc.RoomMsgs, fmt.Sprintf("%s suddenly fades into view.", p.DisplayNameCap()), fmt.Sprintf("%s suddenly becomes visible.", p.DisplayNameCap()))
		}
	case 512: // True Aim
		applyTrueAim(p, mult)
		sc.Messages = append(sc.Messages, "A gleam appears in your eyes.")
		sc.RoomMsgs = append(sc.RoomMsgs, fmt.Sprintf("A gleam appears in %s's eyes.", name))
	case 535: // Form Lock
		lockForm(p)
		sc.Messages = append(sc.Messages, "Your body stiffens as its form becomes fixed!")
	case wolfFormSpellID: // Wolf Form: EDDIE2.SCR's wolf statue and others; the item does the work
		if p.WolfForm || p.MistForm || p.SlimeForm {
			return true
		}
		if msg := formFixedMessage(p); msg != "" {
			sc.Messages = append(sc.Messages, msg)
			return true
		}
		for _, slot := range []**InventoryItem{&p.Wielded, &p.OffHand} {
			if *slot != nil {
				p.Inventory = append(p.Inventory, **slot)
				*slot = nil
			}
		}
		selfMsgs, roomMsg := becomeWolf(p)
		sc.Messages = append(sc.Messages, selfMsgs...)
		sc.RoomMsgs = append(sc.RoomMsgs, roomMsg)
	default:
		// Web, Plant Snare, Wall of Thorns, Creeping Death, Fear, Charm, Slow, Slumber,
		// Silence, Paranoia, Imprison, Wither Limb, Sunray
		if !monsterUtilitySpells[spell.ID] {
			return false
		}
		selfMsgs, roomMsg := e.hostileSpellEffect(p, spell, nil, nil, mult)
		sc.Messages = append(sc.Messages, selfMsgs...)
		if roomMsg != "" {
			sc.RoomMsgs = append(sc.RoomMsgs, roomMsg)
		}
	}
	return true
}
