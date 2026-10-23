package engine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
)

// STALK <player|creature> (LEGENDS.DOC: "Attempt to secretly track someone.").
//
// Nothing else about it survives in the docs, scripts or captures. Reconstructed
// from how the game owner remembers it: a Stealth-skill FOLLOW that doesn't join
// the target's group and makes no echoes. Every time the target moves, the stalker
// rolls Stealth against the target's Perception; a failure loses the trail.
// The target is never told. A hidden stalker stays hidden while following.

const (
	stalkBaseChance        = 50
	stalkMinChance         = 5
	stalkMaxChance         = 95
	stalkMonsterPerception = 50 // monsters have no Perception stat; an average one
)

// stalkChance is the percent chance stalker keeps up with a target of the given
// Perception into destRoom: 50 + Stealth×5 + Agility/10 − Perception/5, plus the
// destination's poor-light concealment (see roomConcealmentBonus).
func (e *GameEngine) stalkChance(stalker *Player, targetPerception, destRoom int) int {
	chance := stalkBaseChance + effectiveStealthSkill(stalker)*5 + stalker.Agility/10 +
		unemoteSkillBonus(stalker) + e.roomConcealmentBonus(destRoom) - targetPerception/5
	return max(stalkMinChance, min(chance, stalkMaxChance))
}

func clearStalk(p *Player) {
	p.StalkingPlayer, p.StalkingMonster, p.StalkingName = "", 0, ""
}

func (p *Player) isStalking() bool {
	return p.StalkingPlayer != "" || p.StalkingMonster != 0
}

// doStalk handles STALK <target>, and bare STALK to stop.
func (e *GameEngine) doStalk(player *Player, args []string) *CommandResult {
	if len(args) == 0 {
		if !player.isStalking() {
			return &CommandResult{Messages: []string{"Stalk whom?"}}
		}
		name := player.StalkingName
		clearStalk(player)
		return &CommandResult{Messages: []string{fmt.Sprintf("You stop stalking %s.", name)}}
	}
	if effectiveStealthSkill(player) < 1 {
		return &CommandResult{Messages: []string{"You need training in Stealth to stalk someone."}}
	}
	if player.Following != "" {
		return &CommandResult{Messages: []string{"You can't stalk anyone while you're following a group."}}
	}
	target := strings.ToLower(strings.Join(args, " "))
	if p := e.findPlayerInRoom(player, target); p != nil {
		if p.StalkingPlayer == player.FirstName {
			return &CommandResult{Messages: []string{"You can't stalk someone who is stalking you."}}
		}
		clearStalk(player)
		player.StalkingPlayer, player.StalkingName = p.FirstName, p.DisplayName()
		return &CommandResult{Messages: []string{fmt.Sprintf("You begin stalking %s.", player.StalkingName)}}
	}
	if inst, def := e.findMonsterInRoom(player, target); inst != nil {
		name := strings.ToLower(FormatMonsterName(def, e.monAdjs))
		clearStalk(player)
		player.StalkingMonster, player.StalkingName = inst.ID, articleFor(name, def.Unique)+name
		return &CommandResult{Messages: []string{fmt.Sprintf("You begin stalking %s.", player.StalkingName)}}
	}
	return &CommandResult{Messages: []string{"You don't see that here."}}
}

// moveStalkers brings along everyone in originalRoom who is stalking the target
// that just moved to destNum (matches(p) picks them out), rolling each one's
// Stealth against targetPerception. Movement is silent both ways.
func (e *GameEngine) moveStalkers(ctx context.Context, originalRoom, destNum, targetPerception int, matches func(*Player) bool) {
	dest := e.rooms[destNum]
	if dest == nil || e.sessions == nil {
		return
	}
	for _, p := range e.sessions.OnlinePlayers() {
		if !matches(p) {
			continue
		}
		if p.RoomNumber != originalRoom || p.Dead {
			clearStalk(p) // got separated some other way (teleport, scripted move...)
			continue
		}
		name := p.StalkingName
		if followBlockedReason(p) != "" || e.isEngagedInMelee(p) || p.Immobilized || len(p.Entangles) > 0 ||
			rand.Intn(100) >= e.stalkChance(p, targetPerception, destNum) {
			clearStalk(p)
			if e.sendToPlayer != nil {
				e.sendToPlayer(p.FirstName, []string{fmt.Sprintf("You lose track of %s.", name)})
			}
			continue
		}
		p.RoomNumber = destNum
		p.Submitting = false
		e.disengageCombat(p)
		p.GuardTargets, p.GuardPortals, p.GuardItems = nil, nil, nil
		e.SavePlayer(ctx, p)
		if e.sendToPlayer != nil {
			msgs := append([]string{fmt.Sprintf("You stealthily follow %s.", name)}, e.doLook(p).Messages...)
			e.sendToPlayer(p.FirstName, msgs)
		}
		entry := &CommandResult{}
		e.applyEntryScripts(ctx, p, dest, entry)
		if len(entry.Messages) > 0 && e.sendToPlayer != nil {
			e.sendToPlayer(p.FirstName, entry.Messages)
		}
	}
}

// movePlayerStalkers is called after a player walks from originalRoom to destNum.
func (e *GameEngine) movePlayerStalkers(ctx context.Context, target *Player, originalRoom, destNum int) {
	e.moveStalkers(ctx, originalRoom, destNum, target.Perception+target.cyberHearingBonus(), func(p *Player) bool {
		return p.StalkingPlayer == target.FirstName
	})
}

// moveMonsterStalkers is called when monster instance monID wanders or flees.
// Monster moves happen inside monsterTick with monsterMgr.mu held, and doLook
// needs that lock, so the stalkers move on their own goroutine once it's free.
func (e *GameEngine) moveMonsterStalkers(monID, originalRoom, destNum int) {
	go func() {
		defer recoverBackground("monster stalkers")
		e.moveStalkers(context.Background(), originalRoom, destNum, stalkMonsterPerception, func(p *Player) bool {
			return p.StalkingMonster == monID
		})
	}()
}
