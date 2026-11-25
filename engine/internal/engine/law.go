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

// Fayd's law: WARRANT, ARREST, the Lockup, fines and lawkeeper NPCs.
//
// LEGENDS.DOC: "Lawkeepers are authorized to ARREST wanted criminals. High-ranking
// Lawkeepers may issue warrants for arrest." ARREST "allows the subject to SUBMIT and
// come along quietly. If they do not SUBMIT, and if you defeat them in combat, they
// are usually incapacitated rather than slain, and taken into the Fayd Hall of
// Justice." SUBMIT: "If you are wanted for arrest, submit and come along quietly."
//
// The warrant number is the class of the crime in the codex of Fayd (codex.txt, read
// out in LAWKEEP.SCR), I to IX, kept in the player's WARRANT script variable (0-9).
// The codex's laws govern "Citizens Within The City Walls, Including The Graveyard
// And Catacombs", and the lawkeeper badge (ITEM2.SCR) and KREETEK.SCR's note about
// the interrogation room both treat rooms 201-399 as where the law reaches.
//
// The cell is LAWKEEP.SCR's Lockup, room 340, behind the iron door in the Lawkeeper HQ
// (338). Its own IFENTRY sets the prisoner's round time to 40 seconds per class of
// crime and clears the warrant, and that's the sentence: the prisoner is let out when
// it's served. Paying the fine (PAY, or PAY <name> from the HQ) or a high-ranking
// lawkeeper's WARRANT <name> 0 lets them out sooner. The sentence only runs while
// they're in the Realms (spell_pause.go). The fines are our own: no document gives them.

const (
	lawkeeperOrg     = 3
	lockupRoom       = 340
	lawkeeperHQ      = 338
	warrantMinRank   = 100 // "Master" and "High Master": the two top ranks (orgRankTitle)
	maxWarrantLevel  = 9
	sentencePerClass = 40 * time.Second // LAWKEEP.SCR room 340: MULT DUMMY1 40
	arrestGrace      = 20 * time.Second
	lawFirstCityRoom = 201
	lawLastCityRoom  = 399
)

// lawMonsters are Fayd's lawkeepers and enforcers (MONSTERS.SCR 13, 18; FAYD.SCR's
// armored enforcer 321 and justice 322, the GUARD each enforcer has).
var lawMonsters = map[int]bool{13: true, 18: true, 321: true, 322: true}

// crimeNames are the codex's classes of crime.
var crimeNames = []string{"",
	"Petty Theft", "Grand Theft", "Minor Dark Necromancy", "Fraud", "Breaking and Entering",
	"Assault, Heresy or Kidnapping", "Major Dark Necromancy, Aiding or Abetting, or Grand Theft II",
	"Murder", "Treason"}

// romanNumeral is I to IX.
func romanNumeral(n int) string {
	return [...]string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX"}[n]
}

// warrantFine is the fine, in copper, for a crime of the given class: 10 gold crowns
// times the class squared, from 10 gold for petty theft to 810 for treason.
func warrantFine(level int) int { return level * level * 10 * 100 }

// lawReaches reports whether a room is within Fayd's walls, where the law applies.
func lawReaches(room int) bool { return room >= lawFirstCityRoom && room <= lawLastCityRoom }

// isLawkeeper reports whether the player belongs to the Lawkeepers.
func isLawkeeper(p *Player) bool { return p.IsGM || p.IsMemberOf(lawkeeperOrg) }

// wanted reports whether the law is after the player.
func wanted(p *Player) bool { return p.Warrant > 0 && !p.Jailed && !p.Dead && !p.IsGM }

// recognizable reports whether a lawkeeper NPC knows the player on sight: not hidden,
// disguised or in another shape (NOWAY.SCR's warrants, too, are only for the undisguised).
func recognizable(p *Player) bool {
	return !p.IsConcealed() && !p.Disguised && !p.WolfForm && !p.MistForm && !p.SlimeForm
}

