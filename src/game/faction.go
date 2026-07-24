package game

type Faction struct {
	Player *Player `json:"player"`
	AI     bool    `json:"ai"`
	Coins  int32   `json:"coins"`
	Points int32   `json:"points"`
}
