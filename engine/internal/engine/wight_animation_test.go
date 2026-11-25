package engine

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
	"github.com/jonradoff/lofp/internal/scriptparser"
)

// TestDustSummons: Wight Animation (336) and Summon Spectral Warrior (353) take their
// own kind of dust, however it was come by, and call up a commandable guard.
func TestDustSummons(t *testing.T) {
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
	const wight, ghoul, meteoric = 547, 546, 1104
	if sp := FindSpellByName("wight animation"); sp == nil || sp.ID != 336 || sp.Level != 17 || sp.School != "Necromancy" {
		t.Fatalf("spell 336: %+v", sp)
	}

	for _, c := range []struct {
		spell, adj, monster int
		dust, guard         string
	}{{336, wight, 53, "wight dust", "greater wight"}, {353, ghoul, 130, "ghoul dust", "spectral warrior"}} {
		p := e.CreateNewPlayer(ctx, "Necro", "Test", RaceHuman, GenderMale, nil)
		p.RoomNumber = 1322
		p.KnownSpells = map[int]bool{c.spell: true}
		p.Skills = map[int]int{23: 20}
		p.Mana, p.MaxMana = 1000, 1000
		prep := func(with string) string {
			p.RoundTimeExpiry = time.Time{}
			p.PreparedSpell = 0
			return strings.Join(e.doPrepareSpell(p, strings.Fields(strconv.Itoa(c.spell)+with)).Messages, "\n")
		}
		if out := prep(""); p.PreparedSpell != 0 || !strings.Contains(out, "requires a reagent") {
			t.Errorf("%d with nothing: %s", c.spell, out)
		}
		wrong := wight + ghoul - c.adj
		p.Inventory = []InventoryItem{{Archetype: 512, Adj2: wrong}, {Archetype: 512, Adj3: meteoric}, {Archetype: 512}}
		if out := prep(" with dust"); p.PreparedSpell != 0 || len(p.Inventory) != 3 {
			t.Errorf("%d took the wrong dust: %s", c.spell, out)
		}
		// Skinned or ground in an urn (Adj2), bought (Adj3), or Adj1.
		for _, dust := range []InventoryItem{{Archetype: 512, Adj2: c.adj}, {Archetype: 512, Adj3: c.adj}, {Archetype: 512, Adj1: c.adj}} {
			p.Inventory = []InventoryItem{{Archetype: 512, Adj2: wrong}, dust}
			out := prep(" with " + c.dust)
			if p.PreparedSpell != c.spell || len(p.Inventory) != 1 || p.Inventory[0].Adj2 != wrong {
				t.Fatalf("%d with %+v: %s", c.spell, dust, out)
			}
			if !strings.Contains(out, "The "+c.dust+" swirls away") {
				t.Errorf("%d: %s", c.spell, out)
			}
		}

		p.IsGM = true // no fumbles
		p.RoundTimeExpiry = time.Time{}
		out := strings.Join(e.ProcessCommand(ctx, p, "cast").Messages, "\n")
		if !strings.Contains(out, c.guard) || p.SummonedCreatureID == 0 {
			t.Fatalf("%d cast: %s", c.spell, out)
		}
		inst := e.monsterMgr.instances[len(e.monsterMgr.instances)-1]
		if inst.DefNumber != c.monster || !inst.IsSummoned || inst.SummonerName != p.FirstName || inst.FollowTarget != p.FirstName || inst.RoomNumber != p.RoomNumber || inst.IsFamiliar {
			t.Errorf("%d summoned %+v", c.spell, inst)
		}
		p.IsGM = false
		p.RoundTimeExpiry = time.Time{}
		e.ProcessCommand(ctx, p, "command guard me")
		if g := e.monsterMgr.instances[len(e.monsterMgr.instances)-1]; !containsString(g.GuardingPlayers, p.FirstName) {
			t.Errorf("%d: COMMAND GUARD ME: %v", c.spell, g.GuardingPlayers)
		}
		p.Inventory = []InventoryItem{{Archetype: 512, Adj2: c.adj}}
		if again := strings.Join(e.castWightAnimation(p).Messages, "\n"); !strings.Contains(again, "already have a summoned creature") {
			t.Errorf("second summon: %s", again)
		}
	}
}