// isLawMonsterName reports whether a killer's name is one of the law's, so a wanted
// player it brings down is taken alive (resolveDirectHitOutcome).
func (e *GameEngine) isLawMonsterName(name string) bool {
	for n := range lawMonsters {
		if def := e.monsters[n]; def != nil && strings.EqualFold(FormatMonsterName(def, e.monAdjs), name) {
			return true
		}
	}
	return false
}

// takenAlive is resolveDirectHitOutcome for a wanted player the law has felled: they're
// knocked out instead of killed, and hauled off to the Lockup by the next lawkeeper to
// act (lawMonsterAct).
func (e *GameEngine) takenAlive(p *Player, killerName string) bool {
	if !wanted(p) || !lawReaches(p.RoomNumber) || !e.isLawMonsterName(killerName) {
		return false
	}
	p.BodyPoints = 0
	p.Unconscious = true
	p.Position = 2
	return true
}

// doWarrant handles WARRANT [<player> <class> | <player> 0].
func (e *GameEngine) doWarrant(ctx context.Context, player *Player, args []string) *CommandResult {
	if !isLawkeeper(player) {
		return &CommandResult{Messages: []string{"Only the Lawkeepers may issue warrants."}}
	}
	if len(args) == 0 {
		return &CommandResult{Messages: e.warrantList()}
	}
	if !player.IsGM && player.RankIn(lawkeeperOrg) < warrantMinRank {
		return &CommandResult{Messages: []string{"Only the highest ranking Lawkeepers may issue warrants."}}
	}
	if len(args) < 2 {
		return &CommandResult{Messages: []string{"Usage: WARRANT <player> <crime class 1-9>, or WARRANT <player> 0 to lift one. WARRANT alone lists the classes and who is wanted."}}
	}
	level, err := strconv.Atoi(args[len(args)-1])
	if err != nil || level < 0 || level > maxWarrantLevel {
		return &CommandResult{Messages: []string{"The class of the crime must be 1 to 9 (0 lifts a warrant)."}}
	}
	target, err := e.resolvePlayerByNameLive(ctx, strings.Join(args[:len(args)-1], " "))
	if err != nil || target == nil {
		return &CommandResult{Messages: []string{"There is no one by that name."}}
	}
	if target.IsGM {
		return &CommandResult{Messages: []string{fmt.Sprintf("%s is beyond the reach of Fayd's law.", target.FirstName)}}
	}
	name := target.FullName()
	if level == 0 {
		msg := fmt.Sprintf("You lift the warrant for %s.", name)
		lifted := target.Warrant > 0 || target.Jailed
		target.Warrant = 0
		if target.Jailed {
			msg = fmt.Sprintf("You order %s released from the Lockup.", name)
			e.releaseFromJail(ctx, target, "By order of the Lawkeepers, you are released.")
		}
		if !lifted {
			return &CommandResult{Messages: []string{fmt.Sprintf("There is no warrant for %s.", name)}}
		}
		e.SavePlayer(ctx, target)
		e.Events.Publish("law", fmt.Sprintf("%s lifted the warrant for %s", player.FirstName, target.FirstName))
		return &CommandResult{Messages: []string{msg}}
	}
	target.Warrant = level
	e.SavePlayer(ctx, target)
	e.Events.Publish("law", fmt.Sprintf("%s issued a class %s warrant (%s) for %s", player.FirstName, romanNumeral(level), crimeNames[level], target.FirstName))
	e.sendToPlayerMsgs(target, []string{fmt.Sprintf("A warrant has been issued for your arrest, for %s (class %s). SUBMIT to a lawkeeper and come along quietly.", crimeNames[level], romanNumeral(level))})
	msgs := []string{fmt.Sprintf("You issue a class %s warrant for %s: %s. The fine is %s.", romanNumeral(level), name, crimeNames[level], formatCurrency(nil, warrantFine(level)))}
	if level >= 3 {
		msgs = append(msgs, "Remember: the codex requires two witnesses besides the victim for a crime of class III or above.")
	}
	return &CommandResult{Messages: msgs}
}

