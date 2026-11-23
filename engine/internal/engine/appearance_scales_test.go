package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/jonradoff/lofp/internal/gameworld"
)

func TestDrakinScaleAppearance(t *testing.T) {
	a := RollCharacterAppearance(RaceDrakin, GenderMale)
	a.EyeColor = "yellow"
	a.PrimaryScaleColor = "olive green"
	a.SecondaryScaleColor = "crimson"
	a.ScalePattern = "stripes"
	if err := ValidateCharacterAppearance(RaceDrakin, GenderMale, a); err != nil {
		t.Fatalf("valid drakin rejected: %v", err)
	}
	withHair := *a
	withHair.HairStyle, withHair.HairColor = "long", "black"
	if ValidateCharacterAppearance(RaceDrakin, GenderMale, &withHair) == nil {
		t.Error("drakin with hair accepted")
	}
	noScales := *a
	noScales.SecondaryScaleColor = ""
	if ValidateCharacterAppearance(RaceDrakin, GenderMale, &noScales) == nil {
		t.Error("drakin without a secondary scale color accepted")
	}
	blueEyes := *a
	blueEyes.EyeColor = "blue"
	if ValidateCharacterAppearance(RaceDrakin, GenderMale, &blueEyes) == nil {
		t.Error("drakin with blue eyes accepted")
	}
	noPattern := *a
	noPattern.ScalePattern = ""
	if ValidateCharacterAppearance(RaceDrakin, GenderMale, &noPattern) == nil {
		t.Error("drakin with two colors and no pattern accepted")
	}
	solid := *a
	solid.SecondaryScaleColor, solid.ScalePattern = "olive green", ""
	if err := ValidateCharacterAppearance(RaceDrakin, GenderMale, &solid); err != nil {
		t.Errorf("solid-colored drakin rejected: %v", err)
	}
	h := RollCharacterAppearance(RaceHuman, GenderMale)
	h.EyeColor, h.SkinColor, h.HairStyle, h.HairColor = "blue", "fair", "long", "black"
	h.PrimaryScaleColor = "violet"
	if ValidateCharacterAppearance(RaceHuman, GenderMale, h) == nil {
		t.Error("human with scales accepted")
	}

	e := NewGameEngine(nil, &gameworld.ParsedData{})
	p := e.CreateNewPlayer(context.Background(), "Scaly", "Test", RaceDrakin, GenderMale, a)
	if p.SkinColor != "" || p.HairStyle != "" || p.HairColor != "" {
		t.Errorf("drakin got skin/hair: %q %q %q", p.SkinColor, p.HairStyle, p.HairColor)
	}
	viewer := e.CreateNewPlayer(context.Background(), "Viewer", "Test", RaceHuman, GenderFemale, nil)
	out := strings.Join(e.examinePlayer(viewer, p).Messages, "\n")
	if !strings.Contains(out, "yellow eyes and olive green scales with crimson stripes") || strings.Contains(out, "hair") {
		t.Errorf("drakin LOOK:\n%s", out)
	}
}
