package engine

import (
	"fmt"
	"testing"
)

// TestJewelryMetalOdds: low treasure is mostly plain/tin/copper with gold rare
// but possible; high treasure favors the precious metals.
func TestJewelryMetalOdds(t *testing.T) {
	const n = 200000
	shares := func(level int) map[int]float64 {
		counts := map[int]int{}
		for i := 0; i < n; i++ {
			counts[rollJewelryMetal(level).adj]++
		}
		out := map[int]float64{}
		for adj, c := range counts {
			out[adj] = 100 * float64(c) / n
		}
		return out
	}
	names := map[int]string{0: "plain", 744: "tin", 68: "copper", 32: "brass", 39: "bronze", 292: "silver", 147: "gold", 563: "platinum", 5: "alzyron"}
	for _, level := range []int{1, 10, 20, 40, 60} {
		s := shares(level)
		line := fmt.Sprintf("treasure %2d:", level)
		for _, m := range jewelryMetals {
			line += fmt.Sprintf(" %s %.1f%%", names[m.adj], s[m.adj])
		}
		t.Log(line)
	}
	low, high := shares(1), shares(60)
	if g := low[147]; g <= 0.3 || g >= 3 {
		t.Errorf("gold at treasure 1 should be rare but possible, got %.2f%%", g)
	}
	if low[0]+low[744]+low[68] < 70 {
		t.Errorf("treasure 1 should be mostly plain/tin/copper")
	}
	if high[292]+high[147]+high[563]+high[5] < 60 {
		t.Errorf("treasure 60 should be mostly precious metals")
	}
}
