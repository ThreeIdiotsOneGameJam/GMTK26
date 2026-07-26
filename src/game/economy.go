package game

var buildingCoinCosts = [...]int32{
	BuildingUnknown:  0,
	BuildingForester: 10,
	BuildingMine:     14,
	BuildingBarracks: 24,
	BuildingFarm:     12,
	BuildingTownhall: 0,
	BuildingBank:     25,
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
		case TileGold:
			return map[ResourceType]uint32{ResourceGold: 1}
		case TileCoal:
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

// FactionRoundIncome returns unconditional production at the start of a round.
// Consumption-driven buildings are resolved by ResolveFactionRoundIncome,
// because their output depends on the faction's current resource stockpile.
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
			if len(BuildingConsumes(cell.BuildingType())) > 0 {
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

// ResolveFactionRoundIncome applies production before consumption and returns
// both the Coin income and a new final resource stockpile. The input map is
// never mutated. Returning final resources instead of an unsigned net delta
// prevents consuming buildings from underflowing Resource values.
func ResolveFactionRoundIncome(
	m *Map,
	owner int8,
	current Resources,
) (int32, Resources) {
	coins, produced := FactionRoundIncome(m, owner)
	resources := cloneResources(current)
	for resource, amount := range produced {
		resources[resource] += amount
	}
	if m == nil {
		return coins, resources
	}

	for x := range m.Grid {
		for y := range m.Grid[x] {
			cell := &m.Grid[x][y]
			if cell.Owner != owner || !cell.HasBuilding() {
				continue
			}
			consumes := BuildingConsumes(cell.BuildingType())
			if len(consumes) == 0 {
				continue
			}
			canRun := true
			for resource, amount := range consumes {
				if resources[resource] < amount {
					canRun = false
					break
				}
			}
			if !canRun {
				continue
			}
			for resource, amount := range consumes {
				resources[resource] -= amount
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
	incomeCoins, projectedResources := ResolveFactionRoundIncome(m, owner, resources)
	if coins+incomeCoins < UnitCost(t) {
		return false
	}
	for resource, amount := range UnitResourceCost(t) {
		if projectedResources[resource] < amount {
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
	incomeCoins, projectedResources := ResolveFactionRoundIncome(m, owner, resources)
	if coins+incomeCoins < BuildingCost(b) {
		return false
	}
	for resource, amount := range BuildingResourceCost(b) {
		if projectedResources[resource] < amount {
			return false
		}
	}
	return true
}

func BuildingCoinsProduces(b BuildingType, tile TileType) int32 {
	if b == BuildingTownhall {
		return 1
	}
	if b == BuildingBank {
		return 5
	}
	if b == BuildingMine && tile == TileGold {
		return 2
	}
	return 0
}

func BuildingConsumes(b BuildingType) Resources {
	if b == BuildingBank {
		return Resources{ResourceGold: 5}
	}
	return make(Resources)
}
