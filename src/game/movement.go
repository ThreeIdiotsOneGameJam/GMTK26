package game

// MovementOrder is a private, persistent destination assigned to one troop.
// Current changes after every successful activation; Destination is retained
// until reached, rerouted, cancelled, or contested.
type MovementOrder struct {
	Current     Hex `json:"current"`
	Destination Hex `json:"destination"`
}

// MovementEvent is public resolved history used to animate authoritative
// movement. Path includes both the starting cell and final cell.
type MovementEvent struct {
	Troop TroopType `json:"troop"`
	Owner int8      `json:"owner"`
	Path  []Hex     `json:"path"`
}

func TroopMovementBudget(troop TroopType) int {
	switch troop {
	case TroopScout:
		return 3
	case TroopPeasant, TroopArcher:
		return 2
	case TroopKnight:
		return 4
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

// FindTroopPath finds the cheapest legal route for a faction's troop.
// Troops may traverse friendly or unclaimed land, but every other troop blocks
// traversal. The starting troop itself is exempt from that blocker.
func (m *Map) FindTroopPath(faction int8, start, goal Hex) ([]Hex, bool) {
	if m == nil {
		return nil, false
	}

	startCell := m.GetCell(start)
	goalCell := m.GetCell(goal)
	if startCell == nil || goalCell == nil ||
		startCell.Troop == TroopUnknown ||
		startCell.TroopOwner != faction ||
		goalCell.Troop != TroopUnknown && goal != start ||
		!cellFriendlyOrUnclaimed(goalCell, faction) {
		return nil, false
	}

	return m.FindPathWithOptions(start, goal, PathOptions{
		Cost: func(_ Hex, destination Hex, cell *Cell) int {
			if !cellFriendlyOrUnclaimed(cell, faction) {
				return 0
			}
			if destination != start && cell.Troop != TroopUnknown {
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

// AdvanceTroopPath returns the portion of path traversed in one activation.
// The first legal step is always allowed, even when it costs more than budget.
func (m *Map) AdvanceTroopPath(path []Hex, budget int) []Hex {
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
		segment := m.AdvanceTroopPath(path[offset:], budget)
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
