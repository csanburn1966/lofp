package engine

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WeatherNames maps weather state IDs to display names (from GM Manual).
var WeatherNames = map[int]string{
	0:  "Sunny",
	1:  "Partly Cloudy",
	2:  "Overcast",
	3:  "Light Rain",
	4:  "Moderate Rain",
	5:  "Heavy Rain",
	6:  "Thunderstorm",
	7:  "Gale",
	8:  "Hurricane",
	9:  "Hail",
	10: "Sleet",
	11: "Snow Flurries",
	12: "Moderate Snow",
	13: "Heavy Snow",
	14: "Blizzard",
}

// Weather transition messages broadcast to outdoor players.
var weatherTransitionMessages = map[[2]int]string{
	{0, 1}:  "A few clouds drift across the sky.",
	{0, 2}:  "The sky darkens and becomes overcast.",
	{1, 0}:  "The clouds overhead drift away, leaving only clear skies.",
	{1, 2}:  "The sky darkens as clouds gather overhead.",
	{2, 0}:  "The clouds break apart, revealing clear blue sky.",
	{2, 1}:  "The overcast sky begins to lighten as gaps appear in the clouds.",
	{2, 3}:  "A light rain begins to fall.",
	{3, 2}:  "The rain tapers off and the clouds begin to thin.",
	{3, 4}:  "The rain intensifies to a steady downpour.",
	{4, 3}:  "The rain lessens to a light drizzle.",
	{4, 5}:  "The rain grows heavier, driven by gusting wind.",
	{5, 4}:  "The downpour eases somewhat.",
	{5, 6}:  "Thunder rumbles in the distance as lightning flashes across the sky.",
	{6, 5}:  "The thunder fades, though heavy rain continues.",
	{6, 2}:  "The storm passes, leaving only overcast skies.",
	{2, 11}: "Snowflakes begin to fall gently from the gray sky.",
	{11, 2}: "The snow stops and the clouds begin to thin.",
	{11, 12}: "The snow picks up, falling more heavily now.",
	{12, 11}: "The snowfall lightens to scattered flurries.",
	{12, 13}: "It is snowing heavily.",
	{13, 12}: "The heavy snow begins to let up.",
	{13, 14}: "A howling blizzard descends, reducing visibility to almost nothing.",
	{14, 13}: "The blizzard eases to heavy snow.",
}

// weatherTransitions defines likely next states from each weather state.
// Each entry is [nextState, weight]. Higher weight = more likely. Tuned (no
// original source) so a PRECIPITATION 0 region is rainy/snowy roughly a third
// of the time; the precipitation states used to drift back to dry so rarely
// that it was two-thirds.
var weatherTransitions = map[int][][2]int{
	0:  {{0, 60}, {1, 30}, {2, 10}},                // Sunny → stay, partly cloudy, overcast
	1:  {{0, 30}, {1, 40}, {2, 30}},                // Partly Cloudy → clear, stay, overcast
	2:  {{1, 30}, {2, 45}, {3, 15}, {11, 10}},      // Overcast → clearing, stay, rain, snow
	3:  {{2, 45}, {3, 35}, {4, 20}},                // Light Rain → clear, stay, moderate
	4:  {{3, 40}, {4, 35}, {5, 25}},                // Moderate Rain → lighter, stay, heavier
	5:  {{4, 45}, {5, 30}, {6, 25}},                // Heavy Rain → moderate, stay, storm
	6:  {{5, 40}, {2, 40}, {6, 20}},                // Thunderstorm → heavy rain, overcast, stay
	7:  {{5, 50}, {6, 30}, {7, 20}},                // Gale → heavy rain, storm, stay
	8:  {{7, 50}, {5, 30}, {8, 20}},                // Hurricane → gale, heavy rain, stay
	9:  {{2, 50}, {3, 30}, {9, 20}},                // Hail → overcast, rain, stay
	10: {{2, 40}, {11, 30}, {10, 30}},              // Sleet → overcast, flurries, stay
	11: {{2, 45}, {11, 35}, {12, 20}},              // Snow Flurries → overcast, stay, moderate
	12: {{11, 40}, {12, 35}, {13, 25}},             // Moderate Snow → flurries, stay, heavy
	13: {{12, 45}, {13, 30}, {14, 25}},             // Heavy Snow → moderate, stay, blizzard
	14: {{13, 50}, {14, 30}, {12, 20}},             // Blizzard → heavy, stay, moderate
}

// GetWeatherDesc returns a weather description for a given region.
func (e *GameEngine) GetWeatherDesc(region int) string {
	if e.RegionWeather == nil {
		return ""
	}
	state, ok := e.RegionWeather[region]
	if !ok {
		state = 0
	}
	if name, ok := WeatherNames[state]; ok {
		return name
	}
	return "Clear"
}

