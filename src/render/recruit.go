package render

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
)

func (r *WorldRenderer) updateRecruitPlacement(m *game.Map, hex game.Hex, place bool) {
	if global.UIBlocksWorldInput ||
		!r.ActionsEnabled ||
		r.MovementAnimating() ||
		r.RecruitToPlace == game.UnitUnknown ||
		r.SelectedHex == nil ||
		r.SelectedKind != SelectionBuilding ||
		!place {
		return
	}

	from := *r.SelectedHex
	if !r.canRecruitAt(m, from, hex, r.RecruitToPlace) {
		return
	}

	if r.OnRecruit == nil || r.OnRecruit(from, hex, r.RecruitToPlace) {
		r.clearPlacementSelection()
		r.ClearQueuedBuilding()
	}
}

func (r *WorldRenderer) canRecruitAt(m *game.Map, from, to game.Hex, unit game.UnitType) bool {
	source := m.GetCell(from)
	target := m.GetCell(to)
	if source == nil || target == nil ||
		source.Owner != r.LocalFaction ||
		!game.HexAdjacent(from, to) ||
		game.TerrainMovementCost(target.Tile) <= 0 ||
		target.Unit != game.UnitUnknown ||
		target.Owner != -1 && target.Owner != r.LocalFaction {
		return false
	}
	if source.Building == game.BuildingTownhall {
		return unit == game.UnitScout
	}
	return source.Building == game.BuildingBarracks &&
		unit >= game.UnitPeasant &&
		unit <= game.UnitScout
}
