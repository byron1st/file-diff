package util

import "testing"

func TestNewRange(t *testing.T) {
	r := NewRange(1, 3, 2, 5)
	if r.Start1 != 1 || r.End1 != 3 || r.Start2 != 2 || r.End2 != 5 {
		t.Fatalf("unexpected range values: %v", r)
	}
}

func TestNewRange_EqualStartEnd(t *testing.T) {
	r := NewRange(0, 0, 0, 0)
	if !r.IsEmpty() {
		t.Fatal("expected empty range")
	}
}

func TestNewRange_InvalidPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid range")
		}
	}()
	NewRange(5, 3, 0, 1)
}

func TestRange_IsEmpty(t *testing.T) {
	if !(NewRange(2, 2, 3, 3).IsEmpty()) {
		t.Fatal("expected empty")
	}
	if NewRange(0, 1, 0, 0).IsEmpty() {
		t.Fatal("expected non-empty")
	}
}

func TestRange_String(t *testing.T) {
	r := NewRange(1, 3, 2, 5)
	expected := "[1, 3) - [2, 5)"
	if r.String() != expected {
		t.Fatalf("expected %q, got %q", expected, r.String())
	}
}
