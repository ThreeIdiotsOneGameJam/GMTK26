package game

//go:generate stringer -type=BuildingType -trimprefix=Building

type BuildingType uint8

const (
	BuildingUnknown BuildingType = iota
	BuildingForester
	BuildingMine
	BuildingBarracks
	BuildingFarm
	BuildingTownhall
)

func BuildingCanPlace(m *Map, building BuildingType, hex Hex) bool {
	if building == BuildingTownhall {
		return false
	}

	if !m.HexInsideBounds(hex) {
		return false
	}

	cell := m.GetCell(hex)
	if cell.HasBuilding() {
		return false
	}

	switch building {
	case BuildingMine:
		return cell.Tile == TileRock || cell.Tile == TileIron || cell.Tile == TileGold
	case BuildingForester:
		return cell.Tile == TileForest || cell.Tile == TileJungle
	case BuildingBarracks:
		return cell.Tile == TilePlains
	case BuildingFarm:
		if cell.Tile != TilePlains {
			return false
		}

		neighbors := m.GetNeighbors(hex)
		isWater := func(cell *Cell) bool {
			if cell == nil {
				return false
			}
			return cell.Tile == TileWater
		}

		return isWater(neighbors.N) || isWater(neighbors.NE) || isWater(neighbors.NW) || isWater(neighbors.S) || isWater(neighbors.SW) || isWater(neighbors.SE)
	}

	return false
}

func BuildingMaxHP(b BuildingType) int8 {
	switch b {
	case BuildingTownhall:
		return 24
	case BuildingBarracks:
		return 16
	case BuildingForester:
		return 10
	case BuildingMine:
		return 12
	case BuildingFarm:
		return 10
	default:
		return 0
	}
}
