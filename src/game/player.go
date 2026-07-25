package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/google/uuid"
	"github.com/threeidiotsonegamejam/gmtk26/src/storage"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type ClientID string

type Player struct {
	ClientID   ClientID `json:"client_id"`
	PlayerName string   `json:"player_name"`
	Color      util.RGB `json:"color"`
}

func (p *Player) UnmarshalJSON(data []byte) error {
	type playerPayload struct {
		ClientID   *ClientID        `json:"client_id"`
		PlayerName *string          `json:"player_name"`
		Color      *json.RawMessage `json:"color"`
	}

	var payload playerPayload
	if err := jsonutil.DecodeStrict(data, &payload); err != nil {
		return err
	}
	if payload.ClientID == nil {
		return fmt.Errorf("player: missing or null client_id")
	}
	if payload.PlayerName == nil {
		return fmt.Errorf("player: missing or null player_name")
	}
	if payload.Color == nil {
		return fmt.Errorf("player: missing or null color")
	}
	rawClientID := string(*payload.ClientID)
	clientID, err := uuid.Parse(rawClientID)
	if err != nil || clientID.String() != rawClientID {
		return fmt.Errorf("player: client_id must be a canonical UUID")
	}

	var colorComponents []uint8
	if err := jsonutil.DecodeStrict(*payload.Color, &colorComponents); err != nil {
		return fmt.Errorf("player color: %w", err)
	}
	if len(colorComponents) != len(util.RGB{}) {
		return fmt.Errorf("player color: got %d components, want 3", len(colorComponents))
	}

	p.ClientID = *payload.ClientID
	p.PlayerName = *payload.PlayerName
	p.Color = util.RGB{colorComponents[0], colorComponents[1], colorComponents[2]}
	return nil
}

var PlayerData = &Player{}

func LoadOrCreatePlayerData() error {
	if os.Getenv("NEWUUID") != "" {
		createPlayerData()
		return SavePlayerData()
	}

	var loadedPlayer Player
	loaded, loadErr := storage.Load("player", &loadedPlayer)
	if loadErr == nil && loaded {
		*PlayerData = loadedPlayer
		return nil
	}

	createPlayerData()
	if loadErr != nil && !errors.Is(loadErr, storage.ErrInvalidData) {
		PlayerData.ClientID = ClientID(uuid.Nil.String())
		return fmt.Errorf("load player data: %w", loadErr)
	}
	if err := SavePlayerData(); err != nil {
		PlayerData.ClientID = ClientID(uuid.Nil.String())
		if loadErr != nil {
			return errors.Join(
				fmt.Errorf("load player data: %w", loadErr),
				err,
			)
		}
		return err
	}

	return nil
}

func createPlayerData() {
	uid, err := uuid.NewRandom()
	if err != nil {
		uid = uuid.Nil
	}

	color := rand.Uint32()
	*PlayerData = Player{
		ClientID:   ClientID(uid.String()),
		PlayerName: "player",
		Color:      util.RGB{uint8(color >> 16), uint8(color >> 8), uint8(color)},
	}
}

func SavePlayerData() error {
	if err := storage.Save("player", PlayerData); err != nil {
		return fmt.Errorf("save player data: %w", err)
	}

	return nil
}
