package screens

import (
	"math"
	"testing"
	"time"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestBuildingButtonTooltipsMatchGameplay(t *testing.T) {
	tests := map[game.BuildingType]string{
		game.BuildingBarracks: "Recruits: Peasant, Archer, Knight, Scout\nPlace on: Plains",
		game.BuildingFarm:     "Produces: Food +2\nPlace on: Plains adjacent to Water",
		game.BuildingMine:     "Produces: Stone, Iron, Coal, or Gold\nPlace on: Rock, Iron, Coal, or Gold",
		game.BuildingForester: "Produces: Wood +2\nPlace on: Forest or Jungle",
	}

	for building, want := range tests {
		if got := buildingButtonTooltip(building); got != want {
			t.Errorf("buildingButtonTooltip(%s) = %q, want %q", building, got, want)
		}
	}
}

func TestRoundCountdownState(t *testing.T) {
	tests := []struct {
		name         string
		remaining    time.Duration
		wantDigit    int
		wantProgress float64
		wantVisible  bool
	}{
		{name: "before countdown", remaining: 3*time.Second + time.Nanosecond},
		{name: "three starts", remaining: 3 * time.Second, wantDigit: 3, wantVisible: true},
		{name: "three halfway", remaining: 2500 * time.Millisecond, wantDigit: 3, wantProgress: 0.5, wantVisible: true},
		{name: "two starts", remaining: 2 * time.Second, wantDigit: 2, wantVisible: true},
		{name: "two halfway", remaining: 1500 * time.Millisecond, wantDigit: 2, wantProgress: 0.5, wantVisible: true},
		{name: "one starts", remaining: time.Second, wantDigit: 1, wantVisible: true},
		{name: "one nearly ends", remaining: time.Nanosecond, wantDigit: 1, wantProgress: 1, wantVisible: true},
		{name: "deadline", remaining: 0},
		{name: "after deadline", remaining: -time.Nanosecond},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			digit, progress, visible := roundCountdownState(test.remaining)
			if digit != test.wantDigit {
				t.Fatalf("digit = %d, want %d", digit, test.wantDigit)
			}
			if visible != test.wantVisible {
				t.Fatalf("visible = %t, want %t", visible, test.wantVisible)
			}
			if math.Abs(progress-test.wantProgress) > 1e-8 {
				t.Fatalf("progress = %f, want %f", progress, test.wantProgress)
			}
		})
	}
}

func TestCountdownSecondsRemainingMatchesDisplayedDigit(t *testing.T) {
	tests := []struct {
		remaining time.Duration
		want      int
	}{
		{remaining: 3*time.Second + time.Nanosecond, want: 4},
		{remaining: 3 * time.Second, want: 3},
		{remaining: 2500 * time.Millisecond, want: 3},
		{remaining: 2 * time.Second, want: 2},
		{remaining: 1500 * time.Millisecond, want: 2},
		{remaining: time.Second, want: 1},
		{remaining: time.Nanosecond, want: 1},
		{remaining: 0, want: 0},
		{remaining: -time.Nanosecond, want: 0},
	}

	for _, test := range tests {
		if got := countdownSecondsRemaining(test.remaining); got != test.want {
			t.Errorf(
				"countdownSecondsRemaining(%s) = %d, want %d",
				test.remaining,
				got,
				test.want,
			)
		}
	}
}

func TestRoundAnnouncementState(t *testing.T) {
	tests := []struct {
		name        string
		round       int32
		remaining   time.Duration
		wantText    string
		wantVisible bool
	}{
		{
			name:        "new round",
			round:       2,
			remaining:   time.Second,
			wantText:    "Round #2",
			wantVisible: true,
		},
		{name: "announcement elapsed", round: 2, remaining: 0},
		{name: "no announced round", remaining: time.Second},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, visible := roundAnnouncementState(test.round, test.remaining)
			if text != test.wantText {
				t.Fatalf("text = %q, want %q", text, test.wantText)
			}
			if visible != test.wantVisible {
				t.Fatalf("visible = %t, want %t", visible, test.wantVisible)
			}
		})
	}
}

