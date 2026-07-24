package world

import "github.com/threeidiotsonegamejam/gmtk26/src/game"

type Cell struct {
	Tile     Tile
	Resource game.ResourceType
	Building Building
	Troop    Troop
}

type Building interface{}
type Troop interface{}
