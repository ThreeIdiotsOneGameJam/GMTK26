package game

var buildingCoinCosts = [...]int32{
	BuildingUnknown:  0,
	BuildingForester: 10,
	BuildingMine:     14,
	BuildingBarracks: 24,
	BuildingFarm:     12,
	BuildingTownhall: 0,
}

var unitCoinCosts = [...]int32{
	UnitUnknown: 0,
	UnitPeasant: 8,
	UnitArcher:  14,
	UnitKnight:  22,
	UnitScout:   8,
}

func BuildingCost(b BuildingType) int32 {
	if int(b) >= len(buildingCoinCosts) {
		return 0
	}
	return buildingCoinCosts[b]
}

func BuildingResourceCost(b BuildingType) Resources {
	switch b {
	case BuildingBarracks:
		return Resources{
			ResourceWood:  6,
			ResourceStone: 4,
		}
	default:
		return make(Resources)
	}
}

func BuildingProduces(b BuildingType, tile TileType) map[ResourceType]uint32 {
	switch b {
	case BuildingForester:
		return map[ResourceType]uint32{ResourceWood: 1}
	case BuildingMine:
		switch tile {
		case TileRock:
			return map[ResourceType]uint32{ResourceStone: 1}
		case TileIron:
			return map[ResourceType]uint32{ResourceIron: 1}
		case TileCoal, TileGold:
			return nil
		default:
			return nil
		}
	case BuildingFarm:
		return map[ResourceType]uint32{ResourceFood: 1}
	default:
		return nil
	}
}

func UnitCost(t UnitType) int32 {
	if int(t) >= len(unitCoinCosts) {
		return 0
	}
	return unitCoinCosts[t]
}

func UnitResourceCost(t UnitType) Resources {
	switch t {
	case UnitPeasant:
		return Resources{ResourceFood: 4}
	case UnitArcher:
		return Resources{
			ResourceFood: 6,
			ResourceWood: 4,
		}
	case UnitKnight:
		return Resources{
			ResourceFood: 8,
			ResourceIron: 4,
		}
	default:
		return make(Resources)
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
			coins += BuildingCoinsProduces(cell.BuildingType(), cell.Tile)
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

func CanAffordBuildingAfterRoundIncome(
	m *Map,
	owner int8,
	b BuildingType,
	coins int32,
	resources Resources,
) bool {
	incomeCoins, incomeResources := FactionRoundIncome(m, owner)
	if coins+incomeCoins < BuildingCost(b) {
		return false
	}
	for resource, amount := range BuildingResourceCost(b) {
		if resources[resource]+incomeResources[resource] < amount {
			return false
		}
	}
	return true
}

func BuildingCoinsProduces(b BuildingType, tile TileType) int32 {
	if b == BuildingTownhall {
		return 1
	}
	if b == BuildingMine && tile == TileGold {
		return 2
	}
	return 0
}
