package engine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Time ratio: 4 real hours = 24 game hours → 6 game minutes per real minute.
const gameTimeRatio = 6

// gameTimeOffset is the total game-minutes elapsed, loaded from MongoDB on startup.
// All time functions derive from this plus elapsed real time since load.
var (
	gameTimeOffset    int       // game-minutes at server start (from DB)
	gameTimeLoadedAt  time.Time // real wall-clock when offset was loaded
)

func initGameTime(offset int) {
	gameTimeOffset = offset
	gameTimeLoadedAt = time.Now()
}

// GameMinutes returns total elapsed game-minutes since the world began.
func GameMinutes() int {
	elapsed := time.Since(gameTimeLoadedAt)
	return gameTimeOffset + int(elapsed.Minutes()*gameTimeRatio)
}

// Calendar shape (GM Pages/andor.html, MANUAL.DOC "The Calendar and Astrology of
// the Realms"): a 343-day year that opens with the 7 Feast Days of the Feast of the
// Great Moon, the New Year's celebration, followed by twelve 28-day months. The
// Feast is not a month, but scripts see it as MONTH 13 (GMSCRIPT.DOC: "Month
// (1-13, 13=feast days)"), so gameDate uses Month 13 for it too. TIME, DATECVT,
// seasons and the script DATE/MONTH/YEAR vars all derive from these constants.
const (
	gameMinutesPerDay = 24 * 60
	gameDaysPerMonth  = 28
	gameMonthsPerYear = 12
	gameFeastDays     = 7
	gameFeastMonth    = 13                                                   // Month value for the Feast Days
	gameDaysPerYear   = gameFeastDays + gameDaysPerMonth*gameMonthsPerYear // 343
	gameDaysPerWeek   = 7
	gameEpochYear     = 1028 // the year game-minute 0 falls in
	gameFeastName     = "the Feast of the Great Moon"
)

// gameCalendarVersion is stored with the saved game time. Version 1 (no field)
// was the old 336-day year with no Feast Days; LoadGameTime converts it.
const gameCalendarVersion = 2

func GameHour() int { return GameMinutes() / 60 % 24 }
func GameDay() int  { return GameMinutes()/gameMinutesPerDay%gameDaysPerYear + 1 } // day of year, 1-343

// gameDate is a point on the in-game calendar.
type gameDate struct {
	Year, Month, Day, Hour, Minute int // Month 1-12, or gameFeastMonth; Day 1-28 (1-7 in the Feast)
}

// gameDateAt converts total elapsed game-minutes to a calendar date.
func gameDateAt(minutes int) gameDate {
	dayOfYear := minutes / gameMinutesPerDay % gameDaysPerYear // 0-based
	d := gameDate{
		Year:   minutes/(gameMinutesPerDay*gameDaysPerYear) + gameEpochYear,
		Hour:   minutes / 60 % 24,
		Minute: minutes % 60,
	}
	if dayOfYear < gameFeastDays {
		d.Month, d.Day = gameFeastMonth, dayOfYear+1
	} else {
		dayOfYear -= gameFeastDays
		d.Month, d.Day = dayOfYear/gameDaysPerMonth+1, dayOfYear%gameDaysPerMonth+1
	}
	return d
}

// minutes is the inverse of gameDateAt: total game-minutes at that moment.
func (d gameDate) minutes() int {
	days := (d.Year - gameEpochYear) * gameDaysPerYear
	if d.Month == gameFeastMonth {
		days += d.Day - 1
	} else {
		days += gameFeastDays + (d.Month-1)*gameDaysPerMonth + d.Day - 1
	}
	return days*gameMinutesPerDay + d.Hour*60 + d.Minute
}

// isFeast reports whether d falls in the Feast of the Great Moon.
func (d gameDate) isFeast() bool { return d.Month == gameFeastMonth }

