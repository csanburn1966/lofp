package engine

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Psionic mastery is a house addition: the original game had spell mastery
// (MASTER) and weapon specialization but nothing for psionicists. It mirrors
// spell mastery:
//   - Cost: 8 BP for rank 1, 4 BP for each additional rank.
//   - Max ranks: discipline level + 1.
//   - Prereq: the discipline's school skill (Mind over Matter or Mind over
//     Mind) at 2 × the discipline's level, the same threshold MASTER uses
//     for a spell's school.
//   - Each rank cuts the psi cost by 1, down to half the base cost.
//   - Damage disciplines deal +1.5 damage per rank.
//   - Haste (the spell or psionic Speed) cancels all of it, as it cancels
//     spell mastery (MAGIC.TXT).

// psiMasteryLevel returns the player's purchased mastery rank for a discipline.
func psiMasteryLevel(player *Player, disc *PsiDiscipline) int {
	if player.PsiMastery == nil {
		return 0
	}
	return player.PsiMastery[disc.ID]
}

// psiMasteryMaxRank is the most mastery ranks a discipline can take.
func psiMasteryMaxRank(disc *PsiDiscipline) int {
	return disc.Level + 1
}

// psiMasteryTotalCost returns the cumulative BP cost to reach rank.
func psiMasteryTotalCost(rank int) int {
	if rank <= 0 {
		return 0
	}
	return 8 + (rank-1)*4
}

// psiSchoolSkill returns the skill ID of a discipline's school.
func psiSchoolSkill(disc *PsiDiscipline) int {
	if disc.School == "Mind over Matter" {
		return 28
	}
	return 27
}

// effectivePsiCost returns a discipline's psi cost after mastery: 1 less per
// rank, never below half the base cost (minimum 1 for a discipline that costs
// anything at all).
func effectivePsiCost(disc *PsiDiscipline, mastery int) int {
	base := disc.PsiCost
	if base <= 0 {
		return base
	}
	minCost := base / 2
	if minCost < 1 {
		minCost = 1
	}
	cost := base - mastery
	if cost < minCost {
		cost = minCost
	}
	return cost
}

// playerPsiCost is the psi the player pays for disc right now: full price
// while hasted, since haste cancels mastery.
func playerPsiCost(player *Player, disc *PsiDiscipline) int {
	if isHasted(player) {
		return disc.PsiCost
	}
	return effectivePsiCost(disc, psiMasteryLevel(player, disc))
}

// psiRankHasBenefit reports whether taking rank newRank of disc gains anything:
// damage disciplines always gain damage; others gain only while the psi cost
// is still dropping.
func psiRankHasBenefit(disc *PsiDiscipline, newRank int) bool {
	if disc.Effect == "damage" {
		return true
	}
	return effectivePsiCost(disc, newRank) < effectivePsiCost(disc, newRank-1)
}

// resolveKnownPsi finds a discipline the player can use by number or name.
func resolveKnownPsi(player *Player, query string) *PsiDiscipline {
	var disc *PsiDiscipline
	if n, err := strconv.Atoi(query); err == nil {
		disc = FindPsiByID(n)
	} else {
		disc = FindPsiByName(query)
	}
	if disc == nil || !psiKnown(player, disc) {
		return nil
	}
	return disc
}

