package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/jonradoff/lofp/internal/gameworld"
	"github.com/jonradoff/lofp/internal/scriptparser"
)

// TestWieldRoundTime checks wielding round time against the original
// captures: none for the main hand, 3 seconds for a shield and 5 for an
// off-hand weapon, less a second per Combat Maneuvering level, minimum 2.
func TestWieldRoundTime(t *testing.T) {
	r, err := scriptparser.ParseConfig("../../../original/scripts/LEGENDS.CFG")
	if err != nil {
		t.Fatalf("parse scripts: %v", err)
	}
	e := NewGameEngine(nil, &gameworld.ParsedData{
		Rooms: r.Rooms, Items: r.Items, Monsters: r.Monsters, Nouns: r.Nouns,
		Adjectives: r.Adjectives, MonsterAdjs: r.MonsterAdjs, Variables: r.Variables,
		StartRoom: r.StartRoom, BumpRoom: r.BumpRoom,
	})
	const sword, shield, mainGauche = 1, 1060, 9110
	ctx := context.Background()
	try := func(cm int, offHand int) []string {
		p := e.CreateNewPlayer(ctx, "Wieldy", "Test", RaceHuman, GenderMale, nil)
		p.Skills[10] = cm
		p.Inventory = []InventoryItem{{Archetype: sword}, {Archetype: offHand}}
		main := e.doWield(ctx, p, []string{e.itemNoun(sword)})
		if p.Wielded == nil || p.Wielded.Archetype != sword || messagesHaveRoundTime(main.Messages) {
			t.Fatalf("main-hand wield: %v", main.Messages)
		}
		p.RoundTimeExpiry = p.RoundTimeExpiry.AddDate(-1, 0, 0)
		off := e.doWield(ctx, p, []string{e.itemNoun(offHand)})
		if p.OffHand == nil || p.OffHand.Archetype != offHand || p.Wielded.Archetype != sword {
			t.Fatalf("off-hand wield (CM %d, item %d): %v", cm, offHand, off.Messages)
		}
		return off.Messages
	}
	for _, c := range []struct {
		cm, item int
		want     string
	}{
		{0, shield, "[Round: 3 sec]"},
		{1, shield, "[Round: 2 sec]"},
		{2, shield, "[Round: 2 sec]"},
		{0, mainGauche, "[Round: 5 sec]"},
		{2, mainGauche, "[Round: 3 sec]"},
		{5, mainGauche, "[Round: 2 sec]"},
	} {
		msgs := try(c.cm, c.item)
		out := strings.Join(msgs, "\n")
		if !strings.Contains(out, c.want) || !strings.HasPrefix(out, "You wield ") {
			t.Errorf("CM %d item %d: want %q, got %q", c.cm, c.item, c.want, out)
		}
	}
}
