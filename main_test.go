package main

import (
	"testing"
)

func TestNewUniverse(t *testing.T) {
	width := 10
	height := 5
	u := NewUniverse(width, height)

	if u.width != width {
		t.Errorf("Expected width %d, got %d", width, u.width)
	}
	if u.height != height {
		t.Errorf("Expected height %d, got %d", height, u.height)
	}
	if len(u.grid) != height {
		t.Errorf("Expected slice height %d, got %d", height, len(u.grid))
	}
	if len(u.grid[0]) != width {
		t.Errorf("Expected slice width %d, got %d", width, len(u.grid[0]))
	}
}

func TestUniverseClone(t *testing.T) {
	u := NewUniverse(5, 5)
	u.Set(2, 2, true)

	clone := u.Clone()
	if !clone.Get(2, 2) {
		t.Errorf("Expected clone to have true at (2,2)")
	}

	// Change clone, original should not change
	clone.Set(2, 2, false)
	if !u.Get(2, 2) {
		t.Errorf("Expected original to remain true at (2,2) after clone change")
	}
}

func TestToroidalWrap(t *testing.T) {
	// 3x3 universe
	u := NewUniverse(3, 3)
	// Place live cell at (0, 0)
	u.Set(0, 0, true)

	// Since it wraps around:
	// (0,0)'s neighbors are all wrap-around:
	// For example, cell at (2, 2) should see (0,0) as a neighbor.
	// Let's check neighbor counts from perspective of cells surrounding the wrapped edges:
	if count := u.Neighbors(2, 2); count != 1 {
		t.Errorf("Expected (2,2) to have 1 wrapped neighbor (0,0), got %d", count)
	}
	if count := u.Neighbors(1, 1); count != 1 {
		t.Errorf("Expected (1,1) to have 1 wrapped neighbor (0,0), got %d", count)
	}
	if count := u.Neighbors(0, 1); count != 1 {
		t.Errorf("Expected (0,1) to have 1 wrapped neighbor (0,0), got %d", count)
	}
	if count := u.Neighbors(1, 0); count != 1 {
		t.Errorf("Expected (1,0) to have 1 wrapped neighbor (0,0), got %d", count)
	}
}

func TestConwayRulesSimple(t *testing.T) {
	// Blinker is a period-2 oscillator:
	// Horiz: (1,2), (2,2), (3,2)
	// Vert after 1 generation: (2,1), (2,2), (2,3)
	u := NewUniverse(5, 5)
	u.Set(1, 2, true)
	u.Set(2, 2, true)
	u.Set(3, 2, true)

	nextU := u.Next()

	// (2,2) survives (2 neighbors)
	if !nextU.Get(2, 2) {
		t.Error("Blinker center (2,2) should survive")
	}
	// (2,1) and (2,3) are born (each had 3 neighbors: (1,2), (2,2), (3,2))
	if !nextU.Get(2, 1) {
		t.Error("Blinker top (2,1) should be born")
	}
	if !nextU.Get(2, 3) {
		t.Error("Blinker bottom (2,3) should be born")
	}
	// (1,2) and (3,2) should die (each had only 1 neighbor)
	if nextU.Get(1, 2) {
		t.Error("Blinker left (1,2) should die")
	}
	if nextU.Get(3, 2) {
		t.Error("Blinker right (3,2) should die")
	}

	// Another transition should bring it back
	nextNextU := nextU.Next()
	if !nextNextU.Get(1, 2) || !nextNextU.Get(2, 2) || !nextNextU.Get(3, 2) {
		t.Error("Blinker should oscillate back to original horizontal state")
	}
}

func TestInsertPatternValidation(t *testing.T) {
	u := NewUniverse(10, 10)

	// Pattern that fits
	patternFits := [][]bool{
		{true, true},
		{true, true},
	}
	err := u.InsertPattern(patternFits)
	if err != nil {
		t.Errorf("Expected pattern to fit, got error: %v", err)
	}

	// Pattern that does not fit (width/height exceeds universe)
	patternTooBig := [][]bool{
		make([]bool, 11),
	}
	err = u.InsertPattern(patternTooBig)
	if err == nil {
		t.Error("Expected error for too big pattern, got nil")
	}
}
