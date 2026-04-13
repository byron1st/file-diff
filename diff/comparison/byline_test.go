package comparison

import (
	"strings"
	"testing"
)

func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func TestCompareLines_Identical(t *testing.T) {
	lines := []string{"aaa", "bbb", "ccc"}
	result := CompareLines(lines, lines, PolicyDefault)
	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes for identical input")
	}
}

func TestCompareLines_SingleInsert(t *testing.T) {
	left := splitLines("aaa\nccc")
	right := splitLines("aaa\nbbb\nccc")
	result := CompareLines(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	ch := changes[0]
	if ch.Start1 != 1 || ch.End1 != 1 || ch.Start2 != 1 || ch.End2 != 2 {
		t.Fatalf("unexpected change: %v", ch)
	}
}

func TestCompareLines_SingleDelete(t *testing.T) {
	left := splitLines("aaa\nbbb\nccc")
	right := splitLines("aaa\nccc")
	result := CompareLines(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	ch := changes[0]
	if ch.Start1 != 1 || ch.End1 != 2 || ch.Start2 != 1 || ch.End2 != 1 {
		t.Fatalf("unexpected change: %v", ch)
	}
}

func TestCompareLines_Modification(t *testing.T) {
	left := splitLines("aaa\nbbb\nccc")
	right := splitLines("aaa\nxxx\nccc")
	result := CompareLines(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	ch := changes[0]
	if ch.Start1 != 1 || ch.End1 != 2 || ch.Start2 != 1 || ch.End2 != 2 {
		t.Fatalf("unexpected change: %v", ch)
	}
}

func TestCompareLines_MultipleChanges(t *testing.T) {
	left := splitLines("a\nb\nc\nd\ne")
	right := splitLines("a\nB\nc\nD\ne")
	result := CompareLines(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
}

func TestCompareLines_AllDifferent(t *testing.T) {
	left := splitLines("a\nb\nc")
	right := splitLines("x\ny\nz")
	result := CompareLines(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) == 0 {
		t.Fatal("expected at least one change")
	}
}

func TestCompareLines_Empty(t *testing.T) {
	result := CompareLines(nil, nil, PolicyDefault)
	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes for empty input")
	}
}

func TestCompareLines_OneEmpty(t *testing.T) {
	left := splitLines("aaa\nbbb")
	result := CompareLines(left, nil, PolicyDefault)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	ch := changes[0]
	if ch.Start1 != 0 || ch.End1 != 2 || ch.Start2 != 0 || ch.End2 != 0 {
		t.Fatalf("unexpected change: %v", ch)
	}
}

func TestCompareLines_TrimWhitespaces(t *testing.T) {
	left := splitLines("  aaa  \n  bbb  ")
	right := splitLines("aaa\nbbb")
	result := CompareLines(left, right, PolicyTrimWhitespaces)

	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes when trimming whitespace")
	}
}

func TestCompareLines_IgnoreWhitespaces(t *testing.T) {
	left := splitLines("a b c\nd e f")
	right := splitLines("abc\ndef")
	result := CompareLines(left, right, PolicyIgnoreWhitespaces)

	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes when ignoring whitespace")
	}
}

func TestCompareLines_TrimWhitespaces_WithRealChange(t *testing.T) {
	left := splitLines("  aaa  \n  bbb  \n  ccc  ")
	right := splitLines("aaa\nxxx\nccc")
	result := CompareLines(left, right, PolicyTrimWhitespaces)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
}

func TestMyersMatcher_ImplementsLineMatcher(t *testing.T) {
	var m LineMatcher = &MyersMatcher{}
	result := m.Match([]string{"a", "b"}, []string{"a", "c"}, PolicyDefault)
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
}

func TestCompareLines_SmallLinesSecondStep(t *testing.T) {
	// Lines shorter than unimportant threshold (<=3 non-space chars) trigger the
	// compareSmart two-step path where "big" lines are diffed first, then gaps filled.
	left := splitLines("a\n{\nb\n}\nc")
	right := splitLines("a\n{\nx\n}\nc")
	result := CompareLines(left, right, PolicyDefault)
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
}

func TestCompareLines_IgnoreWhitespaces_WithRealChange(t *testing.T) {
	left := splitLines("  aaa  \n  bbb  \n  ccc  ")
	right := splitLines("aaa\nxxx\nccc")
	result := CompareLines(left, right, PolicyIgnoreWhitespaces)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
}

func TestCompareLines_FlushSecondStep(t *testing.T) {
	// Construct a case where IW-matched lines differ in exact content,
	// triggering flushSecondStep with a sample.
	left := splitLines("  a\nb\n  a\nc")
	right := splitLines("a\nb\na\nc")
	result := CompareLines(left, right, PolicyTrimWhitespaces)
	// trimmed, all lines should match
	if len(result.Changes()) != 0 {
		t.Fatalf("expected no changes, got %v", result.Changes())
	}
}

func TestCompareLines_LargeFile(t *testing.T) {
	// Generate a large-ish file with a single change in the middle
	n := 200
	left := make([]string, n)
	right := make([]string, n)
	for i := range n {
		line := strings.Repeat("x", i+1)
		left[i] = line
		right[i] = line
	}
	right[100] = "CHANGED_LINE"

	result := CompareLines(left, right, PolicyDefault)
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Start1 != 100 || changes[0].End1 != 101 {
		t.Fatalf("unexpected change position: %v", changes[0])
	}
}
