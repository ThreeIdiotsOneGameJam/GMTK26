package screens

import (
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/uiutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var (
	resultGold       = color.RGBA{R: 0xFF, G: 0xD1, B: 0x66, A: 255}
	resultGoldMuted  = color.RGBA{R: 0x7D, G: 0x63, B: 0x2D, A: 255}
	resultRow        = color.RGBA{R: 0x18, G: 0x19, B: 0x22, A: 230}
	resultRowLocal   = color.RGBA{R: 0x35, G: 0x36, B: 0x62, A: 245}
	resultCard       = color.RGBA{R: 0x13, G: 0x14, B: 0x1D, A: 238}
	resultCardBorder = color.RGBA{R: 0x4D, G: 0x4F, B: 0x68, A: 255}
)

const (
	resultCardWidth = 620
	resultRowWidth  = 560
	resultRowHeight = 42
	resultRowStride = 48
)

func NewGameEndScreen(
	result packets.S2CGameEndPacket,
	localFaction int,
	previousScreen *ui.ScreenElement,
) *ui.ScreenElement {
	rankings := append([]packets.RankEntry(nil), result.Rankings...)
	won := result.WinnerFaction >= 0 && result.WinnerFaction == localFaction
	placement := resultPlacement(rankings, localFaction)
	winnerName := result.WinnerName
	if strings.TrimSpace(winnerName) == "" {
		winnerName = resultFactionName(rankings, result.WinnerFaction)
	}

	title := "DEFEAT"
	titleColor := uiutil.MenuHeaderColor
	subtitle := fmt.Sprintf("%s claims victory", trimResultName(winnerName, 26))
	if won {
		title = "VICTORY"
		titleColor = resultGold
		subtitle = "Your kingdom outlasted every rival"
	} else if result.WinnerFaction < 0 {
		title = "DRAW"
		subtitle = "The leading factions finished level on score"
	}

	summary := resultSummary(placement, len(rankings), localFaction, rankings)
	cardHeight := int32(82 + len(rankings)*resultRowStride)
	cardHeight = max(cardHeight, 178)

	returnToPlay := func() {
		GoToPreviousScreen(previousScreen)
	}

	card := ui.Panel().
		WithBackgroundColor(resultCard).
		WithOutlineColor(resultCardBorder).
		WithOutlineWidth(2).
		WithRoundness(0.045).
		WithSize(vec.Vec2i{X: resultCardWidth, Y: cardHeight}).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePosDynamic(func(el *ui.PanelElement) vec.Vec2i {
			height := el.Parent.Size().Y
			top := max(int32(193), height/3)
			centeredTop := (height - el.Size().Y) / 2
			return vec.Vec2i{Y: top - centeredTop}
		}).
		AddChild(
			ui.Text().
				WithText("FINAL STANDINGS").
				WithTextSize(18).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.TopLeft, anchor.TopLeft).
				WithRelativePos(vec.Vec2i{X: 30, Y: 22}),
		).
		AddChild(
			ui.Text().
				WithText("SCORE").
				WithTextSize(18).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.TopRight, anchor.TopRight).
				WithRelativePos(vec.Vec2i{X: -30, Y: 22}),
		)

	for i, entry := range rankings {
		card.AddChild(resultRankingRow(
			i,
			resultRankAt(rankings, i),
			entry,
			entry.FactionIdx == localFaction,
		))
	}

	screen := ui.Screen().
		WithBackgroundColor(uiutil.MenuScreenBackground).
		WithBack(returnToPlay).
		AddChild(uiutil.MenuBackdrop()).
		AddChild(
			ui.Text().
				WithText(title).
				WithTextSize(76).
				WithTextColor(titleColor).
				WithTextShadow(color.RGBA{R: 0, G: 0, B: 0, A: 170}, vec.Vec2i{X: 4, Y: 5}).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{Y: max(int32(28), el.Parent.Size().Y/14)}
				}),
		).
		AddChild(
			ui.Text().
				WithText(subtitle).
				WithTextSize(25).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{Y: max(int32(112), el.Parent.Size().Y/14+82)}
				}),
		).
		AddChild(card).
		AddChild(
			ui.Text().
				WithText(summary).
				WithTextSize(20).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{
						Y: max(int32(158), el.Parent.Size().Y/14+118),
					}
				}),
		).
		AddChild(
			ui.Button().
				WithText("Back to Play").
				WithTextSize(30).
				WithSize(vec.Vec2i{X: 260, Y: 58}).
				WithAnchors(anchor.Bottom, anchor.Bottom).
				WithRelativePos(vec.Vec2i{Y: -30}).
				WithClick(returnToPlay),
		).
		AddChild(ui.Vignette().WithAlpha(145))

	return screen
}

