package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// New player tutorial: Sergeant Aldren Fairweather of the Fayd gate watch walks
// new characters through the basic commands in the room they first enter
// (201, Main Street, Gatehouse). Design: "LoFP New Player tutorial.docx".
//
//   - Starts automatically for a level 1 character entering the gate (not bots
//     or GMs). Each step waits for the player to use a command, checked
//     against the player's state where it matters (breeches actually worn, a
//     weapon actually wielded, the coin actually picked up), then Aldren speaks
//     the next passage.
//   - Never restrictive: the player can walk away at any time. The lesson
//     pauses while they're gone and picks up when they return. Progress is
//     saved on the player, so it also resumes after a disconnect.
//   - SKIP TUTORIAL ends it. SAY TRAINING at the gate starts it over, for
//     anyone.
//   - After a few commands that don't complete the step, Aldren repeats what
//     he's waiting for.
//
// Aldren is not a monster: monsters can be attacked, killed, and unloaded when
// no one is near. He is drawn into room 201's LOOK output (doLook), answers
// LOOK ALDREN (doLookAt), and turns away attacks (tutorialPreCommand). His
// lines go only to the player taking the lesson, so a busy gate isn't flooded.
//
// All state is the player's own (saved with the player), so nothing here needs
// the hub, except the tutorial coin, which is a normal room item announced
// with notifyRoomChange.

const (
	tutorialRoom     = 201  // Main Street, Gatehouse: where new characters start
	tutorialCoinItem = 9500 // TUTORIAL.SCR: "a tarnished copper coin"
	tutorialCoinAdj1 = 318  // ADJDEF 318 tarnished
	tutorialCoinAdj2 = 68   // ADJDEF 68 copper
	starterDagger    = 12   // ITEM1.SCR dagger, issued to new characters

	tutorialNotStarted = 0
	tutorialActive     = 1
	tutorialFinished   = 2

	// tutorialHintAfter is how many commands that don't complete a step pass
	// before Aldren repeats what he's waiting for.
	tutorialHintAfter = 4
)

const aldrenShortDesc = "A weathered city guardsman stands near the gates, welcoming newly arrived travelers."

var aldrenLongDesc = []string{
	"Sergeant Aldren Fairweather stands beside the city gates, his polished halberd resting against one shoulder. His watchful eyes follow the steady procession of merchants, adventurers, and travelers entering Fayd. Despite his stern appearance, he offers newcomers a reassuring nod and the occasional piece of practical advice.",
	"He is a seasoned city watchman with a weathered face and a neatly trimmed beard, wearing a faded blue-and-silver tabard bearing the crest of Fayd.",
}

// aldrenSays formats one of Aldren's spoken lines.
func aldrenSays(line string) string {
	return fmt.Sprintf("Sergeant Aldren says, \"%s\"", line)
}

// tutorialCmd is a command as the tutorial sees it.
type tutorialCmd struct {
	verb     string // canonical verb (resolveVerb), "" for speech
	args     string // lowercased arguments, with a leading "at " removed
	speech   bool
	speechTx string // what was said, lowercased
}

func parseTutorialCmd(input string) tutorialCmd {
	input = strings.TrimSpace(input)
	if input == "" {
		return tutorialCmd{}
	}
	if strings.HasPrefix(input, "'") || strings.HasPrefix(input, "\"") {
		return tutorialCmd{speech: true, speechTx: strings.ToLower(strings.TrimSpace(input[1:]))}
	}
	fields := strings.Fields(input)
	rest := strings.ToLower(strings.TrimSpace(strings.Join(fields[1:], " ")))
	// Same test as ProcessCommand's: resolveVerb would turn SAY into SAYTO.
	if strings.EqualFold(fields[0], "SAY") {
		return tutorialCmd{speech: true, speechTx: rest}
	}
	verb := resolveVerb(strings.ToUpper(fields[0]))
	rest = strings.TrimPrefix(rest, "at ")
	return tutorialCmd{verb: verb, args: rest}
}

