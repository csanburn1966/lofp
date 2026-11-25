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

type batchEnv struct {
	t      *testing.T
	e      *GameEngine
	ctx    context.Context
	online portalTestSessions
	sent   map[string]string
	room   map[int]string
}

func newBatchEnv(t *testing.T) *batchEnv {
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
	b := &batchEnv{t: t, e: e, ctx: context.Background(), sent: map[string]string{}, room: map[int]string{}}
	e.SetSendToPlayer(func(name string, msgs []string) { b.sent[name] += strings.Join(msgs, "\n") + "\n" })
	e.SetLocalRoomBroadcast(func(n int, msgs []string) { b.room[n] += strings.Join(msgs, "\n") + "\n" })
	return b
}

const batchForest, batchIndoor = 90001, 90002

func (b *batchEnv) player(name string) *Player {
	for _, n := range []int{batchForest, batchIndoor} {
		if b.e.rooms[n] == nil {
			b.e.rooms[n] = &gameworld.Room{Number: n, Name: "Test", Terrain: map[int]string{batchForest: "FOREST", batchIndoor: "INDOOR_FLOOR"}[n],
				Exits: map[string]int{"N": map[int]int{batchForest: batchIndoor, batchIndoor: batchForest}[n]}}
		}
	}
	p := b.e.CreateNewPlayer(b.ctx, name, "Test", RaceHuman, GenderMale, nil)
	p.RoomNumber = batchForest
	p.Skills = map[int]int{23: 40}
	p.Mana, p.MaxMana = 1000, 1000
	p.BodyPoints, p.MaxBodyPoints = 400, 400
	p.KnownSpells = map[int]bool{}
	p.Agility = 0
	b.online = append(b.online, p)
	b.e.SetSessionProvider(b.online)
	return p
}

// cast casts until it takes, as a GM's cast always does.
func (b *batchEnv) cast(p *Player, spell int, args ...string) *CommandResult {
	p.RoundTimeExpiry = time.Time{}
	p.PreparedSpell = spell
	p.Mana = 1000
	p.IsGM = true
	res := b.e.doCastSpell(b.ctx, p, args)
	p.IsGM = false
	return res
}

func (b *batchEnv) monster(def, room int) *MonsterInstance {
	d := b.e.monsters[def]
	b.e.monsterMgr.SpawnOne(def, room, d.Body, 5000, d.Psi)
	m := &b.e.monsterMgr.instances[len(b.e.monsterMgr.instances)-1]
	m.MaxHP = d.Body
	return m
}

func (b *batchEnv) do(p *Player, cmd string) string {
	p.RoundTimeExpiry = time.Time{}
	return strings.Join(b.e.ProcessCommand(b.ctx, p, cmd).Messages, "\n")
}

func (b *batchEnv) want(got, want string) {
	b.t.Helper()
	if !strings.Contains(got, want) {
		b.t.Errorf("want %q in:\n%s", want, got)
	}
}

func out(r *CommandResult) string { return strings.Join(r.Messages, "\n") }

func TestDisjunction(t *testing.T) {
	b := newBatchEnv(t)
	p, q := b.player("Chandra"), b.player("Marcolan")
	for _, x := range []*Player{p, q} {
		applyTimedDefenseBuff(x, FindSpellByID(326), 1)
		applyTimedDefenseBuff(x, FindSpellByID(105), 1)
		applyTimedDefenseBuff(x, FindSpellByID(133), 1)
		applyMysticArmorBuff(x, FindSpellByID(102), 1)
	}
	p.HasteExpiry = time.Now().Add(time.Minute)
	q.ImprisonedExpiry = time.Now().Add(time.Minute)
	q.RideLightningExpiry = time.Now().Add(time.Minute)
	q.PhantomForm = true
	snarePlayer(q, FindSpellByID(webSpellID))
	snarePlayer(q, FindSpellByID(plantSnareSpellID))
	m := b.monster(116, batchForest)
	m.Webbed, m.Imprisoned, m.HasteExpiry = true, true, time.Now().Add(time.Minute)

	got := out(b.cast(p, 230))
	b.want(got, "A wave of deep red light emanates from you.")
	b.want(got, "You return to normal speed.")
	b.want(got, "The ghostly shield protecting you slowly fades.")
	b.want(got, "The brilliant prismatic globe vanishes from around you.")
	b.want(got, "The sticky webbing holding an ursine") // monster, by name
	q.PhantomForm = false
	b.want(b.sent["Marcolan"], "A wave of deep red light emanates from Chandra.")
	b.want(b.sent["Marcolan"], "The blue force bubble vanishes from around you.")
	b.want(b.sent["Marcolan"], "The glowing outline of armor fades from around Chandra.")
	if p.DefenseBonus != 0 || q.DefenseBonus != 0 || len(q.TimedDefenseBuffs) != 0 || len(q.Entangles) != 0 || q.IsImprisoned() || isHasted(p) || spellActive(q.RideLightningExpiry) {
		t.Errorf("spells left: %+v %+v", p.TimedDefenseBuffs, q)
	}
	if m.Webbed || m.Imprisoned || spellActive(m.HasteExpiry) {
		t.Error("monster's spells left")
	}
}

