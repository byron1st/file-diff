package util

import "testing"

func TestLineOffsets_SingleLine(t *testing.T) {
	lo := NewLineOffsets("hello")
	if lo.LineCount() != 1 {
		t.Fatalf("expected 1 line, got %d", lo.LineCount())
	}
	if lo.LineStart(0) != 0 {
		t.Fatalf("expected start 0, got %d", lo.LineStart(0))
	}
	if lo.LineEnd(0) != 5 {
		t.Fatalf("expected end 5, got %d", lo.LineEnd(0))
	}
}

func TestLineOffsets_MultipleLines(t *testing.T) {
	lo := NewLineOffsets("aaa\nbbb\nccc")
	if lo.LineCount() != 3 {
		t.Fatalf("expected 3 lines, got %d", lo.LineCount())
	}

	tests := []struct {
		line      int
		wantStart int
		wantEnd   int
	}{
		{0, 0, 3},
		{1, 4, 7},
		{2, 8, 11},
	}
	for _, tt := range tests {
		if got := lo.LineStart(tt.line); got != tt.wantStart {
			t.Errorf("line %d: start = %d, want %d", tt.line, got, tt.wantStart)
		}
		if got := lo.LineEnd(tt.line); got != tt.wantEnd {
			t.Errorf("line %d: end = %d, want %d", tt.line, got, tt.wantEnd)
		}
	}
}

func TestLineOffsets_LineNumber(t *testing.T) {
	lo := NewLineOffsets("ab\ncd\nef")
	tests := []struct {
		offset   int
		wantLine int
	}{
		{0, 0},
		{1, 0},
		{3, 1},
		{4, 1},
		{6, 2},
		{8, 2}, // textLength
	}
	for _, tt := range tests {
		if got := lo.LineNumber(tt.offset); got != tt.wantLine {
			t.Errorf("offset %d: line = %d, want %d", tt.offset, got, tt.wantLine)
		}
	}
}

func TestLineOffsets_LineEndWithNewline(t *testing.T) {
	lo := NewLineOffsets("ab\ncd\nef")
	// line 0 end without newline = 2, with newline = 3
	if got := lo.LineEndWithNewline(0, false); got != 2 {
		t.Errorf("line 0 no-newline: got %d, want 2", got)
	}
	if got := lo.LineEndWithNewline(0, true); got != 3 {
		t.Errorf("line 0 with-newline: got %d, want 3", got)
	}
	// last line: includeNewline should not add 1
	if got := lo.LineEndWithNewline(2, true); got != 8 {
		t.Errorf("last line with-newline: got %d, want 8", got)
	}
}

func TestLineOffsets_TextLength(t *testing.T) {
	lo := NewLineOffsets("ab\ncd\nef")
	if lo.TextLength() != 8 {
		t.Fatalf("expected text length 8, got %d", lo.TextLength())
	}
}

func TestLineOffsets_TextLength_Empty(t *testing.T) {
	lo := NewLineOffsets("")
	if lo.TextLength() != 0 {
		t.Fatalf("expected text length 0, got %d", lo.TextLength())
	}
}

func TestLineOffsets_CheckLineIndex_Panic(t *testing.T) {
	lo := NewLineOffsets("ab\ncd")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for out-of-range line index")
		}
	}()
	lo.LineStart(5)
}

func TestLineOffsets_CheckLineIndex_NegativePanic(t *testing.T) {
	lo := NewLineOffsets("ab\ncd")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative line index")
		}
	}()
	lo.LineStart(-1)
}

func TestLineOffsets_LineNumber_Panic(t *testing.T) {
	lo := NewLineOffsets("abc")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative offset")
		}
	}()
	lo.LineNumber(-1)
}
