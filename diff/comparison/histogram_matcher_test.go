package comparison

import (
	"strings"
	"testing"
)

func TestHistogramMatcher_ImplementsLineMatcher(t *testing.T) {
	var m LineMatcher = &HistogramMatcher{}
	result := m.Match([]string{"a", "b"}, []string{"a", "c"}, PolicyDefault)
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
}

func TestHistogramMatcher_Identical(t *testing.T) {
	lines := []string{"aaa", "bbb", "ccc"}
	m := &HistogramMatcher{}
	result := m.Match(lines, lines, PolicyDefault)
	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes for identical input")
	}
}

func TestHistogramMatcher_SingleInsert(t *testing.T) {
	left := strings.Split("aaa\nccc", "\n")
	right := strings.Split("aaa\nbbb\nccc", "\n")
	m := &HistogramMatcher{}
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

func TestHistogramMatcher_SingleDelete(t *testing.T) {
	left := strings.Split("aaa\nbbb\nccc", "\n")
	right := strings.Split("aaa\nccc", "\n")
	m := &HistogramMatcher{}
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

func TestHistogramMatcher_Modification(t *testing.T) {
	left := strings.Split("aaa\nbbb\nccc", "\n")
	right := strings.Split("aaa\nxxx\nccc", "\n")
	m := &HistogramMatcher{}
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

func TestHistogramMatcher_TrimWhitespaces(t *testing.T) {
	left := []string{"  aaa  ", "  bbb  "}
	right := []string{"aaa", "bbb"}
	m := &HistogramMatcher{}
	result := m.Match(left, right, PolicyTrimWhitespaces)

	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes when trimming whitespace")
	}
}

func TestHistogramMatcher_IgnoreWhitespaces(t *testing.T) {
	left := []string{"a b c", "d e f"}
	right := []string{"abc", "def"}
	m := &HistogramMatcher{}
	result := m.Match(left, right, PolicyIgnoreWhitespaces)

	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes when ignoring whitespace")
	}
}

func TestHistogramMatcher_Empty(t *testing.T) {
	m := &HistogramMatcher{}
	result := m.Match(nil, nil, PolicyDefault)
	if len(result.Changes()) != 0 {
		t.Fatal("expected no changes for empty input")
	}
}

func TestHistogramMatcher_OneEmpty(t *testing.T) {
	left := []string{"aaa", "bbb"}
	m := &HistogramMatcher{}
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

func TestHistogramMatcher_RepetitiveJSON(t *testing.T) {
	// Scenario where Patience struggles: JSON with repeated braces
	left := strings.Split(`{
  "users": [
    {
      "name": "alice",
      "age": 30
    },
    {
      "name": "bob",
      "age": 25
    }
  ]
}`, "\n")

	right := strings.Split(`{
  "users": [
    {
      "name": "alice",
      "age": 30
    },
    {
      "name": "charlie",
      "age": 35
    },
    {
      "name": "bob",
      "age": 25
    }
  ]
}`, "\n")

	m := &HistogramMatcher{}
	result := m.Match(left, right, PolicyDefault)

	// "alice" and "bob" blocks should be matched correctly
	unchanged := result.Unchanged()
	aliceMatched, bobMatched := false, false
	for _, u := range unchanged {
		for i := range u.End1 - u.Start1 {
			l := left[u.Start1+i]
			if strings.Contains(l, "alice") {
				aliceMatched = true
			}
			if strings.Contains(l, "bob") {
				bobMatched = true
			}
		}
	}
	if !aliceMatched {
		t.Fatal("expected alice block to be matched as unchanged")
	}
	if !bobMatched {
		t.Fatal("expected bob block to be matched as unchanged")
	}
}

func TestHistogramMatcher_TestCodeInsert(t *testing.T) {
	// From CONTEXT.md: test code with repeated patterns
	left := strings.Split(`func TestAdd(t *testing.T) {
    result := Add(1, 2)
    assert.Equal(t, 3, result)
}

func TestSubtract(t *testing.T) {
    result := Subtract(5, 3)
    assert.Equal(t, 2, result)
}`, "\n")

	right := strings.Split(`func TestAdd(t *testing.T) {
    result := Add(1, 2)
    assert.Equal(t, 3, result)
}

func TestMultiply(t *testing.T) {
    result := Multiply(4, 3)
    assert.Equal(t, 12, result)
}

func TestSubtract(t *testing.T) {
    result := Subtract(5, 3)
    assert.Equal(t, 2, result)
}`, "\n")

	m := &HistogramMatcher{}
	result := m.Match(left, right, PolicyDefault)

	// TestAdd and TestSubtract should be matched; TestMultiply is new
	unchanged := result.Unchanged()
	addMatched, subtractMatched := false, false
	for _, u := range unchanged {
		for i := range u.End1 - u.Start1 {
			l := left[u.Start1+i]
			if l == "func TestAdd(t *testing.T) {" {
				addMatched = true
			}
			if l == "func TestSubtract(t *testing.T) {" {
				subtractMatched = true
			}
		}
	}
	if !addMatched {
		t.Fatal("expected TestAdd signature to be matched")
	}
	if !subtractMatched {
		t.Fatal("expected TestSubtract signature to be matched")
	}
}

func TestHistogramMatcher_FunctionRefactoring(t *testing.T) {
	left := strings.Split(
		"func foo() {\n  return 1\n}\nfunc bar() {\n  return 2\n}", "\n")
	right := strings.Split(
		"func bar() {\n  return 2\n}\nfunc baz() {\n  return 3\n}", "\n")

	m := &HistogramMatcher{}
	result := m.Match(left, right, PolicyDefault)

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

func TestHistogramMatcher_AllDifferent(t *testing.T) {
	left := []string{"a", "b", "c"}
	right := []string{"x", "y", "z"}
	m := &HistogramMatcher{}
	result := m.Match(left, right, PolicyDefault)

	changes := result.Changes()
	if len(changes) == 0 {
		t.Fatal("expected at least one change")
	}
}

func TestHistogramMatcher_BraceMisalignment(t *testing.T) {
	left := strings.Split("func alpha() {\n  doA()\n}\n\nfunc beta() {\n  doB()\n}", "\n")
	right := strings.Split("func alpha() {\n  doA()\n  doA2()\n}\n\nfunc beta() {\n  doB()\n}", "\n")

	m := &HistogramMatcher{}
	result := m.Match(left, right, PolicyDefault)

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