func (c tutorialCmd) isLook() bool { return c.verb == "LOOK" && c.args == "" }
func (c tutorialCmd) isExamine() bool {
	return c.verb == "LOOK" || c.verb == "EXAMINE" || c.verb == "INSPECT"
}

// isTrainingRequest: SAY TRAINING (also "training?", "training!").
func (c tutorialCmd) isTrainingRequest() bool {
	return c.speech && strings.Trim(c.speechTx, " .!?") == "training"
}

// targetsAldren reports whether a command argument names Aldren. "guard" is
// left out on purpose: it would steal LOOK GUARD from the gate's real guards.
func targetsAldren(arg string) bool {
	for _, w := range strings.Fields(arg) {
		switch w {
		case "aldren", "sergeant", "fairweather", "guardsman", "sarge":
			return true
		}
	}
	return false
}

func argMentions(arg, word string) bool {
	for _, w := range strings.Fields(arg) {
		if w == word {
			return true
		}
	}
	return false
}

type tutorialStep struct {
	// lines Aldren speaks when the step begins.
	lines []string
	// hint is what he says when the player seems stuck.
	hint string
	// begin runs as the step starts (e.g. putting the coin down).
	begin func(e *GameEngine, ctx context.Context, p *Player)
	// done reports whether this command completes the step; nil means the
	// step completes on its own (the farewell).
	done func(e *GameEngine, p *Player, c tutorialCmd) bool
}

// bothDone is for steps that want two commands in either order: it records
// which of them (bit a or bit b) this command was and reports when both are in.
func bothDone(p *Player, c tutorialCmd, verbA, verbB string) bool {
	switch c.verb {
	case verbA:
		p.TutorialFlags |= 1
	case verbB:
		p.TutorialFlags |= 2
	}
	return p.TutorialFlags == 3
}

