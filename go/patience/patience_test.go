package patience

import (
	"testing"
)

// trivial fallback that returns no matches
func noFallback(_, _ []string) []Anchor { return nil }

func TestDiff_Identical(t *testing.T) {
	lines := []string{"a", "b", "c"}
	matches := Diff(lines, lines, noFallback)
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(matches))
	}
	for i, m := range matches {
		if m.LeftIdx != i || m.RightIdx != i {
			t.Fatalf("match %d: got (%d,%d), want (%d,%d)", i, m.LeftIdx, m.RightIdx, i, i)
		}
	}
}

func TestDiff_SingleInsert(t *testing.T) {
	left := []string{"a", "c"}
	right := []string{"a", "b", "c"}
	matches := Diff(left, right, noFallback)
	// "a" and "c" are unique in both sides
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].LeftIdx != 0 || matches[0].RightIdx != 0 {
		t.Fatalf("first match: got (%d,%d), want (0,0)", matches[0].LeftIdx, matches[0].RightIdx)
	}
	if matches[1].LeftIdx != 1 || matches[1].RightIdx != 2 {
		t.Fatalf("second match: got (%d,%d), want (1,2)", matches[1].LeftIdx, matches[1].RightIdx)
	}
}

func TestDiff_SingleDelete(t *testing.T) {
	left := []string{"a", "b", "c"}
	right := []string{"a", "c"}
	matches := Diff(left, right, noFallback)
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].LeftIdx != 0 || matches[0].RightIdx != 0 {
		t.Fatalf("first match: got (%d,%d), want (0,0)", matches[0].LeftIdx, matches[0].RightIdx)
	}
	if matches[1].LeftIdx != 2 || matches[1].RightIdx != 1 {
		t.Fatalf("second match: got (%d,%d), want (2,1)", matches[1].LeftIdx, matches[1].RightIdx)
	}
}

func TestDiff_EmptyInputs(t *testing.T) {
	matches := Diff(nil, nil, noFallback)
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matches))
	}
	matches = Diff([]string{"a"}, nil, noFallback)
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matches))
	}
}

func TestDiff_DuplicateLines_UseFallback(t *testing.T) {
	// All lines are duplicated — no unique anchors, must use fallback
	left := []string{"{", "x", "{", "x", "}"}
	right := []string{"{", "y", "{", "y", "}"}
	called := false
	fallback := func(l, r []string) []Anchor {
		called = true
		// Simple fallback: match first and last lines
		return []Anchor{{LeftIdx: 0, RightIdx: 0}, {LeftIdx: len(l) - 1, RightIdx: len(r) - 1}}
	}
	matches := Diff(left, right, fallback)
	if !called {
		t.Fatal("expected fallback to be called for non-unique lines")
	}
	if len(matches) < 2 {
		t.Fatalf("expected at least 2 matches from fallback, got %d", len(matches))
	}
}

func TestDiff_FunctionMoveScenario(t *testing.T) {
	// Scenario: function reordering. Patience should anchor on unique signatures.
	left := []string{
		"func foo() {",
		"  return 1",
		"}",
		"func bar() {",
		"  return 2",
		"}",
	}
	right := []string{
		"func bar() {",
		"  return 2",
		"}",
		"func foo() {",
		"  return 1",
		"}",
	}
	matches := Diff(left, right, noFallback)
	// "func foo() {" and "func bar() {" are unique; "return 1", "return 2" are unique
	// "}" is duplicated so not an anchor
	// The LIS should pick the ordering that maximizes matches
	if len(matches) < 2 {
		t.Fatalf("expected at least 2 anchor matches, got %d", len(matches))
	}
}

func TestLIS_Basic(t *testing.T) {
	pairs := []Anchor{
		{LeftIdx: 0, RightIdx: 3},
		{LeftIdx: 1, RightIdx: 1},
		{LeftIdx: 2, RightIdx: 4},
		{LeftIdx: 3, RightIdx: 2},
	}
	result := lis(pairs)
	// LIS by RightIdx: 1, 2 or 1, 4 or 3, 4 — length 2
	if len(result) != 2 {
		t.Fatalf("expected LIS length 2, got %d", len(result))
	}
	// Verify increasing order of RightIdx
	for i := 1; i < len(result); i++ {
		if result[i].RightIdx <= result[i-1].RightIdx {
			t.Fatalf("LIS not increasing at %d: %d <= %d", i, result[i].RightIdx, result[i-1].RightIdx)
		}
	}
}

func TestLIS_AlreadySorted(t *testing.T) {
	pairs := []Anchor{
		{LeftIdx: 0, RightIdx: 0},
		{LeftIdx: 1, RightIdx: 1},
		{LeftIdx: 2, RightIdx: 2},
	}
	result := lis(pairs)
	if len(result) != 3 {
		t.Fatalf("expected LIS length 3, got %d", len(result))
	}
}

func TestLIS_Reversed(t *testing.T) {
	pairs := []Anchor{
		{LeftIdx: 0, RightIdx: 2},
		{LeftIdx: 1, RightIdx: 1},
		{LeftIdx: 2, RightIdx: 0},
	}
	result := lis(pairs)
	if len(result) != 1 {
		t.Fatalf("expected LIS length 1, got %d", len(result))
	}
}

func TestLIS_Empty(t *testing.T) {
	result := lis(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty LIS, got %d", len(result))
	}
}