func TestEaseInExpoStartsSlowAndAccelerates(t *testing.T) {
	firstHalfGrowth := easeInExpo(0.5) - easeInExpo(0)
	secondHalfGrowth := easeInExpo(1) - easeInExpo(0.5)

	if firstHalfGrowth >= secondHalfGrowth {
		t.Fatalf(
			"expected accelerating growth, first half = %f, second half = %f",
			firstHalfGrowth,
			secondHalfGrowth,
		)
	}
}

func TestRoundCountdownPositionUsesGridCellCenters(t *testing.T) {
	renderSize := vec.Vec2i{X: 1200, Y: 800}
	tests := []struct {
		name   string
		anchor settings.CountdownAnchor
		want   vec.Vec2i
	}{
		{name: "top left", anchor: settings.CountdownAnchorAt(0, 0), want: vec.Vec2i{X: 120, Y: 80}},
		{name: "top right", anchor: settings.CountdownAnchorAt(4, 0), want: vec.Vec2i{X: 1080, Y: 80}},
		{name: "center", anchor: settings.CountdownAnchorAt(2, 2), want: vec.Vec2i{X: 600, Y: 400}},
		{name: "bottom left", anchor: settings.CountdownAnchorAt(0, 4), want: vec.Vec2i{X: 120, Y: 720}},
		{name: "bottom right", anchor: settings.CountdownAnchorAt(4, 4), want: vec.Vec2i{X: 1080, Y: 720}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := roundCountdownPosition(test.anchor, renderSize); got != test.want {
				t.Fatalf("roundCountdownPosition() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestRoundCountdownTextSizeAppliesScale(t *testing.T) {
	const renderHeight int32 = 1080

	if got := roundCountdownTextSize(1, renderHeight, 0.25); got != 80 {
		t.Fatalf("25%% size = %d, want 80", got)
	}
	if got := roundCountdownTextSize(1, renderHeight, 0.5); got != 160 {
		t.Fatalf("50%% size = %d, want 160", got)
	}
	if got := roundCountdownTextSize(1, renderHeight, 1); got != 320 {
		t.Fatalf("100%% size = %d, want 320", got)
	}
	if got := roundCountdownTextSize(1, renderHeight, 1.5); got != 480 {
		t.Fatalf("150%% size = %d, want 480", got)
	}
}

func TestCountdownSettingsPreviewPlaysOneCountdown(t *testing.T) {
	startedAt := time.Unix(100, 0)
	tests := []struct {
		name         string
		elapsed      time.Duration
		wantDigit    int
		wantProgress float64
		wantVisible  bool
	}{
		{name: "starts at three", wantDigit: 3, wantVisible: true},
		{name: "three animates", elapsed: 500 * time.Millisecond, wantDigit: 3, wantProgress: 0.5, wantVisible: true},
		{name: "changes to two", elapsed: time.Second, wantDigit: 2, wantVisible: true},
		{name: "changes to one", elapsed: 2 * time.Second, wantDigit: 1, wantVisible: true},
		{name: "ends", elapsed: 3 * time.Second},
		{name: "stays ended", elapsed: 4 * time.Second},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			digit, progress, visible := countdownSettingsPreviewState(
				startedAt,
				startedAt.Add(test.elapsed),
			)
			if digit != test.wantDigit {
				t.Fatalf("digit = %d, want %d", digit, test.wantDigit)
			}
			if math.Abs(progress-test.wantProgress) > 1e-8 {
				t.Fatalf("progress = %f, want %f", progress, test.wantProgress)
			}
			if visible != test.wantVisible {
				t.Fatalf("visible = %t, want %t", visible, test.wantVisible)
			}
		})
	}
}
