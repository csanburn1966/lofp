package engine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Wither Limb (351, Necromancy level 24): withers an arm or a leg at random, for good —
// only Regeneration (343) restores it. A withered arm can't hold a shield or a second
// weapon; a withered leg adds round time to every move. It never takes both arms or both
// legs: cast at someone with a withered arm it takes a leg, and the other way about, and
// once they have one of each it does nothing more.
//
// A monster loses some of its attack with a withered arm, and can't wander off or flee
// on a withered leg. Nothing documents what it did to them.

const (
	witheredLegRoundTime     = 2
	witheredArmAttackPenalty = 15
)

// witheredLimb returns the player's withered "arm" or "leg" ("left arm"), or "".
func witheredLimb(p *Player, kind string) string {
	for _, l := range p.WitheredLimbs {
		if strings.HasSuffix(l, kind) {
			return l
		}
	}
	return ""
}

// witherPlayer withers an arm or a leg at random, whichever the player doesn't already
// have withered, returning its name ("" if they have one of each) and, for an arm, the
// shield or second weapon they had to put away. The caller saves the player.
func (e *GameEngine) witherPlayer(p *Player) (limb, dropped string) {
	var kinds []string
	for _, k := range []string{"arm", "leg"} {
		if witheredLimb(p, k) == "" {
			kinds = append(kinds, k)
		}
	}
	if len(kinds) == 0 {
		return "", ""
	}
	limb = []string{"right ", "left "}[rand.Intn(2)] + kinds[rand.Intn(len(kinds))]
	p.WitheredLimbs = append(p.WitheredLimbs, limb)
	if strings.HasSuffix(limb, "arm") && p.OffHand != nil {
		dropped = "the item in your other hand"
		if def := e.items[p.OffHand.Archetype]; def != nil {
			dropped = e.formatItemName(def, p.OffHand.Adj1, p.OffHand.Adj2, p.OffHand.Adj3, p.OffHand.Tail)
		}
		p.Inventory = append(p.Inventory, *p.OffHand)
		p.OffHand = nil
	}
	return limb, dropped
}

// witherMessages withers one of the player's limbs and describes it.
func (e *GameEngine) witherMessages(p *Player) (selfMsgs []string, roomMsg string) {
	limb, dropped := e.witherPlayer(p)
	if limb == "" {
		return []string{"You feel a chill, but nothing more."}, fmt.Sprintf("Nothing happens to %s.", p.DisplayName())
	}
	selfMsgs = []string{fmt.Sprintf("Your %s withers and shrivels!", limb)}
	if dropped != "" {
		selfMsgs = append(selfMsgs, fmt.Sprintf("You can no longer hold %s, and put it away.", dropped))
	}
	return selfMsgs, fmt.Sprintf("%s's %s withers and shrivels!", p.DisplayNameCap(), limb)
}

// restoreWitheredLimbs makes the player's withered limbs whole, returning their names
// ("left arm and right leg") or "" if there were none.
func restoreWitheredLimbs(p *Player) string {
	restored := joinWithAnd(p.WitheredLimbs)
	p.WitheredLimbs = nil
	return restored
}

// witheredLimbsText is "a withered left arm and a withered right leg".
func witheredLimbsText(p *Player) string {
	var parts []string
	for _, l := range p.WitheredLimbs {
		parts = append(parts, "a withered "+l)
	}
	return joinWithAnd(parts)
}

// witheredArmMessage refuses a shield or second weapon to a player with a withered
// arm, or returns "".
func witheredArmMessage(p *Player) string {
	if arm := witheredLimb(p, "arm"); arm != "" {
		return fmt.Sprintf("You can't hold that with your withered %s.", arm)
	}
	return ""
}

// witheredLegDelay adds the round time a withered leg costs a player who has just
// moved, on top of any the move already gave them. Returns the line to show, or "".
func witheredLegDelay(p *Player) string {
	if witheredLimb(p, "leg") == "" {
		return ""
	}
	rt := applyRoundTime(p, witheredLegRoundTime)
	start := time.Now()
	if p.RoundTimeExpiry.After(start) {
		start = p.RoundTimeExpiry
	}
	p.RoundTimeExpiry = start.Add(time.Duration(rt) * time.Second)
	total := int(time.Until(p.RoundTimeExpiry)/time.Second) + 1
	p.RoundTime = total
	return fmt.Sprintf("Your withered leg slows you. [Round: %d sec]", total)
}

