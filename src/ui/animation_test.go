package ui

import "testing"

func TestMoveTowardsDoesNotOvershoot(t *testing.T) {
	if got := MoveTowards(0, 1, 2); got != 1 {
		t.Fatalf("MoveTowards increasing = %v, want 1", got)
	}
	if got := MoveTowards(1, 0, 2); got != 0 {
		t.Fatalf("MoveTowards decreasing = %v, want 0", got)
	}
}

func TestSmoothstepClampsInput(t *testing.T) {
	if got := Smoothstep(-1); got != 0 {
		t.Fatalf("Smoothstep(-1) = %v, want 0", got)
	}
	if got := Smoothstep(2); got != 1 {
		t.Fatalf("Smoothstep(2) = %v, want 1", got)
	}
}
