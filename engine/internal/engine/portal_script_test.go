package engine

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
	"github.com/jonradoff/lofp/internal/scriptparser"
)

// TestGoPortalScriptsBeforeClosedCheck: IFPREVERB GO runs before the closed
// check, so STEEL.SCR's automatic door (3502 -> 3501) opens for the player.
func TestGoPortalScriptsBeforeClosedCheck(t *testing.T) {
	r, err := scriptparser.ParseConfig("../../../original/scripts/LEGENDS.CFG")
	if err != nil {
		t.Fatalf("parse scripts: %v", err)
	}
	e := NewGameEngine(nil, &gameworld.ParsedData{
		Rooms: r.Rooms, Items: r.Items, Monsters: r.Monsters, Nouns: r.Nouns,
		Adjectives: r.Adjectives, MonsterAdjs: r.MonsterAdjs, Variables: r.Variables,
		Regions: r.Regions, Macros: r.Macros, OrgDefs: r.OrgDefs,
		StartRoom: r.StartRoom, BumpRoom: r.BumpRoom,
	})
	ctx := context.Background()
	p := e.CreateNewPlayer(ctx, "Testy", "Door", RaceHuman, GenderMale, nil)
	broadcasts := map[int]string{}
	e.SetRoomBroadcast(func(room int, msgs []string) { broadcasts[room] += strings.Join(msgs, "\n") + "\n" })
	door := func(room, ref int) *gameworld.RoomItem {
		for i := range e.rooms[room].Items {
			if e.rooms[room].Items[i].Ref == ref {
				return &e.rooms[room].Items[i]
			}
		}
		t.Fatalf("room %d has no item %d", room, ref)
		return nil
	}
	const clank = "The door slides open with a loud"

	for _, state := range []string{"CLOSED", "LOCKED", "CLOSED"} {
		p.RoomNumber = 3502
		door(3502, 0).State = state
		broadcasts = map[int]string{}
		if out := strings.Join(e.ProcessCommand(ctx, p, "open door").Messages, "\n"); !strings.Contains(out, "automatic door") || p.RoomNumber != 3502 {
			t.Errorf("open door: room %d\n%s", p.RoomNumber, out)
		}
		res := e.ProcessCommand(ctx, p, "go door")
		out := strings.Join(res.Messages, "\n")
		if p.RoomNumber != 3501 {
			t.Fatalf("%s: go door left the player in %d:\n%s", state, p.RoomNumber, out)
		}
		if !strings.Contains(out, clank) || strings.Contains(out, "is closed") {
			t.Errorf("%s: go door output:\n%s", state, out)
		}
		if old := strings.Join(res.OldRoomMsg, "\n"); !strings.Contains(old, clank) {
			t.Errorf("bystanders in 3502 didn't hear the door: %q", old)
		}
		if in := strings.Join(res.RoomBroadcast, "\n"); strings.Contains(in, clank) {
			t.Errorf("3502's echo went to 3501: %q", in)
		}
		if !strings.Contains(broadcasts[3501], "just long enough to let someone in") {
			t.Errorf("3501 broadcast: %q", broadcasts[3501])
		}
		// 3501's IFENTRY shuts the door behind them.
		if s := door(3502, 0).State; s != "CLOSED" {
			t.Errorf("door left %q", s)
		}
	}

	// A closed door whose script doesn't open it still stops the player, once.
	p.RoomNumber = 1459
	door(1459, 0).State = "CLOSED"
	if out := strings.Join(e.ProcessCommand(ctx, p, "go door").Messages, "\n"); p.RoomNumber != 1459 || strings.Count(out, "The door is closed.") != 1 {
		t.Errorf("scripted closed door: room %d\n%s", p.RoomNumber, out)
	}
	// No script at all.
	p.RoomNumber = 3501
	door(3501, 0).State = "CLOSED"
	if out := strings.Join(e.ProcessCommand(ctx, p, "go door").Messages, "\n"); p.RoomNumber != 3501 || !strings.Contains(out, "is closed.") {
		t.Errorf("unscripted closed door: room %d\n%s", p.RoomNumber, out)
	}
}

