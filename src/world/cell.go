package world

type Cell struct {
	Tile     Tile
	Resource Resource
	Building Building
	Troop    Troop
}

type Resource interface{}
type Building interface{}
type Troop interface{}
