package game

// MovementOrder is a private, persistent destination assigned to one unit.
// Current changes after every successful activation; Destination is retained
// until reached, rerouted, cancelled, or contested.
type MovementOrder struct {
	Current     Hex `json:"current"`
	Destination Hex `json:"destination"`
}

type AttackOrder struct {
	From       Hex `json:"from"`
	TargetTile Hex `json:"target_tile"`
}

type AttackEvent struct {
	Unit   UnitType `json:"unit"`
	Owner  int8     `json:"owner"`
	From   Hex      `json:"from"`
	Target Hex      `json:"target"`
}

// MovementEvent is public resolved history used to animate authoritative
// movement. Path includes both the starting cell and final cell.
type MovementEvent struct {
	Unit  UnitType `json:"unit"`
	Owner int8     `json:"owner"`
	Path  []Hex    `json:"path"`
}

func UnitMovementBudget(unit UnitType) int {
	switch unit {
	case UnitScout:
		return 4
	case UnitPeasant, UnitArcher:
		return 2
	case UnitKnight:
		return 3
	default:
		return 0
	}
}

func TerrainMovementCost(tile TileType) int {
	switch tile {
	case TilePlains, TileRock, TileIron, TileCoal, TileGold:
		return 1
	case TileForest:
		return 2
	case TileDesert, TileJungle:
		return 3
	default:
		return 0
	}
}

// FindUnitPath finds the cheapest legal route for a faction's unit.
// Units may traverse friendly or unclaimed land, but every other unit blocks
// traversal. The starting unit itself is exempt from that blocker.
func (m *Map) FindUnitPath(faction int8, start, goal Hex) ([]Hex, bool) {
	if m == nil {
		return nil, false
	}

	startCell := m.GetCell(start)
	goalCell := m.GetCell(goal)
	if startCell == nil || goalCell == nil ||
		!startCell.HasUnits() ||
		startCell.Units[0].Owner != faction ||
		goalCell.HasUnits() && goal != start ||
		!cellFriendlyOrUnclaimed(goalCell, faction) {
		return nil, false
	}

	return m.FindPathWithOptions(start, goal, PathOptions{
		Cost: func(_ Hex, destination Hex, cell *Cell) int {
			if !cellFriendlyOrUnclaimed(cell, faction) {
				return 0
			}
			if destination != start && cell.HasUnits() {
				return 0
			}
			return TerrainMovementCost(cell.Tile)
		},
		MinimumCost: 1,
	})
}

func cellFriendlyOrUnclaimed(cell *Cell, faction int8) bool {
	return cell != nil && (cell.Owner == -1 || cell.Owner == faction)
}

// AdvanceUnitPath returns the portion of path traversed in one activation.
// The first legal step is always allowed, even when it costs more than budget.
func (m *Map) AdvanceUnitPath(path []Hex, budget int) []Hex {
	if m == nil || len(path) == 0 {
		return nil
	}
	traversed := []Hex{path[0]}
	remaining := budget
	for i := 1; i < len(path); i++ {
		cell := m.GetCell(path[i])
		if cell == nil {
			break
		}
		cost := TerrainMovementCost(cell.Tile)
		if cost <= 0 {
			break
		}
		if len(traversed) > 1 && cost > remaining {
			break
		}
		traversed = append(traversed, path[i])
		if cost > remaining {
			remaining = 0
		} else {
			remaining -= cost
		}
	}
	return traversed
}

// MovementTurnStops projects activation endpoints for a currently valid
// route. It is used by the UI preview and intentionally ignores future map
// mutations.
func (m *Map) MovementTurnStops(path []Hex, budget int) []Hex {
	if len(path) < 2 {
		return nil
	}
	stops := make([]Hex, 0)
	for offset := 0; offset < len(path)-1; {
		segment := m.AdvanceUnitPath(path[offset:], budget)
		if len(segment) < 2 {
			break
		}
		offset += len(segment) - 1
		stops = append(stops, path[offset])
	}
	return stops
}

func HexAdjacent(a, b Hex) bool {
	return a.Distance(b) == 1
}

func (m *Map) FindAdjacentApproachPath(faction int8, start, target Hex) ([]Hex, Hex, bool) {
	if m == nil || start == target {
		return nil, Hex{}, false
	}
	startCell := m.GetCell(start)
	if startCell == nil || !startCell.HasUnits() || startCell.Units[0].Owner != faction {
		return nil, Hex{}, false
	}
	var bestPath []Hex
	var bestApproach Hex
	var bestDist int64
	found := false
	for _, neighbor := range pathNeighborHexes(target) {
		cell := m.GetCell(neighbor)
		if cell == nil || !cellFriendlyOrUnclaimed(cell, faction) {
			continue
		}
		if neighbor != start && cell.HasUnits() {
			continue
		}
		if TerrainMovementCost(cell.Tile) <= 0 {
			continue
		}
		path, ok := m.FindPathWithOptions(start, neighbor, PathOptions{
			Cost: func(_ Hex, destination Hex, cell *Cell) int {
				if !cellFriendlyOrUnclaimed(cell, faction) {
					return 0
				}
				if destination != start && cell.HasUnits() {
					return 0
				}
				return TerrainMovementCost(cell.Tile)
			},
			MinimumCost: 1,
		})
		if !ok {
			continue
		}
		dist := int64(len(path))
		if !found || dist < bestDist {
			bestPath = path
			bestApproach = neighbor
			bestDist = dist
			found = true
		}
	}
	if !found {
		return nil, Hex{}, false
	}
	return bestPath, bestApproach, true
}