// warrantList is WARRANT with no arguments: the classes of crime, and who is wanted.
func (e *GameEngine) warrantList() []string {
	msgs := []string{"The Crimes Punishable In Fayd By Warrant Number:"}
	for i := 1; i <= maxWarrantLevel; i++ {
		msgs = append(msgs, fmt.Sprintf("  %-4s %-62s fine %s", romanNumeral(i)+":", crimeNames[i], formatCurrency(nil, warrantFine(i))))
	}
	var lines []string
	if e.sessions != nil {
		for _, p := range e.sessions.OnlinePlayers() {
			switch {
			case p.Jailed:
				lines = append(lines, fmt.Sprintf("  %s, in the Lockup (fine %s)", p.FullName(), formatCurrency(nil, p.JailFine)))
			case p.Warrant > 0:
				lines = append(lines, fmt.Sprintf("  %s, class %s: %s", p.FullName(), romanNumeral(p.Warrant), crimeNames[p.Warrant]))
			}
		}
	}
	sort.Strings(lines)
	if len(lines) == 0 {
		return append(msgs, "", "No one in the Realms is wanted right now.")
	}
	return append(append(msgs, "", "Wanted, of those in the Realms:"), lines...)
}

// doArrest handles ARREST <player>.
func (e *GameEngine) doArrest(ctx context.Context, player *Player, args []string) *CommandResult {
	if !isLawkeeper(player) {
		return &CommandResult{Messages: []string{"Only the Lawkeepers and their deputies may arrest anyone."}}
	}
	if len(args) == 0 {
		return &CommandResult{Messages: []string{"Arrest whom?"}}
	}
	if !player.IsGM && !lawReaches(player.RoomNumber) {
		return &CommandResult{Messages: []string{"Your authority doesn't reach beyond the walls of Fayd."}}
	}
	name := strings.ToLower(strings.Join(args, " "))
	target := e.findPlayerInRoom(player, name)
	if target == nil {
		if inst, def := e.findMonsterInRoom(player, name); inst != nil {
			return &CommandResult{Messages: []string{fmt.Sprintf("There is no warrant for %s.", articleFor(FormatMonsterName(def, e.monAdjs), def.Unique)+FormatMonsterName(def, e.monAdjs))}}
		}
		return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here.", strings.Join(args, " "))}}
	}
	if !wanted(target) {
		return &CommandResult{Messages: []string{fmt.Sprintf("There is no warrant for %s.", target.DisplayName())}}
	}
	if target.Submitting {
		msg := fmt.Sprintf("You place %s under arrest and lead %s away to the Lockup.", target.DisplayName(), target.Objective())
		e.jailPlayer(ctx, target, player)
		return &CommandResult{Messages: []string{msg}}
	}
	target.ArrestGraceUntil = time.Now().Add(arrestGrace)
	e.sendToPlayerMsgs(target, []string{fmt.Sprintf("%s places you under arrest! SUBMIT and come along quietly, or be taken by force.", player.DisplayNameCap())})
	return &CommandResult{
		Messages:         []string{fmt.Sprintf("You place %s under arrest. %s may SUBMIT and come along quietly.", target.DisplayName(), target.PronounCap())},
		RoomBroadcast:    []string{fmt.Sprintf("%s places %s under arrest!", player.DisplayNameCap(), target.DisplayName())},
		BroadcastExclude: []string{target.FirstName},
	}
}

// surrender is SUBMIT for a wanted player with the law at hand — a lawkeeper who has
// arrested them, or a lawkeeper or enforcer in the room: they're taken to the Lockup.
// It reports whether that happened.
func (e *GameEngine) surrender(ctx context.Context, p *Player) bool {
	if !wanted(p) || !lawReaches(p.RoomNumber) {
		return false
	}
	law := spellActive(p.ArrestGraceUntil)
	if !law && e.monsterMgr != nil {
		for _, m := range e.monsterMgr.MonstersInRoom(p.RoomNumber) {
			law = law || lawMonsters[m.DefNumber]
		}
	}
	if !law {
		return false
	}
	e.jailPlayer(ctx, p, nil)
	return true
}