func TestDispelInvisibilityUnstunAnalyzeOre(t *testing.T) {
	b := newBatchEnv(t)
	p, q, r := b.player("Stefan"), b.player("Saryon"), b.player("Ghost")
	q.Invisible, r.PhantomForm = true, true
	got := out(b.cast(p, 406))
	b.want(got, "Saryon suddenly becomes visible.")
	if q.Invisible || !r.PhantomForm || strings.Contains(got, "Ghost") {
		t.Errorf("invisible %v phantom %v: %s", q.Invisible, r.PhantomForm, got)
	}
	b.want(b.sent["Ghost"], "Saryon suddenly becomes visible.")
	if n := len(b.cast(p, 406).RoomBroadcast); n != 1 {
		t.Errorf("nothing to reveal: %d room lines", n)
	}

	q.Stunned, q.StunExpiry = true, time.Now().Add(time.Minute)
	res := b.cast(p, 338, "saryon")
	b.want(out(res), "Saryon is no longer stunned.")
	if q.Stunned {
		t.Error("still stunned")
	}

	var ore int
	for n, d := range b.e.items {
		if d.Type == "ORE" && b.e.nouns[d.NameID] == "ore" {
			ore = n
		}
	}
	if ore == 0 {
		t.Fatal("no ore")
	}
	p.Inventory = []InventoryItem{{Archetype: ore, Val1: 73}}
	b.want(out(b.cast(p, 407, "my", "ore")), "Purity: 73%")
	p.Inventory = nil
	b.e.rooms[batchForest].Items = []gameworld.RoomItem{{Ref: 0, Archetype: ore, Val1: 12}}
	b.want(out(b.cast(p, 407, "ore")), "Purity: 12%")
}

func TestTimedDruidSpells(t *testing.T) {
	b := newBatchEnv(t)
	p, q := b.player("Druid"), b.player("Archer")

	// True Aim: +20 ranks for a bow, nothing for a sword; resets rather than adds.
	var bow, sword *gameworld.ItemDef
	for _, d := range b.e.items {
		switch d.Type {
		case "BOW_WEAPON":
			bow = d
		case "SLASH_WEAPON":
			sword = d
		}
	}
	base, baseS := b.e.playerAttackRating(q, bow), b.e.playerAttackRating(q, sword)
	b.cast(p, 512, "archer")
	if d := b.e.playerAttackRating(q, bow) - base; d != 100 {
		t.Errorf("true aim bow +%d", d)
	}
	if b.e.playerAttackRating(q, sword) != baseS {
		t.Error("true aim helped a sword")
	}
	b.cast(p, 512, "archer")
	if left := time.Until(q.TrueAimExpiry); left > 31*time.Second || left < 9*time.Second {
		t.Errorf("true aim left %v", left)
	}
	q.TrueAimExpiry = time.Now().Add(-time.Second)
	b.e.spellEffectTick()
	b.want(b.sent["Archer"], "The gleam leaves your eyes.")

	// Free Action stops snares, and stacks.
	b.cast(p, 528, "archer")
	first := q.FreeActionExpiry
	b.cast(p, 528, "archer")
	if q.FreeActionExpiry.Sub(first) < 19*time.Minute {
		t.Error("free action didn't extend")
	}
	for _, id := range []int{webSpellID, plantSnareSpellID, wallOfThornsSpellID, creepingDeathSpellID} {
		if msg, _ := snareImmunity(q, FindSpellByID(id)); msg == "" {
			t.Errorf("free action let %d hold", id)
		}
	}
	b.cast(p, 500, "archer")
	if len(q.Entangles) != 0 {
		t.Error("plant snare held through free action")
	}

	// Ride the Lightning: no defense, all electricity becomes mana, past the maximum.
	q.Mana, q.MaxMana = 100, 100
	b.cast(p, 532, "archer")
	if q.DefenseBonus != 0 {
		t.Errorf("defense %d", q.DefenseBonus)
	}
	m := b.monster(116, batchForest)
	m.CastSpellID, m.CastTarget = 103, "Archer"
	bp := q.BodyPoints
	for i := 0; i < 60 && q.Mana == 100; i++ {
		m.CastSpellID, m.CastTarget = 103, "Archer"
		b.e.resolveMonsterCast(m, b.e.monsters[m.DefNumber])
	}
	if q.Mana <= 100 || q.BodyPoints != bp {
		t.Errorf("mana %d bp %d/%d", q.Mana, q.BodyPoints, bp)
	}
	if s, _ := absorbLightning(q, 5, "heat"); s != "" {
		t.Error("absorbed heat")
	}

	// Claws of the Elder Wolf: claws, and only claws, hit anything.
	hard := &gameworld.MonsterDef{MagicWeapon: 3}
	claw := b.e.items[clawGrowthWeaponArch]
	q.Wielded = nil
	if checkMagicWeapon(q, nil, claw, hard) {
		t.Error("claws already hit")
	}
	b.cast(p, 534, "archer")
	if !checkMagicWeapon(q, nil, claw, hard) || checkMagicWeapon(q, nil, nil, hard) || checkMagicWeapon(q, &InventoryItem{}, sword, hard) {
		t.Error("elder wolf claws")
	}
	q.WolfForm = true
	if !checkMagicWeapon(q, nil, nil, hard) {
		t.Error("a wolf's own claws")
	}
}

