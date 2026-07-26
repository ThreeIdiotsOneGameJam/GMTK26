package game

func BuildingCost(b BuildingType) int32 {
	switch b {
	case BuildingForester:
		return 10
	case BuildingMine:
		return 15
	case BuildingBarracks:
		return 20
	case BuildingFarm:
		return 12
	default:
		return 0
	}
}

func BuildingProduces(b BuildingType, tile TileType) map[ResourceType]uint32 {
	switch b {
	case BuildingForester:
		return map[ResourceType]uint32{ResourceWood: 2}
	case BuildingMine:
		switch tile {
		case TileIron:
			return map[ResourceType]uint32{ResourceIron: 2}
		case TileCoal:
			return map[ResourceType]uint32{ResourceCoal: 2}
		case TileGold:
			return map[ResourceType]uint32{ResourceGold: 1}
		default:
			return map[ResourceType]uint32{ResourceStone: 2}
		}
	case BuildingFarm:
		return map[ResourceType]uint32{ResourceFood: 2}
	default:
		return nil
	}
}

func UnitCost(t UnitType) int32 {
	switch t {
	case UnitPeasant:
		return 10
	case UnitArcher:
		return 20
	case UnitKnight:
		return 30
	case UnitScout:
		return 10
	default:
		return 0
	}
}

func UnitResourceCost(t UnitType) Resources {
	switch t {
	case UnitPeasant:
		return Resources{ResourceFood: 1}
	case UnitArcher:
		return Resources{ResourceFood: 3}
	case UnitKnight:
		return Resources{ResourceFood: 5}
	default:
		return nil
	}
}

func BuildingCoinsProduces(b BuildingType) int32 {
	if b == BuildingTownhall {
		return 1
	}
	return 0
}