// jailPlayer takes a captured player to the Lockup, telling them and the rooms involved.
// by is the lawkeeper who took them in, nil for the lawkeeps. Must be called without
// monsterMgr.mu held.
func (e *GameEngine) jailPlayer(ctx context.Context, p *Player, by *Player) {
	from := p.RoomNumber
	name := p.DisplayNameCap()
	level := max(p.Warrant, 1)

	e.breakAllMeleeEngagements(p)
	e.breakCarryAsCarried(ctx, p, "")
	e.breakCarryAsCarrier(ctx, p)
	e.removeFromGroup(p)
	e.dismissSummonedCreature(p)
	e.releaseMaelstrom(p.FirstName)
	clearStalk(p)
	p.breakGuard()
	p.Submitting = false
	p.ArrestGraceUntil = time.Time{}
	p.PreparedSpell, p.PreparedPsi = 0, 0
	p.Entangles, p.ImprisonedExpiry, p.ImmobilizedExpiry = nil, time.Time{}, time.Time{}
	p.MistForm, p.SlimeForm, p.WolfForm, p.Hidden, p.Invisible, p.PhantomForm = false, false, false, false, false, false
	p.Stunned, p.KnockedDown, p.SleepExpiry, p.Position = false, false, time.Time{}, 0
	if p.Unconscious || p.BodyPoints <= 0 {
		p.Unconscious = false
		p.BodyPoints = max(1, p.MaxBodyPoints/10)
	}
	for i := range p.Wounds { // the lawkeeps bind their wounds
		p.Wounds[i].Bleeding = false
	}
	p.Bleeding = false

	p.Jailed = true
	p.JailFine = warrantFine(level)
	p.JailUntil = time.Now().Add(time.Duration(level) * sentencePerClass)
	p.RoomNumber = lockupRoom

	// LAWKEEP.SCR's and the arrest macro's (3SYSOP.SCR) own words.
	if by != nil {
		if e.localRoomBroadcastExclude != nil {
			e.localRoomBroadcastExclude(from, by.FirstName, []string{fmt.Sprintf("%s leads %s away to jail.", by.DisplayNameCap(), p.DisplayName())})
		}
	} else if e.localRoomBroadcast != nil {
		e.localRoomBroadcast(from, []string{fmt.Sprintf("%s is hauled off to jail.", name)})
	}

	result := &CommandResult{Messages: []string{"The lawkeeps search you for weapons and haul you off to jail."}}
	look := e.doLook(p)
	result.Messages = append(result.Messages, look.Messages...)
	if room := e.rooms[lockupRoom]; room != nil {
		e.applyEntryScripts(ctx, p, room, result) // its IFENTRY: the sentence, and the warrant cleared
	}
	result.Messages = append(result.Messages, fmt.Sprintf("You will be let out in %s, or sooner if you PAY your fine of %s.", sentenceText(time.Until(p.JailUntil)), formatCurrency(nil, p.JailFine)))
	p.Warrant = 0
	e.SavePlayer(ctx, p)
	e.sendToPlayerMsgs(p, result.Messages)
	if e.localRoomBroadcast != nil {
		e.localRoomBroadcast(lockupRoom, append(result.RoomBroadcast, fmt.Sprintf("%s is tossed in by three of the meanest lawkeeps you've ever seen.", name)))
	}
	e.Events.Publish("law", fmt.Sprintf("%s was jailed (class %s, fine %d copper)", p.FirstName, romanNumeral(level), p.JailFine))
}

