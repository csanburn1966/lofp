package engine

import (
	"strings"
	"testing"
	"time"
)

// TestTurnUndead: the turn line frightens or destroys by tier and never does
// ordinary damage.
func TestTurnUndead(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Turner")
	const skeleton, lich = 1, 16
	if undeadTier(b.e.monsters[skeleton]) != 1 || undeadTier(b.e.monsters[lich]) != 3 || undeadTier(b.e.monsters[210]) != 2 {
		t.Fatal("tier table")
	}
	// Undead that aren't RACE 22, and loaded monsters that reuse an archived undead's number.
	for num, want := range map[int]int{104: 1, 151: 2, 87: 3, 263: 0, 264: 0, 326: 0} {
		if got := undeadTier(b.e.monsters[num]); got != want {
			t.Errorf("monster %d: tier %d, want %d", num, got, want)
		}
	}
	fresh := func(def int) *MonsterInstance {
		b.e.monsterMgr.instances = nil
		m := b.monster(def, p.RoomNumber)
		m.CurrentHP = 100
		return m
	}
	name := func(def int) string { return strings.ToLower(b.e.monsters[def].Name) }

	m := fresh(skeleton)
	o := out(b.cast(p, 301, name(skeleton)))
	if !m.Feared || m.CurrentHP != 100 || !strings.Contains(o, "cowers in fear") {
		t.Errorf("TU I on tier 1: feared=%v hp=%d\n%s", m.Feared, m.CurrentHP, o)
	}
	if d := time.Until(m.FearExpiry); d < 25*time.Second || d > 31*time.Second {
		t.Errorf("fear lasts %v", d)
	}

	m = fresh(lich)
	o = out(b.cast(p, 301, name(lich)))
	if m.Feared || m.CurrentHP != 100 || !strings.Contains(o, "too powerful") {
		t.Errorf("TU I on tier 3: feared=%v hp=%d\n%s", m.Feared, m.CurrentHP, o)
	}
	m = fresh(lich)
	if b.cast(p, 302, name(lich)); !m.Feared || m.CurrentHP != 100 {
		t.Errorf("TU II on tier 3: feared=%v hp=%d", m.Feared, m.CurrentHP)
	}
	m = fresh(lich)
	if b.cast(p, 304, name(lich)); !m.Feared || m.CurrentHP != 100 {
		t.Errorf("TU III on tier 3: feared=%v hp=%d", m.Feared, m.CurrentHP)
	}

	m = fresh(skeleton)
	if o = out(b.cast(p, 304, name(skeleton))); m.Alive || !strings.Contains(o, "crumbles to dust") {
		t.Errorf("TU III on tier 1: alive=%v\n%s", m.Alive, o)
	}

	// TU II on tier 1: destroyed or halved, about half the time each.
	dead, halved := 0, 0
	for i := 0; i < 60; i++ {
		m = fresh(skeleton)
		b.cast(p, 302, name(skeleton))
		switch {
		case !m.Alive:
			dead++
		case m.CurrentHP == 50 && !m.Feared:
			halved++
		default:
			t.Fatalf("TU II on tier 1: alive=%v hp=%d", m.Alive, m.CurrentHP)
		}
	}
	if dead < 10 || halved < 10 {
		t.Errorf("TU II on tier 1: %d destroyed, %d halved", dead, halved)
	}

	// Not undead at all.
	b.e.monsterMgr.instances = nil
	for num, d := range b.e.monsters {
		if d.Race != 22 && undeadTier(d) == 0 && d.Body > 0 {
			m = b.monster(num, p.RoomNumber)
			m.CurrentHP = 100
			if o = out(b.cast(p, 301, strings.ToLower(d.Name))); m.Feared || m.CurrentHP != 100 || !strings.Contains(o, "only the undead") {
				t.Errorf("TU I on living %s: %s", d.Name, o)
			}
			break
		}
	}
}

// TestScrollNeedsSpellcraft: only a Spellcraft student can read a scroll, and
// reading it names the scroll for its spell.
func TestScrollNeedsSpellcraft(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Reader")
	p.Skills = map[int]int{}
	p.Inventory = []InventoryItem{{Archetype: 168, Val3: 301}}
	o := b.do(p, "examine scroll")
	if strings.Contains(o, "Turn Undead") || !strings.Contains(o, "do not understand") || p.Inventory[0].Tail != "" {
		t.Errorf("no spellcraft: tail %q\n%s", p.Inventory[0].Tail, o)
	}
	if o = b.do(p, "learn scroll"); strings.Contains(o, "Turn Undead") || len(p.Inventory) != 1 {
		t.Errorf("learn without spellcraft: %s", o)
	}

	p.Skills[23] = 1
	b.want(b.do(p, "examine scroll"), "The scroll contains the spell 'Turn Undead I'")
	if p.Inventory[0].Tail != "of Turn Undead I" {
		t.Fatalf("tail %q", p.Inventory[0].Tail)
	}
	b.want(b.e.formatItemName(b.e.items[168], p.Inventory[0].Adj1, 0, 0, p.Inventory[0].Tail), "scroll of Turn Undead I")
	b.want(b.do(p, "inventory"), "scroll of Turn Undead I")
	b.do(p, "examine scroll")
	if p.Inventory[0].Tail != "of Turn Undead I" {
		t.Errorf("tail after second look %q", p.Inventory[0].Tail)
	}

	// A scroll inside an open carried container gets named too.
	ran := false
	for num, d := range b.e.items {
		if d != nil && isContainerDef(d) && d.Type != "LIQCONTAINER" {
			p.Inventory = []InventoryItem{{Archetype: num, State: "OPEN", Contents: []InventoryItem{{Archetype: 168, Val3: 302}}}}
			if !containerIsOpen(d, "OPEN") {
				continue
			}
			ran = true
			b.do(p, "examine scroll")
			if tail := p.Inventory[0].Contents[0].Tail; tail != "of Turn Undead II" {
				t.Errorf("scroll in %d: tail %q", num, tail)
			}
			break
		}
	}
	if !ran {
		t.Error("no open container to test")
	}
}
