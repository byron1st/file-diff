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