// daysInMonth is how many days d's month has (the Feast has 7).
func gameDaysInMonth(month int) int {
	if month == gameFeastMonth {
		return gameFeastDays
	}
	return gameDaysPerMonth
}

// weekday is the 1-based day of the week. The Feast is exactly one week and each
// month exactly four (343 = 49 weeks), so every month starts a new week.
func (d gameDate) weekday() int { return (d.Day-1)%gameDaysPerWeek + 1 }

// text is the date as players see it: "Abra 5, 1046", or "the third day of the
// Feast of the Great Moon, 1046".
func (d gameDate) text() string {
	if d.isFeast() {
		return fmt.Sprintf("the %s day of %s, %d", ordinalWord(d.Day), gameFeastName, d.Year)
	}
	return fmt.Sprintf("%s %d, %d", gameMonthName(d.Month), d.Day, d.Year)
}

// GameDateText returns today's date as players see it (see gameDate.text).
func GameDateText() string { return gameDateAt(GameMinutes()).text() }

// GameDayOfMonth returns today's day of the month (1-28, or 1-7 in the Feast).
func GameDayOfMonth() int { return gameDateAt(GameMinutes()).Day }

// convertFeastlessGameMinutes maps a game-minute count saved under the old
// 336-day calendar to the same date and time on the 343-day calendar, so adding
// the Feast Days doesn't move the current date.
func convertFeastlessGameMinutes(old int) int {
	const oldDaysPerYear = gameDaysPerMonth * gameMonthsPerYear
	days, rem := old/gameMinutesPerDay, old%gameMinutesPerDay
	year, dayOfYear := days/oldDaysPerYear, days%oldDaysPerYear
	return (year*gameDaysPerYear+gameFeastDays+dayOfYear)*gameMinutesPerDay + rem
}

// gameMinutesAtReal maps a real moment to game-minutes using the same anchor
// GameMinutes uses (the offset loaded at startup plus real time since). It assumes
// the game clock runs continuously, but the clock pauses while the server is down,
// so moments far from now are projections.
func gameMinutesAtReal(t time.Time) float64 {
	return float64(gameTimeOffset) + t.Sub(gameTimeLoadedAt).Minutes()*gameTimeRatio
}

// realTimeAtGameMinutes is the inverse of gameMinutesAtReal.
func realTimeAtGameMinutes(gm float64) time.Time {
	return gameTimeLoadedAt.Add(time.Duration((gm - float64(gameTimeOffset)) / gameTimeRatio * float64(time.Minute)))
}

// TimePeriod returns the current time of day, used by the TIME command to give
// players a more specific read than day/night. "dawn" is a single hour (6),
// not a band — scripts key exact-moment events off IFVAR TIM = <hour> (e.g. the
// JEZRAEL.SC1 fountain's dawn dance requires TIM = 6 exactly, SALOON.SCR's sunset
// vista requires TIM = 19 exactly), so the displayed label must match the hour a
// script actually checks or those events become nearly untriggerable by players
// going solely off what TIME tells them. midnight (0) and noon (12) are the same
// kind of exact-moment anchor and already sat correctly in the middle of their
// bands. The two hours freed up by narrowing dawn to a point are absorbed by its
// immediate neighbors so all 24 hours stay covered.
func TimePeriod() string { return timePeriodForHour(GameHour()) }

// timePeriodForHour is TimePeriod for an arbitrary game hour (0-23).
func timePeriodForHour(h int) string {
	switch {
	case h < 2 || h >= 23:
		return "midnight"
	case h < 6:
		return "very early morning"
	case h == 6:
		return "dawn"
	case h < 11:
		return "mid morning"
	case h < 14:
		return "noon"
	case h < 17:
		return "afternoon"
	case h < 20:
		return "evening"
	default:
		return "night"
	}
}

func GameMonth() int { return gameDateAt(GameMinutes()).Month }

