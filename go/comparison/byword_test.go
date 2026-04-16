// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import (
	"testing"
)

func TestGetInlineChunks_SimpleWords(t *testing.T) {
	chunks := GetInlineChunks("hello world")
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	w1, ok := chunks[0].(*WordChunk)
	if !ok {
		t.Fatal("expected WordChunk")
	}
	if w1.content() != "hello" {
		t.Fatalf("expected 'hello', got %q", w1.content())
	}
	w2, ok := chunks[1].(*WordChunk)
	if !ok {
		t.Fatal("expected WordChunk")
	}
	if w2.content() != "world" {
		t.Fatalf("expected 'world', got %q", w2.content())
	}
}

func TestGetInlineChunks_WithNewline(t *testing.T) {
	chunks := GetInlineChunks("a\nb")
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks (word, newline, word), got %d", len(chunks))
	}
	if _, ok := chunks[1].(*NewlineChunk); !ok {
		t.Fatal("expected NewlineChunk at index 1")
	}
}

func TestGetInlineChunks_WithPunctuation(t *testing.T) {
	chunks := GetInlineChunks("a.b")
	// 'a' is a word, '.' is punctuation (not a chunk), 'b' is a word
	if len(chunks) != 2 {
		t.Fatalf("expected 2 word chunks, got %d", len(chunks))
	}
}

func TestGetInlineChunks_Underscores(t *testing.T) {
	chunks := GetInlineChunks("my_var")
	// '_' is not punctuation, so "my_var" is a single word
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk for 'my_var', got %d", len(chunks))
	}
	w, ok := chunks[0].(*WordChunk)
	if !ok {
		t.Fatal("expected WordChunk")
	}
	if w.content() != "my_var" {
		t.Fatalf("expected 'my_var', got %q", w.content())
	}
}

func TestGetInlineChunks_CJK(t *testing.T) {
	// Each CJK character should be a separate WordChunk
	chunks := GetInlineChunks("中文")
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks for CJK characters, got %d", len(chunks))
	}
}

func TestGetInlineChunks_Empty(t *testing.T) {
	chunks := GetInlineChunks("")
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks, got %d", len(chunks))
	}
}