// walkingVerbs are the commands that move a player on their own feet, besides the
// compass directions.
var walkingVerbs = map[string]bool{
	"GO": true, "CLIMB": true, "FLEE": true, "UP": true, "DOWN": true, "OUT": true,
	"NORTH": true, "SOUTH": true, "EAST": true, "WEST": true,
	"NORTHEAST": true, "NORTHWEST": true, "SOUTHEAST": true, "SOUTHWEST": true,
	"N": true, "S": true, "E": true, "W": true, "NE": true, "NW": true, "SE": true, "SW": true,
	"U": true, "D": true, "O": true,
	"SNEAK": true, "SWIM": true, "JUMP": true, "CRAWL": true, "ASCEND": true, "DESCEND": true,
}

// isWalkingCommand reports whether a typed command moves the player themselves, as
// opposed to a spell, a script or a GM moving them.
func isWalkingCommand(input string) bool {
	f := strings.Fields(input)
	if len(f) == 0 {
		return false
	}
	verb := strings.ToUpper(f[0])
	return walkingVerbs[verb] || walkingVerbs[resolveVerb(verb)]
}

// castWitherLimb handles Wither Limb (351) cast at a monster or a player.
func (e *GameEngine) castWitherLimb(player *Player, spell *SpellDef, args []string, castRoll, castChance int) *CommandResult {
	targetName := strings.Join(args, " ")
	if targetName == "" {
		targetName = e.autoTargetMonsterName(player)
	}
	if targetName == "" {
		return &CommandResult{Messages: []string{"Cast Wither Limb at what?"}, TargetNotFound: true}
	}
	inst, def, refusal := e.findHostileMonsterInRoom(player, targetName)
	if refusal != nil {
		return refusal
	}
	if inst == nil {
		target, refusal := e.hostileSpellAtPlayer(player, targetName)
		if refusal != nil {
			return refusal
		}
		selfMsgs, roomMsg := e.witherMessages(target)
		e.SavePlayer(context.Background(), target)
		if target == player {
			return &CommandResult{
				Messages:      append([]string{"You gesture."}, selfMsgs...),
				RoomBroadcast: []string{fmt.Sprintf("%s gestures.", player.DisplayNameCap()), roomMsg},
			}
		}
		return &CommandResult{
			Messages:      []string{fmt.Sprintf("You gesture at %s.", target.DisplayName()), roomMsg},
			RoomBroadcast: []string{fmt.Sprintf("%s gestures at %s.", player.DisplayNameCap(), target.DisplayName()), roomMsg},
			TargetName:    target.FirstName,
			TargetMsg:     append([]string{fmt.Sprintf("%s gestures at you.", player.DisplayNameCap())}, selfMsgs...),
		}
	}

	name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
	article := articleFor(name, def.Unique)
	gestureSelf := fmt.Sprintf("You gesture at %s%s.", article, name)
	gestureRoom := fmt.Sprintf("%s gestures at %s%s.", player.DisplayNameCap(), article, name)

	var can []string
	if !inst.WitheredArm && monsterHasArms(def) {
		can = append(can, "arm")
	}
	if !inst.WitheredLeg && monsterHasLegs(def) {
		can = append(can, "leg")
	}
	if def.Discorporate || len(can) == 0 {
		line := fmt.Sprintf("%s%s seems unaffected.", capArticle(article), name)
		return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}}
	}
	if magicResisted(player, castRoll, castChance, def.MagicResist) {
		line := fmt.Sprintf("The %s resists the spell!", name)
		return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}, Resisted: true}
	}
	limb := can[rand.Intn(len(can))]
	e.updateMonster(inst.ID, func(m *MonsterInstance) {
		if limb == "arm" {
			m.WitheredArm = true
		} else {
			m.WitheredLeg = true
		}
	})
	line := fmt.Sprintf("One of %s%s's %ss withers and shrivels!", article, name, limb)
	return &CommandResult{Messages: []string{gestureSelf, line}, RoomBroadcast: []string{gestureRoom, line}}
}

// monsterHasArms reports whether a creature has arms (GMSCRIPT.DOC parts 5 and 6).
func monsterHasArms(def *gameworld.MonsterDef) bool {
	return def.BodyType == "HUMAN" || containsInt(def.Parts, 5) || containsInt(def.Parts, 6)
}
