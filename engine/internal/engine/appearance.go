package engine

import (
	"fmt"
	"math/rand"
)

// raceHeightWeightRanges holds [minHeight, maxHeight, minWeight, maxWeight] per race
// (height in inches, weight in pounds). Shared by CreateNewPlayer, LoadPlayer's
// legacy-character backfill, and RollHeightWeight so there's a single source of truth.
var raceHeightWeightRanges = map[int][4]int{
	RaceHuman:      {62, 76, 120, 220},
	RaceAelfen:     {66, 80, 100, 170},
	RaceHighlander: {48, 58, 130, 200},
	RaceWolfling:   {64, 74, 130, 200},
	RaceMurg:       {62, 74, 150, 230},
	RaceDrakin:     {68, 82, 150, 250},
	RaceMechanoid:  {60, 74, 150, 250},
	RaceEphemeral:  {58, 72, 80, 130},
}

// RaceAgeRanges holds [minAge, maxAge] per race. There's no canonical source for
// these in LEGENDS.DOC/MANUAL.DOC (only that displayed Age fluctuates +2 to +7
// above the true value) — these are reasonable fantasy lifespans, all starting
// at 16 (young adult), varying only the max by race.
var RaceAgeRanges = map[int][2]int{
	RaceHuman:      {16, 70},
	RaceAelfen:     {16, 300},
	RaceHighlander: {16, 90},
	RaceWolfling:   {16, 65},
	RaceMurg:       {16, 60},
	RaceDrakin:     {16, 150},
	RaceMechanoid:  {16, 80},
	RaceEphemeral:  {16, 500},
}

// EyeColors, SkinColors, HairColors and HairStyles are the fixed lists a player
// chooses from during character creation. HairStyles includes the sentinel
// "bald", which skips hair color selection entirely.
var EyeColors = []string{
	"blue", "green", "brown", "hazel", "gray", "amber", "violet", "yellow", "silver", "black", "red",
}

var SkinColors = []string{
	"fair", "pale", "tan", "olive", "bronze", "dark", "ebony", "golden", "ashen", "ruddy", "copper", "porcelain",
}

// ScaleColors replace skin and hair for Drakin, who choose a primary and a
// secondary scale color (the original CHARGEN.SCR list).
var ScaleColors = []string{
	"olive green", "sea green", "light green", "violet", "crimson",
	"orange", "brown", "ochre", "jet black",
}

// DrakinEyeColors are the only eye colors a drakin can choose (CHARGEN.SCR).
var DrakinEyeColors = []string{"red", "black", "yellow", "green"}

// ScalePatterns are how a drakin's secondary scale color is laid over the
// primary (CHARGEN.SCR). Chosen only when the two colors differ.
var ScalePatterns = []string{"bands", "stripes", "spots", "splotches", "blotches", "streaks"}

// EyeColorsFor returns the eye colors a race chooses from at creation.
func EyeColorsFor(race int) []string {
	if HasScales(race) {
		return DrakinEyeColors
	}
	return EyeColors
}

// HasScales reports whether a race has scales instead of skin and hair.
func HasScales(race int) bool { return race == RaceDrakin }

// scaleDescription is the scales part of a LOOK description, e.g. "olive
// green scales with crimson stripes". Empty when no primary color is known.
func scaleDescription(primary, secondary, pattern string) string {
	if primary == "" {
		return ""
	}
	if secondary == "" || secondary == primary {
		return primary + " scales"
	}
	if pattern == "" {
		return fmt.Sprintf("%s scales marked with %s", primary, secondary)
	}
	return fmt.Sprintf("%s scales with %s %s", primary, secondary, pattern)
}

var HairColors = []string{
	"black", "brown", "dark brown", "chestnut", "auburn", "red", "strawberry blond",
	"blond", "golden blond", "dark-blond", "sandy", "gray", "silver", "white",
}

// PlayableRaces is the canonical ordered list of playable race IDs, used for
// disguise race selection and random NPC appearance rolls.
var PlayableRaces = []int{
	RaceHuman, RaceAelfen, RaceHighlander, RaceWolfling,
	RaceMurg, RaceDrakin, RaceMechanoid, RaceEphemeral,
}

