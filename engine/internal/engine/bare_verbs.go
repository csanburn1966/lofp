package engine

import "context"

// Room scripts react to a bare verb with IFPREVERB/IFVERB <verb> -1 (e.g.
// KARTHEM.SCR room 3404's "IFPREVERB DIG -1" frees a counterweight). Several
// command handlers never reach those scripts for a bare verb:
//
//   - Verbs that need an object refuse first ("Dig what?", "Climb what?"), so
//     neither the IFPREVERB nor the IFVERB block ever runs.
//   - Posture verbs and RECALL run the room's IFVERB blocks after acting, but
//     never its IFPREVERB blocks, so a room can't stop or replace them.
//
// runBareVerbRoomScripts covers both, for a verb typed with no arguments. It
// returns a finished result when the scripts handled the command, or nil plus
// any messages a non-blocking IFPREVERB block printed, for the caller to show
// ahead of the normal command output.

// bareVerbNeedsObject: handlers that refuse a bare verb without running any
// room script. Both the IFPREVERB and the IFVERB blocks run here, and if they
// print anything, that replaces the "<verb> what?" refusal.
var bareVerbNeedsObject = map[string]bool{
	"DIG": true, "CLIMB": true, "DRINK": true, "DROP": true, "GET": true, "TAKE": true,
	"PULL": true, "PUSH": true, "READ": true, "RUB": true, "THUMP": true, "TURN": true,
	"USE": true, "YELL": true, "GO": true, "BUY": true, "GIVE": true, "WIELD": true,
}

// bareVerbPreverbOnly: handlers that run the room's IFVERB blocks themselves
// but skip IFPREVERB. Only the IFPREVERB blocks run here.
var bareVerbPreverbOnly = map[string]bool{
	"SIT": true, "KNEEL": true, "LAY": true, "STAND": true, "PRAY": true, "RECALL": true,
}

// bareVerbAliases lists the spellings room scripts use for the same verb.
func bareVerbAliases(verb string) []string {
	switch verb {
	case "GET", "TAKE":
		return []string{"GET", "TAKE"}
	}
	return []string{verb}
}

func (e *GameEngine) runBareVerbRoomScripts(ctx context.Context, player *Player, verb string) (*CommandResult, []string, []string) {
	needsObject := bareVerbNeedsObject[verb]
	if !needsObject && !bareVerbPreverbOnly[verb] {
		return nil, nil, nil
	}
	room := e.rooms[player.RoomNumber]
	if room == nil {
		return nil, nil, nil
	}
	sc := &ScriptContext{Player: player, Room: room, Engine: e, activeRef: "-1"}
	run := func() {
		for _, alias := range bareVerbAliases(verb) {
			sc.activeVerb = alias
			for _, block := range room.Scripts {
				if scriptBlockHandlesVerb(block, alias) {
					sc.execBlock(block)
				}
				if sc.MoveTo > 0 || sc.MoveGroupTo > 0 {
					return
				}
			}
		}
	}
	sc.preverbOnly = true
	run()
	sc.preverbOnly = false
	if needsObject && !sc.Blocked && sc.MoveTo == 0 && sc.MoveGroupTo == 0 {
		sc.verbOnly = true
		run()
		sc.verbOnly = false
	}

	if len(sc.DeferredSegments) > 0 {
		e.scheduleScriptSegments(player, sc.DeferredSegments)
	}
	if sc.MoveGroupTo > 0 {
		return e.finishPreverbMoveGroup(ctx, player, sc), nil, nil
	}
	handled := sc.Blocked || sc.MoveTo > 0 ||
		(needsObject && (len(sc.Messages) > 0 || len(sc.RoomMsgs) > 0))
	if !handled {
		if sc.NeedsSave {
			e.SavePlayer(ctx, player)
		}
		return nil, sc.Messages, sc.RoomMsgs
	}
	if sc.NeedsSave || sc.MoveTo > 0 {
		e.SavePlayer(ctx, player)
	}
	result := &CommandResult{Messages: sc.Messages, RoomBroadcast: sc.RoomMsgs, GMBroadcast: sc.GMMsgs, PlayerState: player}
	if sc.MoveTo > 0 {
		e.applySayMove(ctx, player, sc, result)
	}
	if len(result.Messages) == 0 {
		result.Messages = []string{"Nothing happens."}
	}
	return result, nil, nil
}
