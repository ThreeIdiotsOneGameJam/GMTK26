package settings

import "testing"

func TestNormalizedMigratesCountdownDefaults(t *testing.T) {
	got := normalized(SettingsStore{})

	if got.CountdownScale != DefaultCountdownScale {
		t.Fatalf("CountdownScale = %v, want %v", got.CountdownScale, DefaultCountdownScale)
	}
	if got.CountdownAnchor != DefaultCountdownAnchor {
		t.Fatalf(
			"CountdownAnchor = %q, want %q",
			got.CountdownAnchor,
			DefaultCountdownAnchor,
		)
	}
}

func TestNormalizedClampsCountdownScale(t *testing.T) {
	tests := []struct {
		name  string
		scale float32
		want  float32
	}{
		{name: "below minimum", scale: 0.1, want: MinCountdownScale},
		{name: "within range", scale: 1.25, want: 1.25},
		{name: "above maximum", scale: 2, want: MaxCountdownScale},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := normalized(SettingsStore{
				CountdownScale:  test.scale,
				CountdownAnchor: CountdownCenter,
			})
			if got.CountdownScale != test.want {
				t.Fatalf("CountdownScale = %v, want %v", got.CountdownScale, test.want)
			}
		})
	}
}

func TestNormalizedRejectsUnknownCountdownAnchor(t *testing.T) {
	got := normalized(SettingsStore{
		CountdownScale:  DefaultCountdownScale,
		CountdownAnchor: "somewhere",
	})

	if got.CountdownAnchor != DefaultCountdownAnchor {
		t.Fatalf(
			"CountdownAnchor = %q, want %q",
			got.CountdownAnchor,
			DefaultCountdownAnchor,
		)
	}
}

func TestNormalizedMigratesLegacyCountdownAnchor(t *testing.T) {
	got := normalized(SettingsStore{
		CountdownScale:  DefaultCountdownScale,
		CountdownAnchor: "top_left",
	})

	if got.CountdownAnchor != CountdownAnchorAt(1, 1) {
		t.Fatalf(
			"CountdownAnchor = %q, want %q",
			got.CountdownAnchor,
			CountdownAnchorAt(1, 1),
		)
	}
}

func TestCountdownAnchorGridPositions(t *testing.T) {
	for row := int32(0); row < CountdownGridSize; row++ {
		for column := int32(0); column < CountdownGridSize; column++ {
			value := CountdownAnchorAt(column, row)
			gotColumn, gotRow, valid := CountdownAnchorGridPosition(value)
			if !valid || gotColumn != column || gotRow != row {
				t.Fatalf(
					"CountdownAnchorGridPosition(%q) = (%d, %d, %t), want (%d, %d, true)",
					value,
					gotColumn,
					gotRow,
					valid,
					column,
					row,
				)
			}
		}
	}
}
