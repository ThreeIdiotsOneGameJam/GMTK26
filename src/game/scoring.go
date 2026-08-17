package game

func BuildingControlScore(building BuildingType, tile TileType) int32 {
	switch building {
	case BuildingFarm, BuildingForester:
		return 2
	case BuildingMine:
		if tile == TileGold {
			return 5
		}
		return 3
	case BuildingBarracks:
		return 3
	case BuildingBank:
		return 3
	default:
		return 0
	}
}

func BuildingDestructionScore(building BuildingType, tile TileType) int32 {
	switch building {
	case BuildingTownhall:
		return 30
	case BuildingFarm, BuildingForester:
		return 8
	case BuildingMine:
		if tile == TileGold {
			return 12
		}
		return 10
	case BuildingBarracks:
		return 15
	case BuildingBank:
		return 12
	default:
		return 0
	}
}

func UnitDestructionScore(unit UnitType) int32 {
	switch unit {
	case UnitScout:
		return 3
	case UnitPeasant:
		return 5
	case UnitArcher:
		return 8
	case UnitKnight:
		return 12
	default:
		return 0
	}
}

func FactionControlScore(m *Map, owner int8) int32 {
	if m == nil {
		return 0
	}

	var score int32
	for x := range m.Grid {
		for y := range m.Grid[x] {
			cell := &m.Grid[x][y]
			if cell.Owner != owner || !cell.HasBuilding() {
				continue
			}
			score += BuildingControlScore(cell.BuildingType(), cell.Tile)
		}
	}
	return score
}
