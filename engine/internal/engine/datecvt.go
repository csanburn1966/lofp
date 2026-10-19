package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DATECVT converts between real-world dates and the in-game calendar, in either
// direction, for planning events ("what in-game day will the festival fall on?").
// The original game had a command by this name; no capture of it survives, so the
// syntax and output here are our own.
//
//	DATECVT                          the current moment both ways, plus usage
//	DATECVT <m-d-yyyy> [time]        real date (server time zone) -> in-game
//	DATECVT NOW | TODAY              same, for right now / all of today
//	DATECVT GAME <m-d-yyyy> [time]   in-game date -> real
//	DATECVT GAME <month> <day> <year> [time]   (month by name or number)
//	DATECVT GAME FEAST <day> <year> [time]     (the 7 Feast Days; also month 13)
//
// Every conversion goes through gameMinutesAtReal / realTimeAtGameMinutes, the
// same anchor GameMinutes (and so TIME) uses, so the two always agree.

// datecvtRange is how far from now (in real time) DATECVT will convert.
const datecvtRange = 10 * 365 * 24 * time.Hour

const (
	datecvtRealLayout = "Monday, January 2, 2006"
	datecvtClock      = "3:04 PM"
)

func (e *GameEngine) doDateCvt(args []string) *CommandResult {
	if len(args) == 0 {
		now := time.Now()
		return &CommandResult{Messages: []string{
			fmt.Sprintf("Right now it is %s, which is %s.", formatRealMoment(now), formatGameMoment(gameDateAt(int(gameMinutesAtReal(now))))),
			"Usage:",
			"  DATECVT <month-day-year> [time]       real date to in-game, e.g. DATECVT 10-25-2026 3:30pm",
			"  DATECVT GAME <month-day-year> [time]  in-game date to real, e.g. DATECVT GAME 3-14-1042 15:00",
			"  DATECVT GAME <month> <day> <year>     e.g. DATECVT GAME Manta 14 1042",
			"  DATECVT GAME FEAST <day> <year>       one of the 7 Feast Days that open the year",
			fmt.Sprintf("Real dates use the server's time zone (%s). Leave the time off to see the whole day.", now.Format("MST")),
		}}
	}
	switch strings.ToUpper(args[0]) {
	case "GAME", "INGAME", "IG":
		return datecvtFromGame(args[1:])
	case "REAL":
		args = args[1:]
	}
	return datecvtFromReal(args)
}

// datecvtFromReal converts a real date (and optional time) to the in-game calendar.
func datecvtFromReal(args []string) *CommandResult {
	if len(args) == 0 {
		return datecvtUsage()
	}
	now := time.Now()
	var start time.Time
	hasTime := true
	switch strings.ToUpper(args[0]) {
	case "NOW":
		start = now
	case "TODAY":
		y, m, d := now.Date()
		start, hasTime = time.Date(y, m, d, 0, 0, 0, 0, time.Local), false
	default:
		nums, clock, ok := splitDateArgs(args)
		if !ok {
			return datecvtUsage()
		}
		mo, d, y := nums[0], nums[1], nums[2]
		if y < 100 {
			y += 2000
		}
		start = time.Date(y, time.Month(mo), d, 0, 0, 0, 0, time.Local)
		if mo < 1 || mo > 12 || start.Day() != d || start.Year() != y {
			return &CommandResult{Messages: []string{fmt.Sprintf("There is no such date as %d-%d-%d.", mo, d, y)}}
		}
		if clock == "" {
			hasTime = false
		} else {
			h, m, ok := parseClock(clock)
			if !ok {
				return &CommandResult{Messages: []string{fmt.Sprintf("I don't understand the time \"%s\". Try 3:30pm or 15:30.", strings.ToLower(clock))}}
			}
			start = time.Date(y, time.Month(mo), d, h, m, 0, 0, time.Local)
		}
	}
	if msg := datecvtOutOfRange(start, now); msg != "" {
		return &CommandResult{Messages: []string{msg}}
	}

	startGM := gameMinutesAtReal(start)
	if startGM < 0 {
		return &CommandResult{Messages: []string{fmt.Sprintf("That is before the Realms' calendar began (%s 1, %d).", gameMonthName(1), gameEpochYear)}}
	}
	var msgs []string
	if hasTime {
		msgs = append(msgs, fmt.Sprintf("%s is %s.", formatRealMoment(start), formatGameMoment(gameDateAt(int(startGM)))))
	} else {
		y, m, d := start.Date()
		end := time.Date(y, m, d+1, 0, 0, 0, 0, time.Local) // DST-safe: not start+24h
		from := gameDateAt(int(startGM))
		to := gameDateAt(int(gameMinutesAtReal(end)))
		msgs = append(msgs,
			fmt.Sprintf("%s (%s) runs from %s at %s to %s at %s.",
				start.Format(datecvtRealLayout), start.Format("MST"),
				formatGameDay(from), formatGameClock(from), formatGameDay(to), formatGameClock(to)),
			fmt.Sprintf("One real day is %d in-game days.", gameTimeRatio))
	}
	return &CommandResult{Messages: append(msgs, datecvtProjectionNote(start, now)...)}
}

