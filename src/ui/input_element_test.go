package ui

import (
	"strings"
	"testing"
)

func TestInputConfigurationRevalidatesText(t *testing.T) {
	input := Input().
		WithText("a1b2c3").
		WithInputTransformer(strings.ToUpper).
		WithCharset("ABC123").
		WithMaxTextLength(4)

	if got, want := input.Value(), "A1B2"; got != want {
		t.Fatalf("input text = %q, want %q", got, want)
	}
}

func TestInputRevalidatesExistingTextWhenCharsetChanges(t *testing.T) {
	input := Input().WithText("abc123").WithCharset("123")

	if got, want := input.Value(), "123"; got != want {
		t.Fatalf("input text = %q, want %q", got, want)
	}
}