type portalTestSessions []*Player

func (s portalTestSessions) OnlinePlayers() []*Player { return s }

// TestEphemeralSqueezesUnderClosedDoors: ITEM1.SCR's doors let an Ephemeral (and
// only an Ephemeral) past when shut, alone, unless sealed or a room script refuses.
func TestEphemeralSqueezesUnderClosedDoors(t *testing.T) {
	r, err := scriptparser.ParseConfig("../../../original/scripts/LEGENDS.CFG")
	if err != nil {
		t.Fatalf("parse scripts: %v", err)
	}
	e := NewGameEngine(nil, &gameworld.ParsedData{
		Rooms: r.Rooms, Items: r.Items, Monsters: r.Monsters, Nouns: r.Nouns,
		Adjectives: r.Adjectives, MonsterAdjs: r.MonsterAdjs, Variables: r.Variables,
		Regions: r.Regions, Macros: r.Macros, OrgDefs: r.OrgDefs,
		StartRoom: r.StartRoom, BumpRoom: r.BumpRoom,
	})
	ctx := context.Background()
	const squeeze = "squeeze your semi-solid self"
	player := func(name string, race int) *Player {
		return e.CreateNewPlayer(ctx, name, "Door", race, GenderMale, nil)
	}
	goDoor := func(p *Player, room, ref int, state string) (*CommandResult, string) {
		t.Helper()
		p.RoomNumber = room
		for i := range e.rooms[room].Items {
			ri := &e.rooms[room].Items[i]
			if def := e.items[ri.Archetype]; ri.Ref == ref && def != nil {
				ri.State = state
				res := e.doGoPortal(ctx, p, e.rooms[room], ri, def, "GO")
				return res, strings.Join(res.Messages, "\n")
			}
		}
		t.Fatalf("room %d has no item %d", room, ref)
		return nil, ""
	}
	// 1322 ref 0: an archetype 65 door to 1321, no room scripts.
	const room, ref, dest = 1322, 0, 1321
	eph := player("Wisp", RaceEphemeral)
	for _, state := range []string{"CLOSED", "LOCKED"} {
		res, out := goDoor(eph, room, ref, state)
		if eph.RoomNumber != dest || !strings.Contains(out, squeeze) || strings.Contains(out, "is closed") {
			t.Errorf("%s: room %d\n%s", state, eph.RoomNumber, out)
		}
		if old := strings.Join(res.OldRoomMsg, "\n"); !strings.Contains(old, "squeezes his semi-solid body") || strings.Contains(old, "goes through") {
			t.Errorf("%s: left-room echo %q", state, old)
		}
		if in := strings.Join(res.RoomBroadcast, "\n"); !strings.Contains(in, "Suddenly the ephemeral") || strings.Contains(in, "arrives") || strings.Contains(in, "squeezes") {
			t.Errorf("%s: arrival echo %q", state, in)
		}
	}
	// Doors standing open are walked through.
	for _, state := range []string{"OPEN", "UNLOCKED", ""} {
		if _, out := goDoor(eph, room, ref, state); eph.RoomNumber != dest || strings.Contains(out, squeeze) {
			t.Errorf("open door (%q): room %d\n%s", state, eph.RoomNumber, out)
		}
	}
	// Nobody else gets past.
	human := player("Solid", RaceHuman)
	if _, out := goDoor(human, room, ref, "LOCKED"); human.RoomNumber != room || !strings.Contains(out, "is closed") {
		t.Errorf("human: room %d\n%s", human.RoomNumber, out)
	}
	// A sealed door (3501 ref 0, archetype 943) stops an Ephemeral.
	if _, out := goDoor(eph, 3501, 0, "CLOSED"); eph.RoomNumber != 3501 {
		t.Errorf("sealed door: room %d\n%s", eph.RoomNumber, out)
	}
	// Room scripts that refuse entry (1459's cell door) still do.
	if _, out := goDoor(eph, 1459, 0, "CLOSED"); eph.RoomNumber != 1459 || strings.Contains(out, squeeze) {
		t.Errorf("cell door: room %d\n%s", eph.RoomNumber, out)
	}
	// ...and one that moves the player isn't followed by a squeeze (it shuts the door behind them).
	if _, out := goDoor(eph, 1459, 0, "OPEN"); eph.RoomNumber != 1460 || strings.Contains(out, squeeze) {
		t.Errorf("open cell door: room %d\n%s", eph.RoomNumber, out)
	}
	// Past a shut door, only Ephemerals and those in Mist or Slime Form keep up
	// with the leader; the rest are dropped from the group and left behind.
	human2, eph2 := player("Plain", RaceHuman), player("Haze", RaceEphemeral)
	mist, slime := player("Fog", RaceVampire), player("Ooze", RaceHuman)
	mist.MistForm, slime.SlimeForm = true, true
	group := portalTestSessions{eph, human, human2, eph2, mist, slime}
	e.SetSessionProvider(group)
	gather := func(leader *Player) {
		leader.Following, leader.IsGroupLeader, leader.GroupMembers = "", true, nil
		for _, p := range group {
			if p.RoomNumber = room; p != leader {
				p.Following, p.IsGroupLeader, p.GroupMembers = leader.FirstName, false, nil
				leader.GroupMembers = append(leader.GroupMembers, p.FirstName)
			}
		}
	}
	for _, leader := range []*Player{eph, mist, slime} {
		for _, state := range []string{"CLOSED", "LOCKED"} {
			gather(leader)
			_, out := goDoor(leader, room, ref, state)
			for _, p := range group {
				wantRoom, wantFollowing := dest, leader.FirstName
				if p == leader {
					wantFollowing = ""
				} else if p == human || p == human2 {
					wantRoom, wantFollowing = room, ""
				}
				if p.RoomNumber != wantRoom || p.Following != wantFollowing {
					t.Errorf("%s leads, %s: %s in %d following %q\n%s", leader.FirstName, state, p.FirstName, p.RoomNumber, p.Following, out)
				}
			}
			if len(leader.GroupMembers) != 3 {
				t.Errorf("%s's group: %v", leader.FirstName, leader.GroupMembers)
			}
		}
		gather(leader)
		goDoor(leader, room, ref, "OPEN")
		for _, p := range group {
			if p.RoomNumber != dest {
				t.Errorf("%s leads through an open door: %s left in %d", leader.FirstName, p.FirstName, p.RoomNumber)
			}
		}
	}
	// With nobody who can keep up, the leader has no group left.
	e.SetSessionProvider(portalTestSessions{mist, human})
	human.RoomNumber, human.Following = room, mist.FirstName
	mist.IsGroupLeader, mist.GroupMembers = true, []string{human.FirstName}
	goDoor(mist, room, ref, "LOCKED")
	if mist.IsGroupLeader || human.RoomNumber != room {
		t.Errorf("mist alone: leader %v follower in %d", mist.IsGroupLeader, human.RoomNumber)
	}

	// Nobody can guard a door against the insubstantial; anyone else can be stopped.
	for _, p := range group {
		p.Following, p.IsGroupLeader, p.GroupMembers = "", false, nil
	}
	guard := player("Guard", RaceHuman)
	guard.Skills = map[int]int{10: 100}
	takeGuard := func(in int) {
		t.Helper()
		guard.RoomNumber = in
		guard.breakGuard()
		if e.ProcessCommand(ctx, guard, "guard door"); !guard.guardsPortal(65) {
			t.Fatalf("guard not taken up in %d: %v", in, guard.GuardPortals)
		}
	}
	takeGuard(room)
	e.SetSessionProvider(append(portalTestSessions{guard}, group...))
	for _, state := range []string{"OPEN", "LOCKED"} {
		for _, p := range []*Player{eph, mist, slime} {
			if _, out := goDoor(p, room, ref, state); p.RoomNumber != dest || strings.Contains(out, "blocks your path") {
				t.Errorf("guarded %s door: %s in %d\n%s", state, p.FirstName, p.RoomNumber, out)
			}
		}
		eph.EphemeralPhysicalUntil = time.Now().Add(time.Minute)
		if _, out := goDoor(eph, room, ref, state); eph.RoomNumber != room || !strings.Contains(out, "blocks your path") || strings.Contains(out, squeeze) {
			t.Errorf("guarded %s door, substantial Ephemeral: room %d\n%s", state, eph.RoomNumber, out)
		}
		eph.EphemeralPhysicalUntil = time.Time{}
	}
	if _, out := goDoor(human, room, ref, "OPEN"); human.RoomNumber != room || !strings.Contains(out, "blocks your path") {
		t.Errorf("guarded door, human: room %d\n%s", human.RoomNumber, out)
	}

	// A script MOVE past a guarded door is stopped by the guard in the room being
	// left, before the script acts, and not by one in the destination.
	// 757 ref 1: "IFPREVERB GO 1" MOVEs INTNUM3 284 to 1695.
	human.IntNums = map[int]int{3: 284}
	takeGuard(757)
	if _, out := goDoor(human, 757, 1, "OPEN"); human.RoomNumber != 757 || !strings.Contains(out, "blocks your path") || strings.Contains(out, "bedroom") {
		t.Errorf("guarded script MOVE: room %d\n%s", human.RoomNumber, out)
	}
	guard.RoomNumber, guard.GuardRoom = 1695, 1695
	if _, out := goDoor(human, 757, 1, "OPEN"); human.RoomNumber != 1695 || strings.Contains(out, "blocks your path") {
		t.Errorf("guard in the destination: room %d\n%s", human.RoomNumber, out)
	}

	// Guarding breaks on leaving the room, however the guard leaves it.
	takeGuard(room)
	goDoor(guard, room, ref, "OPEN")
	if guard.RoomNumber != dest || guard.onGuard() || guard.GuardPortals != nil {
		t.Errorf("walking kept the guard: %v", guard.GuardPortals)
	}
	takeGuard(room)
	guard.RoomNumber = dest // moved without clearing it (script MOVE, spell, ...)
	if guard.onGuard() || guard.guardsPortal(65) {
		t.Error("guard holds in another room")
	}
	if _, out := goDoor(human, room, ref, "OPEN"); human.RoomNumber != dest {
		t.Errorf("absent guard still blocks\n%s", out)
	}
	e.ProcessCommand(ctx, guard, "look")
	guard.RoomNumber = room
	if guard.guardsPortal(65) {
		t.Error("guard resumed on coming back")
	}
	takeGuard(room)
	guard.Following = human.FirstName
	human.IsGroupLeader, human.GroupMembers = true, []string{guard.FirstName}
	human.RoomNumber = room
	e.moveGroupFollowers(ctx, human, room, dest, false)
	if guard.RoomNumber != dest || guard.GuardPortals != nil {
		t.Errorf("following kept the guard: room %d %v", guard.RoomNumber, guard.GuardPortals)
	}
	guard.Following, human.IsGroupLeader, human.GroupMembers = "", false, nil

	// A carried player comes past a shut door only if they could slip through it
	// too; through an open one, always.
	e.SetSessionProvider(group)
	for _, c := range []struct {
		carried *Player
		state   string
		comes   bool
	}{{human, "LOCKED", false}, {eph2, "LOCKED", true}, {slime, "CLOSED", true}, {human, "OPEN", true}} {
		for _, carrier := range []*Player{mist, eph} {
			carrier.RoomNumber, c.carried.RoomNumber = room, room
			carrier.Carrying, c.carried.CarriedBy = c.carried.FirstName, carrier.FirstName
			c.carried.Following = carrier.FirstName
			carrier.IsGroupLeader, carrier.GroupMembers = true, []string{c.carried.FirstName}
			res, out := goDoor(carrier, room, ref, c.state)
			came := c.carried.RoomNumber == dest
			if carrier.RoomNumber != dest || came != c.comes || (carrier.Carrying != "") != c.comes || (c.carried.CarriedBy != "") != c.comes || (c.carried.Following != "") != c.comes {
				t.Errorf("%s carrying %s, %s: carried in %d, carrying %q, by %q, following %q\n%s", carrier.FirstName, c.carried.FirstName, c.state,
					c.carried.RoomNumber, carrier.Carrying, c.carried.CarriedBy, c.carried.Following, out)
			}
			if strings.Contains(strings.Join(res.RoomBroadcast, "\n"), "carries") != c.comes {
				t.Errorf("%s: carry echo %v", c.state, res.RoomBroadcast)
			}
			carrier.Carrying, c.carried.CarriedBy, c.carried.Following = "", "", ""
			carrier.IsGroupLeader, carrier.GroupMembers = false, nil
		}
	}

	// Summoned and controlled creatures stay behind and stop following and guarding.
	var defNum int
	for n := range e.monsters {
		defNum = n
		break
	}
	for _, leader := range []*Player{eph, mist, slime} {
		for _, state := range []string{"LOCKED", "OPEN"} {
			e.monsterMgr.instances = []MonsterInstance{
				{ID: 1, DefNumber: defNum, RoomNumber: room, Alive: true, IsSummoned: true, SummonerName: leader.FirstName, FollowTarget: leader.FirstName, GuardingPlayers: []string{leader.FirstName, human.FirstName}},
				{ID: 2, DefNumber: defNum, RoomNumber: room, Alive: true, IsSummoned: true, SummonerName: leader.FirstName, FollowTarget: leader.FirstName, ControlExpiry: time.Now().Add(time.Hour)},
			}
			_, out := goDoor(leader, room, ref, state)
			for _, c := range e.monsterMgr.instances {
				want, follow := room, ""
				if state == "OPEN" {
					want, follow = dest, leader.FirstName
				}
				if c.RoomNumber != want || c.FollowTarget != follow || containsString(c.GuardingPlayers, leader.FirstName) != (state == "OPEN" && c.ID == 1) {
					t.Errorf("%s, %s: creature %d in %d following %q guarding %v\n%s", leader.FirstName, state, c.ID, c.RoomNumber, c.FollowTarget, c.GuardingPlayers, out)
				}
				if c.SummonerName != leader.FirstName || !c.IsSummoned {
					t.Errorf("creature %d lost its controller", c.ID)
				}
			}
			if strings.Contains(out, "is left behind") != (state == "LOCKED") {
				t.Errorf("%s: %s", state, out)
			}
		}
	}

	// MISTFORM is the form, not the race.
	vamp := player("Fang", RaceVampire)
	vamp.MistForm = true
	for p, want := range map[*Player]int{eph: 0, human: 0, vamp: 1} {
		if got := (&ScriptContext{Player: p, Engine: e}).getVar("MISTFORM"); got != want {
			t.Errorf("%s MISTFORM = %d, want %d", p.FirstName, got, want)
		}
	}
}

