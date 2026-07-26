package ui

import (
	"fmt"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/render"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

func GameBuildingDetailsPanel() *GameBuildingDetailsPanelElement {
	el := &GameBuildingDetailsPanelElement{
		world: nil,
	}
	el.BaseElement = NewBaseElement(el)
	return el
}

func (el *GameBuildingDetailsPanelElement) WithWorld(w *GameWorldElement) *GameBuildingDetailsPanelElement {
	el.world = w
	return el
}

type buildingDetailsLayout struct {
	panelX, panelY, panelW, panelH int32
	headerText                     string
	contentLines                   []string
	buttons                        []buttonRect
}

type buttonRect struct {
	x, y, w, h int32
	unit       game.UnitType
	label      string
}

type GameBuildingDetailsPanelElement struct {
	BaseElement[*GameBuildingDetailsPanelElement]
	world *GameWorldElement
	lay   buildingDetailsLayout
}

var buildingResourceDisplayOrder = []game.ResourceType{
	game.ResourceFood,
	game.ResourceWood,
	game.ResourceStone,
	game.ResourceCoal,
	game.ResourceIron,
	game.ResourceSteel,
	game.ResourceGold,
}

func (el *GameBuildingDetailsPanelElement) canAffordUnit(unit game.UnitType) bool {
	if el.world == nil {
		return false
	}
	return game.CanAffordUnitAfterRoundIncome(
		&el.world.Map,
		int8(gameNet.LocalGameState.FactionIdx),
		unit,
		gameNet.LocalGameState.GetCoins(),
		gameNet.LocalGameState.Resources,
	)
}

func (el *GameBuildingDetailsPanelElement) update(deltaNano int64) {
	if el.world == nil {
		return
	}

	r := &el.world.Renderer
	m := &el.world.Map

	if r.SelectedHex == nil || r.SelectedKind != render.SelectionBuilding {
		el.lay = buildingDetailsLayout{}
		return
	}

	selectedCell := m.GetCell(*r.SelectedHex)
	if selectedCell == nil || selectedCell.Building == game.BuildingUnknown {
		el.lay = buildingDetailsLayout{}
		return
	}
	if int8(gameNet.LocalGameState.FactionIdx) != selectedCell.Owner {
		el.lay = buildingDetailsLayout{}
		return
	}

	el.lay = el.computeLayout(selectedCell)

	for _, btn := range el.lay.buttons {
		if (elementRect{X: btn.x, Y: btn.y, Width: btn.w, Height: btn.h}).
			contains(mousePosition()) {
			claimPointer(rl.MouseCursorPointingHand)
			canAfford := el.canAffordUnit(btn.unit)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && canAfford {
				r.RecruitToPlace = btn.unit
				r.BuildingToPlace = game.BuildingUnknown
				r.ClearQueuedBuilding()
			}
		}
	}
}

func (el *GameBuildingDetailsPanelElement) draw() {
	if el.world == nil {
		return
	}

	r := &el.world.Renderer
	m := &el.world.Map

	if tileHoverTooltipVisible(global.UIBlocksWorldInput, global.UIModalBlocksInput) {
		hoveredCell := m.GetCell(r.HoveredHex)
		hoverLines := tileHoverLines(hoveredCell, r.HoveredHex, global.DebugEnabled)
		if len(hoverLines) > 0 {
			el.drawHover(hoverLines)
		}
	}

	if el.lay.panelH == 0 {
		return
	}

	el.drawPanel()
}

func tileHoverTooltipVisible(uiBlocksWorldInput, uiModalBlocksInput bool) bool {
	return !uiBlocksWorldInput && !uiModalBlocksInput
}

func (el *GameBuildingDetailsPanelElement) drawHover(lines []string) {
	if len(lines) == 0 {
		return
	}

	textSize := int32(18)
	lineH := int32(22)
	textW := int32(0)
	for _, line := range lines {
		textW = max(textW, rl.MeasureText(line, textSize))
	}
	textH := textSize + int32(len(lines)-1)*lineH
	pad := int32(6)
	bgW := textW + pad*2
	bgH := textH + pad*2
	mx := int32(global.MousePosition.X)
	my := int32(global.MousePosition.Y)
	x := mx - bgW/2
	y := my - bgH - 10

	if x < 0 {
		x = 0
	}
	if maxX := int32(rl.GetRenderWidth()) - bgW; x > maxX {
		x = max(0, maxX)
	}
	if y < 0 {
		y = my + 10
	}

	rl.DrawRectangle(x, y, bgW, bgH, util.ColorOpacity(rl.Black, 0.6))
	for i, line := range lines {
		rl.DrawText(line, x+pad, y+pad+int32(i)*lineH, textSize, rl.White)
	}
}

func tileHoverLines(cell *game.Cell, hex game.Hex, showCoordinates bool) []string {
	if cell == nil {
		return nil
	}

	var lines []string
	if cell.Tile != game.TileUnknown {
		lines = append(lines, "Tile: "+cell.Tile.String())
	}
	if resource := tileResourceLabel(cell.Tile); resource != "" {
		lines = append(lines, "Resource: "+resource)
	}
	lines = append(lines, "Territory: "+factionLabel(cell.Owner))
	if showCoordinates {
		lines = append(lines, fmt.Sprintf("Coordinates: (%d, %d)", hex.X, hex.Y))
	}

	var contents []string
	if cell.Building != game.BuildingUnknown {
		building := "Building: " + buildingLabel(cell.Building)
		if output := buildingOutputText(cell.Building, cell.Tile); output != "" {
			building += " - " + output
		}
		contents = append(contents, building)
	}
	if cell.Unit != game.UnitUnknown {
		contents = append(contents, fmt.Sprintf("Unit: %s - %s", cell.Unit, factionLabel(cell.UnitOwner)))
	}
	if len(lines) > 0 && len(contents) > 0 {
		lines = append(lines, "")
	}
	lines = append(lines, contents...)
	return lines
}

func factionLabel(owner int8) string {
	if owner < 0 {
		return "Unclaimed"
	}
	return fmt.Sprintf("Faction %d", owner+1)
}

func tileResourceLabel(tile game.TileType) string {
	switch tile {
	case game.TileForest, game.TileJungle:
		return game.ResourceWood.String()
	case game.TileRock:
		return game.ResourceStone.String()
	case game.TileCoal:
		return game.ResourceCoal.String()
	case game.TileIron:
		return game.ResourceIron.String()
	case game.TileGold:
		return game.ResourceGold.String()
	default:
		return ""
	}
}

func buildingOutputText(building game.BuildingType, tile game.TileType) string {
	var outputs []string
	if amount := game.BuildingCoinsProduces(building); amount > 0 {
		outputs = append(outputs, fmt.Sprintf("Coin x %d", amount))
	}

	produces := game.BuildingProduces(building, tile)
	for _, resource := range buildingResourceDisplayOrder {
		if amount := produces[resource]; amount > 0 {
			outputs = append(outputs, fmt.Sprintf("%s x %d", resource, amount))
		}
	}
	return strings.Join(outputs, ", ")
}

func (el *GameBuildingDetailsPanelElement) drawPanel() {
	lay := &el.lay
	pad := int32(10)
	lineH := int32(24)

	rl.DrawRectangle(lay.panelX, lay.panelY, lay.panelW, lay.panelH, util.ColorOpacity(rl.DarkGray, 0.85))
	rl.DrawRectangleLines(lay.panelX, lay.panelY, lay.panelW, lay.panelH, rl.White)

	rl.DrawText(lay.headerText, lay.panelX+pad, lay.panelY+pad, 20, rl.White)
	y := lay.panelY + 20 + pad*2
	for _, line := range lay.contentLines {
		rl.DrawText(line, lay.panelX+pad, y, 18, rl.LightGray)
		y += lineH
	}

	for _, btn := range lay.buttons {
		canAfford := el.canAffordUnit(btn.unit)
		isHovered := (elementRect{X: btn.x, Y: btn.y, Width: btn.w, Height: btn.h}).
			contains(mousePosition())

		bgCol := rl.Gray
		if canAfford {
			bgCol = rl.DarkGreen
		}
		if isHovered {
			bgCol = *util.ColorAdd(bgCol, 25)
		}
		rl.DrawRectangle(btn.x, btn.y, btn.w, btn.h, bgCol)
		rl.DrawText(btn.label, btn.x+4, btn.y+4, int32(18), rl.White)
	}
}

func (el *GameBuildingDetailsPanelElement) computeLayout(cell *game.Cell) buildingDetailsLayout {
	textSize := int32(20)
	pad := int32(10)
	lineH := textSize + 4
	btnH := int32(26)
	btnGap := int32(4)

	label := buildingLabel(cell.Building)
	prod := buildingProductionText(cell.Building, cell.Tile)

	lines := []string{label}
	if prod != "" {
		lines = append(lines, "Produces: "+prod)
	}

	panelW := int32(280)
	headerH := textSize + pad*2
	contentH := int32(len(lines))*lineH + pad
	panelH := headerH + contentH

	winW := int32(rl.GetRenderWidth())
	winH := int32(rl.GetRenderHeight())
	bgX := winW - panelW - 20
	bgY := (winH - panelH) / 2

	var buttons []buttonRect

	if cell.Building == game.BuildingBarracks || cell.Building == game.BuildingTownhall {
		y := bgY + headerH + contentH + pad
		units := []struct {
			label string
			t     game.UnitType
		}{
			{"Peasant 10c 1f", game.UnitPeasant},
			{"Archer 20c 3f", game.UnitArcher},
			{"Knight 30c 5f", game.UnitKnight},
			{"Scout 10c", game.UnitScout},
		}
		if cell.Building == game.BuildingTownhall {
			units = units[3:]
		}
		panelH += pad*2 + int32(len(units))*(btnH+btnGap)
		bgY = (winH - panelH) / 2
		y = bgY + headerH + contentH + pad
		for _, tp := range units {
			buttons = append(buttons, buttonRect{
				x:     bgX + pad,
				y:     y,
				w:     panelW - pad*2,
				h:     btnH,
				unit:  tp.t,
				label: tp.label,
			})
			y += btnH + btnGap
		}
	}

	return buildingDetailsLayout{
		panelX:       bgX,
		panelY:       bgY,
		panelW:       panelW,
		panelH:       panelH,
		headerText:   label,
		contentLines: lines[1:],
		buttons:      buttons,
	}
}

func buildingLabel(b game.BuildingType) string {
	switch b {
	case game.BuildingTownhall:
		return "Townhall"
	case game.BuildingForester:
		return "Forester"
	case game.BuildingMine:
		return "Mine"
	case game.BuildingBarracks:
		return "Barracks"
	case game.BuildingFarm:
		return "Farm"
	default:
		return "Unknown"
	}
}

func buildingProductionText(b game.BuildingType, tile game.TileType) string {
	coinOutput := game.BuildingCoinsProduces(b)
	text := ""
	if coinOutput > 0 {
		text = fmt.Sprintf("Coin +%d", coinOutput)
	}

	produces := game.BuildingProduces(b, tile)
	for _, resType := range buildingResourceDisplayOrder {
		amount := produces[resType]
		if amount == 0 {
			continue
		}
		if text != "" {
			text += ", "
		}
		text += fmt.Sprintf("%s +%d", resType.String(), amount)
	}
	return text
}