func resultRankingRow(rowIndex, rank int, entry packets.RankEntry, local bool) *ui.PanelElement {
	background := resultRow
	outline := color.RGBA{}
	if local {
		background = resultRowLocal
		outline = resultGoldMuted
	}

	row := ui.Panel().
		WithBackgroundColor(background).
		WithOutlineColor(outline).
		WithOutlineWidth(1).
		WithRoundness(0.12).
		WithSize(vec.Vec2i{X: resultRowWidth, Y: resultRowHeight}).
		WithAnchors(anchor.Top, anchor.Top).
		WithRelativePos(vec.Vec2i{Y: int32(58 + rowIndex*resultRowStride)}).
		AddChild(
			ui.Text().
				WithText(fmt.Sprintf("%d", rank)).
				WithTextSize(22).
				WithTextColor(resultRankColor(rank-1)).
				WithAnchors(anchor.Left, anchor.Left).
				WithRelativePos(vec.Vec2i{X: 18}),
		).
		AddChild(
			ui.Text().
				WithText(trimResultName(resultEntryName(entry), 24)).
				WithTextSize(22).
				WithTextColor(uiutil.MenuHeaderColor).
				WithAnchors(anchor.Left, anchor.Left).
				WithRelativePos(vec.Vec2i{X: 60}),
		).
		AddChild(
			ui.Text().
				WithText(fmt.Sprintf("%d", entry.Points)).
				WithTextSize(22).
				WithTextColor(uiutil.MenuHeaderColor).
				WithAnchors(anchor.Right, anchor.Right).
				WithRelativePos(vec.Vec2i{X: -18}),
		)

	if local {
		row.AddChild(
			ui.Text().
				WithText("YOU").
				WithTextSize(15).
				WithTextColor(resultGold).
				WithAnchors(anchor.Right, anchor.Right).
				WithRelativePos(vec.Vec2i{X: -90}),
		)
	}
	return row
}

func resultRankColor(rank int) color.RGBA {
	switch rank {
	case 0:
		return resultGold
	case 1:
		return color.RGBA{R: 0xD6, G: 0xDC, B: 0xE8, A: 255}
	case 2:
		return color.RGBA{R: 0xD0, G: 0x8C, B: 0x60, A: 255}
	default:
		return uiutil.MenuMutedColor
	}
}

func resultPlacement(rankings []packets.RankEntry, localFaction int) int {
	for i, entry := range rankings {
		if entry.FactionIdx == localFaction {
			return resultRankAt(rankings, i)
		}
	}
	return 0
}

func resultRankAt(rankings []packets.RankEntry, index int) int {
	if index < 0 || index >= len(rankings) {
		return 0
	}
	rank := 1
	for i := 1; i <= index; i++ {
		previous := rankings[i-1]
		current := rankings[i]
		if previous.Alive != current.Alive || previous.Points != current.Points {
			rank = i + 1
		}
	}
	return rank
}

func resultSummary(placement, total, localFaction int, rankings []packets.RankEntry) string {
	if placement == 0 {
		return "Final results"
	}
	points := int32(0)
	for _, entry := range rankings {
		if entry.FactionIdx == localFaction {
			points = entry.Points
			break
		}
	}
	return fmt.Sprintf("%s place of %d  |  %d points", ordinal(placement), total, points)
}

func ordinal(value int) string {
	if value%100 >= 11 && value%100 <= 13 {
		return fmt.Sprintf("%dth", value)
	}
	switch value % 10 {
	case 1:
		return fmt.Sprintf("%dst", value)
	case 2:
		return fmt.Sprintf("%dnd", value)
	case 3:
		return fmt.Sprintf("%drd", value)
	default:
		return fmt.Sprintf("%dth", value)
	}
}

func resultFactionName(rankings []packets.RankEntry, faction int) string {
	for _, entry := range rankings {
		if entry.FactionIdx == faction {
			return resultEntryName(entry)
		}
	}
	return "A rival kingdom"
}

func resultEntryName(entry packets.RankEntry) string {
	name := strings.TrimSpace(entry.PlayerName)
	if name != "" {
		return name
	}
	return fmt.Sprintf("Faction %d", entry.FactionIdx+1)
}

func trimResultName(name string, maxRunes int) string {
	if utf8.RuneCountInString(name) <= maxRunes {
		return name
	}
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(name)
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
