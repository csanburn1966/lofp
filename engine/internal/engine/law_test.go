package engine

import (
	"strings"
	"testing"
	"time"
)

const mainStreet = 253 // FAYD.SCR, within the walls

func TestWarrantArrestJail(t *testing.T) {
	b := newBatchEnv(t)
	chief, rookie, crook, friend := b.player("Chief"), b.player("Rookie"), b.player("Crook"), b.player("Friend")
	for _, p := range []*Player{chief, rookie, crook, friend} {
		p.RoomNumber = mainStreet
	}
	chief.AddOrg(lawkeeperOrg, 150)
	rookie.AddOrg(lawkeeperOrg, 40)

	b.want(b.do(friend, "warrant crook 3"), "Only the Lawkeepers")
	b.want(b.do(rookie, "warrant crook 3"), "highest ranking")
	b.want(b.do(rookie, "warrant"), "Petty Theft")
	b.want(b.do(chief, "warrant crook 12"), "1 to 9")
	got := b.do(chief, "warrant crook 3")
	b.want(got, "class III warrant for Crook Test: Minor Dark Necromancy")
	b.want(got, "two witnesses")
	b.want(b.sent["Crook"], "A warrant has been issued for your arrest")
	b.want(b.do(rookie, "warrant"), "Crook Test, class III")
	b.want(b.do(crook, "status"), "There is a warrant for your arrest")
	b.want(b.do(friend, "war"), "wary combat stance")

	// Arrest: anyone can't, and not outside the walls.
	b.want(b.do(friend, "arrest crook"), "Only the Lawkeepers")
	b.want(b.do(rookie, "arrest friend"), "no warrant for Friend")
	rookie.RoomNumber, crook.RoomNumber = batchForest, batchForest
	b.want(b.do(rookie, "arrest crook"), "beyond the walls")
	rookie.RoomNumber, crook.RoomNumber = mainStreet, mainStreet

	b.want(b.do(rookie, "arrest crook"), "may SUBMIT")
	b.want(b.sent["Crook"], "places you under arrest!")
	b.want(b.do(crook, "submit"), "come along quietly")
	if !crook.Jailed || crook.RoomNumber != lockupRoom || crook.Warrant != 0 || crook.JailFine != warrantFine(3) || crook.Submitting {
		t.Fatalf("jailed %v room %d warrant %d fine %d", crook.Jailed, crook.RoomNumber, crook.Warrant, crook.JailFine)
	}
	if left := time.Until(crook.RoundTimeExpiry); left < 115*time.Second || left > 120*time.Second {
		t.Errorf("LAWKEEP.SCR's sentence: %v", left) // 40 seconds per class
	}
	if left := time.Until(crook.JailUntil); left < 115*time.Second || left > 120*time.Second {
		t.Errorf("sentence %v", left)
	}
	b.want(b.sent["Crook"], "[Lockup]")
	b.want(b.sent["Crook"], "You will be let out in 2 minutes, or sooner if you PAY your fine of 90 gold")
	b.want(b.do(crook, "status"), "to be let out in")
	b.e.spellEffectTick()
	if !crook.Jailed {
		t.Fatal("let out early")
	}
	b.want(b.room[mainStreet], "Crook is hauled off to jail.")

	// The Lockup holds them.
	crook.KnownSpells[213] = true
	for _, cmd := range []string{"prepare 213", "psi 2"} {
		b.want(b.do(crook, cmd), "Lockup's wards")
	}
	for _, cmd := range []string{"go door", "open door", "out"} {
		b.want(b.do(crook, cmd), "Serve your time, or PAY your fine.")
	}
	crook.RoomNumber = mainStreet // a script or a spell takes them out
	b.do(crook, "look")
	if crook.RoomNumber != lockupRoom {
		t.Error("left the Lockup")
	}
	crook.Dead = true
	b.do(crook, "depart")
	if crook.RoomNumber != lockupRoom || crook.Dead {
		t.Error("departed the Lockup")
	}

	// The fine: not enough on them, then enough.
	crook.Gold, crook.Silver, crook.Copper = 50, 0, 0
	crook.RoundTimeExpiry = time.Time{}
	b.want(strings.Join(b.e.ProcessCommand(b.ctx, crook, "pay").Messages, ""), "don't have that much")
	crook.Gold = 100
	b.e.ProcessCommand(b.ctx, crook, "pay")
	if crook.Jailed || crook.RoomNumber != lawkeeperHQ || crook.Gold != 10 {
		t.Fatalf("paid: jailed %v room %d gold %d", crook.Jailed, crook.RoomNumber, crook.Gold)
	}
	b.want(b.sent["Crook"], "You pay your fine of 90 gold")

	// Someone else can pay it, from the Lawkeeper HQ, and a chief can lift it.
	crook.RoomNumber, crook.Warrant = mainStreet, 1
	crook.Submitting = false
	b.want(b.do(rookie, "arrest crook"), "may SUBMIT")
	b.want(b.do(crook, "submit"), "come along quietly")
	friend.RoomNumber, friend.Gold, friend.Silver, friend.Copper = lawkeeperHQ, 20, 0, 0
	b.want(b.do(friend, "pay crook"), "You pay Crook's fine of 10 gold")
	if crook.Jailed || friend.Gold != 10 {
		t.Errorf("friend's payment: jailed %v gold %d", crook.Jailed, friend.Gold)
	}
	crook.RoomNumber, crook.Warrant = mainStreet, 9
	b.do(rookie, "arrest crook")
	b.do(crook, "submit")
	b.want(b.do(chief, "warrant crook 0"), "released from the Lockup")
	if crook.Jailed || crook.RoomNumber != lawkeeperHQ {
		t.Error("pardon")
	}
	b.want(b.do(chief, "warrant crook 0"), "There is no warrant")

	// The sentence served, they're let out.
	crook.RoomNumber, crook.Warrant = mainStreet, 2
	b.do(rookie, "arrest crook")
	b.do(crook, "submit")
	crook.JailUntil = time.Now().Add(-time.Second)
	b.e.spellEffectTick()
	if crook.Jailed || crook.RoomNumber != lawkeeperHQ || !crook.RoundTimeExpiry.IsZero() {
		t.Error("sentence served")
	}
	b.want(b.sent["Crook"], "You have served your sentence.")
	// Time logged off doesn't count.
	crook.RoomNumber, crook.Warrant = mainStreet, 1
	b.do(rookie, "arrest crook")
	b.do(crook, "submit")
	crook.LoggedOutAt = time.Now().Add(-time.Hour)
	crook.pauseSpellTimersOffline()
	if time.Until(crook.JailUntil) < time.Hour {
		t.Error("served the sentence offline")
	}

	// A script sets it (NOWAY.SCR: ADD WARRANT 1), up to 9.
	sc := &ScriptContext{Player: crook, Engine: b.e}
	sc.doAdd([]string{"WARRANT", "12"})
	if crook.Warrant != 9 {
		t.Errorf("script warrant %d", crook.Warrant)
	}
}