var tutorialSteps = []tutorialStep{
	// 0 — Scene 1: welcome, LOOK
	{
		lines: []string{
			"As you step through the massive gates of Fayd, a grizzled city guardsman raises one hand, signaling for you to stop.",
			aldrenSays("Hail and well met, traveler! Welcome to the fair city of Fayd. Name's Sergeant Aldren Fairweather, and it's my duty to see that newcomers find their feet before they lose their purses, or their lives."),
			aldrenSays("First time through these gates, is it? Well, before you go wandering off, take a moment to have a proper look around [LOOK] or [L]. You'll find that keeping your eyes open is the first rule of surviving these lands."),
			"[You can leave the lesson at any time with SKIP TUTORIAL.]",
		},
		hint: "Have a proper look around first. Type [LOOK], or just [L].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool { return c.isLook() },
	},
	// 1 — LOOK ALDREN
	{
		lines: []string{
			aldrenSays("Good! Now, if something catches your eye, you can take a closer look at it. Try inspecting me [LOOK ALDREN]. I assure you, I'm considerably more handsome than the other guards."),
		},
		hint: "Take a closer look at me. Type [LOOK ALDREN].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool {
			return c.isExamine() && targetsAldren(c.args)
		},
	},
	// 2 — Scene 2: INVENTORY
	{
		lines: []string{
			"Aldren looks you over with a raised eyebrow.",
			aldrenSays("By the gods, traveler! Have you come all this way dressed like that? Surely you've got something decent in your belongings. Best check what you're carrying [INVENTORY] before you embarrass us both."),
		},
		hint: "Check what you're carrying. Type [INVENTORY], or just [I].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool { return c.verb == "INVENTORY" },
	},
	// 3 — WEAR BREECHES
	{
		lines: []string{
			aldrenSays("Ah, there we are! I suggest you put some breeches on [WEAR BREECHES]. Fayd is a respectable city, at least during daylight hours."),
		},
		hint: "Put your breeches on. Type [WEAR BREECHES]. If you've none, [WEAR] whatever clothing you do have.",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool {
			if c.verb != "WEAR" {
				return false
			}
			// With no breeches to wear, any WEAR will do.
			return e.tutorialWearsNoun(p, "breeches") || !e.tutorialCarriesNoun(p, "breeches")
		},
	},
	// 4 — Scene 3: WIELD DAGGER
	{
		lines: []string{
			aldrenSays("Much better! You can wear other clothing and armor the same way. If you need to take something off, use [REMOVE BREECHES], though I'd strongly advise against doing that here."),
			aldrenSays("And if you ever forget what you've got on, [INV] or [I] should remind you. Saves having to look down every few minutes."),
			"Aldren taps the haft of his halberd against the cobblestones.",
			aldrenSays("Now, the roads beyond Fayd and folk in the Poor Quarter aren't nearly as civilized as the streets within. If you've a weapon among your belongings, you'd best know how to use it. Draw and ready your weapon [WIELD DAGGER], assuming that's what you've been issued."),
		},
		hint: "Ready your weapon. Type [WIELD DAGGER], or WIELD whatever weapon you carry.",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool {
			if c.verb != "WIELD" {
				return false
			}
			// Unarmed: there's nothing to wield, so trying counts.
			return p.Wielded != nil || !e.tutorialCarriesWeapon(p)
		},
	},
	// 5 — STAT and FATIGUE
	{
		lines: []string{
			aldrenSays("Good. Remember, having a weapon in your pack isn't the same as having one in your hand."),
			aldrenSays("Now, tell me, are you strong enough to swing that thing? Quickest way to find out what sort of adventurer you are is to examine your abilities [STAT], and [FATIGUE] to find out what your fatigue, health, mana, and psi levels are. Strength, intelligence, and the rest of your talents all have their uses."),
		},
		hint: "Look over your abilities. Type [STAT], and then [FATIGUE].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool {
			return bothDone(p, c, "STATUS", "FATIGUE")
		},
	},
	// 6 — Scene 4: GET COIN
	{
		lines: []string{
			aldrenSays("And don't go believing that every problem can be solved with a sharp blade. Some of the most dangerous people I've met couldn't swing a sword to save their lives."),
			"Aldren notices a small copper coin lying near the gatehouse and gestures toward it.",
			aldrenSays("Well, would you look at that! Someone's dropped a copper. Finders keepers, as they say. Go on, pick it up [GET COIN]."),
		},
		hint:  "Pick up that coin by the gatehouse. Type [GET COIN].",
		begin: func(e *GameEngine, ctx context.Context, p *Player) { e.placeTutorialCoin(p) },
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool {
			return c.verb == "GET" && e.tutorialCarriesCoin(p)
		},
	},
	// 7 — EXAMINE COIN
	{
		lines: []string{
			aldrenSays("There you are! If you ever want to inspect something closely, try [EXAMINE COIN]. Sometimes an ordinary-looking object is worth more than you think."),
		},
		hint: "Have a close look at that coin. Type [EXAMINE COIN].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool {
			return c.isExamine() && argMentions(c.args, "coin")
		},
	},
	// 8 — DROP COIN
	{
		lines: []string{
			aldrenSays("You can also put things down [DROP COIN], though I wouldn't recommend making a habit of leaving valuables lying about. Fayd has its share of light-fingered citizens."),
		},
		hint: "Try putting the coin down. Type [DROP COIN].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool {
			return c.verb == "DROP" && argMentions(c.args, "coin") && !e.tutorialCarriesCoin(p)
		},
	},
	// 9 — Scene 5: SAY
	{
		lines: []string{
			"Aldren folds his arms and studies you for a moment.",
			aldrenSays("You've been awfully quiet, traveler. Lost your tongue on the road? Come now, let's hear your voice. Say something to me [SAY Greetings, Sergeant!]."),
		},
		hint: "Let's hear your voice. Type [SAY Greetings, Sergeant!]",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool { return c.speech },
	},
	// 10 — SMILE or BOW
	{
		lines: []string{
			aldrenSays("Ha! There you are! A little conversation goes a long way in this city. You can start a sentence with just a single quotation mark ['], too. You'll find merchants, adventurers, and all manner of interesting folk willing to speak with you."),
			aldrenSays("If you wish to know who else is about the realm, [WHO] will give you an idea. And should you need to speak privately with someone, [WHISPER PLAYER MESSAGE] is commonly used for that purpose. Replace PLAYER with their name, of course."),
			aldrenSays("You'll also find that a gesture sometimes says more than words. A friendly [SMILE] or [BOW] rarely goes amiss, unless you're smiling at someone who's just drawn a sword."),
		},
		hint: "Go on, give us a friendly gesture. Type [SMILE] or [BOW].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool {
			return c.verb == "SMILE" || c.verb == "BOW"
		},
	},
	// 11 — Scene 6: finding your way, LOOK
	{
		lines: []string{
			"Aldren gestures toward the busy streets stretching into Fayd.",
			aldrenSays("Now, Fayd is a fine city, but it's easy enough to lose your bearings."),
			aldrenSays("Most travelers move by direction. [NORTH], [SOUTH], [EAST], and [WEST] will take you where the roads allow. You'll also encounter paths [GO PATH], doors, entrances [GO ENT], arches, and crevices leading up, down, or into buildings or other spaces."),
			aldrenSays("You can check out the newsboard [READ NEWSBOARD] in the Bazaar Center a short ways west of us here to find out what's happening in the city or who's looking for workers."),
			aldrenSays("If you lose track of your surroundings, [LOOK] will remind you where you stand. A simple habit, but one that might save your life someday. Go on, humor me."),
		},
		hint: "Remind yourself where you stand. Type [LOOK].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool { return c.isLook() },
	},
	// 12 — Scene 7: SKILLS and EXP
	{
		lines: []string{
			"Aldren's expression grows a little more serious.",
			aldrenSays("One last bit of advice. There's more to surviving these lands than wandering about and hoping for the best."),
			aldrenSays("Take stock of your talents [SKILLS] and mark your progress towards your next build point [EXP]. A clever adventurer learns what they're capable of before testing those abilities against something with teeth."),
		},
		hint: "Take stock of yourself. Type [SKILLS], and then [EXP].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool {
			return bothDone(p, c, "SKILLS", "EXPERIENCE")
		},
	},
	// 13 — HELP
	{
		lines: []string{
			aldrenSays("And if you find yourself confused, there's no shame in asking for guidance. The realm has a wealth of information available [HELP]. If you're curious about something particular, you can often ask directly [HELP COMBAT], for example."),
		},
		hint: "Have a look at the help that's available. Type [HELP].",
		done: func(e *GameEngine, p *Player, c tutorialCmd) bool { return c.verb == "HELP" },
	},
	// 14 — FLEE, QUIT, Scene 8: farewell. Completes on its own.
	{
		lines: []string{
			aldrenSays("Should you find yourself in danger, remember that retreat is sometimes wiser than fighting. Learn how to escape a losing battle [FLEE] before you need to."),
			aldrenSays("And don't forget to learn how to leave the realm safely [QUIT]. No sense surviving a dragon only to lose your belongings through carelessness."),
			"Aldren steps aside, allowing you a clear view of Fayd's bustling streets.",
			aldrenSays("Well, traveler, I reckon you're ready to find your own way. There's work to be had, coin to be earned, and plenty of trouble waiting for anyone foolish enough to go looking for it."),
			aldrenSays("Speak with the townsfolk, learn the lay of the land, and don't be afraid to ask questions. Every great adventurer started somewhere."),
			aldrenSays("If you forget what I've taught you, come back and ask me about [SAY TRAINING]. I'll gladly run through the basics again. Also, you can abbreviate many of the commands I've spoken about. You'll have to try them out yourself to see how short you can make them."),
			aldrenSays("You can talk to Tad, that loitering vagabond in the Center of the Bazaar, if you want to know more about the city. He's smarter than he looks. And Tanner in the Adventurer's Guild is always looking for some basic work to be done in the city for coin and experience."),
			aldrenSays("Now off with you! Welcome to Fayd, and may fortune favor your travels!"),
			"Aldren offers a crisp salute before returning his attention to the steady stream of travelers entering the city.",
		},
	},
}