var HairStyles = []string{
	"long, flowing", "long, straight", "long, messy", "long, curly",
	"short, flowing", "short, straight", "short, messy", "short, curly",
	"mohawk", "braided", "shaved sides", "bald",
	"short", "long", "unkempt", "spiky",
}

// CharacterAppearance holds a character's rolled stats/build and chosen colors
// during creation. A zero value for a stat/height/weight/age means "not yet
// rolled" (all racial ranges start at 1 or above, so 0 is never a legitimate
// roll). An empty EyeColor/SkinColor/HairStyle means "not yet chosen". HairColor
// is legitimately empty when HairStyle is "bald".
type CharacterAppearance struct {
	Strength     int `json:"strength,omitempty"`
	Agility      int `json:"agility,omitempty"`
	Quickness    int `json:"quickness,omitempty"`
	Constitution int `json:"constitution,omitempty"`
	Perception   int `json:"perception,omitempty"`
	Willpower    int `json:"willpower,omitempty"`
	Empathy      int `json:"empathy,omitempty"`
	Height       int `json:"height,omitempty"`
	Weight       int `json:"weight,omitempty"`
	Age          int `json:"age,omitempty"`

	EyeColor  string `json:"eyeColor,omitempty"`
	HairColor string `json:"hairColor,omitempty"`
	HairStyle string `json:"hairStyle,omitempty"`
	SkinColor string `json:"skinColor,omitempty"`

	// Drakin only, in place of SkinColor/HairStyle/HairColor.
	PrimaryScaleColor   string `json:"primaryScaleColor,omitempty"`
	SecondaryScaleColor string `json:"secondaryScaleColor,omitempty"`
	ScalePattern        string `json:"scalePattern,omitempty"` // empty when the colors match
}

// IsEmpty reports whether every field is at its zero value — used to distinguish
// "client sent no appearance choices at all" (roll everything server-side, for
// backward compatibility with older/simpler callers) from "client submitted a
// specific set of choices" (must be validated in full).
func (a *CharacterAppearance) IsEmpty() bool {
	if a == nil {
		return true
	}
	return a.Strength == 0 && a.Agility == 0 && a.Quickness == 0 && a.Constitution == 0 &&
		a.Perception == 0 && a.Willpower == 0 && a.Empathy == 0 &&
		a.Height == 0 && a.Weight == 0 && a.Age == 0 &&
		a.EyeColor == "" && a.HairColor == "" && a.HairStyle == "" && a.SkinColor == "" &&
		a.PrimaryScaleColor == "" && a.SecondaryScaleColor == "" && a.ScalePattern == ""
}

// RollStats rolls STR/AGI/QUI/CON/PER/WIL/EMP within race's stat ranges.
func RollStats(race int) (str, agi, qui, con, per, wil, emp int) {
	ranges := RaceStatRanges[race]
	rollStat := func(idx int) int {
		r := ranges[idx]
		return r[0] + rand.Intn(r[1]-r[0]+1)
	}
	return rollStat(0), rollStat(1), rollStat(2), rollStat(3), rollStat(4), rollStat(5), rollStat(6)
}

// RollStatsSeeded rolls STR/AGI/QUI/CON/PER/WIL/EMP the same way RollStats does, but
// deterministically from seed instead of the global RNG — the same (seed, race) pair
// always reproduces the same stats. Used by the reroll charm (modern_fixes.scr) so a
// preview rolled on RUB shows exactly what CONCENTRATE will later apply, without
// needing to persist all seven stat values on the item itself (just the seed, in
// ItemVal1).
func RollStatsSeeded(seed int64, race int) (str, agi, qui, con, per, wil, emp int) {
	ranges := RaceStatRanges[race]
	rng := rand.New(rand.NewSource(seed))
	rollStat := func(idx int) int {
		r := ranges[idx]
		return r[0] + rng.Intn(r[1]-r[0]+1)
	}
	return rollStat(0), rollStat(1), rollStat(2), rollStat(3), rollStat(4), rollStat(5), rollStat(6)
}

