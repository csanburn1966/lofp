package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NumberReservation records that a block of room/item/monster numbers has been
// set aside for a GM's work. Purely advisory — nothing in the engine or the GM
// script pipeline enforces it, so GMs can still collaborate on the same numbers.
// It exists so the Number Ranges view (number_ranges.go) can show who is
// working where. Stored in MongoDB only (no in-memory copy), so every machine
// sees the same list.
type NumberReservation struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Kind      string        `bson:"kind" json:"kind"` // room, item, monster
	Start     int           `bson:"start" json:"start"`
	End       int           `bson:"end" json:"end"`
	GM        string        `bson:"gm" json:"gm"` // who the block is for
	Note      string        `bson:"note,omitempty" json:"note,omitempty"`
	CreatedBy string        `bson:"createdBy" json:"createdBy"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedBy string        `bson:"updatedBy,omitempty" json:"updatedBy,omitempty"`
	UpdatedAt time.Time     `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

const numberReservationsCollection = "number_reservations"

// Normalize trims and validates a reservation's user-supplied fields.
func (r *NumberReservation) Normalize() error {
	r.Kind = strings.ToLower(strings.TrimSpace(r.Kind))
	r.GM = strings.TrimSpace(r.GM)
	r.Note = strings.TrimSpace(r.Note)
	switch r.Kind {
	case "room", "item", "monster":
	default:
		return fmt.Errorf("kind must be room, item or monster")
	}
	if r.Start < 0 || r.End < r.Start {
		return fmt.Errorf("range must be non-negative with end >= start")
	}
	if r.GM == "" {
		return fmt.Errorf("GM name is required")
	}
	if len(r.Note) > 500 {
		return fmt.Errorf("note is too long (max 500 characters)")
	}
	return nil
}

// ListNumberReservations returns every reservation, ordered by kind then start.
func (e *GameEngine) ListNumberReservations(ctx context.Context) ([]NumberReservation, error) {
	if e.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	opts := options.Find().SetSort(bson.D{{Key: "kind", Value: 1}, {Key: "start", Value: 1}})
	cursor, err := e.db.Collection(numberReservationsCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	out := []NumberReservation{}
	if err := cursor.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveNumberReservation inserts (zero ID) or updates a reservation. by is the
// GM making the change. Returns the stored reservation.
func (e *GameEngine) SaveNumberReservation(ctx context.Context, res *NumberReservation, by string) (*NumberReservation, error) {
	if e.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	if err := res.Normalize(); err != nil {
		return nil, err
	}
	coll := e.db.Collection(numberReservationsCollection)
	now := time.Now()
	if res.ID.IsZero() {
		res.ID = bson.NewObjectID()
		res.CreatedBy, res.CreatedAt = by, now
		res.UpdatedBy, res.UpdatedAt = "", time.Time{}
		if _, err := coll.InsertOne(ctx, res); err != nil {
			return nil, err
		}
		return res, nil
	}
	update := bson.M{"$set": bson.M{
		"kind": res.Kind, "start": res.Start, "end": res.End, "gm": res.GM, "note": res.Note,
		"updatedBy": by, "updatedAt": now,
	}}
	var stored NumberReservation
	err := coll.FindOneAndUpdate(ctx, bson.M{"_id": res.ID}, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&stored)
	if err != nil {
		return nil, fmt.Errorf("reservation not found")
	}
	return &stored, nil
}

// DeleteNumberReservation removes a reservation by ID.
func (e *GameEngine) DeleteNumberReservation(ctx context.Context, id bson.ObjectID) error {
	if e.db == nil {
		return fmt.Errorf("database not available")
	}
	res, err := e.db.Collection(numberReservationsCollection).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("reservation not found")
	}
	return nil
}
