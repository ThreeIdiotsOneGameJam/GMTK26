package game

import "fmt"

// Funds is an immutable-by-convention snapshot used by action validation.
// Constructors in this file always copy Resources so callers cannot mutate a
// faction or a shared balance table through a validation result.
type Funds struct {
	Coins     int32
	Resources Resources
}

type ActionValidation struct {
	Valid        bool
	Status       ActionResultStatus
	Message      string
	CoinCost     int32
	ResourceCost Resources
}

func CurrentFunds(faction Faction) Funds {
	return Funds{
		Coins:     faction.Coins,
		Resources: cloneResources(faction.Resources),
	}
}

// ProjectedRoundFunds mirrors the authoritative resolution order: income is
// credited before a pending manual action is validated.
func ProjectedRoundFunds(m *Map, owner int8, faction Faction) Funds {
	funds := CurrentFunds(faction)
	coins, resources := FactionRoundIncome(m, owner)
	funds.Coins += coins
	for resource, amount := range resources {
		funds.Resources[resource] += amount
	}
	return funds
}

func ValidateBuildAction(
	m *Map,
	owner int8,
	payload BuildActionPayload,
	funds Funds,
) ActionValidation {
	validation := actionCostValidation(
		BuildingCost(payload.Building),
		BuildingResourceCost(payload.Building),
	)
	source, target := cellsForAction(m, payload.From, payload.To)
	if source == nil || target == nil ||
		!source.HasUnits() ||
		source.Units[0].Type != UnitScout ||
		source.Units[0].Owner != owner {
		return validation.invalid("A friendly Scout is required to build")
	}
	if payload.From != payload.To && !HexAdjacent(payload.From, payload.To) {
		return validation.invalid("Scout may only build on its tile or an adjacent tile")
	}
	if target.Owner != -1 && target.Owner != owner {
		return validation.invalid("Cannot build in enemy territory")
	}
	if target.HasUnits() && target.Units[0].Owner != owner {
		return validation.invalid("Cannot build beneath an enemy unit")
	}
	if !BuildingCanPlace(m, payload.Building, payload.To) {
		return validation.invalid("Building cannot be placed on that terrain")
	}
	return validation.affordable(funds, fmt.Sprintf("%s constructed", payload.Building))
}

func ValidateRecruitAction(
	m *Map,
	owner int8,
	payload RecruitActionPayload,
	funds Funds,
) ActionValidation {
	validation := actionCostValidation(
		UnitCost(payload.Unit),
		UnitResourceCost(payload.Unit),
	)
	source, target := cellsForAction(m, payload.From, payload.To)
	if source == nil || target == nil || source.Owner != owner {
		return validation.invalid("Recruitment source is not friendly")
	}
	if !HexAdjacent(payload.From, payload.To) {
		return validation.invalid("Recruitment target must be adjacent")
	}
	if TerrainMovementCost(target.Tile) <= 0 ||
		target.Owner != -1 && target.Owner != owner ||
		target.HasUnits() {
		return validation.invalid("Recruitment target is not available")
	}
	validUnit := payload.Unit >= UnitPeasant && payload.Unit <= UnitScout
	if !validUnit ||
		source.BuildingType() == BuildingTownhall && payload.Unit != UnitScout ||
		source.BuildingType() != BuildingTownhall && source.BuildingType() != BuildingBarracks {
		return validation.invalid("Building cannot recruit that unit")
	}
	return validation.affordable(funds, fmt.Sprintf("%s recruited", payload.Unit))
}

func ValidateAdjacentAttackAction(
	m *Map,
	owner int8,
	payload AttackActionPayload,
) ActionValidation {
	validation := ActionValidation{
		Status:       ActionResultInvalid,
		ResourceCost: make(Resources),
	}
	source, target := cellsForAction(m, payload.From, payload.To)
	if source == nil || target == nil ||
		!source.HasUnits() ||
		source.Units[0].Type == UnitScout ||
		source.Units[0].Owner != owner {
		return validation.invalid("A friendly non-Scout unit is required to attack")
	}
	if !HexAdjacent(payload.From, payload.To) {
		return validation.invalid("Attack target must be adjacent")
	}
	if !cellHasEnemy(target, owner) {
		return validation.invalid("Target has no enemy units or buildings")
	}
	validation.Valid = true
	validation.Status = ActionResultSucceeded
	validation.Message = "Attack ordered"
	return validation
}

func actionCostValidation(coins int32, resources Resources) ActionValidation {
	return ActionValidation{
		Status:       ActionResultInvalid,
		CoinCost:     coins,
		ResourceCost: cloneResources(resources),
	}
}

func (validation ActionValidation) affordable(funds Funds, success string) ActionValidation {
	if funds.Coins < validation.CoinCost {
		validation.Status = ActionResultInsufficientCoins
		validation.Message = "Not enough coins"
		return validation
	}
	for resource, amount := range validation.ResourceCost {
		if funds.Resources[resource] < amount {
			validation.Message = fmt.Sprintf("Not enough %s", resource.String())
			return validation
		}
	}
	validation.Valid = true
	validation.Status = ActionResultSucceeded
	validation.Message = success
	return validation
}

func (validation ActionValidation) invalid(message string) ActionValidation {
	validation.Valid = false
	validation.Status = ActionResultInvalid
	validation.Message = message
	return validation
}

func cellsForAction(m *Map, from, to Hex) (*Cell, *Cell) {
	if m == nil {
		return nil, nil
	}
	return m.GetCell(from), m.GetCell(to)
}

func cellHasEnemy(cell *Cell, owner int8) bool {
	if cell == nil {
		return false
	}
	for _, unit := range cell.Units {
		if unit.Owner != owner {
			return true
		}
	}
	return cell.HasBuilding() && cell.Owner >= 0 && cell.Owner != owner
}

func cloneResources(resources Resources) Resources {
	cloned := make(Resources, len(resources))
	for resource, amount := range resources {
		cloned[resource] = amount
	}
	return cloned
}