// tutorialEligible: who gets the lesson without asking for it.
func tutorialEligible(p *Player) bool {
	return p.Level == 1 && !p.IsBot && !p.IsGM
}

// startTutorial (re)starts the lesson at step 0 and returns the opening lines.
func (e *GameEngine) startTutorial(ctx context.Context, p *Player) []string {
	p.TutorialStatus = tutorialActive
	p.TutorialStep = 0
	p.TutorialFlags = 0
	p.TutorialMisses = 0
	p.TutorialPrompted = true
	p.TutorialAway = false
	e.SavePlayer(ctx, p)
	return e.beginTutorialStep(ctx, p)
}

// beginTutorialStep runs the current step's setup and returns its lines. The
// last step has no command to wait for, so the lesson finishes here.
func (e *GameEngine) beginTutorialStep(ctx context.Context, p *Player) []string {
	if p.TutorialStep < 0 || p.TutorialStep >= len(tutorialSteps) {
		p.TutorialStatus = tutorialFinished
		return nil
	}
	step := tutorialSteps[p.TutorialStep]
	if step.begin != nil {
		step.begin(e, ctx, p)
	}
	lines := append([]string(nil), step.lines...)
	if step.done == nil {
		p.TutorialStatus = tutorialFinished
		p.TutorialStep = 0
		p.TutorialFlags = 0
	}
	return lines
}