// TestClimbPastGuard: a guarded portal's climber has to get past the guard, whether
// a CLIMB script moves them over it or they just climb through.
func TestClimbPastGuard(t *testing.T) {
	r, err := scriptparser.ParseConfig("../../../original/scripts/LEGENDS.CFG")
	if err != nil {
		t.Fatalf("parse scripts: %v", err)
	}
	e := NewGameEngine(nil, &gameworld.ParsedData{
		Rooms: r.Rooms, Items: r.Items, Monsters: r.Monsters, Nouns: r.Nouns,
		Adjectives: r.Adjectives, MonsterAdjs: r.MonsterAdjs, Variables: r.Variables,
		Regions: r.Regions, Macros: r.Macros, OrgDefs: r.OrgDefs,
		StartRoom: r.StartRoom, BumpRoom: r.BumpRoom,
	})
	// CAT.SCR's climbable gates (2301 <-> 2302) aren't in LEGENDS.CFG.
	content, err := os.ReadFile("../../../original/scripts/CAT.SCR")
	if err != nil {
		t.Fatal(err)
	}
	cat, err := scriptparser.ParseScriptContent(string(content), "CAT.SCR")
	if err != nil {
		t.Fatal(err)
	}
	for i := range cat.Rooms {
		if n := cat.Rooms[i].Number; n == 2301 || n == 2302 {
			e.rooms[n] = &cat.Rooms[i]
		}
	}
	ctx := context.Background()
	player := func(name string, race int) *Player {
		return e.CreateNewPlayer(ctx, name, "Climber", race, GenderMale, nil)
	}
	human, eph, mist, guard := player("Solid", RaceHuman), player("Wisp", RaceEphemeral), player("Fog", RaceVampire), player("Guard", RaceHuman)
	mist.MistForm = true
	e.SetSessionProvider(portalTestSessions{guard, human, eph, mist})

	// 2301's gate is scripted (CLEARVERB, MOVE); 1140's door has no CLIMB script.
	for _, c := range []struct {
		room, dest int
		noun       string
		climbs     string
	}{{2301, 2302, "gate", "easily climb up the gate"}, {1140, 1142, "door", ""}} {
		for i := range e.rooms[c.room].Items {
			if ri := &e.rooms[c.room].Items[i]; e.items[ri.Archetype] != nil && isPortal(e.items[ri.Archetype].Type) {
				ri.State = "OPEN"
			}
		}
		climb := func(p *Player) (*CommandResult, string) {
			p.RoomNumber = c.room
			res := e.doClimb(ctx, p, []string{c.noun})
			return res, strings.Join(res.Messages, "\n")
		}
		// skill 100: nobody gets past. -100: everybody does.
		setGuard := func(cm int) {
			t.Helper()
			guard.RoomNumber = c.room
			guard.breakGuard()
			guard.Skills = map[int]int{10: 100}
			e.ProcessCommand(ctx, guard, "guard "+c.noun)
			if !guard.onGuard() {
				t.Fatalf("%d: guard not taken up", c.room)
			}
			guard.Skills = map[int]int{10: cm}
		}

		if _, out := climb(human); human.RoomNumber != c.dest || !strings.Contains(out, c.climbs) || strings.Contains(out, "guard") {
			t.Errorf("%d unguarded: room %d\n%s", c.room, human.RoomNumber, out)
		}
		setGuard(100)
		res, out := climb(human)
		if human.RoomNumber != c.room || !strings.Contains(out, "blocks your path") || (c.climbs != "" && strings.Contains(out, c.climbs)) {
			t.Errorf("%d blocked: room %d\n%s", c.room, human.RoomNumber, out)
		}
		if in := strings.Join(res.RoomBroadcast, "\n"); !strings.Contains(in, "was blocked") {
			t.Errorf("%d blocked echo: %q", c.room, in)
		}
		for _, p := range []*Player{eph, mist} {
			if _, out := climb(p); p.RoomNumber != c.dest || strings.Contains(out, "blocks your path") {
				t.Errorf("%d: %s stopped in %d\n%s", c.room, p.FirstName, p.RoomNumber, out)
			}
		}
		setGuard(-100)
		res, out = climb(human)
		if human.RoomNumber != c.dest || strings.Count(out, "slip past Guard's guard") != 1 || !strings.Contains(out, c.climbs) {
			t.Errorf("%d slipped past: room %d\n%s", c.room, human.RoomNumber, out)
		}
		if old := strings.Join(res.OldRoomMsg, "\n"); res.OldRoom != c.room || strings.Count(old, "slips past Guard's guard") != 1 || !strings.Contains(old, "climbs") {
			t.Errorf("%d left-room echo (%d): %q", c.room, res.OldRoom, old)
		}
		if in := strings.Join(res.RoomBroadcast, "\n"); strings.Contains(in, "guard") {
			t.Errorf("%d arrival echo: %q", c.room, in)
		}
	}
}
