package game

import "github.com/threeidiotsonegamejam/gmtk26/src/util/vec"

type Map struct {
	Seed     int64     `json:"seed"`
	Grid     [][]Cell  `json:"grid"`
	GridSize vec.Vec2i `json:"grid_size"`
}

type Cell struct {
	Tile     TileType     `json:"tile"`
	Building BuildingType `json:"building,omitempty"`
	Troop    TroopType    `json:"troop,omitempty"`
}

type Neighbors struct {
	NW *Cell
	N  *Cell
	NE *Cell
	SE *Cell
	S  *Cell
	SW *Cell
}

func (m *Map) Generate() {
	if m.GridSize == (vec.Vec2i{}) {
		m.GridSize = vec.Vec2i{X: 96, Y: 96}
	}

	m.Grid = generateMap(m.GridSize, m.Seed)
	spreadResources(m, m.Seed)
}

func (m *Map) GetNeighbors(pos Hex) Neighbors {
	if pos.X%2 == 0 {
		return Neighbors{
			NW: m.GetCell(pos.Add(NewHex(-1, -1))),
			N:  m.GetCell(pos.Add(NewHex(0, -1))),
			NE: m.GetCell(pos.Add(NewHex(1, -1))),
			SW: m.GetCell(pos.Add(NewHex(-1, 0))),
			S:  m.GetCell(pos.Add(NewHex(0, 1))),
			SE: m.GetCell(pos.Add(NewHex(1, 0))),
		}
	}

	return Neighbors{
		NW: m.GetCell(pos.Add(NewHex(-1, 0))),
		N:  m.GetCell(pos.Add(NewHex(0, -1))),
		NE: m.GetCell(pos.Add(NewHex(1, 0))),
		SW: m.GetCell(pos.Add(NewHex(-1, 1))),
		S:  m.GetCell(pos.Add(NewHex(0, 1))),
		SE: m.GetCell(pos.Add(NewHex(1, 1))),
	}
}

func (m *Map) GetCell(pos Hex) *Cell {
	if pos.X < 0 || pos.Y < 0 || int(pos.X) >= len(m.Grid) || int(pos.Y) >= len(m.Grid[pos.X]) {
		return nil
	}

	return &m.Grid[pos.X][pos.Y]
}

func (m *Map) HexInsideBounds(hex Hex) bool {
	if hex.X < 0 || hex.X >= m.GridSize.X || hex.Y < 0 || hex.Y >= m.GridSize.Y {
		return false
	}

	if m.GetCell(hex).Tile == TileVoid {
		return false
	}

	return true
}