// datecvtFromGame converts an in-game date (and optional hour) to real time.
func datecvtFromGame(args []string) *CommandResult {
	if len(args) == 0 {
		return datecvtUsage()
	}
	// "FEAST OF THE GREAT MOON 3 1046" / "FEAST DAY 3 1046" -> "FEAST 3 1046"
	if mo, ok := parseGameMonth(args[0]); ok && mo == gameFeastMonth {
		rest := []string{args[0]}
		i := 1
		for i < len(args) && feastFillerWords[strings.ToUpper(args[i])] {
			i++
		}
		args = append(rest, args[i:]...)
	}
	var gd gameDate
	var clock string
	if mo, ok := parseGameMonth(args[0]); ok && !strings.ContainsAny(args[0], "-/") {
		// <month name|number> <day> <year> [time]
		if len(args) < 3 {
			return datecvtUsage()
		}
		d, err1 := strconv.Atoi(strings.Trim(args[1], ","))
		y, err2 := strconv.Atoi(strings.Trim(args[2], ","))
		if err1 != nil || err2 != nil {
			return datecvtUsage()
		}
		gd = gameDate{Year: y, Month: mo, Day: d}
		clock = strings.Join(args[3:], " ")
	} else {
		// <m-d-y>[-time] [time], where m may also be a month name
		parts := strings.SplitN(strings.ReplaceAll(args[0], "/", "-"), "-", 4)
		if len(parts) < 3 {
			return datecvtUsage()
		}
		mo, ok := parseGameMonth(parts[0])
		d, err1 := strconv.Atoi(parts[1])
		y, err2 := strconv.Atoi(parts[2])
		if !ok || err1 != nil || err2 != nil {
			return datecvtUsage()
		}
		gd = gameDate{Year: y, Month: mo, Day: d}
		if len(parts) == 4 {
			clock = parts[3]
		}
		clock = strings.TrimSpace(clock + " " + strings.Join(args[1:], " "))
	}
	if gd.Day < 1 || gd.Day > gameDaysInMonth(gd.Month) {
		if gd.Month == gameFeastMonth {
			return &CommandResult{Messages: []string{fmt.Sprintf("There is no Feast Day %d. The Feast of the Great Moon lasts %d days.", gd.Day, gameFeastDays)}}
		}
		return &CommandResult{Messages: []string{fmt.Sprintf("There is no such date as %s %d. Every month has %d days.", gameMonthName(gd.Month), gd.Day, gameDaysPerMonth)}}
	}
	if gd.Year < gameEpochYear {
		return &CommandResult{Messages: []string{fmt.Sprintf("That is before the Realms' calendar began (%s 1, %d).", gameMonthName(1), gameEpochYear)}}
	}
	hasTime := clock != ""
	if hasTime {
		h, m, ok := parseClock(clock)
		if !ok {
			return &CommandResult{Messages: []string{fmt.Sprintf("I don't understand the time \"%s\". Try 15:00 or 3pm.", strings.ToLower(clock))}}
		}
		gd.Hour, gd.Minute = h, m
	}

	now := time.Now()
	start := realTimeAtGameMinutes(float64(gd.minutes())).Round(time.Minute)
	if msg := datecvtOutOfRange(start, now); msg != "" {
		return &CommandResult{Messages: []string{msg}}
	}
	var msgs []string
	if hasTime {
		msgs = append(msgs, fmt.Sprintf("%s is %s.", capFirst(formatGameMoment(gd)), formatRealMoment(start)))
	} else {
		end := realTimeAtGameMinutes(float64(gd.minutes() + gameMinutesPerDay)).Round(time.Minute)
		endText := end.Format(datecvtClock)
		if end.YearDay() != start.YearDay() || end.Year() != start.Year() {
			endText = end.Format(datecvtRealLayout + " at " + datecvtClock)
		}
		msgs = append(msgs,
			fmt.Sprintf("%s runs from %s to %s %s.", capFirst(formatGameDay(gd)), start.Format(datecvtRealLayout+" at "+datecvtClock), endText, end.Format("MST")),
			fmt.Sprintf("One in-game day is %d real hours.", 24/gameTimeRatio))
	}
	return &CommandResult{Messages: append(msgs, datecvtProjectionNote(start, now)...)}
}

func datecvtUsage() *CommandResult {
	return &CommandResult{Messages: []string{
		"Usage: DATECVT <month-day-year> [time]  or  DATECVT GAME <month-day-year> [time]",
		"Examples: DATECVT 10-25-2026, DATECVT 10-25-2026 3:30pm, DATECVT GAME Manta 14 1042 15:00",
	}}
}

