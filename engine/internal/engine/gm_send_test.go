package engine

import (
	"slices"
	"testing"
)

// TestGMSendWhileConcealed: @snd and @exclude send their text even while the GM
// is @hide or @invis (api.go drops a concealed actor's RoomBroadcast unless the
// result sets ConcealedBroadcastOK).
func TestGMSendWhileConcealed(t *testing.T) {
	b := newBatchEnv(t)
	gm := b.player("Gamemaster")
	bob := b.player("Bob")
	sly := b.player("Sly")
	sly.Hidden = true
	gm.IsGM, gm.GMInvis, gm.Hidden = true, true, true
	if !gm.IsConcealed() {
		t.Fatal("GM not concealed")
	}

	r := b.e.ProcessCommand(b.ctx, gm, "@snd The ground trembles.")
	if !r.ConcealedBroadcastOK || !slices.Equal(r.RoomBroadcast, []string{"The ground trembles."}) {
		t.Errorf("@snd: %+v", r)
	}

	for _, name := range []string{"bob", "sly"} { // Sly is hidden
		r = b.e.ProcessCommand(b.ctx, gm, "@exclude "+name+" A chill runs down your spine.")
		want := map[string]string{"bob": bob.FirstName, "sly": sly.FirstName}[name]
		if !r.ConcealedBroadcastOK || !slices.Equal(r.RoomBroadcast, []string{"A chill runs down your spine."}) || !slices.Equal(r.BroadcastExclude, []string{want}) {
			t.Errorf("@exclude %s: %+v", name, r)
		}
	}
	r = b.e.ProcessCommand(b.ctx, gm, "@exclude nobody Hello.")
	if len(r.RoomBroadcast) != 0 {
		t.Errorf("@exclude nobody: %+v", r)
	}
	bob.RoomNumber = batchIndoor
	if r = b.e.ProcessCommand(b.ctx, gm, "@exclude bob Hello."); len(r.RoomBroadcast) != 0 {
		t.Errorf("@exclude someone elsewhere: %+v", r)
	}
}
