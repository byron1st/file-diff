package util

import "testing"

func TestEnumerator_UniqueIDs(t *testing.T) {
	e := NewEnumerator(4)
	ids := e.Enumerate([]string{"a", "b", "c", "a"})

	if ids[0] != ids[3] {
		t.Fatalf("same string should get same ID: %d != %d", ids[0], ids[3])
	}
	if ids[0] == ids[1] || ids[1] == ids[2] {
		t.Fatalf("different strings should get different IDs: %v", ids)
	}
}

func TestEnumerator_StartsFromOne(t *testing.T) {
	e := NewEnumerator(2)
	ids := e.Enumerate([]string{"x"})
	if ids[0] != 1 {
		t.Fatalf("expected first ID to be 1, got %d", ids[0])
	}
}
