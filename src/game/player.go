package game

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/threeidiotsonegamejam/gmtk26/src/storage"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

type ClientID string

type Player struct {
	ClientID   ClientID `json:"client_id"`
	PlayerName string   `json:"player_name"`
	Color      util.RGB `json:"color"`
}

var PlayerData = &Player{}

func LoadOrCreatePlayerData() error {
	loaded, err := storage.Load("player", PlayerData)
	if err != nil {
		return fmt.Errorf("load player data: %w", err)
	}

	if loaded {
		return nil
	}

	uid, err := uuid.NewRandom()
	if err != nil {
		uid = uuid.Nil // should have different server handling
	}

	PlayerData.ClientID = ClientID(uid.String())

	PlayerData.PlayerName = "player"

	color := rand.Uint32()
	PlayerData.Color = util.RGB{uint8(color >> 16), uint8(color >> 8), uint8(color)}

	if err := SavePlayerData(); err != nil {
		return fmt.Errorf("save new player data: %w", err)
	}

	return nil
}

func SavePlayerData() error {
	if err := storage.Save("player", PlayerData); err != nil {
		return fmt.Errorf("save player data: %w", err)
	}

	return nil
}