// tutorialResumeLines: what Aldren says when a player comes back to the gate,
// or reconnects, in the middle of the lesson.
func (e *GameEngine) tutorialResumeLines(ctx context.Context, p *Player) []string {
	if p.TutorialStep < 0 || p.TutorialStep >= len(tutorialSteps) {
		p.TutorialStatus = tutorialFinished
		return nil
	}
	step := tutorialSteps[p.TutorialStep]
	if step.begin != nil {
		step.begin(e, ctx, p) // e.g. the coin may be gone by now
	}
	lines := []string{"Sergeant Aldren spots you and nods."}
	if p.TutorialStep == 0 {
		return append(lines, step.lines[1:]...)
	}
	return append(lines,
		aldrenSays("Back again, traveler? Let's pick up where we left off."),
		aldrenSays(step.hint))
}

// tutorialOnEnter runs when a player enters a room, from applyEntryScripts.
// scriptDelayed says the room's IFENTRY left delayed (PLREVENT) messages, such
// as room 201's new-character clothing, so Aldren waits for them to go out.
func (e *GameEngine) tutorialOnEnter(ctx context.Context, p *Player, room *gameworld.Room, result *CommandResult, scriptDelayed bool) {
	if p.TutorialStatus == tutorialFinished {
		return
	}
	if p.RoomNumber != room.Number {
		return // an entry script already moved them on
	}
	if room.Number != tutorialRoom {
		if p.TutorialStatus == tutorialActive {
			p.TutorialAway = true
		}
		return
	}
	var lines []string
	switch {
	case p.TutorialStatus == tutorialNotStarted:
		if !tutorialEligible(p) {
			return
		}
		lines = e.startTutorial(ctx, p)
	case !p.TutorialPrompted || p.TutorialAway:
		p.TutorialPrompted = true
		p.TutorialAway = false
		lines = e.tutorialResumeLines(ctx, p)
	default:
		return
	}
	e.SavePlayer(ctx, p)
	if len(lines) == 0 {
		return
	}
	if e.sendToPlayer == nil {
		result.Messages = append(result.Messages, lines...)
		return
	}
	// Deliver after the room look and entry messages rather than in the middle
	// of them. If the player enters another command first, it delivers them at
	// the top of its output instead (ProcessCommand, tutorialAfterCommand).
	delay := 1500 * time.Millisecond
	if scriptDelayed {
		delay = 5 * time.Second
	}
	tutorialMu.Lock()
	p.TutorialPending = lines
	tutorialMu.Unlock()
	name := p.FirstName
	time.AfterFunc(delay, func() {
		if pending := takeTutorialPending(p); len(pending) > 0 {
			e.sendToPlayer(name, pending)
		}
	})
}

