package comparison

import (
	"strings"
	"testing"
)

func TestPatienceMatcher_ImplementsLineMatcher(t *testing.T) {
	var m LineMatcher = &PatienceMatcher{}
	result := m.Match([]string{"a", "b"}, []string{"a", "c"}, PolicyDefault)
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
}

func TestPatienceMatcher_Identical(t *testing.T) {
	lines := []string{"aaa", "bbb", "ccc"}
	m := &PatienceMatcher{}
	result := m.Match(lines, lines, PolicyDefault)
	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes for identical input")
	}
}

func TestPatienceMatcher_SingleInsert(t *testing.T) {
	left := strings.Split("aaa\nccc", "\n")
	right := strings.Split("aaa\nbbb\nccc", "\n")
	m := &PatienceMatcher{}
	result := m.Match(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	ch := changes[0]
	if ch.Start1 != 1 || ch.End1 != 1 || ch.Start2 != 1 || ch.End2 != 2 {
		t.Fatalf("unexpected change: %v", ch)
	}
}

func TestPatienceMatcher_SingleDelete(t *testing.T) {
	left := strings.Split("aaa\nbbb\nccc", "\n")
	right := strings.Split("aaa\nccc", "\n")
	m := &PatienceMatcher{}
	result := m.Match(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	ch := changes[0]
	if ch.Start1 != 1 || ch.End1 != 2 || ch.Start2 != 1 || ch.End2 != 1 {
		t.Fatalf("unexpected change: %v", ch)
	}
}

func TestPatienceMatcher_Modification(t *testing.T) {
	left := strings.Split("aaa\nbbb\nccc", "\n")
	right := strings.Split("aaa\nxxx\nccc", "\n")
	m := &PatienceMatcher{}
	result := m.Match(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	ch := changes[0]
	if ch.Start1 != 1 || ch.End1 != 2 || ch.Start2 != 1 || ch.End2 != 2 {
		t.Fatalf("unexpected change: %v", ch)
	}
}

func TestPatienceMatcher_TrimWhitespaces(t *testing.T) {
	left := []string{"  aaa  ", "  bbb  "}
	right := []string{"aaa", "bbb"}
	m := &PatienceMatcher{}
	result := m.Match(left, right, PolicyTrimWhitespaces)

	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes when trimming whitespace")
	}
}

func TestPatienceMatcher_IgnoreWhitespaces(t *testing.T) {
	left := []string{"a b c", "d e f"}
	right := []string{"abc", "def"}
	m := &PatienceMatcher{}
	result := m.Match(left, right, PolicyIgnoreWhitespaces)

	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes when ignoring whitespace")
	}
}

func TestPatienceMatcher_FunctionRefactoring(t *testing.T) {
	// Classic scenario where Patience shines: function reordering.
	// Myers often misaligns closing braces; Patience anchors on unique signatures.
	left := strings.Split(
		"func foo() {\n  return 1\n}\nfunc bar() {\n  return 2\n}", "\n")
	right := strings.Split(
		"func bar() {\n  return 2\n}\nfunc baz() {\n  return 3\n}", "\n")

	m := &PatienceMatcher{}
	result := m.Match(left, right, PolicyDefault)

	// "func bar() {" and "  return 2" should be matched as unchanged
	unchanged := result.Unchanged()
	found := false
	for _, u := range unchanged {
		for i := u.Start1; i < u.End1; i++ {
			if left[i] == "func bar() {" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected 'func bar() {' to be matched as unchanged")
	}
}

func TestPatienceMatcher_AllDifferent(t *testing.T) {
	left := []string{"a", "b", "c"}
	right := []string{"x", "y", "z"}
	m := &PatienceMatcher{}
	result := m.Match(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) == 0 {
		t.Fatal("expected at least one change")
	}
}

func TestPatienceMatcher_Empty(t *testing.T) {
	m := &PatienceMatcher{}
	result := m.Match(nil, nil, PolicyDefault)
	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes for empty input")
	}
}

func TestPatienceMatcher_OneEmpty(t *testing.T) {
	left := []string{"aaa", "bbb"}
	m := &PatienceMatcher{}
	result := m.Match(left, nil, PolicyDefault)

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	ch := changes[0]
	if ch.Start1 != 0 || ch.End1 != 2 || ch.Start2 != 0 || ch.End2 != 0 {
		t.Fatalf("unexpected change: %v", ch)
	}
}

func TestPatienceMatcher_BraceMisalignment(t *testing.T) {
	// Scenario from PLAN.md: Myers can misalign braces between functions.
	// Patience should anchor on the unique function signatures.
	left := strings.Split("func alpha() {\n  doA()\n}\n\nfunc beta() {\n  doB()\n}", "\n")
	right := strings.Split("func alpha() {\n  doA()\n  doA2()\n}\n\nfunc beta() {\n  doB()\n}", "\n")

	m := &PatienceMatcher{}
	result := m.Match(left, right, PolicyDefault)

	// Verify that "func alpha() {" is matched to "func alpha() {"
	// and "func beta() {" to "func beta() {"
	unchanged := result.Unchanged()
	alphaMatched, betaMatched := false, false
	for _, u := range unchanged {
		for i := range u.End1 - u.Start1 {
			l := left[u.Start1+i]
			r := right[u.Start2+i]
			if l == "func alpha() {" && r == "func alpha() {" {
				alphaMatched = true
			}
			if l == "func beta() {" && r == "func beta() {" {
				betaMatched = true
			}
		}
	}
	if !alphaMatched {
		t.Fatal("expected 'func alpha() {' to be correctly matched")
	}
	if !betaMatched {
		t.Fatal("expected 'func beta() {' to be correctly matched")
	}
}