func TestFormLockAndWolfForm(t *testing.T) {
	b := newBatchEnv(t)
	p, q := b.player("Druid"), b.player("Mage")
	q.MistForm = true
	b.cast(p, 535, "some mist")
	if left := time.Until(q.FormLockExpiry); left < 290*time.Second || left > 5*time.Minute {
		t.Fatalf("form lock %v", left)
	}
	b.cast(p, 535, "some mist")
	if time.Until(q.FormLockExpiry) > 5*time.Minute {
		t.Error("form lock stacked")
	}
	b.do(q, "transform")
	if !q.MistForm {
		t.Error("transformed while locked")
	}
	q.FormLockExpiry = time.Time{}
	b.do(q, "transform")
	if q.MistForm {
		t.Error("couldn't transform once free")
	}

	// Wolf Form: good druids only, self only, and TRANSFORM ends it.
	if sp := FindSpellByID(536); sp == nil || sp.Level != 26 || sp.School != "Druidic" {
		t.Fatalf("536: %+v", sp)
	}
	p.KnownSpells[536] = true
	for _, al := range []int{-5, 0} {
		p.Alignment = al
		b.want(b.do(p, "prepare 536"), "good heart")
		p.PreparedSpell = 536
		p.RoundTimeExpiry = time.Time{}
		b.e.doCastSpell(b.ctx, p, nil)
		if p.WolfForm {
			t.Errorf("alignment %d became a wolf", al)
		}
	}
	p.Alignment = 1
	b.do(p, "prepare 536")
	if p.PreparedSpell != 536 {
		t.Fatal("couldn't prepare")
	}
	snarePlayer(p, FindSpellByID(webSpellID))
	b.cast(p, 536, "mage")
	if !p.WolfForm || q.WolfForm || len(p.Entangles) != 0 || p.DisplayName() != "a wolf" {
		t.Errorf("wolf %v %v", p.WolfForm, q.WolfForm)
	}
	b.do(p, "transform")
	if p.WolfForm {
		t.Error("still a wolf")
	}
	b.want(b.do(p, "transform"), "Only wolflings")
}

func TestTreeDoor(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Druid")
	p.Marks = map[int]int{1: batchIndoor, 2: 90003}
	b.e.rooms[90003] = &gameworld.Room{Number: 90003, Name: "Jungle", Terrain: "JUNGLE", Exits: map[string]int{}}
	b.want(out(b.cast(p, 531, "1")), "no tree at that mark")
	if p.PreparedSpell != 531 || p.RoomNumber != batchForest {
		t.Error("a refused tree door was spent")
	}
	var res *CommandResult
	for p.RoomNumber == batchForest { // not as a GM, whose comings and goings aren't announced
		p.RoundTimeExpiry, p.PreparedSpell, p.Mana, p.BodyPoints = time.Time{}, 531, 1000, 400
		res = b.e.doCastSpell(b.ctx, p, []string{"2"})
	}
	b.want(out(res), "The tree accepts you as you become one with it...")
	if p.RoomNumber != 90003 || res.OldRoom != batchForest || len(res.OldRoomMsg) != 1 || !strings.Contains(res.OldRoomMsg[0], "sinks into the tree and is gone.") {
		t.Fatalf("room %d", p.RoomNumber)
	}
	// A tree standing in an indoor room will do, at either end.
	b.e.rooms[batchIndoor].Items = []gameworld.RoomItem{{Archetype: 1457}}
	b.cast(p, 531, "1")
	if p.RoomNumber != batchIndoor {
		t.Error("tree object at the mark")
	}
	b.cast(p, 531, "2")
	if p.RoomNumber != 90003 {
		t.Error("tree object here")
	}
	b.e.rooms[batchIndoor].Items = nil
	p.RoomNumber = batchIndoor
	b.want(out(b.cast(p, 531, "2")), "no tree here")
}