// tutorialMu guards Player.TutorialPending, which a timer and the player's
// own commands both take from.
var tutorialMu sync.Mutex

func takeTutorialPending(p *Player) []string {
	tutorialMu.Lock()
	defer tutorialMu.Unlock()
	lines := p.TutorialPending
	p.TutorialPending = nil
	return lines
}

// tutorialPreCommand handles commands that belong to the tutorial before the
// normal command dispatch sees them. A non-nil result is the command's output.
func (e *GameEngine) tutorialPreCommand(ctx context.Context, p *Player, input string) *CommandResult {
	fields := strings.Fields(strings.ToUpper(input))
	if len(fields) == 0 {
		return nil
	}
	// SKIP TUTORIAL (SKIP isn't otherwise a verb).
	if len(fields) == 2 && fields[0] == "SKIP" && strings.HasPrefix("TUTORIAL", fields[1]) && len(fields[1]) >= 3 {
		if p.TutorialStatus != tutorialActive {
			return &CommandResult{Messages: []string{"You aren't taking the tutorial. Sergeant Aldren at the city gates will run through the basics with you if you SAY TRAINING there."}}
		}
		p.TutorialStatus = tutorialFinished
		p.TutorialStep = 0
		p.TutorialFlags = 0
		e.SavePlayer(ctx, p)
		msgs := []string{"You skip the tutorial."}
		if p.RoomNumber == tutorialRoom {
			msgs = append(msgs, aldrenSays("Suit yourself, traveler. If you ever want a refresher, come back and ask me about [SAY TRAINING]."))
		} else {
			msgs = append(msgs, "If you ever want a refresher, Sergeant Aldren at the city gates will run through the basics again if you SAY TRAINING there.")
		}
		return &CommandResult{Messages: msgs}
	}
	if p.RoomNumber != tutorialRoom || len(fields) < 2 {
		return nil
	}
	// Aldren can't be fought or robbed.
	switch resolveVerb(fields[0]) {
	case "ATTACK", "KILL", "FEINT", "STEAL", "ROB", "PICKPOCKET", "AMBUSH", "BACKSTAB", "PUNCH", "KICK":
		if targetsAldren(strings.ToLower(strings.Join(fields[1:], " "))) {
			return &CommandResult{Messages: []string{
				"Sergeant Aldren steps back, his halberd sweeping down to bar your way.",
				aldrenSays("I'd think very carefully about that, traveler. Very carefully."),
			}}
		}
	}
	return nil
}

