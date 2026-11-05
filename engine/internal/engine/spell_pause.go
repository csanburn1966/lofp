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
	since := p.logoutTime()
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

// logoutTime is when the player's last session ended: LoggedOutAt, or the last
// save when the session never closed cleanly.
func (p *Player) logoutTime() time.Time {
	if !p.LoggedOutAt.IsZero() {
		return p.LoggedOutAt
	}
	return p.UpdatedAt
}

// timerRunningAtLogout reports whether expiry was still in the future when the
// player logged out, i.e. whether pauseSpellTimersOffline will carry it over into
// the next session. Only meaningful before that has run.
func (p *Player) timerRunningAtLogout(expiry time.Time) bool {
	since := p.logoutTime()
	return !expiry.IsZero() && !since.IsZero() && expiry.After(since)
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
		&p.ImmobilizedExpiry,
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