func TestWitherLimb(t *testing.T) {
	b := newBatchEnv(t)
	p, q := b.player("Necro"), b.player("Victim")
	// It can be resisted, and submitting makes that much less likely; no need to submit.
	resists := func() (n int) {
		for i := 0; i < 400; i++ {
			q.WitheredLimbs = nil
			p.PreparedSpell, p.Mana, p.RoundTimeExpiry, p.BodyPoints = 351, 1000, time.Time{}, 400
			p.Wounds, p.Unconscious, p.Dead = nil, false, false
			if res := b.e.doCastSpell(b.ctx, p, []string{"victim"}); res.Resisted {
				n++
				if len(q.WitheredLimbs) != 0 || !strings.Contains(out(res), "Failure.") {
					t.Fatal("withered though resisted")
				}
			}
		}
		return n
	}
	p.Skills[23] = 10
	q.Level, q.Willpower = 30, 100
	unwilling := resists()
	q.Submitting = true
	willing := resists()
	if unwilling < 80 || unwilling > 380 || willing > 45 {
		t.Errorf("of 400 casts resisted %d, submitting %d", unwilling, willing)
	}
	q.WitheredLimbs, p.Skills[23], q.Submitting = nil, 40, false
	var shield, weapon int
	for n, d := range b.e.items {
		if d.Type == "SHIELD" && b.e.nouns[d.NameID] == "shield" {
			shield = n
		}
		if d.Type == "SLASH_WEAPON" {
			weapon = n
		}
	}
	q.Wielded = &InventoryItem{Archetype: weapon}
	q.OffHand = &InventoryItem{Archetype: shield}
	for i := 0; i < 3; i++ {
		b.cast(p, 351, "victim")
	}
	arms, legs := 0, 0
	for _, l := range q.WitheredLimbs {
		if strings.HasSuffix(l, "arm") {
			arms++
		} else {
			legs++
		}
	}
	if arms != 1 || legs != 1 {
		t.Fatalf("withered %v", q.WitheredLimbs)
	}
	if q.OffHand != nil || q.Wielded == nil || len(q.Inventory) != 1 {
		t.Error("shield not put away")
	}
	b.want(b.do(q, "wield shield"), "withered")
	if q.OffHand != nil {
		t.Error("wielded a shield")
	}
	q.Inventory = nil
	look := out(b.e.doLookAt(b.ctx, p, []string{"victim"}))
	b.want(look, "withered "+q.WitheredLimbs[0])
	b.want(look, "withered "+q.WitheredLimbs[1])

	q.Submitting = false
	from := q.RoomNumber
	b.want(b.do(q, "north"), "Your withered leg slows you. [Round: 2 sec]")
	if left := time.Until(q.RoundTimeExpiry); q.RoomNumber == from || left < time.Second || left > 2*time.Second {
		t.Errorf("room %d round time %v", q.RoomNumber, left)
	}
	if strings.Contains(b.do(q, "look"), "slows you") {
		t.Error("looking cost round time")
	}
	b.do(q, "north")
	res := b.cast(p, 343, "victim")
	b.want(strings.Join(res.TargetMsg, "\n"), "Life flows back into your")
	if len(q.WitheredLimbs) != 0 {
		t.Error("regeneration didn't restore")
	}

	// A follower's withered leg slows them too.
	q.WitheredLimbs = []string{"left leg"}
	p.IsGroupLeader, p.GroupMembers, q.Following = true, []string{"Victim"}, "Necro"
	q.RoundTimeExpiry = time.Time{}
	b.do(p, "north")
	if q.RoomNumber != p.RoomNumber || !q.RoundTimeExpiry.After(time.Now()) {
		t.Error("follower")
	}

	// Monsters: an arm and a leg, one of each.
	m := b.monster(49, p.RoomNumber) // kvantar/goblin shaman: HUMAN body
	id := m.ID
	for i := 0; i < 4; i++ {
		b.cast(p, 351, "shaman")
	}
	m = &b.e.monsterMgr.instances[b.e.monsterMgr.indexOfID(id)]
	if !m.WitheredArm || !m.WitheredLeg || monsterEffectiveAttack(b.e.monsters[49], m) != b.e.monsters[49].Attack1-witheredArmAttackPenalty {
		t.Errorf("monster %v %v", m.WitheredArm, m.WitheredLeg)
	}
}

