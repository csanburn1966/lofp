package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/jonradoff/lofp/internal/gameworld"
	"github.com/jonradoff/lofp/internal/scriptparser"
)

// TestTutorialWalkthrough plays Sergeant Aldren's lesson start to finish
// against the real scripts (no database).
func TestTutorialWalkthrough(t *testing.T) {
	r, err := scriptparser.ParseConfig("../../../original/scripts/LEGENDS.CFG")
	if err != nil {
		t.Fatalf("parse scripts: %v", err)
	}
	e := NewGameEngine(nil, &gameworld.ParsedData{
		Rooms: r.Rooms, Items: r.Items, Monsters: r.Monsters, Nouns: r.Nouns,
		Adjectives: r.Adjectives, MonsterAdjs: r.MonsterAdjs, BreakMods: r.BreakMods,
		Variables: r.Variables, Regions: r.Regions, MonsterLists: r.MonsterLists,
		SeasonalMonsterLists: r.SeasonalMonsterLists, SeasonalRooms: r.SeasonalRooms,
		CEvents: r.CEvents, Macros: r.Macros, MoneyDefs: r.MoneyDefs,
		ForageDefs: r.ForageDefs, MineDefs: r.MineDefs, OrgDefs: r.OrgDefs,
		StartRoom: r.StartRoom, BumpRoom: r.BumpRoom,
	})
	if e.items[tutorialCoinItem] == nil {
		t.Fatal("tutorial coin (TUTORIAL.SCR) not loaded")
	}
	ctx := context.Background()
	p := e.CreateNewPlayer(ctx, "Testy", "Newcomer", 1, 0, nil)
	if len(p.Inventory) != 0 {
		t.Fatalf("new character should start empty-handed, got %+v", p.Inventory)
	}

	all := []string{}
	enter := e.EnterRoom(ctx, p)
	all = append(all, enter.Messages...)
	if p.TutorialStatus != tutorialActive {
		t.Fatalf("tutorial didn't start on entry: status %d", p.TutorialStatus)
	}
	run := func(cmd string) []string {
		res := e.ProcessCommand(ctx, p, cmd)
		all = append(all, "> "+cmd)
		all = append(all, res.Messages...)
		return res.Messages
	}
	want := func(cmd string, step int) {
		t.Helper()
		run(cmd)
		if p.TutorialStatus == tutorialActive && p.TutorialStep != step {
			t.Fatalf("after %q: step %d, want %d\n%s", cmd, p.TutorialStep, step, strings.Join(all, "\n"))
		}
	}
	if out := strings.Join(run("look"), "\n"); !strings.Contains(out, aldrenShortDesc) {
		t.Errorf("LOOK doesn't show Aldren:\n%s", out)
	}
	want("l aldren", 2)
	if out := strings.Join(run("attack aldren"), "\n"); !strings.Contains(out, "halberd") {
		t.Errorf("attack Aldren not refused: %s", out)
	}
	want("i", 3)
	want("wear breeches", 4)
	if !p.TutorialDagger || !e.tutorialCarriesWeapon(p) {
		t.Fatalf("Aldren didn't hand over a dagger at the WIELD step: %+v", p.Inventory)
	}
	want("wield dagger", 5)
	want("stat", 5)
	want("fatigue", 6)
	want("get coin", 7)
	want("examine coin", 8)
	want("drop coin", 9)
	want("say Greetings, Sergeant!", 10)
	want("bow", 11)
	want("look", 12)
	want("exp", 12)
	want("skills", 13)
	want("help", 14)
	if p.TutorialStatus != tutorialFinished {
		t.Fatalf("tutorial not finished: status %d step %d", p.TutorialStatus, p.TutorialStep)
	}
	run("say training")
	if p.TutorialStatus != tutorialActive || p.TutorialStep != 0 {
		t.Fatalf("SAY TRAINING didn't restart: status %d step %d", p.TutorialStatus, p.TutorialStep)
	}
	// Leaving pauses the lesson; coming back resumes it.
	want("look", 1)
	run("west")
	if p.RoomNumber == tutorialRoom || !p.TutorialAway {
		t.Fatalf("west didn't leave the gate (room %d, away %v)", p.RoomNumber, p.TutorialAway)
	}
	if out := strings.Join(run("east"), "\n"); !strings.Contains(out, "pick up where we left off") {
		t.Errorf("no resume on return:\n%s", out)
	}
	// Stuck: a hint after a few commands that don't complete the step.
	var hinted bool
	for i := 0; i < tutorialHintAfter; i++ {
		hinted = hinted || strings.Contains(strings.Join(run("smile"), "\n"), "[LOOK ALDREN]")
	}
	if !hinted {
		t.Errorf("no hint after %d misses", tutorialHintAfter)
	}
	run("skip tutorial")
	if p.TutorialStatus != tutorialFinished {
		t.Fatalf("SKIP TUTORIAL didn't end it")
	}
	t.Log("\n" + strings.Join(all, "\n"))
}