// RollHeightWeight rolls height (inches) and weight (pounds) within race/gender ranges.
func RollHeightWeight(race, gender int) (height, weight int) {
	hw := raceHeightWeightRanges[race]
	if hw == [4]int{} {
		hw = [4]int{62, 76, 120, 220}
	}
	height = hw[0] + rand.Intn(hw[1]-hw[0]+1)
	weight = hw[2] + rand.Intn(hw[3]-hw[2]+1)
	if gender == GenderFemale {
		height -= 2 + rand.Intn(3)
		weight -= 10 + rand.Intn(20)
		if height < hw[0]-4 {
			height = hw[0] - 4
		}
		if weight < hw[2]-20 {
			weight = hw[2] - 20
		}
	}
	return height, weight
}

// RollAge rolls an age within the race's age range.
func RollAge(race int) int {
	r, ok := RaceAgeRanges[race]
	if !ok {
		r = [2]int{16, 70}
	}
	return r[0] + rand.Intn(r[1]-r[0]+1)
}

func RandomEyeColor() string { return EyeColors[rand.Intn(len(EyeColors))] }

// RandomEyeColorFor picks from the race's own eye colors (EyeColorsFor).
func RandomEyeColorFor(race int) string {
	c := EyeColorsFor(race)
	return c[rand.Intn(len(c))]
}

// RandomScalePattern picks a pattern, or "" when the two colors match.
func RandomScalePattern(primary, secondary string) string {
	if primary == secondary {
		return ""
	}
	return ScalePatterns[rand.Intn(len(ScalePatterns))]
}
func RandomSkinColor() string  { return SkinColors[rand.Intn(len(SkinColors))] }
func RandomHairColor() string  { return HairColors[rand.Intn(len(HairColors))] }
func RandomHairStyle() string  { return HairStyles[rand.Intn(len(HairStyles))] }
func RandomScaleColor() string { return ScaleColors[rand.Intn(len(ScaleColors))] }

// RollCharacterAppearance rolls the stats/height/weight/age portion of character
// creation (the part the original game let players reroll "until they liked what
// they saw"). Eye/hair/skin color and hair style are always an explicit player
// choice, never randomized-then-shown, so they're not included here.
func RollCharacterAppearance(race, gender int) *CharacterAppearance {
	str, agi, qui, con, per, wil, emp := RollStats(race)
	height, weight := RollHeightWeight(race, gender)
	return &CharacterAppearance{
		Strength: str, Agility: agi, Quickness: qui, Constitution: con,
		Perception: per, Willpower: wil, Empathy: emp,
		Height: height, Weight: weight, Age: RollAge(race),
	}
}