func GameYear() int  { return gameDateAt(GameMinutes()).Year }
func IsNight() bool  { h := GameHour(); return h < 5 || h > 19 }
func IsDay() bool    { return !IsNight() }

// seasonHold pins the season regardless of the calendar (GM @season hold/set),
// e.g. so a festival's seasonal room overrides aren't swapped out mid-event.
// "" = follow the calendar. Persisted in game_state {_id: "season"}.
var seasonHold struct {
	sync.RWMutex
	key string
}

// HeldSeason returns the GM-held season key, or "" when following the calendar.
func HeldSeason() string {
	seasonHold.RLock()
	defer seasonHold.RUnlock()
	return seasonHold.key
}

func setHeldSeason(key string) {
	seasonHold.Lock()
	seasonHold.key = key
	seasonHold.Unlock()
}

// GameSeason returns the current season script key: the GM-held season if one is
// set (@season), otherwise the calendar season (CalendarSeason).
func GameSeason() string {
	if held := HeldSeason(); held != "" {
		return held
	}
	return CalendarSeason()
}

// CalendarSeason returns the season the game calendar says it is, ignoring any
// GM hold. Months 1-3 = Spring (PSCRIPT), 4-6 = Summer (SSCRIPT),
// 7-9 = Autumn (ASCRIPT), 10-12 and the Feast Days that follow Folster = Winter
// (WSCRIPT).
func CalendarSeason() string {
	m := GameMonth()
	switch {
	case m >= 1 && m <= 3:
		return "PSCRIPT"
	case m >= 4 && m <= 6:
		return "SSCRIPT"
	case m >= 7 && m <= 9:
		return "ASCRIPT"
	default:
		return "WSCRIPT"
	}
}

// SeasonName returns a human-readable season name.
func SeasonName() string {
	return seasonKeyName(GameSeason())
}

// seasonKeyName maps a season script key (PSCRIPT...) to its display name.
func seasonKeyName(key string) string {
	switch key {
	case "PSCRIPT":
		return "Spring"
	case "SSCRIPT":
		return "Summer"
	case "ASCRIPT":
		return "Autumn"
	case "WSCRIPT":
		return "Winter"
	}
	return "Spring"
}

var MonthNames = []string{
	"", "Abra", "Brama", "Manta", "Dretmar", "Alabea", "Phobrus",
	"Melma", "Banamea", "Olum", "Mixus", "Farnum", "Folster",
}

func GameMonthName() string { return gameMonthName(GameMonth()) }

func gameMonthName(m int) string {
	if m == gameFeastMonth {
		return gameFeastName
	}
	if m >= 1 && m < len(MonthNames) {
		return MonthNames[m]
	}
	return "Folster"
}

// broadcastOutdoor sends a message to all players in outdoor rooms.
func (e *GameEngine) broadcastOutdoor(msg string) {
	if e.localRoomBroadcast == nil {
		return
	}
	for num, room := range e.rooms {
		if isOutdoorTerrain(room.Terrain) {
			e.localRoomBroadcast(num, []string{msg})
		}
	}
}

