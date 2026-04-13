package fragment

import "testing"

func TestNewDiffFragment(t *testing.T) {
	f := NewDiffFragment(0, 5, 0, 3)
	if f.StartOffset1 != 0 || f.EndOffset1 != 5 || f.StartOffset2 != 0 || f.EndOffset2 != 3 {
		t.Fatalf("unexpected fragment: %v", f)
	}
}

func TestNewDiffFragment_EmptyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty fragment")
		}
	}()
	NewDiffFragment(0, 0, 0, 0)
}

func TestNewDiffFragment_InvalidPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid fragment")
		}
	}()
	NewDiffFragment(5, 3, 0, 1)
}

func TestNewLineFragment(t *testing.T) {
	f := NewLineFragment(0, 1, 0, 1, 0, 10, 0, 8, nil)
	if f.StartLine1 != 0 || f.EndLine1 != 1 || f.InnerFragments != nil {
		t.Fatalf("unexpected line fragment: %v", f)
	}
}

func TestNewLineFragment_DropsWholeInner(t *testing.T) {
	inner := []DiffFragment{NewDiffFragment(0, 10, 0, 8)}
	f := NewLineFragment(0, 1, 0, 1, 0, 10, 0, 8, inner)
	if f.InnerFragments != nil {
		t.Fatal("expected inner fragments to be dropped when spanning the whole range")
	}
}

func TestNewLineFragment_KeepsPartialInner(t *testing.T) {
	inner := []DiffFragment{NewDiffFragment(2, 5, 2, 4)}
	f := NewLineFragment(0, 1, 0, 1, 0, 10, 0, 8, inner)
	if f.InnerFragments == nil || len(f.InnerFragments) != 1 {
		t.Fatal("expected inner fragments to be kept for partial range")
	}
}

func TestNewLineFragment_EmptyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty line fragment")
		}
	}()
	NewLineFragment(1, 1, 1, 1, 0, 0, 0, 0, nil)
}

func TestNewLineFragment_InvalidPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid line fragment")
		}
	}()
	NewLineFragment(2, 1, 0, 1, 0, 10, 0, 8, nil)
}

func TestDiffFragment_String(t *testing.T) {
	f := NewDiffFragment(0, 5, 0, 3)
	s := f.String()
	if s != "[0, 5) - [0, 3)" {
		t.Fatalf("unexpected string: %q", s)
	}
}

func TestLineFragment_String(t *testing.T) {
	f := NewLineFragment(0, 1, 0, 1, 0, 10, 0, 8, nil)
	s := f.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
	// nil inner fragments should produce -1
	expected := "Lines [0, 1) - [0, 1); Offsets [0, 10) - [0, 8); Inner -1"
	if s != expected {
		t.Fatalf("unexpected string: %q, want %q", s, expected)
	}
}

func TestLineFragment_StringWithInner(t *testing.T) {
	inner := []DiffFragment{NewDiffFragment(2, 5, 2, 4)}
	f := NewLineFragment(0, 1, 0, 1, 0, 10, 0, 8, inner)
	s := f.String()
	expected := "Lines [0, 1) - [0, 1); Offsets [0, 10) - [0, 8); Inner 1"
	if s != expected {
		t.Fatalf("unexpected string: %q, want %q", s, expected)
	}
}

func TestNewLineFragment_InsertionOnly(t *testing.T) {
	f := NewLineFragment(0, 0, 0, 2, 0, 0, 0, 16, nil)
	if f.StartLine2 != 0 || f.EndLine2 != 2 {
		t.Fatalf("unexpected line fragment: %v", f)
	}
}

func TestNewLineFragment_DeletionOnly(t *testing.T) {
	f := NewLineFragment(0, 2, 0, 0, 0, 16, 0, 0, nil)
	if f.StartLine1 != 0 || f.EndLine1 != 2 {
		t.Fatalf("unexpected line fragment: %v", f)
	}
}

func TestNewLineFragment_MultipleInnerFragments(t *testing.T) {
	inner := []DiffFragment{
		NewDiffFragment(0, 3, 0, 3),
		NewDiffFragment(5, 8, 5, 7),
	}
	f := NewLineFragment(0, 1, 0, 1, 0, 10, 0, 8, inner)
	if len(f.InnerFragments) != 2 {
		t.Fatalf("expected 2 inner fragments, got %d", len(f.InnerFragments))
	}
}