// ValidateCharacterAppearance checks that a's fields are legal for the given
// race/gender: each stat within RaceStatRanges, height/weight within
// raceHeightWeightRanges (widened by the same gender-reduction tolerance
// RollHeightWeight applies), age within RaceAgeRanges, and EyeColor/SkinColor/
// HairStyle/HairColor each a member of their fixed list (HairColor must be
// empty when HairStyle is "bald", and non-empty and listed otherwise). Drakin
// have both scale colors instead, and no skin or hair.
func ValidateCharacterAppearance(race, gender int, a *CharacterAppearance) error {
	ranges := RaceStatRanges[race]
	stats := []struct {
		name string
		val  int
		idx  int
	}{
		{"Strength", a.Strength, 0}, {"Agility", a.Agility, 1}, {"Quickness", a.Quickness, 2},
		{"Constitution", a.Constitution, 3}, {"Perception", a.Perception, 4},
		{"Willpower", a.Willpower, 5}, {"Empathy", a.Empathy, 6},
	}
	for _, s := range stats {
		r := ranges[s.idx]
		if s.val < r[0] || s.val > r[1] {
			return fmt.Errorf("%s must be between %d and %d for this race", s.name, r[0], r[1])
		}
	}

	hw := raceHeightWeightRanges[race]
	if hw == [4]int{} {
		hw = [4]int{62, 76, 120, 220}
	}
	minH, maxH := hw[0], hw[1]
	minW, maxW := hw[2], hw[3]
	if gender == GenderFemale {
		minH -= 4
		minW -= 20
	}
	if a.Height < minH || a.Height > maxH {
		return fmt.Errorf("height must be between %d and %d for this race", minH, maxH)
	}
	if a.Weight < minW || a.Weight > maxW {
		return fmt.Errorf("weight must be between %d and %d for this race", minW, maxW)
	}

	ageRange, ok := RaceAgeRanges[race]
	if !ok {
		ageRange = [2]int{16, 70}
	}
	if a.Age < ageRange[0] || a.Age > ageRange[1] {
		return fmt.Errorf("age must be between %d and %d for this race", ageRange[0], ageRange[1])
	}

	if !containsString(EyeColorsFor(race), a.EyeColor) {
		return fmt.Errorf("invalid eye color")
	}
	if HasScales(race) {
		if a.SkinColor != "" || a.HairStyle != "" || a.HairColor != "" {
			return fmt.Errorf("drakin have scales, not skin or hair")
		}
		if !containsString(ScaleColors, a.PrimaryScaleColor) {
			return fmt.Errorf("invalid primary scale color")
		}
		if !containsString(ScaleColors, a.SecondaryScaleColor) {
			return fmt.Errorf("invalid secondary scale color")
		}
		if a.PrimaryScaleColor == a.SecondaryScaleColor {
			if a.ScalePattern != "" {
				return fmt.Errorf("matching scale colors have no pattern")
			}
		} else if !containsString(ScalePatterns, a.ScalePattern) {
			return fmt.Errorf("invalid scale pattern")
		}
		return nil
	}
	if a.PrimaryScaleColor != "" || a.SecondaryScaleColor != "" || a.ScalePattern != "" {
		return fmt.Errorf("only drakin have scales")
	}
	if !containsString(SkinColors, a.SkinColor) {
		return fmt.Errorf("invalid skin color")
	}
	if !containsString(HairStyles, a.HairStyle) {
		return fmt.Errorf("invalid hair style")
	}
	if a.HairStyle == "bald" {
		if a.HairColor != "" {
			return fmt.Errorf("bald characters have no hair color")
		}
	} else if !containsString(HairColors, a.HairColor) {
		return fmt.Errorf("invalid hair color")
	}

	return nil
}

// tierIndex returns which tier (0-based) a value falls into given a race's [min,max]
// range and a set of ascending percentile thresholds (0-100). Used by the
// descriptor helpers below to turn a raw stat into a flavor word.
func tierIndex(val, min, max int, thresholds []int) int {
	if max <= min {
		return 0
	}
	pct := (val - min) * 100 / (max - min)
	for i, t := range thresholds {
		if pct < t {
			return i
		}
	}
	return len(thresholds)
}

// ageDescriptor returns a flavor word for age, relative to the race's lifespan.
func ageDescriptor(age, race int) string {
	r, ok := RaceAgeRanges[race]
	if !ok {
		r = [2]int{16, 70}
	}
	words := []string{"young", "mature", "aging", "old"}
	return words[tierIndex(age, r[0], r[1], []int{30, 60, 85})]
}

// heightDescriptor returns a flavor phrase for height, relative to the race's range.
func heightDescriptor(height, race int) string {
	hw := raceHeightWeightRanges[race]
	if hw == [4]int{} {
		hw = [4]int{62, 76, 120, 220}
	}
	words := []string{"short", "of medium height", "tall"}
	return words[tierIndex(height, hw[0], hw[1], []int{33, 67})]
}

// weightDescriptor returns a flavor phrase for weight, relative to the race's range.
func weightDescriptor(weight, race int) string {
	hw := raceHeightWeightRanges[race]
	if hw == [4]int{} {
		hw = [4]int{62, 76, 120, 220}
	}
	words := []string{"light weight", "of average weight", "heavy"}
	return words[tierIndex(weight, hw[2], hw[3], []int{33, 67})]
}

// buildDescriptor returns a physique flavor word derived from Strength relative
// to the race's range, so it reflects the character's actual stats rather than
// needing its own stored/random field.
func buildDescriptor(strength, race int) string {
	ranges, ok := RaceStatRanges[race]
	if !ok {
		ranges = RaceStatRanges[RaceHuman] // GM-set races (Undead, Troll...) have no stat ranges of their own
	}
	r := ranges[0]
	words := []string{"slender", "wiry", "of average build", "robust", "powerfully built"}
	return words[tierIndex(strength, r[0], r[1], []int{25, 50, 75, 90})]
}