func TestSnares(t *testing.T) {
	b := newBatchEnv(t)
	p, q := b.player("Druid"), b.player("Friend")
	m := b.monster(117, batchForest) // ursine
	id := m.ID
	mon := func() *MonsterInstance { return &b.e.monsterMgr.instances[b.e.monsterMgr.indexOfID(id)] }

	// Creeping Death holds and hurts; lasts a minute.
	b.want(out(b.cast(p, 526, "ursine")), "coil around an ursine")
	hp := mon().CurrentHP
	b.e.insectSwarmStep()
	if d := hp - mon().CurrentHP; d < 3 || d > 7 || !mon().heldInPlace() {
		t.Errorf("creeping death did %d", d)
	}
	mon().CreepExpiry = time.Now().Add(-time.Second)
	b.e.insectSwarmStep()
	if mon().Creeping {
		t.Error("creeping death didn't end")
	}

	// Plant Snare on a monster: held, can't close in, can fight what's already here.
	mon().Swarmed, mon().SwarmExpiry = true, time.Now().Add(time.Minute)
	got := out(b.cast(p, 500, "ursine"))
	b.want(got, "Plant tendrils shoot up from the ground and ensnare an ursine!")
	b.want(got, "The insects swarming over an ursine drop off and die.")
	b.want(strings.Join(b.e.monsterNamesInRoom(batchForest), ""), "ursine (ensnared)")
	b.e.monsterCheckAggro(q, batchForest)
	if mon().Target != "" || !mon().heldInPlace() || mon().Swarmed {
		t.Error("plant snare doesn't hold it")
	}
	// It still fights whoever has closed with it (its attack, or its special attack
	// at anyone in the room).
	mon().Target = "Friend"
	b.e.monsterMgr.mu.Lock()
	b.sent["Friend"], b.sent["Druid"], b.room[batchForest] = "", "", ""
	b.e.monsterCombatTick(mon(), b.e.monsters[117])
	b.e.monsterMgr.mu.Unlock()
	b.want(b.sent["Friend"]+b.sent["Druid"]+b.room[batchForest], "ursine")
	for _, x := range []*Player{p, q} {
		x.BodyPoints, x.Position, x.Stunned, x.KnockedDown, x.Wounds, x.Unconscious, x.Dead = 400, 0, false, false, nil, false, false
	}
	mon().Target = ""
	mon().SnareExpiry = time.Now().Add(-time.Second)
	b.e.insectSwarmStep()
	b.want(b.room[batchForest], "The plant tendrils ensnaring an ursine rapidly wither and turn to dust.")
	if mon().heldInPlace() {
		t.Error("plant snare never ends")
	}
	p.RoomNumber = batchIndoor
	b.want(out(b.cast(p, 500, "ursine")), "only works outdoors")
	p.RoomNumber = batchForest

	// Wall of Thorns on a monster: held, can't close in, hurt by attacking.
	b.cast(p, 516, "ursine")
	if !mon().heldInPlace() || !b.e.snareHoldsBack(mon(), b.e.monsters[117]) {
		t.Fatal("thorns don't hold it")
	}
	b.e.monsterCheckAggro(q, batchForest)
	if mon().Target != "" {
		t.Error("advanced through the thorns")
	}
	hp = mon().CurrentHP
	mon().Target = "Friend"
	for i := 0; i < 50 && mon().CurrentHP == hp; i++ { // a special attack doesn't cost it anything
		q.Dead, q.Unconscious, q.Position, q.BodyPoints = false, false, 0, 400
		b.e.monsterMgr.mu.Lock()
		b.e.monsterCombatTick(mon(), b.e.monsters[117])
		b.e.monsterMgr.mu.Unlock()
	}
	if d := hp - mon().CurrentHP; d < 3 || d > 9 {
		t.Errorf("thorns did %d", d)
	}
	mon().Target = ""
	q.Dead, q.Unconscious, q.Position, q.Stunned, q.KnockedDown, q.BodyPoints, q.Wounds = false, false, 0, false, false, 400, nil

	// On a player: no leaving, no advancing, no reach, hurt fighting what closed in.
	q.BodyPoints = 400
	snarePlayer(q, FindSpellByID(516))
	for _, cmd := range []string{"north", "advance ursine", "go north"} {
		b.want(b.do(q, cmd), "bound by a wall of thorns")
	}
	b.want(b.do(q, "attack ursine"), "melee range")
	mon().Target = "Friend"
	b.want(b.do(q, "attack ursine"), "The thorns tear at you")
	b.want(b.do(q, "retreat"), "wall of thorns")
	if q.BodyPoints >= 400 || q.BodyPoints < 390 {
		t.Errorf("bp %d", q.BodyPoints)
	}
	var bow, pole int
	for n, d := range b.e.items {
		if d.Type == "POLE_WEAPON" {
			pole = n
		}
		if d.Type == "THROWN_WEAPON" {
			bow = n
		}
	}
	mon().Target = ""
	bp := q.BodyPoints
	q.Wielded = &InventoryItem{Archetype: pole}
	b.want(b.do(q, "attack ursine"), "wall of thorns keeps you from reaching")
	q.Wielded = &InventoryItem{Archetype: bow}
	got = b.do(q, "attack ursine")
	if strings.Contains(got, "thorns") || q.BodyPoints != bp {
		t.Errorf("throwing cost: %s", got)
	}
	// It runs out, and portals are closed to the held as well.
	q.Entangles[0].Expiry = time.Now().Add(-time.Second)
	b.e.insectSwarmStep()
	if len(q.Entangles) != 0 {
		t.Error("thorns never end")
	}
}