func TestCompareWords_Identical(t *testing.T) {
	fragments, err := CompareWords("hello world", "hello world", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 0 {
		t.Fatalf("expected no changes, got %v", fragments)
	}
}

func TestCompareWords_SingleWordChange(t *testing.T) {
	fragments, err := CompareWords("hello world", "hello earth", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) == 0 {
		t.Fatal("expected at least one change")
	}
	// The change should cover 'world' -> 'earth'
	f := fragments[0]
	if f.StartOffset1 > 6 || f.EndOffset1 < 11 {
		t.Fatalf("expected change to cover 'world' region, got %v", f)
	}
}

func TestCompareWords_VariableNameSpelling(t *testing.T) {
	// Phase 3 test requirement: short variable name spelling change should be detected
	fragments, err := CompareWords("int coutner = 0;", "int counter = 0;", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) == 0 {
		t.Fatal("expected change for variable spelling difference")
	}
	// Should detect only 'coutner' -> 'counter' change
	found := false
	for _, f := range fragments {
		// The change should be within the word region
		if f.StartOffset1 >= 4 && f.EndOffset1 <= 11 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected change in variable name region, got %v", fragments)
	}
}

func TestCompareWords_BracketChange(t *testing.T) {
	// Phase 3 test requirement: bracket/punctuation differences at char level
	fragments, err := CompareWords("(a + b)", "[a + b]", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) == 0 {
		t.Fatal("expected changes for bracket differences")
	}
	// Should detect '(' vs '[' and ')' vs ']' changes, but 'a', '+', 'b' should match
	totalChanged := 0
	for _, f := range fragments {
		totalChanged += (f.EndOffset1 - f.StartOffset1) + (f.EndOffset2 - f.StartOffset2)
	}
	// The unchanged parts (a, +, b, spaces) should be larger than changes (brackets)
	if totalChanged > 8 {
		t.Fatalf("too many changes detected, expected mostly matching: %v", fragments)
	}
}

func TestCompareWords_WhitespaceChange(t *testing.T) {
	// Space difference in DEFAULT policy should be detected
	fragments, err := CompareWords("a  b", "a b", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) == 0 {
		t.Fatal("expected change for whitespace difference in DEFAULT policy")
	}
}

func TestCompareWords_IgnoreWhitespaces(t *testing.T) {
	fragments, err := CompareWords("a  b", "a b", PolicyIgnoreWhitespaces)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 0 {
		t.Fatalf("expected no changes when ignoring whitespace, got %v", fragments)
	}
}

func TestCompareWords_TrimWhitespaces(t *testing.T) {
	// Leading/trailing whitespace on a line should be ignored
	fragments, err := CompareWords("  hello  ", "hello", PolicyTrimWhitespaces)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 0 {
		t.Fatalf("expected no changes when trimming whitespace, got %v", fragments)
	}
}

func TestCompareWords_Insertion(t *testing.T) {
	fragments, err := CompareWords("a b", "a x b", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) == 0 {
		t.Fatal("expected changes for word insertion")
	}
}

func TestCompareWords_MultipleChanges(t *testing.T) {
	fragments, err := CompareWords("foo bar baz", "foo qux baz", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) == 0 {
		t.Fatal("expected changes for word replacement")
	}
	// 'foo' and 'baz' should match, 'bar' -> 'qux' is the change
	for _, f := range fragments {
		text1 := "foo bar baz"[f.StartOffset1:f.EndOffset1]
		if text1 == "foo" || text1 == "baz" {
			t.Fatalf("matched word should not appear in changes: %q", text1)
		}
	}
}

func TestCompareWords_EmptyInput(t *testing.T) {
	fragments, err := CompareWords("", "hello", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 1 {
		t.Fatalf("expected 1 change for empty vs non-empty, got %d", len(fragments))
	}
}

func TestCompareWords_BothEmpty(t *testing.T) {
	fragments, err := CompareWords("", "", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 0 {
		t.Fatalf("expected no changes for empty vs empty, got %v", fragments)
	}
}

func TestNewlineChunk_Offsets(t *testing.T) {
	nc := newNewlineChunk(5)
	if nc.Offset1() != 5 {
		t.Fatalf("expected Offset1=5, got %d", nc.Offset1())
	}
	if nc.Offset2() != 6 {
		t.Fatalf("expected Offset2=6, got %d", nc.Offset2())
	}
}

func TestNewlineChunk_ChunkKey(t *testing.T) {
	nc := newNewlineChunk(0)
	key := nc.chunkKey()
	if key.isWord {
		t.Fatal("expected isWord=false for NewlineChunk")
	}
	if key.content != "\n" {
		t.Fatalf("expected content='\\n', got %q", key.content)
	}
}

func TestCompareWords_MultiLine(t *testing.T) {
	// Multi-line input exercises newline chunk handling
	fragments, err := CompareWords("hello\nworld", "hello\nearth", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) == 0 {
		t.Fatal("expected changes for multi-line word difference")
	}
}

func TestCompareWords_Deletion(t *testing.T) {
	fragments, err := CompareWords("a b c", "a c", PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) == 0 {
		t.Fatal("expected changes for word deletion")
	}
}

func TestCompareWords_UTF8FragmentOffsets(t *testing.T) {
	left := "go 한글 test"
	right := "go 한 test"

	fragments, err := CompareWords(left, right, PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 1 {
		t.Fatalf("expected 1 change, got %d", len(fragments))
	}

	f := fragments[0]
	if got := left[f.StartOffset1:f.EndOffset1]; got != "한글" {
		t.Fatalf("expected left fragment to be '한글', got %q", got)
	}
	if got := right[f.StartOffset2:f.EndOffset2]; got != "한" {
		t.Fatalf("expected right fragment to be '한', got %q", got)
	}
	if f.StartOffset1 != len("go ") || f.EndOffset1 != len("go 한글") {
		t.Fatalf("unexpected UTF-8 offsets for left fragment: %v", f)
	}
	if f.StartOffset2 != len("go ") || f.EndOffset2 != len("go 한") {
		t.Fatalf("unexpected UTF-8 offsets for right fragment: %v", f)
	}
}

func TestCompareWords_PunctuationInsertionAroundRepeatedWords(t *testing.T) {
	left := "foo(bar) baz"
	right := "foo(bar)(bar) baz"

	fragments, err := CompareWords(left, right, PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 1 {
		t.Fatalf("expected 1 change, got %d", len(fragments))
	}

	f := fragments[0]
	if got := right[f.StartOffset2:f.EndOffset2]; got != "(bar)" {
		t.Fatalf("expected inserted punctuation-wrapped word, got %q", got)
	}
}
