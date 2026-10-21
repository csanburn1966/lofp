package engine

import (
	"context"
	"sort"
	"strings"

	"github.com/jonradoff/lofp/internal/gameworld"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Number ranges — a GM-facing view of which room/item/monster numbers are taken
// and which contiguous blocks are free, so blocks can be handed out to GMs
// working in parallel without collisions.
//
// There is no engine-side maximum: rooms, items and monsters are all map[int]
// keyed, so anything above a kind's current max is free. Draft GM scripts are
// parsed (never applied) and included, since they've already claimed their
// numbers even though they aren't in the live maps. ANTI.SCR's placeholder items
// (INUMBER 1-1989, see project_anti_scr_fallback memory) are treated as free —
// redefining one of those numbers is exactly what the placeholders exist for.
//
// The vocabulary tables — item adjectives (ADJDEF), monster adjectives (MADJDEF)
// and nouns (NOUNDEF) — are numbered the same way and get the same report, plus
// the word each number maps to so GMs can look up IDs while writing scripts.

// vocabKinds are the report kinds backed by a vocabulary table, in display order.
var vocabKinds = []string{"adjective", "monadj", "noun"}

// NumberBlock is a contiguous run of used numbers that all come from one source.
type NumberBlock struct {
	Start  int    `json:"start"`
	End    int    `json:"end"`
	Source string `json:"source"`           // script filename
	GM     bool   `json:"gm,omitempty"`     // source is a GM-uploaded script
	Draft  bool   `json:"draft,omitempty"`  // source is a draft GM script (not in game)
	Author string `json:"author,omitempty"` // GM script uploader
}

// NumberGap is a contiguous run of free numbers between used ones.
type NumberGap struct {
	Start int `json:"start"`
	End   int `json:"end"`
	Size  int `json:"size"`
}

// NumberConflict is a number a draft script defines that something else also
// defines — applying the draft would overwrite (or be overwritten by) it.
type NumberConflict struct {
	Number int      `json:"number"`
	Draft  string   `json:"draft"`
	Others []string `json:"others"`
}

// NumberKindRanges is the full picture for one kind (room, item or monster).
type NumberKindRanges struct {
	Kind      string           `json:"kind"`
	Count     int              `json:"count"`
	Min       int              `json:"min"`
	Max       int              `json:"max"` // highest used; everything above is free
	Blocks    []NumberBlock    `json:"blocks"`
	Gaps      []NumberGap      `json:"gaps"`
	Conflicts []NumberConflict `json:"conflicts,omitempty"`
	Names     map[int]string   `json:"names,omitempty"` // vocabulary kinds only: ID -> word
}

// NumberRangesReport is the response for GET /api/gm/ranges.
type NumberRangesReport struct {
	Kinds  []NumberKindRanges `json:"kinds"`
	Drafts []string           `json:"drafts"` // draft scripts folded into the report
}

// gmScriptMeta is the slice of a GMScript document the ranges report needs.
type gmScriptMeta struct {
	Filename   string `bson:"_id"`
	Content    string `bson:"content"`
	Draft      bool   `bson:"draft"`
	UploadedBy string `bson:"uploadedBy"`
}

type numberClaim struct {
	source string
	draft  bool
}

// NumberRanges builds the used/free report from the live world plus draft GM scripts.
func (e *GameEngine) NumberRanges(ctx context.Context) (*NumberRangesReport, error) {
	// GM script metadata: which source filenames are GM scripts, who uploaded
	// them, and the content of drafts (live ones are already in the maps).
	gmScripts := map[string]gmScriptMeta{} // keyed by lowercased filename
	if e.db != nil {
		opts := options.Find().SetProjection(bson.M{"history": 0})
		cursor, err := e.db.Collection(gmScriptsCollection).Find(ctx, bson.M{}, opts)
		if err != nil {
			return nil, err
		}
		var metas []gmScriptMeta
		if err := cursor.All(ctx, &metas); err != nil {
			return nil, err
		}
		for _, m := range metas {
			gmScripts[strings.ToLower(m.Filename)] = m
		}
	}

	claims := map[string]map[int][]numberClaim{"room": {}, "item": {}, "monster": {}}
	names := map[string]map[int]string{}
	for _, kind := range vocabKinds {
		claims[kind] = map[int][]numberClaim{}
		names[kind] = map[int]string{}
	}
	for kind, table := range map[string]map[int]string{"adjective": e.adjectives, "monadj": e.monAdjs, "noun": e.nouns} {
		for n, name := range table {
			claims[kind][n] = append(claims[kind][n], numberClaim{source: e.vocabSources[kind][n]})
			names[kind][n] = name
		}
	}
	for n, r := range e.rooms {
		claims["room"][n] = append(claims["room"][n], numberClaim{source: r.SourceFile})
	}
	for n, it := range e.items {
		if strings.EqualFold(it.SourceFile, "ANTI.SCR") {
			continue
		}
		claims["item"][n] = append(claims["item"][n], numberClaim{source: it.SourceFile})
	}
	for n, m := range e.monsters {
		claims["monster"][n] = append(claims["monster"][n], numberClaim{source: m.SourceFile})
	}

	report := &NumberRangesReport{Drafts: []string{}}
	if ScriptParser != nil {
		for _, m := range gmScripts {
			if !m.Draft || m.Content == "" {
				continue
			}
			parsed, err := ScriptParser(m.Content, m.Filename)
			if err != nil {
				continue // an unparseable draft claims nothing usable
			}
			report.Drafts = append(report.Drafts, m.Filename)
			dc := numberClaim{source: m.Filename, draft: true}
			for _, r := range parsed.Rooms {
				claims["room"][r.Number] = append(claims["room"][r.Number], dc)
			}
			for _, it := range parsed.Items {
				claims["item"][it.Number] = append(claims["item"][it.Number], dc)
			}
			for _, mon := range parsed.Monsters {
				claims["monster"][mon.Number] = append(claims["monster"][mon.Number], dc)
			}
			addDraftVocab := func(kind string, id int, name string) {
				claims[kind][id] = append(claims[kind][id], dc)
				if _, live := names[kind][id]; !live {
					names[kind][id] = name
				}
			}
			for _, a := range parsed.Adjectives {
				addDraftVocab("adjective", a.ID, a.Name)
			}
			for _, a := range parsed.MonsterAdjs {
				addDraftVocab("monadj", a.ID, a.Name)
			}
			for _, nd := range parsed.Nouns {
				addDraftVocab("noun", nd.ID, nd.Name)
			}
		}
		sort.Strings(report.Drafts)
	}

	for _, kind := range []string{"room", "item", "monster"} {
		report.Kinds = append(report.Kinds, buildKindRanges(kind, claims[kind], gmScripts))
	}
	for _, kind := range vocabKinds {
		kr := buildKindRanges(kind, claims[kind], gmScripts)
		kr.Names = names[kind]
		report.Kinds = append(report.Kinds, kr)
	}
	return report, nil
}

// recordVocabSources notes which script file defined each noun/adjective/monster
// adjective in parsed, for the Number Ranges report. Later definitions win, the
// same as the name tables themselves.
func (e *GameEngine) recordVocabSources(parsed *gameworld.ParsedData) {
	if e.vocabSources == nil {
		e.vocabSources = map[string]map[int]string{}
		for _, kind := range vocabKinds {
			e.vocabSources[kind] = map[int]string{}
		}
	}
	for _, d := range parsed.Adjectives {
		e.vocabSources["adjective"][d.ID] = d.SourceFile
	}
	for _, d := range parsed.MonsterAdjs {
		e.vocabSources["monadj"][d.ID] = d.SourceFile
	}
	for _, d := range parsed.Nouns {
		e.vocabSources["noun"][d.ID] = d.SourceFile
	}
}

func buildKindRanges(kind string, claims map[int][]numberClaim, gmScripts map[string]gmScriptMeta) NumberKindRanges {
	out := NumberKindRanges{Kind: kind, Blocks: []NumberBlock{}, Gaps: []NumberGap{}}
	nums := make([]int, 0, len(claims))
	for n := range claims {
		nums = append(nums, n)
	}
	if len(nums) == 0 {
		return out
	}
	sort.Ints(nums)
	out.Count, out.Min, out.Max = len(nums), nums[0], nums[len(nums)-1]

	// Display owner of a number: the live definition if there is one, otherwise
	// the (first) draft claiming it. Extra claims become conflicts.
	owner := func(n int) numberClaim {
		cs := claims[n]
		for _, c := range cs {
			if !c.draft {
				return c
			}
		}
		return cs[0]
	}
	for _, n := range nums {
		cs := claims[n]
		if len(cs) < 2 {
			continue
		}
		for i, c := range cs {
			if !c.draft {
				continue
			}
			var others []string
			for j, o := range cs {
				if j != i && o.source != c.source {
					others = append(others, o.source)
				}
			}
			if len(others) > 0 {
				out.Conflicts = append(out.Conflicts, NumberConflict{Number: n, Draft: c.source, Others: others})
			}
		}
	}

	newBlock := func(n int, c numberClaim) NumberBlock {
		b := NumberBlock{Start: n, End: n, Source: c.source, Draft: c.draft}
		if m, ok := gmScripts[strings.ToLower(c.source)]; ok {
			b.GM, b.Author = true, m.UploadedBy
		}
		return b
	}
	cur := newBlock(nums[0], owner(nums[0]))
	for _, n := range nums[1:] {
		c := owner(n)
		if n == cur.End+1 && c.source == cur.Source && c.draft == cur.Draft {
			cur.End = n
			continue
		}
		out.Blocks = append(out.Blocks, cur)
		if n > cur.End+1 {
			out.Gaps = append(out.Gaps, NumberGap{Start: cur.End + 1, End: n - 1, Size: n - cur.End - 1})
		}
		cur = newBlock(n, c)
	}
	out.Blocks = append(out.Blocks, cur)
	return out
}