// releaseFromJail lets a prisoner out of the Lockup into the Lawkeeper HQ.
func (e *GameEngine) releaseFromJail(ctx context.Context, p *Player, why string) {
	p.Jailed = false
	p.JailFine = 0
	p.JailUntil = time.Time{}
	if p.RoomNumber != lockupRoom {
		e.SavePlayer(ctx, p)
		return
	}
	p.RoundTimeExpiry = time.Time{}
	p.RoomNumber = lawkeeperHQ
	result := &CommandResult{Messages: []string{why, "The iron door swings open and you step out of the Lockup."}}
	look := e.doLook(p)
	result.Messages = append(result.Messages, look.Messages...)
	if room := e.rooms[lawkeeperHQ]; room != nil {
		e.applyEntryScripts(ctx, p, room, result)
	}
	e.SavePlayer(ctx, p)
	e.sendToPlayerMsgs(p, result.Messages)
	if e.localRoomBroadcast != nil {
		e.localRoomBroadcast(lawkeeperHQ, []string{fmt.Sprintf("The iron door swings open and %s steps out of the Lockup.", p.DisplayName())})
	}
	e.Events.Publish("law", fmt.Sprintf("%s was released from jail", p.FirstName))
}

// payFine handles PAY in the Lockup, and PAY <name> in the Lawkeeper HQ. It returns nil
// for any other PAY.
func (e *GameEngine) payFine(ctx context.Context, player *Player, args []string) *CommandResult {
	prisoner := player
	switch {
	case player.Jailed && player.RoomNumber == lockupRoom:
	case player.RoomNumber == lawkeeperHQ && len(args) > 0:
		prisoner = e.findOnlinePlayerByPrefix(strings.Join(args, " "))
		if prisoner == nil || !prisoner.Jailed {
			return &CommandResult{Messages: []string{"There is no one by that name in the Lockup."}}
		}
	default:
		return nil
	}
	fine := prisoner.JailFine
	if !e.spendFunds(player, nil, fine) {
		return &CommandResult{Messages: []string{fmt.Sprintf("The fine is %s, and you don't have that much on you.", formatCurrency(nil, fine))}}
	}
	if prisoner == player {
		e.releaseFromJail(ctx, player, fmt.Sprintf("You pay your fine of %s.", formatCurrency(nil, fine)))
		return &CommandResult{PlayerState: player}
	}
	e.SavePlayer(ctx, player)
	e.releaseFromJail(ctx, prisoner, fmt.Sprintf("%s has paid your fine.", player.FirstName))
	return &CommandResult{
		Messages:      []string{fmt.Sprintf("You pay %s's fine of %s.", prisoner.FirstName, formatCurrency(nil, fine))},
		RoomBroadcast: []string{fmt.Sprintf("%s pays %s's fine.", player.DisplayNameCap(), prisoner.FirstName)},
		PlayerState:   player,
	}
}

// jailRefusal stops a prisoner from doing what the Lockup doesn't allow, or returns "".
// The room is OUTOFBODY (GMSCRIPT.DOC: "the player has no apparent skills, spells,
// psionic disciplines or inventory"), and only the fine opens its door.
func jailRefusal(p *Player, verb string) string {
	if !p.Jailed || p.RoomNumber != lockupRoom || p.IsGM {
		return ""
	}
	switch verb {
	case "CAST", "PREPARE", "INVOKE", "CHANT", "PSI", "PROJECT", "RUB", "CONCENTRATE", "ANSWER", "TRANSFORM", "DROP", "GIVE":
		return "The lawkeeps took everything of use from you, and the Lockup's wards smother any power you have."
	case "GO", "CLIMB", "OPEN", "UNLOCK", "PICK", "BREAK", "FLEE", "OUT", "O", "SNEAK":
		return "The stout iron door is shut fast, and there's no lock on this side. Serve your time, or PAY your fine."
	}
	return ""
}