// tutorialAfterCommand runs once a command has been handled: SAY TRAINING
// (re)starts the lesson, otherwise a command that completes the current step
// advances it and Aldren speaks the next passage.
//
// carried holds Aldren's entry lines that were still waiting on their delay
// when this command started (tutorialOnEnter); they go out first. Lines queued
// by this command's own room entry are left to their timer, so they follow the
// room description.
func (e *GameEngine) tutorialAfterCommand(ctx context.Context, p *Player, input string, result *CommandResult, carried []string) {
	if result == nil {
		return
	}
	if len(carried) > 0 {
		result.Messages = append(carried, result.Messages...)
	}
	if p.RoomNumber != tutorialRoom || p.Dead || p.Unconscious {
		return
	}
	c := parseTutorialCmd(input)
	if c.isTrainingRequest() {
		var msgs []string
		if p.TutorialStatus == tutorialFinished || p.Level > 1 {
			msgs = append(msgs,
				"Sergeant Aldren grins and plants the butt of his halberd on the cobblestones.",
				aldrenSays("Back for a refresher, are we? No shame in that. The best soldiers I know drill the basics every morning. Let's start from the top."))
		}
		msgs = append(msgs, e.startTutorial(ctx, p)[1:]...) // skip the "as you step through the gates" line
		result.Messages = append(result.Messages, msgs...)
		return
	}
	if p.TutorialStatus != tutorialActive || p.TutorialStep < 0 || p.TutorialStep >= len(tutorialSteps) {
		return
	}
	p.TutorialAway = false
	step := tutorialSteps[p.TutorialStep]
	if step.done == nil {
		return
	}
	if !step.done(e, p, c) {
		p.TutorialMisses++
		if p.TutorialMisses == tutorialHintAfter && step.hint != "" {
			result.Messages = append(result.Messages, aldrenSays(step.hint))
		}
		return
	}
	p.TutorialStep++
	p.TutorialFlags = 0
	p.TutorialMisses = 0
	p.TutorialPrompted = true
	result.Messages = append(result.Messages, e.beginTutorialStep(ctx, p)...)
	e.SavePlayer(ctx, p)
}

// aldrenLookLines adds Aldren to a LOOK in the gate room.
func aldrenLookLines(roomNum int) []string {
	if roomNum != tutorialRoom {
		return nil
	}
	return []string{aldrenShortDesc}
}

// lookAtAldren answers LOOK/EXAMINE ALDREN in the gate room.
func lookAtAldren(p *Player, target string) *CommandResult {
	if p.RoomNumber != tutorialRoom || !targetsAldren(target) {
		return nil
	}
	return &CommandResult{Messages: append([]string(nil), aldrenLongDesc...)}
}

// --- state checks ---

func (e *GameEngine) itemNoun(arch int) string {
	if def := e.items[arch]; def != nil {
		return strings.ToLower(e.nouns[def.NameID])
	}
	return ""
}

func (e *GameEngine) tutorialWearsNoun(p *Player, noun string) bool {
	for _, ii := range p.Worn {
		if e.itemNoun(ii.Archetype) == noun {
			return true
		}
	}
	return false
}

func (e *GameEngine) tutorialCarriesNoun(p *Player, noun string) bool {
	if e.tutorialWearsNoun(p, noun) {
		return true
	}
	for _, ii := range p.Inventory {
		if e.itemNoun(ii.Archetype) == noun {
			return true
		}
	}
	return false
}

func (e *GameEngine) tutorialCarriesWeapon(p *Player) bool {
	if p.Wielded != nil {
		return true
	}
	for _, ii := range p.Inventory {
		if def := e.items[ii.Archetype]; def != nil && isWeapon(def.Type) {
			return true
		}
	}
	return false
}

func (e *GameEngine) tutorialCarriesCoin(p *Player) bool {
	for _, ii := range p.Inventory {
		if ii.Archetype == tutorialCoinItem {
			return true
		}
	}
	return p.Wielded != nil && p.Wielded.Archetype == tutorialCoinItem
}

// placeTutorialCoin puts a tutorial coin by the gatehouse unless one is
// already lying there or the player already has one.
func (e *GameEngine) placeTutorialCoin(p *Player) {
	room := e.rooms[tutorialRoom]
	if room == nil || e.items[tutorialCoinItem] == nil || e.tutorialCarriesCoin(p) {
		return
	}
	for _, ri := range room.Items {
		if ri.Archetype == tutorialCoinItem {
			return
		}
	}
	coin := gameworld.RoomItem{
		Ref:       nextRoomItemRef(room),
		Archetype: tutorialCoinItem,
		Adj1:      tutorialCoinAdj1,
		Adj2:      tutorialCoinAdj2,
	}
	room.Items = append(room.Items, coin)
	e.notifyRoomChange(RoomChange{RoomNumber: tutorialRoom, Type: "item_add", Item: &coin})
}