func TestLawkeeperNPCs(t *testing.T) {
	b := newBatchEnv(t)
	crook := b.player("Crook")
	crook.RoomNumber, crook.Warrant = mainStreet, 6
	m := b.monster(13, mainStreet) // lawkeeper
	id := m.ID
	mon := func() *MonsterInstance { return &b.e.monsterMgr.instances[b.e.monsterMgr.indexOfID(id)] }
	def := b.e.monsters[13]

	// Disguised, they're not known; otherwise the lawkeeper goes for them.
	crook.Disguised = true
	b.e.monsterMgr.mu.Lock()
	spotted := b.e.lawSpotsWanted(mon(), def)
	b.e.monsterMgr.mu.Unlock()
	if spotted {
		t.Error("saw through a disguise")
	}
	crook.Disguised = false
	b.e.monsterMgr.mu.Lock()
	b.e.lawSpotsWanted(mon(), def)
	b.e.monsterMgr.mu.Unlock()
	if mon().Target != "Crook" {
		t.Fatal("didn't go for the wanted player")
	}
	b.want(b.sent["Crook"], "Submit, in the name of the law!")

	// Beaten, they're taken alive and hauled off.
	crook.BodyPoints = -40
	if _, died := b.e.resolveDirectHitOutcome(crook, -40, FormatMonsterName(def, b.e.monAdjs)); died || !crook.Unconscious {
		t.Fatal("killed by the law")
	}
	if _, died := b.e.resolveDirectHitOutcome(&Player{FirstName: "X", BodyPoints: -40, MaxBodyPoints: 10}, -40, "lawkeeper"); !died {
		t.Error("an honest citizen spared")
	}
	b.e.monsterMgr.mu.Lock()
	b.e.monsterCombatTick(mon(), def)
	b.e.monsterMgr.mu.Unlock()
	if !crook.Jailed || crook.RoomNumber != lockupRoom || crook.BodyPoints < 1 || crook.Unconscious || mon().Target != "" {
		t.Fatalf("hauled off: jailed %v room %d bp %d", crook.Jailed, crook.RoomNumber, crook.BodyPoints)
	}

	// Submitting with one of them in the room is surrender.
	other := b.player("Other")
	other.RoomNumber, other.Warrant = mainStreet, 2
	b.want(b.do(other, "submit"), "come along quietly")
	if !other.Jailed {
		t.Error("surrender to an NPC")
	}
	// Without the law at hand, submitting is only submitting.
	third := b.player("Third")
	third.RoomNumber, third.Warrant = mainStreet+1, 2
	b.do(third, "submit")
	if third.Jailed || !third.Submitting {
		t.Error("jailed with no one to take them")
	}

	// A lawkeeper's badge calls for reinforcements (ITEM2.SCR).
	cop := b.player("Cop")
	cop.RoomNumber = mainStreet
	cop.AddOrg(lawkeeperOrg, 50)
	cop.Inventory = []InventoryItem{{Archetype: 652}}
	b.want(b.do(cop, "poke my badge"), "You call for reinforcements")
}

