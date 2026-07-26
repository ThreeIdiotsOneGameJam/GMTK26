package screens

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

func TestResultPlacement(t *testing.T) {
	rankings := []packets.RankEntry{
		{FactionIdx: 2},
		{FactionIdx: 0},
		{FactionIdx: 1},
	}

	if got := resultPlacement(rankings, 0); got != 2 {
		t.Fatalf("resultPlacement() = %d, want 2", got)
	}
	if got := resultPlacement(rankings, 3); got != 0 {
		t.Fatalf("missing resultPlacement() = %d, want 0", got)
	}
}

func TestOrdinal(t *testing.T) {
	tests := map[int]string{
		1:   "1st",
		2:   "2nd",
		3:   "3rd",
		4:   "4th",
		11:  "11th",
		12:  "12th",
		13:  "13th",
		21:  "21st",
		102: "102nd",
	}
	for value, want := range tests {
		if got := ordinal(value); got != want {
			t.Errorf("ordinal(%d) = %q, want %q", value, got, want)
		}
	}
}

func TestResultEntryNameFallback(t *testing.T) {
	entry := packets.RankEntry{FactionIdx: 2}
	if got := resultEntryName(entry); got != "Faction 3" {
		t.Fatalf("resultEntryName() = %q, want %q", got, "Faction 3")
	}
}

func TestTrimResultName(t *testing.T) {
	if got := trimResultName("A very long kingdom name", 12); got != "A very lo..." {
		t.Fatalf("trimResultName() = %q, want %q", got, "A very lo...")
	}
	if got := trimResultName("short", 12); got != "short" {
		t.Fatalf("short trimResultName() = %q, want %q", got, "short")
	}
}