// doMasterPsi handles MASTER PSI <discipline> (and MASTER <discipline name>
// when the name isn't one of the player's spells).
func (e *GameEngine) doMasterPsi(ctx context.Context, player *Player, query string) *CommandResult {
	if query == "" {
		return &CommandResult{Messages: []string{"Master which discipline? (MASTER PSI <discipline name or number>)"}}
	}
	if player.Skills[26] == 0 && !player.IsGM {
		return &CommandResult{Messages: []string{"You have no training in Psionics."}}
	}
	disc := resolveKnownPsi(player, query)
	if disc == nil {
		return &CommandResult{Messages: []string{"You don't know that discipline."}}
	}

	current := psiMasteryLevel(player, disc)
	maxRanks := psiMasteryMaxRank(disc)
	if current >= maxRanks {
		return &CommandResult{Messages: []string{fmt.Sprintf("You have fully mastered %s (%d/%d ranks).", disc.Name, current, maxRanks)}}
	}
	if !psiRankHasBenefit(disc, current+1) {
		return &CommandResult{Messages: []string{fmt.Sprintf("You have mastered %s as far as it can be taken (%d ranks).", disc.Name, current)}}
	}

	schoolID := psiSchoolSkill(disc)
	required := 2 * disc.Level
	if player.Skills[schoolID] < required {
		return &CommandResult{Messages: []string{fmt.Sprintf("You need %s skill %d to master %s (rank %d). You have %d.", disc.School, required, disc.Name, current+1, player.Skills[schoolID])}}
	}

	bpCost := 8
	if current > 0 {
		bpCost = 4
	}
	if player.BuildPoints < bpCost {
		return &CommandResult{Messages: []string{fmt.Sprintf("You need %d build points to master %s (you have %d).", bpCost, disc.Name, player.BuildPoints)}}
	}

	if player.PsiMastery == nil {
		player.PsiMastery = make(map[int]int)
	}
	newRank := current + 1
	player.PsiMastery[disc.ID] = newRank
	player.BuildPoints -= bpCost

	msgs := []string{fmt.Sprintf("You meditate on %s, achieving rank %d of mastery. (%s) [-%d BP, %d BP remaining]",
		disc.Name, newRank, strings.Repeat("*", newRank), bpCost, player.BuildPoints)}
	if disc.Effect == "damage" {
		msgs = append(msgs, fmt.Sprintf("Mastery damage bonus: +%d (when not hasted).", newRank*3/2))
	}
	if disc.PsiCost > 0 {
		msgs = append(msgs, fmt.Sprintf("Psi cost for %s is now %d (when not hasted).", disc.Name, effectivePsiCost(disc, newRank)))
	}

	e.SavePlayer(ctx, player)
	return &CommandResult{Messages: msgs, PlayerState: player}
}

// gmPsiMastery handles @MASTERY <player> PSI <discipline# or name> <level>.
// args is the full @MASTERY argument list (args[1] is "psi"). Like @MASTERY for
// spells it bypasses the skill and BP prerequisites, and refunds BP when a
// rank is lowered.
func (e *GameEngine) gmPsiMastery(ctx context.Context, target *Player, args []string) *CommandResult {
	if len(args) < 4 {
		return &CommandResult{Messages: []string{"Usage: @mastery <player> psi <discipline# or name> <level>"}}
	}
	query := strings.Join(args[2:len(args)-1], " ")
	var disc *PsiDiscipline
	if id, convErr := strconv.Atoi(query); convErr == nil {
		disc = FindPsiByID(id)
	} else {
		disc = FindPsiByName(query)
	}
	if disc == nil {
		return &CommandResult{Messages: []string{fmt.Sprintf("Unknown discipline: %s", query)}}
	}
	level, err := strconv.Atoi(args[len(args)-1])
	if err != nil || level < 0 {
		return &CommandResult{Messages: []string{"Level must be a non-negative integer (use 0 to clear mastery)."}}
	}
	maxRank := psiMasteryMaxRank(disc)
	if level > maxRank {
		return &CommandResult{Messages: []string{fmt.Sprintf("Max mastery for %s (level %d discipline) is %d.", disc.Name, disc.Level, maxRank)}}
	}
	if target.PsiMastery == nil {
		target.PsiMastery = make(map[int]int)
	}
	oldRank := target.PsiMastery[disc.ID]

	var bpRefund int
	if level < oldRank {
		bpRefund = psiMasteryTotalCost(oldRank) - psiMasteryTotalCost(level)
		target.BuildPoints += bpRefund
	}
	if level == 0 {
		delete(target.PsiMastery, disc.ID)
	} else {
		target.PsiMastery[disc.ID] = level
	}
	e.SavePlayer(ctx, target)

	msg := fmt.Sprintf("Set %s psionic mastery for %s to rank %d (max %d).", disc.Name, target.FullName(), level, maxRank)
	if bpRefund > 0 {
		msg += fmt.Sprintf(" Refunded %d BP (%d BP total).", bpRefund, target.BuildPoints)
	}
	return &CommandResult{Messages: []string{msg}}
}

// psiMasteryListLines returns the @MASTERY list-mode lines for psionic mastery.
func psiMasteryListLines(target *Player) []string {
	var ids []int
	for id, rank := range target.PsiMastery {
		if rank > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	sort.Ints(ids)
	lines := []string{fmt.Sprintf("Psionic mastery for %s:", target.FullName())}
	for _, id := range ids {
		name, maxRank := fmt.Sprintf("[unknown #%d]", id), 0
		if disc := FindPsiByID(id); disc != nil {
			name, maxRank = disc.Name, psiMasteryMaxRank(disc)
		}
		lines = append(lines, fmt.Sprintf("  #%-4d %-35s rank %d / %d", id, name, target.PsiMastery[id], maxRank))
	}
	return lines
}
