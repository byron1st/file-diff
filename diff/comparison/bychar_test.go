// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import (
	"testing"

	"github.com/byron1st/file-diff/diff/util"
)

func TestCompareChars_Identical(t *testing.T) {
	result, err := CompareChars("abc", "abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes()) != 0 {
		t.Fatalf("expected no changes, got %v", result.Changes())
	}
}

func TestCompareChars_SingleCharChange(t *testing.T) {
	result, err := CompareChars("abc", "axc")
	if err != nil {
		t.Fatal(err)
	}
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	assertRange(t, changes[0], 1, 2, 1, 2) // 'b' -> 'x'
}

func TestCompareChars_Insertion(t *testing.T) {
	result, err := CompareChars("ac", "abc")
	if err != nil {
		t.Fatal(err)
	}
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	assertRange(t, changes[0], 1, 1, 1, 2) // insert 'b'
}

func TestCompareChars_Deletion(t *testing.T) {
	result, err := CompareChars("abc", "ac")
	if err != nil {
		t.Fatal(err)
	}
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	assertRange(t, changes[0], 1, 2, 1, 1) // delete 'b'
}

func TestCompareChars_CompletelyDifferent(t *testing.T) {
	result, err := CompareChars("abc", "xyz")
	if err != nil {
		t.Fatal(err)
	}
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	assertRange(t, changes[0], 0, 3, 0, 3)
}

func TestCompareChars_Empty(t *testing.T) {
	result, err := CompareChars("", "abc")
	if err != nil {
		t.Fatal(err)
	}
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	assertRange(t, changes[0], 0, 0, 0, 3)
}

func TestCompareCharsTwoStep_SpaceDifference(t *testing.T) {
	result, err := CompareCharsTwoStep("a b", "a  b")
	if err != nil {
		t.Fatal(err)
	}
	// Non-space chars 'a' and 'b' should match; space difference is in the gap
	unchanged := result.Unchanged()
	if len(unchanged) < 2 {
		t.Fatalf("expected at least 2 unchanged regions, got %d: %v", len(unchanged), unchanged)
	}
}

func TestCompareCharsIgnoreWhitespaces_SameContent(t *testing.T) {
	result, err := CompareCharsIgnoreWhitespaces("a b c", "a  b  c")
	if err != nil {
		t.Fatal(err)
	}
	changes := result.Changes()
	if len(changes) != 0 {
		t.Fatalf("expected no changes when ignoring whitespace, got %v", changes)
	}
}

func TestCompareCharsIgnoreWhitespaces_RealDifference(t *testing.T) {
	result, err := CompareCharsIgnoreWhitespaces("a x c", "a y c")
	if err != nil {
		t.Fatal(err)
	}
	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
}

func TestComparePunctuation_MatchBrackets(t *testing.T) {
	result, err := ComparePunctuation("(a + b)", "[a + b]")
	if err != nil {
		t.Fatal(err)
	}
	// '+' should match; '(' vs '[' and ')' vs ']' are changes
	unchanged := result.Unchanged()
	hasPlus := false
	for _, u := range unchanged {
		if u.End1-u.Start1 == 1 {
			hasPlus = true
		}
	}
	if !hasPlus {
		t.Fatalf("expected '+' to be matched, unchanged: %v", unchanged)
	}
}

func TestGetAllCodePoints(t *testing.T) {
	cp := getAllCodePoints("hello")
	if len(cp.codePoints) != 5 {
		t.Fatalf("expected 5 code points, got %d", len(cp.codePoints))
	}
	if cp.codePoints[0] != 'h' || cp.codePoints[4] != 'o' {
		t.Fatalf("unexpected code points: %v", cp.codePoints)
	}
}

func TestGetNonSpaceCodePoints(t *testing.T) {
	cp := getNonSpaceCodePoints("a b c")
	if len(cp.codePoints) != 3 {
		t.Fatalf("expected 3 non-space code points, got %d", len(cp.codePoints))
	}
	if cp.offsets[0] != 0 || cp.offsets[1] != 2 || cp.offsets[2] != 4 {
		t.Fatalf("unexpected offsets: %v", cp.offsets)
	}
}

func TestGetPunctuationChars(t *testing.T) {
	cp := getPunctuationChars("a.b;c")
	if len(cp.codePoints) != 2 {
		t.Fatalf("expected 2 punctuation chars, got %d", len(cp.codePoints))
	}
	if cp.codePoints[0] != '.' || cp.codePoints[1] != ';' {
		t.Fatalf("unexpected punctuation: %v", cp.codePoints)
	}
}

func TestCompareCharsTrimWhitespaces_SameContent(t *testing.T) {
	result, err := CompareCharsTrimWhitespaces("  hello  ", "hello")
	if err != nil {
		t.Fatal(err)
	}
	changes := result.Changes()
	if len(changes) != 0 {
		t.Fatalf("expected no changes when trimming whitespace, got %v", changes)
	}
}

func TestCompareCharsTrimWhitespaces_RealDifference(t *testing.T) {
	result, err := CompareCharsTrimWhitespaces("  abc  ", "  axc  ")
	if err != nil {
		t.Fatal(err)
	}
	changes := result.Changes()
	if len(changes) == 0 {
		t.Fatal("expected changes for real difference")
	}
}

func TestCompareChars_BothEmpty(t *testing.T) {
	result, err := CompareChars("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes()) != 0 {
		t.Fatalf("expected no changes, got %v", result.Changes())
	}
}

func TestCompareCharsIgnoreWhitespaces_BothEmpty(t *testing.T) {
	result, err := CompareCharsIgnoreWhitespaces("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes()) != 0 {
		t.Fatalf("expected no changes, got %v", result.Changes())
	}
}

func TestCompareCharsIgnoreWhitespaces_OnlySpaces(t *testing.T) {
	result, err := CompareCharsIgnoreWhitespaces("   ", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes()) != 0 {
		t.Fatalf("expected no changes for only-space comparison, got %v", result.Changes())
	}
}

func TestCompareChars_MultiByteRuneChange(t *testing.T) {
	result, err := CompareChars("a한b", "a😀b")
	if err != nil {
		t.Fatal(err)
	}

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}

	assertRange(t, changes[0], 1, len("a한"), 1, len("a😀"))
}

func TestCompareCharsIgnoreWhitespaces_InsertionRange(t *testing.T) {
	result, err := CompareCharsIgnoreWhitespaces("ab", "a x b")
	if err != nil {
		t.Fatal(err)
	}

	changes := result.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}

	assertRange(t, changes[0], 1, 1, 2, 3)
}

func TestCompareCharsTrimWhitespaces_UTF8LeadingAndTrailingSpaces(t *testing.T) {
	result, err := CompareCharsTrimWhitespaces("  한글  ", "한글")
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Changes()) != 0 {
		t.Fatalf("expected no changes when trimming UTF-8 surrounding spaces, got %v", result.Changes())
	}
}

func assertRange(t *testing.T, r util.Range, start1, end1, start2, end2 int) {
	t.Helper()
	if r.Start1 != start1 || r.End1 != end1 || r.Start2 != start2 || r.End2 != end2 {
		t.Fatalf("expected [%d, %d) - [%d, %d), got %v", start1, end1, start2, end2, r)
	}
}