func TestCapacitance(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Psion")
	p.Skills[26], p.Skills[28] = 10, 10
	p.Psi, p.MaxPsi = 20, 20
	p.Willpower = 500 // never fails
	if d := FindPsiByNumber(6); d == nil || d.Name != "Capacitance" || FindPsiByNumber(2).Name != "Kinetic Thrust" || FindPsiByNumber(22).Name != "Psychic Blast" || FindPsiByNumber(53).Name != "Psychic Blast" {
		t.Fatalf("numbers: %+v", d)
	}
	b.want(b.do(p, "psi"), "6    Capacitance")
	for _, c := range []struct{ level, per int }{{1, 5}, {4, 5}, {5, 6}, {10, 6}, {11, 7}, {20, 7}, {21, 8}, {31, 9}} {
		if got := capacitancePerCharge(c.level); got != c.per {
			t.Errorf("level %d: %d per charge, want %d", c.level, got, c.per)
		}
	}
	ring, sword := 0, 0
	for n, d := range b.e.items { // the lowest-numbered of each, for a test that doesn't vary
		if itemCastsItsSpell(d) && b.e.nouns[d.NameID] == "ring" && (ring == 0 || n < ring) {
			ring = n
		}
		if d.Type == "SLASH_WEAPON" && b.e.nouns[d.NameID] == "sword" && (sword == 0 || n < sword) {
			sword = n
		}
	}
	project := func(target string) string {
		p.RoundTimeExpiry = time.Time{}
		b.do(p, "psi 6")
		if p.PreparedPsi != capacitanceID {
			t.Fatal("psi 6 isn't Capacitance")
		}
		p.RoundTimeExpiry = time.Time{}
		return b.do(p, "project "+target)
	}
	// Six charges of a level 20 spell (Call Meteor is 20): 42, past the maximum.
	p.Inventory = []InventoryItem{{Archetype: ring, Val2: 6, Val3: 112}, {Archetype: sword, Val2: 5, Val3: 4}}
	got := project("ring")
	b.want(got, "[+42 Psi, 62/20]")
	if p.Psi != 62 || p.Inventory[0].Val2 != 0 {
		t.Errorf("psi %d charges %d", p.Psi, p.Inventory[0].Val2)
	}
	b.want(project("ring"), "its magic is spent")
	b.want(project("sword"), "no spell imbued")
	if p.Psi != 62 || p.Inventory[1].Val2 != 5 {
		t.Errorf("a failed projection cost psi (%d) or charges", p.Psi)
	}
	// Regeneration doesn't take the overcharge away.
	b.e.regenTick()
	if p.Psi < 62 {
		t.Errorf("regen took psi: %d", p.Psi)
	}
}

func TestRideTheLightningStorm(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Druid")
	p.Mana, p.MaxMana = 50, 50
	region := b.e.rooms[batchForest].Region
	b.e.RegionWeather = map[int]int{region: 6}
	if b.e.stormStrikesRider(p) {
		t.Fatal("struck without the spell")
	}
	p.RideLightningExpiry = time.Now().Add(time.Hour)
	bp := p.BodyPoints
	if !b.e.stormStrikesRider(p) || p.Mana < 58 || p.Mana > 80 || p.BodyPoints != bp {
		t.Fatalf("mana %d bp %d", p.Mana, p.BodyPoints)
	}
	b.want(b.sent["Druid"], "A bolt of lightning streaks down from the storm and strikes you!")
	b.e.RegionWeather[region] = 4
	if b.e.stormStrikesRider(p) {
		t.Error("struck in plain rain")
	}
	b.e.RegionWeather[region] = 6
	p.RoomNumber = batchIndoor
	if b.e.stormStrikesRider(p) {
		t.Error("struck indoors")
	}
	p.RoomNumber = batchForest
	mana := p.Mana
	b.e.regenTick()
	if p.Mana <= mana {
		t.Error("regenTick doesn't call the lightning")
	}
}

// TestMonsterPsionics: monsters' DISCIPLINE lines are the original numbers.
func TestMonsterPsionics(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Target")
	p.RoomNumber = batchForest
	def := *b.e.monsters[13]
	def.Disciplines, def.PsiUse = []int{41, 15, 38}, 100 // Pain; Force Field and Psychic Fortress
	b.e.monsters[99999] = &def
	def.Number = 99999
	if got := monsterPsiDefenseBonus(def.Disciplines); got != 125 {
		t.Errorf("defense %d", got)
	}
	b.e.monsterMgr.SpawnOne(99999, batchForest, 100, 0, 500)
	m := &b.e.monsterMgr.instances[len(b.e.monsterMgr.instances)-1]
	if !b.e.monsterTryStartPsi(m, &def) || m.PsiCastDiscID != 65 {
		t.Fatalf("cast %d", m.PsiCastDiscID) // Pain, internally 65
	}
	// Every monster's damage disciplines are ones the engine knows.
	for _, d := range b.e.monsters {
		for _, n := range d.Disciplines {
			if FindPsiByNumber(n) == nil && n != 1 && n != 18 && (n < 24 || n > 27) && n != 30 && n != 35 && n != 40 {
				t.Errorf("monster %d (%s): discipline %d", d.Number, d.Name, n)
			}
		}
	}
}
