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

func BuildingProduces(b BuildingType) map[ResourceType]uint32 {
	switch b {
	case BuildingForester:
		return map[ResourceType]uint32{ResourceWood: 2}
	case BuildingMine:
		return map[ResourceType]uint32{ResourceStone: 2, ResourceCoal: 1}
	case BuildingFarm:
		return map[ResourceType]uint32{ResourceWood: 1}
	default:
		return nil
	}
}