func TestEnergyMaelstrom(t *testing.T) {
	b := newBatchEnv(t)
	p, q := b.player("Chandra"), b.player("Rion")
	p.KnownSpells[121] = true
	for i := 0; i < 8; i++ {
		b.monster(117, batchForest)
	}
	got := out(b.cast(p, 138))
	b.want(got, "A swirling vortex of black mists and crackling energy arises!")
	b.want(got, "Rion is knocked over!")
	b.want(got, "An ursine is knocked over!")
	b.want(b.sent["Rion"], "You are knocked over!")
	if q.Position != 2 || p.Position != 0 {
		t.Error("positions")
	}
	b.want(b.do(p, "prepare 121"), "You cannot prepare a spell while you maintain your current magic.")
	b.want(strings.Join(b.e.monsterNamesInRoom(batchForest), ", "), "(laying)")
	p.RoomNumber = batchIndoor // it stays behind
	hp := 0
	for _, m := range b.e.monsterMgr.MonstersInRoom(batchForest) {
		hp += m.CurrentHP
	}
	bp := q.BodyPoints
	b.room[batchForest] = ""
	b.e.maelstromTick()
	b.want(b.room[batchForest], "A black maelstrom of wind and energy rages!")
	after := 0
	for _, m := range b.e.monsterMgr.MonstersInRoom(batchForest) {
		after += m.CurrentHP
	}
	if after >= hp || q.BodyPoints != bp || strings.Contains(b.room[batchIndoor], "maelstrom") {
		t.Errorf("hp %d -> %d, bp %d", hp, after, q.BodyPoints)
	}
	b.e.maelstroms.start("Chandra", maelstrom{room: batchForest, expiry: time.Now().Add(-time.Second)})
	b.e.maelstromTick()
	b.do(p, "prepare 121")
	if p.PreparedSpell != 121 {
		t.Error("still maintaining")
	}
}

func TestChaoticSummons(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Conjurer")
	p.KnownSpells[144] = true
	p.Inventory = []InventoryItem{{Archetype: 521, Adj2: 1}}
	b.do(p, "prepare 144 with tongue")
	if p.PreparedSpell != 0 {
		t.Fatal("any tongue will do")
	}
	p.Inventory = []InventoryItem{{Archetype: 521, Adj2: 868}}
	b.want(b.do(p, "prepare 144 with ursine tongue"), "You prepare the Tindarath's Chaotic Summons spell.")
	if len(p.Inventory) != 0 {
		t.Error("tongue not consumed")
	}
	b.want(out(b.cast(p, 144)), "fire tiger")
	b.want(b.do(p, "command guard me"), "acknowledges your command")
}

