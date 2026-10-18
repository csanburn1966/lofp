package engine

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GM-initiated server restart (@REBOOT), for GMs without access to the host.
//
// The process saves everything that isn't already saved as it changes (online
// players, game time, global PVALs) and exits with rebootExitCode. Something
// outside the process has to start it again: the production server runs under
// systemd, whose unit needs Restart=on-failure (or always) for that — a non-zero
// exit counts as a failure. A local `start.bat`/`go run` does not restart itself.

const (
	rebootExitCode   = 3
	rebootMinSeconds = 10
	rebootMaxSeconds = 1800
)

var rebootState struct {
	sync.Mutex
	at     time.Time
	by     string
	timers []*time.Timer
}

// rebootExit is how the process exits; a variable so it can be stubbed.
var rebootExit = os.Exit

// gmReboot handles @REBOOT. It must be typed in full (no abbreviation) and with an
// argument, so a stray "@reb" can never restart the world:
//   @reboot                 show status / usage
//   @reboot <seconds>       restart after a countdown (10-1800 seconds)
//   @reboot now             restart after the minimum 10-second warning
//   @reboot cancel          cancel a pending restart
func (e *GameEngine) gmReboot(player *Player, args []string, rawInput string) *CommandResult {
	typed := strings.ToUpper(strings.Fields(rawInput)[0])
	if typed != "@REBOOT" {
		return &CommandResult{Messages: []string{"For safety, @reboot must be typed in full."}}
	}

	rebootState.Lock()
	pending, pendingBy := rebootState.at, rebootState.by
	rebootState.Unlock()

	if len(args) == 0 {
		msgs := []string{
			"Usage: @reboot <seconds> | @reboot now | @reboot cancel",
			"Saves all players and the game clock, warns everyone online, then restarts the game server.",
			"Everyone is disconnected and can log back in about a minute later. Monsters respawn; items lying on the ground may be lost.",
		}
		if !pending.IsZero() {
			msgs = append([]string{fmt.Sprintf("A restart requested by %s is pending in %d seconds.", pendingBy, int(time.Until(pending).Seconds()))}, msgs...)
		}
		return &CommandResult{Messages: msgs}
	}

	switch strings.ToLower(args[0]) {
	case "cancel", "abort", "stop":
		if pending.IsZero() {
			return &CommandResult{Messages: []string{"No restart is pending."}}
		}
		cancelReboot()
		e.announceAll("*** The server restart has been cancelled. ***")
		e.Events.Publish("admin", fmt.Sprintf("%s cancelled the server restart", player.FirstName))
		log.Printf("REBOOT cancelled by %s", player.FirstName)
		return &CommandResult{
			Messages:    []string{"Restart cancelled."},
			GMBroadcast: []string{fmt.Sprintf("[GM] %s cancelled the server restart.", player.FirstName)},
		}
	}

	if !pending.IsZero() {
		return &CommandResult{Messages: []string{fmt.Sprintf("A restart requested by %s is already pending in %d seconds. Use @reboot cancel first.", pendingBy, int(time.Until(pending).Seconds()))}}
	}
	seconds := rebootMinSeconds
	if strings.ToLower(args[0]) != "now" {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return &CommandResult{Messages: []string{"Usage: @reboot <seconds> | @reboot now | @reboot cancel"}}
		}
		if n < rebootMinSeconds || n > rebootMaxSeconds {
			return &CommandResult{Messages: []string{fmt.Sprintf("Seconds must be between %d and %d.", rebootMinSeconds, rebootMaxSeconds)}}
		}
		seconds = n
	}

	e.scheduleReboot(player.FirstName, seconds)
	return &CommandResult{
		Messages:    []string{fmt.Sprintf("Server restart scheduled in %d seconds. Use @reboot cancel to stop it.", seconds)},
		GMBroadcast: []string{fmt.Sprintf("[GM] %s scheduled a server restart in %d seconds.", player.FirstName, seconds)},
	}
}

func (e *GameEngine) scheduleReboot(by string, seconds int) {
	rebootState.Lock()
	defer rebootState.Unlock()
	rebootState.at = time.Now().Add(time.Duration(seconds) * time.Second)
	rebootState.by = by
	log.Printf("REBOOT scheduled by %s in %d seconds", by, seconds)
	e.Events.Publish("admin", fmt.Sprintf("%s scheduled a server restart in %d seconds", by, seconds))

	e.announceAll(fmt.Sprintf("*** The server will restart in %s. Your character will be saved automatically; you can log back in about a minute after the restart. ***", rebootDuration(seconds)))
	// Reminders at fixed marks before the restart.
	for _, mark := range []int{300, 120, 60, 30, 10} {
		if mark >= seconds {
			continue
		}
		mark := mark
		rebootState.timers = append(rebootState.timers, time.AfterFunc(time.Duration(seconds-mark)*time.Second, func() {
			e.announceAll(fmt.Sprintf("*** The server will restart in %s. ***", rebootDuration(mark)))
		}))
	}
	rebootState.timers = append(rebootState.timers, time.AfterFunc(time.Duration(seconds)*time.Second, e.performReboot))
}

func cancelReboot() {
	rebootState.Lock()
	defer rebootState.Unlock()
	for _, t := range rebootState.timers {
		t.Stop()
	}
	rebootState.timers = nil
	rebootState.at = time.Time{}
	rebootState.by = ""
}

// performReboot saves state and exits so the host restarts the server.
func (e *GameEngine) performReboot() {
	rebootState.Lock()
	by := rebootState.by
	if rebootState.at.IsZero() { // cancelled at the last moment
		rebootState.Unlock()
		return
	}
	rebootState.Unlock()

	e.announceAll("*** The server is restarting now. Your character has been saved; please log back in in a minute or so. ***")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	saved := 0
	if e.sessions != nil {
		for _, p := range e.sessions.OnlinePlayers() {
			e.SavePlayer(ctx, p)
			saved++
		}
	}
	SaveGameTime(e.db)
	e.savePVals()
	log.Printf("REBOOT by %s: saved %d online players, game time and PVALs; exiting with code %d", by, saved, rebootExitCode)
	// Give the final message a moment to reach clients before connections drop.
	time.Sleep(2 * time.Second)
	rebootExit(rebootExitCode)
}

// announceAll sends a line to every online player.
func (e *GameEngine) announceAll(msg string) {
	if e.sessions == nil || e.sendToPlayer == nil {
		return
	}
	for _, p := range e.sessions.OnlinePlayers() {
		e.sendToPlayer(p.FirstName, []string{msg})
	}
}

func rebootDuration(seconds int) string {
	switch {
	case seconds >= 120 && seconds%60 == 0:
		return fmt.Sprintf("%d minutes", seconds/60)
	case seconds == 60:
		return "1 minute"
	default:
		return fmt.Sprintf("%d seconds", seconds)
	}
}