// weatherRoomDesc maps weather state IDs to immersive room description lines.
var weatherRoomDesc = map[int]string{
	1:  "A few white clouds drift lazily across an otherwise blue sky.",
	2:  "The sky above is heavy with gray clouds, blocking out the sun.",
	3:  "A light rain falls steadily, pattering softly on the ground.",
	4:  "A steady rain comes down, soaking everything in sight.",
	5:  "Heavy rain lashes down, driven by gusting winds.",
	6:  "Thunder rumbles overhead and lightning flickers through the roiling storm clouds.",
	7:  "A powerful gale tears through, bending trees and howling between every gap.",
	8:  "A furious hurricane rages, the wind a deafening roar and the rain nearly horizontal.",
	9:  "Chunks of ice clatter down from a bruised sky, bouncing off the ground in every direction.",
	10: "Wet sleet stings any exposed skin, half-rain half-ice, coating surfaces in a thin glaze.",
	11: "Light snowflakes drift down from gray clouds, dusting the ground in white.",
	12: "Snow falls steadily, blanketing everything in a deepening layer of white.",
	13: "Heavy snow swirls down in thick curtains, making it hard to see far.",
	14: "A howling blizzard rages, the world reduced to a churning wall of white.",
}

// GetRoomWeather returns a weather line for an outdoor room, or "" for indoor.
func (e *GameEngine) GetRoomWeather(roomNum int) string {
	room := e.rooms[roomNum]
	if room == nil {
		return ""
	}
	if !isOutdoorTerrain(room.Terrain) {
		return ""
	}
	region := room.Region
	state, ok := e.RegionWeather[region]
	if !ok {
		state = 0
	}
	if state == 0 { // Sunny — no extra line
		return ""
	}
	if desc, ok := weatherRoomDesc[state]; ok {
		return desc
	}
	return ""
}

// temperatureBySeason is the fallback base temperature (Fahrenheit) for a
// season, used when neither the region nor region 0 has a REGIONDEF
// <season>TEMP. The values are region 0's own (WINTERTEMP 10, SPRINGTEMP 45,
// SUMMERTEMP 60, AUTUMNTEMP 29).
var temperatureBySeason = map[string]int{
	"PSCRIPT": 45, // Spring
	"SSCRIPT": 60, // Summer
	"ASCRIPT": 29, // Autumn
	"WSCRIPT": 10, // Winter
}

// seasonTempKey maps a season script name to its REGIONDEF temperature key.
var seasonTempKey = map[string]string{
	"PSCRIPT": "SPRINGTEMP",
	"SSCRIPT": "SUMMERTEMP",
	"ASCRIPT": "AUTUMNTEMP",
	"WSCRIPT": "WINTERTEMP",
}

// temperatureByPeriod adjusts the day's temperature for time of day (see TimePeriod).
var temperatureByPeriod = map[string]int{
	"midnight":            -12,
	"very early morning":  -10,
	"dawn":                -5,
	"mid morning":         0,
	"noon":                8,
	"afternoon":           6,
	"evening":             -2,
	"night":               -8,
}

// dailyTemp caches each region's rolled temperature for the current game day.
type dailyTemp struct {
	day    int // absolute game day (GameMinutes / gameMinutesPerDay)
	season string
	temp   int
}

var (
	dailyTempMu sync.Mutex
	dailyTemps  = map[int]dailyTemp{}
)

// regionBaseTemp returns the region's base temperature for the current season:
// its REGIONDEF <season>TEMP, else region 0's, else temperatureBySeason.
func (e *GameEngine) regionBaseTemp(region int, season string) int {
	key := seasonTempKey[season]
	for _, id := range []int{region, 0} {
		if r := e.regions[id]; r != nil {
			if v, ok := r.Properties[key]; ok {
				if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
					return n
				}
			}
		}
	}
	return temperatureBySeason[season]
}

// GetTemperature returns the current temperature (degrees Fahrenheit) for a region.
// Per GMSCRIPT.DOC (REGIONDEF x <season>TEMP), the temperature is rolled once a day
// as base_temp + rnd(15) + rnd(15); the time of day then shifts it a little.
//
// It used to be season baseline + time of day + a weather offset (-22 for Moderate
// Snow), which ran backwards: the weather set the temperature, so spring snow made
// the day colder rather than needing a cold day to snow. Weather now follows the
// temperature instead (see precipitationForTemp).
func (e *GameEngine) GetTemperature(region int) int {
	season := GameSeason()
	day := GameMinutes() / gameMinutesPerDay
	dailyTempMu.Lock()
	dt, ok := dailyTemps[region]
	if !ok || dt.day != day || dt.season != season {
		dt = dailyTemp{day: day, season: season,
			temp: e.regionBaseTemp(region, season) + 1 + rand.Intn(15) + 1 + rand.Intn(15)}
		dailyTemps[region] = dt
	}
	dailyTempMu.Unlock()
	return dt.temp + temperatureByPeriod[TimePeriod()]
}

