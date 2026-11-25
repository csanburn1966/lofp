package engine

// Guard duty (GUARD, player_commands.go) belongs to the room it was taken up in
// and breaks when the guard leaves it by any means: walking, following, a script
// MOVE, a spell, being carried. Walking and following clear it on the spot; the
// rest are caught by GuardRoom, which keeps a guard from counting in another room,
// and by breakStaleGuard at the guard's next command.

// breakGuard ends all the player's guard duty.
func (p *Player) breakGuard() {
	p.GuardTargets, p.GuardPortals, p.GuardItems, p.GuardRoom = nil, nil, nil, 0
}

// breakStaleGuard ends guard duty taken up in a room the player has since left.
func (p *Player) breakStaleGuard() {
	if p.GuardRoom != 0 && p.GuardRoom != p.RoomNumber {
		p.breakGuard()
	}
}

// onGuard reports whether the player is guarding anyone or anything here.
func (p *Player) onGuard() bool {
	return p.GuardRoom == p.RoomNumber &&
		(len(p.GuardTargets) > 0 || len(p.GuardPortals) > 0 || len(p.GuardItems) > 0)
}

func (p *Player) guardsPlayer(firstName string) bool {
	return p.GuardRoom == p.RoomNumber && containsString(p.GuardTargets, firstName)
}

func (p *Player) guardsPortal(arch int) bool {
	return p.GuardRoom == p.RoomNumber && containsInt(p.GuardPortals, arch)
}

func (p *Player) guardsItem(arch int) bool {
	return p.GuardRoom == p.RoomNumber && containsInt(p.GuardItems, arch)
}
