package game

type Map struct {
	Seed     int64    `json:"seed"`
	Grid     [][]Cell `json:"grid"`
	GridSize GridSize `json:"grid_size"`
}

type GridSize struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
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
	if m.GridSize == (GridSize{}) {
		m.GridSize = GridSize{X: 96, Y: 96}
	}

	m.Grid = generateMap(m.GridSize, m.Seed)
	spreadResources(m, m.Seed)
}

func (m *Map) GetNeighbors(pos Hex) Neighbors {
	if pos.X%2 == 0 {
		return Neighbors{
			NW: m.GetCell(pos.Add(Hex{X: -1, Y: -1})),
			N:  m.GetCell(pos.Add(Hex{X: 0, Y: -1})),
			NE: m.GetCell(pos.Add(Hex{X: 1, Y: -1})),
			SW: m.GetCell(pos.Add(Hex{X: -1, Y: 0})),
			S:  m.GetCell(pos.Add(Hex{X: 0, Y: 1})),
			SE: m.GetCell(pos.Add(Hex{X: 1, Y: 0})),
		}
	}

	return Neighbors{
		NW: m.GetCell(pos.Add(Hex{X: -1, Y: 0})),
		N:  m.GetCell(pos.Add(Hex{X: 0, Y: -1})),
		NE: m.GetCell(pos.Add(Hex{X: 1, Y: 0})),
		SW: m.GetCell(pos.Add(Hex{X: -1, Y: 1})),
		S:  m.GetCell(pos.Add(Hex{X: 0, Y: 1})),
		SE: m.GetCell(pos.Add(Hex{X: 1, Y: 1})),
	}
}

func (m *Map) GetCell(pos Hex) *Cell {
	if pos.X < 0 || pos.Y < 0 || int(pos.X) >= len(m.Grid) || int(pos.Y) >= len(m.Grid[pos.X]) {
		return nil
	}

	return &m.Grid[pos.X][pos.Y]
}