// TemperatureDesc returns a short descriptive word for a Fahrenheit temperature.
func TemperatureDesc(f int) string {
	switch {
	case f < 20:
		return "frigid"
	case f < 32:
		return "freezing"
	case f < 45:
		return "cold"
	case f < 60:
		return "cool"
	case f < 75:
		return "mild"
	case f < 85:
		return "warm"
	case f < 95:
		return "hot"
	default:
		return "sweltering"
	}
}

// isOutdoorTerrain returns true if the terrain type is outdoors.
func isOutdoorTerrain(terrain string) bool {
	switch terrain {
	case "FOREST", "MOUNTAIN", "PLAIN", "SWAMP", "JUNGLE",
		"WASTE", "OUTDOOR_OTHER", "OUTDOOR_FLOOR", "AERIAL":
		return true
	}
	return false
}

// snowStates are weather states that only make sense when it's cold enough for
// precipitation to fall as ice or snow rather than rain: Sleet, Snow Flurries,
// Moderate Snow, Heavy Snow, Blizzard.
var snowStates = map[int]bool{10: true, 11: true, 12: true, 13: true, 14: true}

// Rain and snow counterparts of matching intensity, for switching precipitation
// type when the temperature crosses freezing.
var (
	rainToSnow = map[int]int{3: 11, 4: 12, 5: 13, 6: 13}
	snowToRain = map[int]int{10: 3, 11: 3, 12: 4, 13: 5, 14: 5}
)

// precipitationForTemp turns rain into snow at or below freezing, and snow or sleet
// into rain once it's clearly above freezing (37°F+). 33-36°F is left alone —
// either can fall there.
func precipitationForTemp(state, temp int) int {
	if temp <= 32 {
		if s, ok := rainToSnow[state]; ok {
			return s
		}
	} else if temp >= 37 {
		if s, ok := snowToRain[state]; ok {
			return s
		}
	}
	return state
}

// weatherWetness ranks states by how much precipitation they bring: 0-2 dry (sky
// cover only), 3+ precipitating.
var weatherWetness = map[int]int{
	0: 0, 1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6, 7: 6, 8: 7,
	9: 4, 10: 3, 11: 3, 12: 4, 13: 5, 14: 6,
}

// precipitationLevel returns the region's REGIONDEF PRECIPITATION level (0 = normal).
func (e *GameEngine) precipitationLevel(region int) int {
	if r := e.regions[region]; r != nil {
		if n, err := strconv.Atoi(strings.TrimSpace(r.Properties["PRECIPITATION"])); err == nil {
			return n
		}
	}
	return 0
}

// biasTransitions applies a REGIONDEF PRECIPITATION level to a state's transitions.
// Per GMSCRIPT.DOC: closer to 10 makes precipitation less likely to start and more
// likely to stop (10 = it probably never rains); negative makes it more likely.
// Moves into wetter precipitation are scaled by wet, moves out of precipitation
// toward drier states by 1/wet. The weights are scaled by 10 so fractions survive.
func biasTransitions(oldState int, transitions [][2]int, level int) [][2]int {
	wet := 1.0
	if level >= 0 {
		wet = float64(10-min(level, 10)) / 10
	} else {
		wet = 1 + float64(-level)/10
	}
	out := make([][2]int, 0, len(transitions))
	for _, t := range transitions {
		w := float64(t[1] * 10)
		from, to := weatherWetness[oldState], weatherWetness[t[0]]
		switch {
		case to > from && to >= 3:
			w *= wet
		case to < from && from >= 3 && wet > 0:
			w /= wet
		}
		if w >= 1 {
			out = append(out, [2]int{t[0], int(w)})
		}
	}
	return out
}

// advanceWeather randomly transitions weather for all regions and broadcasts changes.
//
// The random walk is biased by each region's REGIONDEF PRECIPITATION, and its rain
// or snow is then matched to the region's temperature (precipitationForTemp). Snow
// was once only blocked in Summer, so Spring snowed as readily as Winter.
func (e *GameEngine) advanceWeather() {
	if e.RegionWeather == nil {
		return
	}

	// Collect all regions that have outdoor rooms
	regionSet := make(map[int]bool)
	for _, room := range e.rooms {
		if room.Region > 0 {
			regionSet[room.Region] = true
		}
	}
	// Always include region 0 (default)
	regionSet[0] = true

	for region := range regionSet {
		oldState := e.RegionWeather[region]
		newState := oldState

		transitions := biasTransitions(oldState, weatherTransitions[oldState], e.precipitationLevel(region))
		totalWeight := 0
		for _, t := range transitions {
			totalWeight += t[1]
		}
		if totalWeight > 0 {
			roll := rand.Intn(totalWeight)
			for _, t := range transitions {
				roll -= t[1]
				if roll < 0 {
					newState = t[0]
					break
				}
			}
		}

		// Also re-checks a state that didn't move, so lingering snow turns to rain
		// once the day warms up (and the other way round).
		newState = precipitationForTemp(newState, e.GetTemperature(region))

		if newState != oldState {
			e.RegionWeather[region] = newState
			e.broadcastWeatherChange(region, oldState, newState)
		}
	}
}

