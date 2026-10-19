package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Player export/import — lets an admin download a character from one server
// (e.g. production) and load it into another (e.g. a GM's local copy).
//
// The player document is carried as MongoDB relaxed Extended JSON of the
// stored BSON, not the Player struct's JSON encoding: dozens of persisted
// fields are tagged json:"-", so a plain JSON round-trip would silently lose
// them. Extended JSON also keeps dates and ObjectIDs typed.

// PlayerExportFormat identifies a player export file.
const PlayerExportFormat = "lofp-player"

// PlayerExportVersion is bumped if the envelope layout ever changes.
const PlayerExportVersion = 1

// playerExportStripped are stored fields that never leave the server: the
// identity/ownership of this copy and the character's API key.
var playerExportStripped = map[string]bool{
	"_id": true, "accountId": true, "deletedAt": true, "apiKeyHash": true, "apiKeyPrefix": true,
}

type playerExportEnvelope struct {
	Format      string          `json:"format"`
	Version     int             `json:"version"`
	GameVersion string          `json:"gameVersion"`
	ExportedAt  time.Time       `json:"exportedAt"`
	Name        string          `json:"name"`
	Player      json.RawMessage `json:"player"`
}

// ExportPlayer returns the export file for the active character firstName,
// as stored at its last save.
func (e *GameEngine) ExportPlayer(ctx context.Context, firstName string) ([]byte, string, error) {
	if e.db == nil {
		return nil, "", fmt.Errorf("no database connection")
	}
	player, err := e.resolvePlayerByName(ctx, firstName)
	if err != nil {
		return nil, "", err
	}
	var doc bson.D
	if err := e.db.Collection("players").FindOne(ctx, bson.M{"_id": player.ID}).Decode(&doc); err != nil {
		return nil, "", fmt.Errorf("player '%s' not found", firstName)
	}
	data, err := encodePlayerExport(doc, player.FullName())
	if err != nil {
		return nil, "", err
	}
	return data, player.FirstName, nil
}

// encodePlayerExport wraps a stored player document in the export envelope,
// minus the playerExportStripped fields.
func encodePlayerExport(doc bson.D, name string) ([]byte, error) {
	var kept bson.D
	for _, el := range doc {
		if !playerExportStripped[el.Key] {
			kept = append(kept, el)
		}
	}
	ext, err := bson.MarshalExtJSON(kept, false, false)
	if err != nil {
		return nil, err
	}
	out, err := json.Marshal(playerExportEnvelope{
		Format:      PlayerExportFormat,
		Version:     PlayerExportVersion,
		GameVersion: GameVersion,
		ExportedAt:  time.Now().UTC(),
		Name:        name,
		Player:      ext,
	})
	if err != nil {
		return nil, err
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, out, "", "  "); err != nil {
		return nil, err
	}
	return pretty.Bytes(), nil
}

// ImportPlayer creates a new character on accountID from an export file.
// newFirstName, if set, renames the character (needed when the name is taken).
// The import always gets a fresh database ID and no API key.
func (e *GameEngine) ImportPlayer(ctx context.Context, data []byte, accountID, newFirstName string) (*Player, error) {
	if e.db == nil {
		return nil, fmt.Errorf("no database connection")
	}
	player, err := decodePlayerExport(data)
	if err != nil {
		return nil, err
	}

	if newFirstName = strings.TrimSpace(newFirstName); newFirstName != "" {
		if !validNamePattern.MatchString(newFirstName) {
			return nil, fmt.Errorf("first name must be 3-20 letters (may include ' and -)")
		}
		player.FirstName = newFirstName
	}
	taken, err := e.IsFirstNameTaken(ctx, player.FirstName)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, fmt.Errorf("a character named '%s' already exists on this server; import it under a new first name", player.FirstName)
	}

	player.ID = bson.NewObjectID()
	player.AccountID = accountID
	player.DeletedAt = nil
	player.APIKeyHash = ""
	player.APIKeyPrefix = ""
	player.UpdatedAt = time.Now()
	// A room that only exists on the source server (e.g. from a GM-uploaded
	// script) would leave the character "nowhere"; use the new-character start.
	if e.rooms[player.RoomNumber] == nil {
		player.RoomNumber = 201
	}
	if _, err := e.db.Collection("players").InsertOne(ctx, player); err != nil {
		return nil, err
	}
	return player, nil
}

// decodePlayerExport validates an export file and returns its player.
func decodePlayerExport(data []byte) (*Player, error) {
	var env playerExportEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("not a valid player export file: %v", err)
	}
	if env.Format != PlayerExportFormat || len(env.Player) == 0 {
		return nil, fmt.Errorf("not a LoFP player export file")
	}
	if env.Version > PlayerExportVersion {
		return nil, fmt.Errorf("export file version %d is newer than this server supports (%d)", env.Version, PlayerExportVersion)
	}
	var player Player
	if err := bson.UnmarshalExtJSON(env.Player, false, &player); err != nil {
		return nil, fmt.Errorf("could not read player data: %v", err)
	}
	if player.FirstName == "" {
		return nil, fmt.Errorf("export file has no character name")
	}
	return &player, nil
}
