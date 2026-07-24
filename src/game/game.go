package game

type Game struct {
	GameID      uint64     `json:"game_id"`
	GameCode    string     `json:"game_code"`
	Public      bool       `json:"public"`
	HostID      ClientID   `json:"host_id"`
	Factions    [4]Faction `json:"factions"`
	Multiplayer bool       `json:"multiplayer"`
	MaxPlayers  uint8      `json:"max_players"` // MaxPlayers has to be one of {1, 2, 3, 4}. Can only be 1 with Multiplayer == false. When less than 4, the remaining (4-MaxPlayers) factions will be automatically set as AI.

}
