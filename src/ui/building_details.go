package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

func BuildingDetailsPanel() *BuildingDetailsPanelElement {
	el := &BuildingDetailsPanelElement{
		world: nil,
	}
	el.BaseElement = NewBaseElement(el)
	return el
}

func (el *BuildingDetailsPanelElement) WithWorld(w *WorldElement) *BuildingDetailsPanelElement {
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
	troop      game.TroopType
	label      string
}

type BuildingDetailsPanelElement struct {
	BaseElement[*BuildingDetailsPanelElement]
	world *WorldElement
	lay   buildingDetailsLayout
}

func (el *BuildingDetailsPanelElement) update(deltaNano int64) {
	if el.world == nil {
		return
	}

	r := &el.world.Renderer
	m := &el.world.Map

	if r.SelectedHex == nil {
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
		mx := int32(global.MousePosition.X)
		my := int32(global.MousePosition.Y)
		if mx >= btn.x && mx <= btn.x+btn.w && my >= btn.y && my <= btn.y+btn.h {
			global.UIBlocksWorldInput = true
			canAfford := gameNet.LocalGameState.GetCoins() >= game.TroopCost(btn.troop)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && canAfford {
				if err := gameNet.SendDispatchAction(gameNet.LocalGameState.GetRound(), *r.SelectedHex, *r.SelectedHex, btn.troop); err != nil {
					fmt.Printf("failed to send dispatch action: %v\n", err)
				} else {
					// The server keeps one action per faction, so dispatching
					// replaces any building that was queued this round.
					r.ClearQueuedBuilding()
				}
			}
		}
	}
}

func (el *BuildingDetailsPanelElement) draw() {
	if el.world == nil {
		return
	}

	r := &el.world.Renderer
	m := &el.world.Map

	hoveredCell := m.GetCell(r.HoveredHex)

	showHover := hoveredCell != nil && hoveredCell.Building != game.BuildingUnknown

	if showHover {
		el.drawHover()
	}

	if el.lay.panelH == 0 {
		return
	}

	el.drawPanel()
}

func (el *BuildingDetailsPanelElement) drawHover() {
	text := el.hoverText()
	if text == "" {
		return
	}

	textSize := int32(18)
	textW := rl.MeasureText(text, textSize)
	textH := textSize
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
	if y < 0 {
		y = my + 10
	}

	rl.DrawRectangle(x, y, bgW, bgH, util.ColorOpacity(rl.Black, 0.6))
	rl.DrawText(text, x+pad, y+pad, textSize, rl.White)
}

func (el *BuildingDetailsPanelElement) hoverText() string {
	cell := el.world.Map.GetCell(el.world.Renderer.HoveredHex)
	if cell == nil || cell.Building == game.BuildingUnknown {
		return ""
	}
	label := buildingLabel(cell.Building)
	production := buildingProductionText(cell.Building, cell.Tile)
	if production != "" {
		return label + ": " + production
	}
	return label
}

func (el *BuildingDetailsPanelElement) drawPanel() {
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
		cost := game.TroopCost(btn.troop)
		canAfford := gameNet.LocalGameState.GetCoins() >= cost
		mx := int32(global.MousePosition.X)
		my := int32(global.MousePosition.Y)
		isHovered := mx >= btn.x && mx <= btn.x+btn.w && my >= btn.y && my <= btn.y+btn.h

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

func (el *BuildingDetailsPanelElement) computeLayout(cell *game.Cell) buildingDetailsLayout {
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

	if cell.Building == game.BuildingBarracks {
		panelH += pad*2 + 3*(btnH+btnGap)
	}

	winW := int32(rl.GetRenderWidth())
	winH := int32(rl.GetRenderHeight())
	bgX := winW - panelW - 20
	bgY := (winH - panelH) / 2

	var buttons []buttonRect

	if cell.Building == game.BuildingBarracks {
		y := bgY + headerH + contentH + pad
		troops := []struct {
			label string
			t     game.TroopType
		}{
			{"Peasant 10c", game.TroopPeasant},
			{"Archer 20c", game.TroopArcher},
			{"Knight 30c", game.TroopKnight},
		}
		for _, tp := range troops {
			buttons = append(buttons, buttonRect{
				x:     bgX + pad,
				y:     y,
				w:     panelW - pad*2,
				h:     btnH,
				troop: tp.t,
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
	resTypes := []game.ResourceType{game.ResourceWood, game.ResourceStone, game.ResourceCoal, game.ResourceIron, game.ResourceSteel, game.ResourceGold}
	for _, resType := range resTypes {
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
