package world

type Cell struct {
	Tile     Tile
	Building Building
	Troop    Troop
}

type Building interface{}
type Troop interface{}