// keepInJail returns a prisoner who has somehow got out of the Lockup without being
// released (a script, a spell cast at them) to it. It reports whether it did.
func (e *GameEngine) keepInJail(ctx context.Context, p *Player) bool {
	if !p.Jailed || p.RoomNumber == lockupRoom || p.IsGM || p.Dead {
		return false
	}
	p.RoomNumber = lockupRoom
	e.SavePlayer(ctx, p)
	look := e.doLook(p)
	e.sendToPlayerMsgs(p, append([]string{"The Lockup's wards drag you back into your cell!"}, look.Messages...))
	return true
}

// lawMonsterAct is a lawkeeper's or enforcer's turn against a wanted player: one who
// has submitted, or lies helpless, is hauled off to the Lockup. Called from
// monsterCombatTick with monsterMgr.mu held; reports whether it took its turn.
func (e *GameEngine) lawMonsterAct(inst *MonsterInstance, target *Player) bool {
	if !lawMonsters[inst.DefNumber] || !wanted(target) || !lawReaches(target.RoomNumber) {
		return false
	}
	if !target.Submitting && !target.Unconscious {
		return false
	}
	inst.Target = ""
	e.monsterMgr.mu.Unlock()
	e.jailPlayer(context.Background(), target, nil)
	e.monsterMgr.mu.Lock()
	return true
}

// lawSpotsWanted is a lawkeeper or enforcer noticing a wanted player it recognizes, in
// monsterTick: it advances on them as a hostile monster would. Called with
// monsterMgr.mu held; reports whether it found one.
func (e *GameEngine) lawSpotsWanted(inst *MonsterInstance, def *gameworld.MonsterDef) bool {
	if !lawMonsters[inst.DefNumber] || inst.Target != "" || inst.IsSummoned || e.sessions == nil || !lawReaches(inst.RoomNumber) {
		return false
	}
	for _, p := range e.sessions.OnlinePlayers() {
		if p.RoomNumber != inst.RoomNumber || !wanted(p) || !recognizable(p) || p.IsImprisoned() || spellActive(p.ArrestGraceUntil) {
			continue
		}
		name := FormatMonsterName(def, e.monAdjs)
		capName := capArticle(articleFor(name, def.Unique)) + name
		inst.Target = p.FirstName
		if p.Submitting || p.Unconscious {
			return true // taken in on its next turn
		}
		e.sendToPlayerMsgs(p, []string{fmt.Sprintf("%s recognizes you and advances towards you, shouting, \"Submit, in the name of the law!\"", capName)})
		e.broadcastCombatRoom(inst.RoomNumber, p, []string{fmt.Sprintf("%s recognizes %s and advances, shouting, \"Submit, in the name of the law!\"", capName, p.DisplayName())})
		return true
	}
	return false
}

// lawStatusLine is the STATUS line for a wanted or jailed player, or "".
func lawStatusLine(p *Player) string {
	switch {
	case p.Jailed:
		return fmt.Sprintf("You are in the Lockup, to be let out in %s or when you pay your fine of %s.", sentenceText(time.Until(p.JailUntil)), formatCurrency(nil, p.JailFine))
	case p.Warrant > 0:
		return fmt.Sprintf("There is a warrant for your arrest: %s (class %s).", crimeNames[p.Warrant], romanNumeral(p.Warrant))
	}
	return ""
}

// sentenceText is "2 minutes", "1 minute 20 seconds", "40 seconds".
func sentenceText(d time.Duration) string {
	secs := max(int((d+time.Second-1)/time.Second), 1)
	m, s := secs/60, secs%60
	plural := func(n int, unit string) string {
		if n == 1 {
			return "1 " + unit
		}
		return fmt.Sprintf("%d %ss", n, unit)
	}
	switch {
	case m == 0:
		return plural(s, "second")
	case s == 0:
		return plural(m, "minute")
	}
	return plural(m, "minute") + " " + plural(s, "second")
}

// sentenceServed lets out a prisoner whose time is up, from spellEffectTick.
func (e *GameEngine) sentenceServed(p *Player) {
	if p.Jailed && !p.JailUntil.IsZero() && time.Now().After(p.JailUntil) {
		e.releaseFromJail(context.Background(), p, "You have served your sentence.")
	}
}