func datecvtOutOfRange(t, now time.Time) string {
	if d := t.Sub(now); d > datecvtRange || d < -datecvtRange {
		return "DATECVT only works for dates within ten years of today."
	}
	return ""
}

// datecvtProjectionNote warns that conversions away from today assume the game
// clock runs nonstop. It pauses while the server is down, so they can drift.
func datecvtProjectionNote(t, now time.Time) []string {
	if d := t.Sub(now); d < 24*time.Hour && d > -24*time.Hour {
		return nil
	}
	return []string{"(This assumes the game clock keeps running without interruption. It pauses while the server is down, so far-off dates may drift.)"}
}

// splitDateArgs reads "m-d-y" (or m/d/y) with an optional time either attached
// as a fourth dash part ("10-25-2026-15:00") or as the following words.
func splitDateArgs(args []string) (nums [3]int, clock string, ok bool) {
	parts := strings.SplitN(strings.ReplaceAll(args[0], "/", "-"), "-", 4)
	if len(parts) < 3 {
		return nums, "", false
	}
	for i := range 3 {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return nums, "", false
		}
		nums[i] = n
	}
	if len(parts) == 4 {
		clock = parts[3]
	}
	return nums, strings.TrimSpace(clock + " " + strings.Join(args[1:], " ")), true
}

// parseGameMonth accepts a month number (1-12) or name, or a name's first 3+ letters.
func parseGameMonth(s string) (int, bool) {
	s = strings.ToUpper(strings.Trim(s, ","))
	if n, err := strconv.Atoi(s); err == nil {
		return n, (n >= 1 && n <= gameMonthsPerYear) || n == gameFeastMonth
	}
	if len(s) < 3 {
		return 0, false
	}
	if strings.HasPrefix("FEAST", s) {
		return gameFeastMonth, true
	}
	for i := 1; i < len(MonthNames); i++ {
		if strings.HasPrefix(strings.ToUpper(MonthNames[i]), s) {
			return i, true
		}
	}
	if strings.HasPrefix("BANAMAE", s) { // YARIN.SCR's spelling of Banamea
		return 8, true
	}
	return 0, false
}

// parseClock reads "15:30", "15", "3pm", "3:30 pm", "3:30p", "noon", "midnight".
func parseClock(s string) (hour, minute int, ok bool) {
	s = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
	s = strings.ReplaceAll(s, ".", "")
	switch s {
	case "NOON":
		return 12, 0, true
	case "MIDNIGHT":
		return 0, 0, true
	}
	ampm := ""
	for _, suf := range []string{"AM", "PM", "A", "P"} {
		if strings.HasSuffix(s, suf) {
			ampm, s = suf[:1], strings.TrimSuffix(s, suf)
			break
		}
	}
	hs, ms, hasMin := strings.Cut(s, ":")
	h, err := strconv.Atoi(hs)
	if err != nil {
		return 0, 0, false
	}
	m := 0
	if hasMin {
		if m, err = strconv.Atoi(ms); err != nil || len(ms) != 2 {
			return 0, 0, false
		}
	}
	if m < 0 || m > 59 {
		return 0, 0, false
	}
	switch ampm {
	case "":
		if h < 0 || h > 23 {
			return 0, 0, false
		}
	default:
		if h < 1 || h > 12 {
			return 0, 0, false
		}
		h %= 12
		if ampm == "P" {
			h += 12
		}
	}
	return h, m, true
}

func formatRealMoment(t time.Time) string {
	return t.Format(datecvtRealLayout + " at " + datecvtClock + " MST")
}

// feastFillerWords are skipped after FEAST so the full name can be typed out.
var feastFillerWords = map[string]bool{"OF": true, "THE": true, "GREAT": true, "MOON": true, "DAY": true}

// formatGameDay adds the day of the week. A Feast Day number is already its day
// of the week, so it isn't repeated there.
func formatGameDay(d gameDate) string {
	if d.isFeast() {
		return d.text()
	}
	return fmt.Sprintf("%s (%s day of the week)", d.text(), ordinalWord(d.weekday()))
}

// capFirst capitalizes a game date that opens a sentence ("The third day of the Feast...").
func capFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func formatGameClock(d gameDate) string {
	return fmt.Sprintf("%02d:%02d", d.Hour, d.Minute)
}

func formatGameMoment(d gameDate) string {
	return fmt.Sprintf("%s at %s (%s)", formatGameDay(d), formatGameClock(d), timePeriodForHour(d.Hour))
}

func ordinalWord(n int) string {
	words := []string{"", "first", "second", "third", "fourth", "fifth", "sixth", "seventh"}
	if n >= 1 && n < len(words) {
		return words[n]
	}
	return strconv.Itoa(n) + "th"
}