// broadcastWeatherChange announces a weather transition to all outdoor rooms
// in the given region, picking a scripted transition line if one exists for
// this exact oldState->newState pair, falling back to a generic line otherwise.
func (e *GameEngine) broadcastWeatherChange(region, oldState, newState int) {
	if e.localRoomBroadcast == nil {
		return
	}
	msg := weatherTransitionMessages[[2]int{oldState, newState}]
	if msg == "" {
		if newState == 0 {
			msg = "The skies clear."
		} else if desc, ok := weatherRoomDesc[newState]; ok {
			msg = desc
		} else {
			msg = "The weather shifts."
		}
	}
	for num, room := range e.rooms {
		if room.Region == region && isOutdoorTerrain(room.Terrain) {
			e.localRoomBroadcast(num, []string{msg})
		}
	}
}

// weatherValueList returns a comma-separated "id=Name" listing of all known
// weather states, in ID order, for GM command usage/error text.
func weatherValueList() string {
	ids := make([]int, 0, len(WeatherNames))
	for id := range WeatherNames {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%d=%s", id, WeatherNames[id])
	}
	return strings.Join(parts, ", ")
}

// gmWeather handles @weather [value] — shows the current weather (and
// temperature) for the GM's region with no args, or sets the region's
// weather state and broadcasts the transition when given a value.
func (e *GameEngine) gmWeather(player *Player, args []string) *CommandResult {
	room := e.rooms[player.RoomNumber]
	region := 0
	if room != nil {
		region = room.Region
	}

	if len(args) == 0 {
		state := 0
		if e.RegionWeather != nil {
			state = e.RegionWeather[region]
		}
		temp := e.GetTemperature(region)
		return &CommandResult{Messages: []string{
			fmt.Sprintf("Region %d weather: %d (%s). Temperature: %d degrees (%s).",
				region, state, WeatherNames[state], temp, TemperatureDesc(temp)),
		}}
	}

	newState, err := strconv.Atoi(args[0])
	if err != nil {
		return &CommandResult{Messages: []string{"Usage: @weather [value]. Valid values: " + weatherValueList()}}
	}
	name, ok := WeatherNames[newState]
	if !ok {
		return &CommandResult{Messages: []string{"Invalid weather value. Valid values: " + weatherValueList()}}
	}

	if e.RegionWeather == nil {
		e.RegionWeather = make(map[int]int)
	}
	oldState := e.RegionWeather[region]
	e.RegionWeather[region] = newState
	if newState != oldState {
		e.broadcastWeatherChange(region, oldState, newState)
	}

	return &CommandResult{Messages: []string{fmt.Sprintf("Region %d weather set to %d (%s).", region, newState, name)}}
}

// doWeather handles the WEATHER command, showing the player the current
// conditions and temperature for their region.
func (e *GameEngine) doWeather(player *Player) *CommandResult {
	room := e.rooms[player.RoomNumber]
	region := 0
	if room != nil {
		region = room.Region
	}
	temp := e.GetTemperature(region)
	tempDesc := TemperatureDesc(temp)

	if room == nil || !isOutdoorTerrain(room.Terrain) {
		return &CommandResult{Messages: []string{
			fmt.Sprintf("You can't see the sky from here, but it feels %s, around %d degrees.", tempDesc, temp),
		}}
	}

	state, ok := e.RegionWeather[region]
	if !ok {
		state = 0
	}
	msgs := []string{fmt.Sprintf("The weather is currently %s.", WeatherNames[state])}
	if desc, ok := weatherRoomDesc[state]; ok {
		msgs = append(msgs, desc)
	}
	msgs = append(msgs, fmt.Sprintf("It feels %s, around %d degrees.", tempDesc, temp))
	return &CommandResult{Messages: msgs}
}

// StartWeatherCycle starts a background goroutine that changes weather periodically.
func (e *GameEngine) StartWeatherCycle() {
	go func() {
		// Weather changes every 5-10 game hours (5-10 real minutes)
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			e.advanceWeather()
		}
	}()
}
