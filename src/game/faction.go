package game

type Faction struct {
	Index     int       `json:"index"`
	Player    *Player   `json:"player"`
	AI        bool      `json:"ai"`
	Coins     int32     `json:"coins"`
	Points    int32     `json:"points"`
	Resources Resources `json:"resources"`
	Alive     bool      `json:"alive"`
}