// LoadGameTime loads the persisted game time from MongoDB.
// Returns 0 if no saved state exists (fresh world starts at minute 0).
func LoadGameTime(db *mongo.Database) int {
	if db == nil {
		initGameTime(0)
		return 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var doc struct {
		GameMinutes int `bson:"gameMinutes"`
		Calendar    int `bson:"calendar"`
	}
	err := db.Collection("game_state").FindOne(ctx, bson.M{"_id": "gametime"}).Decode(&doc)
	if err != nil {
		log.Printf("No saved game time found, starting at minute 0")
		initGameTime(0)
		return 0
	}
	loadSeasonHold(db)
	if doc.Calendar < gameCalendarVersion {
		// Saved before the Feast Days existed: keep the same date and time. The
		// next SaveGameTime stores the converted value with the new version.
		converted := convertFeastlessGameMinutes(doc.GameMinutes)
		log.Printf("Converting game time to the 343-day Feast calendar: %d -> %d minutes", doc.GameMinutes, converted)
		doc.GameMinutes = converted
	}
	log.Printf("Loaded game time: %d minutes", doc.GameMinutes)
	initGameTime(doc.GameMinutes)
	log.Printf("Game time: %s (%s, %s)",
		GameDateText(), SeasonName(),
		func() string { if IsNight() { return "night" }; return "day" }())
	return doc.GameMinutes
}

// loadSeasonHold restores a GM season hold saved by @season (called from
// LoadGameTime, before the engine picks its starting season).
func loadSeasonHold(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var doc struct {
		Held string `bson:"held"`
	}
	if err := db.Collection("game_state").FindOne(ctx, bson.M{"_id": "season"}).Decode(&doc); err != nil {
		return
	}
	if doc.Held != "" {
		setHeldSeason(doc.Held)
		log.Printf("Season held by GM: %s (calendar season: %s)", seasonKeyName(doc.Held), seasonKeyName(CalendarSeason()))
	}
}

// saveSeasonHold persists the GM season hold ("" clears it).
func saveSeasonHold(db *mongo.Database, key string) error {
	if db == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Collection("game_state").ReplaceOne(ctx, bson.M{"_id": "season"},
		bson.M{"_id": "season", "held": key, "updatedAt": time.Now()}, options.Replace().SetUpsert(true))
	return err
}

// realMinutesToNextCalendarSeason estimates how many real minutes remain until the
// calendar season next changes. Seasons start on Abra 1, Dretmar 1, Melma 1 and
// Mixus 1; winter also covers the Feast Days at the start of the year.
func realMinutesToNextCalendarSeason() int {
	const yearMinutes = gameDaysPerYear * gameMinutesPerDay
	intoYear := GameMinutes() % yearMinutes
	next := yearMinutes + gameFeastDays*gameMinutesPerDay // next year's Abra 1
	for _, month := range []int{1, 4, 7, 10} {
		start := (gameFeastDays + (month-1)*gameDaysPerMonth) * gameMinutesPerDay
		if start > intoYear {
			next = start
			break
		}
	}
	return (next - intoYear + gameTimeRatio - 1) / gameTimeRatio
}

// SaveGameTime persists the current game time to MongoDB.
func SaveGameTime(db *mongo.Database) {
	if db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mins := GameMinutes()
	opts := options.Replace().SetUpsert(true)
	_, err := db.Collection("game_state").ReplaceOne(ctx, bson.M{"_id": "gametime"},
		bson.M{"_id": "gametime", "gameMinutes": mins, "calendar": gameCalendarVersion, "updatedAt": time.Now()}, opts)
	if err != nil {
		log.Printf("Failed to save game time: %v", err)
	}
}

// StartTimeCycle starts a background goroutine that publishes time change events
// and periodically saves game time to MongoDB.
func (e *GameEngine) StartTimeCycle() {
	go func() {
		lastHour := GameHour()
		wasNight := IsNight()
		saveCounter := 0
		ticker := time.NewTicker(10 * time.Second) // check every 10 real seconds
		defer ticker.Stop()
		for range ticker.C {
			// Save game time to DB every ~5 minutes (30 ticks × 10 sec)
			saveCounter++
			if saveCounter >= 30 {
				SaveGameTime(e.db)
				saveCounter = 0
			}

			// Season changes apply whether or not anyone is watching the admin event
			// stream (this used to sit below the subscriber check, so the world only
			// changed season while that stream was open).
			e.CheckSeasonChange()

			if !e.Events.HasSubscribers() {
				lastHour = GameHour()
				wasNight = IsNight()
				continue
			}
			hour := GameHour()
			night := IsNight()

			if hour != lastHour {
				period := "day"
				if night { period = "night" }
				e.Events.Publish("time", fmt.Sprintf("Hour %d:00 — %s of %s",
					hour, period, GameDateText()))
				if night != wasNight {
					if night {
						e.Events.Publish("time", "Night falls across the Shattered Realms.")
					} else {
						e.Events.Publish("time", "Dawn breaks across the Shattered Realms.")
					}
				}
				// Broadcast time-of-day transitions to outdoor players
				if e.localRoomBroadcast != nil {
					switch hour {
					case 5:
						e.broadcastOutdoor("The sun begins to rise in the east.")
					case 6:
						e.broadcastOutdoor("The sun rises in the east.")
					case 18:
						e.broadcastOutdoor("The sun begins to set in the west.")
					case 19:
						e.broadcastOutdoor("The sun sets in the west.")
					}
				}
				lastHour = hour
				wasNight = night
			}
		}
	}()
}

// gmSeason handles @season:
//   @season                 show the season, the calendar season, and time to the next change
//   @season hold            freeze the current season (time of day and dates keep moving)
//   @season set <season>    switch to spring/summer/autumn/winter now and hold it
//   @season release         follow the calendar again (catches up immediately if due)
func (e *GameEngine) gmSeason(player *Player, args []string) *CommandResult {
	calendar := CalendarSeason()
	held := HeldSeason()
	if len(args) == 0 {
		mins := realMinutesToNextCalendarSeason()
		next := fmt.Sprintf("about %dh %dm real time", mins/60, mins%60)
		msgs := []string{fmt.Sprintf("Season: %s. Game date: %s.", SeasonName(), GameDateText())}
		if held != "" {
			msgs = append(msgs,
				fmt.Sprintf("HELD by a GM. The calendar says %s; its next change is in %s.", seasonKeyName(calendar), next),
				"Use @season release to follow the calendar again.")
		} else {
			msgs = append(msgs, fmt.Sprintf("Following the calendar. Next season change in %s.", next),
				"Use @season hold to freeze it, or @season set <spring|summer|autumn|winter>.")
		}
		return &CommandResult{Messages: msgs}
	}

	var key string
	switch strings.ToLower(args[0]) {
	case "hold", "freeze", "lock":
		key = GameSeason()
	case "set":
		if len(args) < 2 {
			return &CommandResult{Messages: []string{"Usage: @season set <spring|summer|autumn|winter>"}}
		}
		switch strings.ToLower(args[1]) {
		case "spring":
			key = "PSCRIPT"
		case "summer":
			key = "SSCRIPT"
		case "autumn", "fall":
			key = "ASCRIPT"
		case "winter":
			key = "WSCRIPT"
		default:
			return &CommandResult{Messages: []string{"Unknown season. Use spring, summer, autumn or winter."}}
		}
	case "release", "unhold", "unlock", "calendar":
		key = ""
	default:
		return &CommandResult{Messages: []string{"Usage: @season [hold | set <spring|summer|autumn|winter> | release]"}}
	}

	if err := saveSeasonHold(e.db, key); err != nil {
		return &CommandResult{Messages: []string{fmt.Sprintf("Could not save the season setting: %v", err)}}
	}
	setHeldSeason(key)
	before := e.currentSeason
	e.CheckSeasonChange() // applies seasonal rooms/MLISTs and announces it outdoors if it changed

	var msg string
	switch {
	case key == "":
		msg = fmt.Sprintf("The season now follows the calendar: %s.", SeasonName())
	default:
		msg = fmt.Sprintf("The season is held at %s until someone uses @season release (this survives restarts).", SeasonName())
	}
	if e.currentSeason != before {
		msg += fmt.Sprintf(" Changed from %s; seasonal rooms and creatures were swapped.", seasonKeyName(before))
	}
	e.Events.Publish("time", fmt.Sprintf("%s: %s", player.FirstName, msg))
	return &CommandResult{Messages: []string{msg}, GMBroadcast: []string{fmt.Sprintf("[GM] %s: %s", player.FirstName, msg)}}
}