// TestMonsterUtilitySpells: monsters' non-damage spells have their own effects instead
// of going off as a 1-point bolt of energy.
func TestMonsterUtilitySpells(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Moordread")
	spider := b.monster(20, batchForest)
	def := b.e.monsters[20]
	seen := ""
	resolve := func(spell int) (string, string) {
		for i := 0; i < 300; i++ {
			spider.CastSpellID, spider.CastTarget = spell, "Moordread"
			pm, rm, _ := b.e.resolveMonsterCast(spider, def)
			room := strings.Join(rm, "\n") + b.sent["Moordread"]
			b.sent["Moordread"] = ""
			if strings.Contains(room, "The web misses!") {
				seen = room
				continue
			}
			return strings.Join(pm, "\n"), room
		}
		t.Fatal("never succeeded")
		return "", ""
	}
	bp := p.BodyPoints
	self, room := resolve(127)
	b.want(self, "A giant spider shoots its web at Moordread.")
	b.want(self, "You are covered with strands of sticky webbing!")
	b.want(room, "Moordread is covered with strands of sticky webbing!")
	b.want(b.do(p, "north"), "bound by sticky webbing")
	for i := 0; i < 300 && seen == ""; i++ {
		resolve(127)
	}
	b.want(seen, "The web misses!")
	_, room = resolve(230)
	b.want(room, "A wave of deep red light emanates from a giant spider.")
	if len(p.Entangles) != 0 {
		t.Error("disjunction left the web")
	}
	p.RepelPlantsAndWebsExpiry = time.Now().Add(time.Minute)
	resolve(127)
	if len(p.Entangles) != 0 {
		t.Error("webbed through Repel Plants and Webs")
	}
	self, _ = resolve(200)
	b.want(self, "You are filled with terror!")
	resolve(211)
	if !spellActive(p.SlowExpiry) {
		t.Error("slow")
	}
	resolve(201)
	b.want(b.do(p, "attack spider"), "can't bring yourself to harm")
	p.CharmExpiry = time.Time{}
	self, _ = resolve(218)
	b.want(self, "You fall asleep.")
	b.want(b.do(p, "north"), "unconscious")
	if wakeFromSlumber(p); p.Unconscious || p.Position != 2 {
		t.Error("waking")
	}
	resolve(216) // 80 body points at most
	if p.Unconscious {
		t.Error("slumber I felled 400 body points")
	}
	p.Position = 0
	resolve(231)
	if !p.IsImprisoned() || len(b.e.monsterAttackTargetPool(spider)) != 0 {
		t.Error("imprison")
	}
	p.ImprisonedExpiry = time.Time{}
	resolve(519)
	if !p.Stunned {
		t.Error("sunray")
	}
	resolve(219)
	if !p.IsSilenced() {
		t.Error("silence")
	}
	for _, id := range []int{102, 105, 326, 511, 210} {
		if !b.e.monsterCanUseSpell(spider, FindSpellByID(id)) {
			t.Errorf("can't use %d", id)
		}
		before := monsterEffectiveDefense(def, spider)
		for i := 0; i < 300 && b.e.monsterCanUseSpell(spider, FindSpellByID(id)); i++ {
			spider.CastSpellID, spider.CastTarget = id, "Moordread"
			b.e.resolveMonsterCast(spider, def)
		}
		if b.e.monsterCanUseSpell(spider, FindSpellByID(id)) || (id != 210 && monsterEffectiveDefense(def, spider) <= before) {
			t.Errorf("self buff %d", id)
		}
	}
	if p.BodyPoints != bp {
		t.Errorf("took %d damage", bp-p.BodyPoints)
	}
	if b.e.monsterCanUseSpell(spider, FindSpellByID(225)) {
		t.Error("would cast Invisibility")
	}
	// Every spell a monster knows is either usable or one we know it can't use.
	for _, d := range b.e.monsters {
		for _, id := range d.Spells {
			sp := FindSpellByID(id)
			fresh := MonsterInstance{RoomNumber: batchForest}
			if sp != nil && id != 225 && !b.e.monsterCanUseSpell(&fresh, sp) {
				t.Errorf("monster %d (%s) can't use %s (%d)", d.Number, d.Name, sp.Name, id)
			}
		}
	}
}

// TestScriptCastSpells: SPELL n in a script or item gives the player these effects.
func TestScriptCastSpells(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Player")
	p.Wielded = &InventoryItem{Archetype: 1}
	spell := func(id int) string {
		sc := &ScriptContext{Player: p, Engine: b.e, Room: b.e.rooms[batchForest]}
		sc.doScriptSpell([]string{strconv.Itoa(id), "30"})
		return strings.Join(sc.Messages, "\n")
	}
	spell(351)
	if len(p.WitheredLimbs) != 1 {
		t.Error("wither")
	}
	spell(343)
	if len(p.WitheredLimbs) != 0 {
		t.Error("regeneration")
	}
	for _, id := range []int{528, 532, 534, 512, 535} {
		spell(id)
	}
	if !spellActive(p.FreeActionExpiry) || !spellActive(p.RideLightningExpiry) || !spellActive(p.ElderWolfExpiry) || !spellActive(p.TrueAimExpiry) || !spellActive(p.FormLockExpiry) {
		t.Error("timed effects")
	}
	spell(536)
	if p.WolfForm {
		t.Error("wolf form under form lock")
	}
	b.want(spell(230), "A wave of deep red light emanates from you.")
	if spellActive(p.FormLockExpiry) || spellActive(p.FreeActionExpiry) {
		t.Error("disjunction")
	}
	spell(536)
	if !p.WolfForm || p.Wielded != nil {
		t.Error("wolf form")
	}
	p.Invisible = true
	spell(406)
	if p.Invisible {
		t.Error("dispel invisibility")
	}
	spell(516)
	if !hasEntangle(p, 516) {
		t.Error("thorns")
	}
	b.want(spell(127), "You are covered with strands of sticky webbing!")
	b.want(spell(500), "Plant tendrils shoot up from the ground and ensnare you!")
	b.want(spell(200), "You are filled with terror!")
	b.want(spell(201), "can't imagine wanting to harm anyone")
	if !charmedBy(p, 7) {
		t.Error("charm")
	}
	b.want(spell(211), "You slow down considerably.")
	if left := time.Until(p.SlowExpiry); left < 26*time.Minute { // cast level 30
		t.Errorf("slow %v", left)
	}
	b.want(spell(219), "you have been silenced")
	b.want(spell(231), "A blue force bubble envelops you!")
	b.want(spell(519), "you reel, stunned!")
	if got := spell(226); got != "" || !spellActive(p.ParanoiaExpiry) {
		t.Errorf("paranoia %q", got)
	}
	p.MaxBodyPoints, p.BodyPoints, p.Stunned = 150, 150, false
	b.want(spell(217), "You fall asleep.")
	if wakeFromUnconscious(p); !p.Unconscious {
		t.Error("regeneration cut the sleep short")
	}
	p.SleepExpiry = time.Now().Add(-time.Second)
	b.e.spellEffectTick()
	b.want(b.sent["Player"], "You wake up!")
	// All of it goes with a disjunction.
	spell(230)
	if len(p.Entangles) != 0 || spellActive(p.FearExpiry) || charmedBy(p, 7) || spellActive(p.SlowExpiry) || p.IsSilenced() || p.IsImprisoned() || spellActive(p.ParanoiaExpiry) {
		t.Error("left after disjunction")
	}
	// Cloak Mind wards off the mind-affecting ones.
	applyTimedDefenseBuff(p, FindSpellByID(235), 1)
	for _, id := range []int{200, 201, 217, 226} {
		spell(id)
	}
	if spellActive(p.FearExpiry) || spellActive(p.CharmExpiry) || p.Unconscious || spellActive(p.ParanoiaExpiry) {
		t.Error("cloak mind")
	}
	// Ride the Lightning takes a script's lightning too.
	spell(532)
	bp := p.BodyPoints
	if spell(103); p.BodyPoints != bp {
		t.Error("script lightning hurt")
	}
}

