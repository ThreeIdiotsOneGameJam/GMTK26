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

// FactionRoundIncome returns the resources generated at the start of a round
// by the buildings currently owned by a faction.
func FactionRoundIncome(m *Map, owner int8) (int32, Resources) {
	resources := make(Resources)
	if m == nil {
		return 0, resources
	}

	var coins int32
	for x := range m.Grid {
		for y := range m.Grid[x] {
			cell := &m.Grid[x][y]
			if cell.Owner != owner || !cell.HasBuilding() {
				continue
			}
			for resource, amount := range BuildingProduces(cell.BuildingType(), cell.Tile) {
				resources[resource] += amount
			}
			coins += BuildingCoinsProduces(cell.BuildingType())
		}
	}
	return coins, resources
}

// CanAffordUnitAfterRoundIncome mirrors round resolution, where building
// production is credited before a submitted recruitment action is validated.
func CanAffordUnitAfterRoundIncome(
	m *Map,
	owner int8,
	t UnitType,
	coins int32,
	resources Resources,
) bool {
	incomeCoins, incomeResources := FactionRoundIncome(m, owner)
	if coins+incomeCoins < UnitCost(t) {
		return false
	}
	for resource, amount := range UnitResourceCost(t) {
		if resources[resource]+incomeResources[resource] < amount {
			return false
		}
	}
	return true
}

func BuildingCoinsProduces(b BuildingType) int32 {
	if b == BuildingTownhall {
		return 1
	}
	return 0
}
