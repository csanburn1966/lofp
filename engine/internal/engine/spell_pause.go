package engine

import "time"

// pauseSpellTimersOffline makes timed spell effects stop counting down while the
// player is logged off. Every expiry that was still in the future at logout is
// pushed forward by the time spent offline, so a 60-minute buff with 40 minutes
// left at logout still has 40 minutes left at the next login. Called from
// RestoreTransientState, which every login path (web, API key, telnet, SSH) runs.
//
// The logout time is LoggedOutAt (HandlePlayerDisconnect), falling back to the last
// save (UpdatedAt) when the session never closed cleanly, e.g. a server restart.
//
// Covers negative effects too (Silence, Imprison, Slow, entangles), so logging off
// can't be used to wait one out. Timers on the player's own items (Storm/Inferno/
// Winter Blade imbues, temporary Tempering) pause as well; items left lying in a
// room keep counting. A new timed effect needs adding to spellTimers.
func (p *Player) pauseSpellTimersOffline() {
	since := p.LoggedOutAt
	if since.IsZero() {
		since = p.UpdatedAt
	}
	p.LoggedOutAt = time.Time{}
	offline := time.Since(since)
	if since.IsZero() || offline <= 0 {
		return
	}

	shift := func(t *time.Time) {
		if t.After(since) {
			*t = t.Add(offline)
		}
	}
	for _, t := range p.spellTimers() {
		shift(t)
	}
	for i := range p.TimedDefenseBuffs {
		shift(&p.TimedDefenseBuffs[i].Expiry)
	}
	for i := range p.Entangles {
		shift(&p.Entangles[i].Expiry)
	}

	sinceUnix, offlineSecs := since.Unix(), int64(offline/time.Second)
	var shiftItem func(ii *InventoryItem)
	shiftItem = func(ii *InventoryItem) {
		if ii == nil {
			return
		}
		shift(&ii.BladeSpellExpiry)
		if ii.TemperedUntil > sinceUnix {
			ii.TemperedUntil += offlineSecs
		}
		for i := range ii.Contents {
			shiftItem(&ii.Contents[i])
		}
	}
	shiftItem(p.Wielded)
	shiftItem(p.OffHand)
	for _, items := range [][]InventoryItem{p.Inventory, p.Worn, p.BankItems} {
		for i := range items {
			shiftItem(&items[i])
		}
	}
}

// spellTimers lists the player's persisted single-value spell expiries.
func (p *Player) spellTimers() []*time.Time {
	return []*time.Time{
		&p.TelepathyExpiry,
		&p.StrengthBuffExpiry,
		&p.AgilityBuffExpiry,
		&p.MysticArmorExpiry,
		&p.HasteExpiry,
		&p.SlowExpiry,
		&p.ParanoiaExpiry,
		&p.SilencedExpiry,
		&p.ImprisonedExpiry,
		&p.FlyExpiry,
		&p.HeatShieldExpiry,
		&p.ColdShieldExpiry,
		&p.CamouflageExpiry,
		&p.LightSpellExpiry,
		&p.NightVisionExpiry,
		&p.DivineBlessingExpiry,
		&p.ResistWeatherExpiry,
		&p.ClawGrowthExpiry,
		&p.RepelPlantsExpiry,
		&p.RepelPlantsAndWebsExpiry,
	}
}