// TestScriptCuresAndForms: the cures, and the spells random potions can hold, work from
// a script and from a drunk potion.
func TestScriptCuresAndForms(t *testing.T) {
	b := newBatchEnv(t)
	p := b.player("Drinker")
	spell := func(id int) string {
		sc := &ScriptContext{Player: p, Engine: b.e, Room: b.e.rooms[batchForest]}
		sc.doScriptSpell([]string{strconv.Itoa(id), "30"})
		return strings.Join(sc.Messages, "\n") + "/" + strings.Join(sc.RoomMsgs, "\n")
	}
	poisonPlayer(p, 9)
	diseasePlayer(p, 7)
	b.want(spell(303), "A blue light flashes around you.\nThe poison leaves you.")
	if p.Poisoned || p.PoisonLevel != 0 || !p.Diseased {
		t.Error("cure poison")
	}
	b.want(spell(319), "A purple light flashes around you.\nThe disease leaves you.")
	if p.Diseased || p.DiseaseLevel != 0 || strings.Contains(spell(319), "leaves you") {
		t.Error("cure disease")
	}
	b.want(spell(403), "You feel the power of your mind expanding.")
	if !p.TelepathyActive {
		t.Error("mindlink")
	}
	b.want(spell(518), "Wicked claws erupt")
	b.want(spell(518), "minutes remaining")

	// Every spell a random potion can hold does something when drunk.
	drop := b.e.randomPotionDrop(40)
	drink := func(id int) string {
		p.Inventory = []InventoryItem{{Archetype: drop.Archetype, Val2: 3, Val3: id, State: "OPEN"}}
		return b.do(p, "drink "+b.e.nouns[b.e.items[drop.Archetype].NameID])
	}
	for _, id := range potionSpellIDs {
		q := b.player("Other")
		*p, q.RoomNumber = *q, batchIndoor
		p.FirstName = "Drinker"
		p.BodyPoints = 200
		if got := drink(id); strings.Contains(got, "The item casts "+FindSpellByID(id).Name+".\n") ||
			strings.HasSuffix(got, "The item casts "+FindSpellByID(id).Name+".") || strings.Contains(got, "coming soon") || !strings.Contains(got, "sip") {
			t.Errorf("potion of %s: %s", FindSpellByID(id).Name, got)
		}
		switch {
		case id == 232 && !p.MistForm, id == 245 && !p.SlimeForm, id == 248 && !p.PhantomForm,
			id == 403 && !p.TelepathyActive, id == 518 && !spellActive(p.ClawGrowthExpiry):
			t.Errorf("potion of %s had no effect", FindSpellByID(id).Name)
		}
	}

	// A form cast on someone held frees them; a locked form can't be changed; TRANSFORM ends it.
	*p = *b.player("Again")
	snarePlayer(p, FindSpellByID(webSpellID))
	b.want(spell(245), "Again suddenly becomes a slime!")
	b.want(spell(535), "fixed")
	b.want(spell(232), "already in another form")
	if len(p.Entangles) != 0 || !p.SlimeForm || p.MistForm {
		t.Error("slime form")
	}
	p.FormLockExpiry = time.Time{}
	b.do(p, "transform")
	lockForm(p)
	if spell(232); p.SlimeForm || p.MistForm {
		t.Error("form lock or transform")
	}
	// The cast spell still works.
	p.FormLockExpiry = time.Time{}
	res := b.cast(p, 232)
	b.want(out(res), "you are a cloud of mist!")
	b.want(strings.Join(res.RoomBroadcast, ""), "Again suddenly becomes a cloud of mist!")
}
