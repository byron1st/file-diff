package histogram

import (
	"testing"
)

func simpleFallback(left, right []string) []Anchor {
	// Naive LCS fallback for testing: match identical lines greedily
	var result []Anchor
	used := make([]bool, len(right))
	for i, l := range left {
		for j, r := range right {
			if !used[j] && l == r {
				result = append(result, Anchor{LeftIdx: i, RightIdx: j})
				used[j] = true
				break
			}
		}
	}
	return result
}

func TestDiff_Identical(t *testing.T) {
	lines := []string{"a", "b", "c"}
	result := Diff(lines, lines, simpleFallback)
	if len(result) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(result))
	}
	for i, a := range result {
		if a.LeftIdx != i || a.RightIdx != i {
			t.Fatalf("match %d: expected (%d,%d), got (%d,%d)", i, i, i, a.LeftIdx, a.RightIdx)
		}
	}
}

func TestDiff_Empty(t *testing.T) {
	result := Diff(nil, nil, simpleFallback)
	if len(result) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(result))
	}
}

func TestDiff_OneEmpty(t *testing.T) {
	result := Diff([]string{"a"}, nil, simpleFallback)
	if len(result) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(result))
	}
}

func TestDiff_NoCommonLines(t *testing.T) {
	result := Diff([]string{"a", "b"}, []string{"x", "y"}, simpleFallback)
	if len(result) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(result))
	}
}

func TestDiff_SingleInsert(t *testing.T) {
	left := []string{"a", "c"}
	right := []string{"a", "b", "c"}
	result := Diff(left, right, simpleFallback)
	if len(result) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(result))
	}
	if result[0].LeftIdx != 0 || result[0].RightIdx != 0 {
		t.Fatalf("first match: expected (0,0), got (%d,%d)", result[0].LeftIdx, result[0].RightIdx)
	}
	if result[1].LeftIdx != 1 || result[1].RightIdx != 2 {
		t.Fatalf("second match: expected (1,2), got (%d,%d)", result[1].LeftIdx, result[1].RightIdx)
	}
}

func TestDiff_RepetitiveLines(t *testing.T) {
	// Scenario where Patience fails: no unique lines
	left := []string{"{", "  a: 1", "}", "{", "  b: 2", "}"}
	right := []string{"{", "  a: 1", "}", "{", "  c: 3", "}", "{", "  b: 2", "}"}

	result := Diff(left, right, simpleFallback)

	// Should match "  a: 1" and "  b: 2" as anchors (unique in both sides)
	// and the surrounding braces should be handled appropriately
	matchedLines := make(map[string]bool)
	for _, a := range result {
		matchedLines[left[a.LeftIdx]] = true
	}
	if !matchedLines["  a: 1"] {
		t.Fatal("expected '  a: 1' to be matched")
	}
	if !matchedLines["  b: 2"] {
		t.Fatal("expected '  b: 2' to be matched")
	}
}

func TestDiff_JSONLikeStructure(t *testing.T) {
	// JSON-like structure with many repeated braces and colons
	left := []string{
		"{",
		`  "name": "alice"`,
		`  "age": 30`,
		"}",
	}
	right := []string{
		"{",
		`  "name": "alice"`,
		`  "email": "alice@example.com"`,
		`  "age": 30`,
		"}",
	}

	result := Diff(left, right, simpleFallback)

	// "name" and "age" lines are unique — should be matched
	nameMatched, ageMatched := false, false
	for _, a := range result {
		if left[a.LeftIdx] == `  "name": "alice"` {
			nameMatched = true
		}
		if left[a.LeftIdx] == `  "age": 30` {
			ageMatched = true
		}
	}
	if !nameMatched {
		t.Fatal("expected '\"name\": \"alice\"' to be matched")
	}
	if !ageMatched {
		t.Fatal("expected '\"age\": 30' to be matched")
	}
}

func TestDiff_AllRepeated(t *testing.T) {
	// Every line repeats — histogram should still find the best anchor
	left := []string{"x", "x", "y", "x"}
	right := []string{"x", "y", "x", "x"}

	result := Diff(left, right, simpleFallback)

	// "y" appears once in each — should be matched
	yMatched := false
	for _, a := range result {
		if left[a.LeftIdx] == "y" {
			yMatched = true
			if a.RightIdx != 1 {
				t.Fatalf("expected y at right index 1, got %d", a.RightIdx)
			}
		}
	}
	if !yMatched {
		t.Fatal("expected 'y' to be matched")
	}
}
