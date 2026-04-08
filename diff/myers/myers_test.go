package myers

import "testing"

func TestMyersLCS_Identical(t *testing.T) {
	a := []int{1, 2, 3}
	lcs := NewMyersLCS(a, a)
	lcs.Execute()
	ch := lcs.Changes()
	for i := range len(a) {
		if ch[0].Get(i) || ch[1].Get(i) {
			t.Fatalf("expected no changes for identical arrays at index %d", i)
		}
	}
}

func TestMyersLCS_CompletelyDifferent(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{4, 5, 6}
	lcs := NewMyersLCS(a, b)
	lcs.Execute()
	ch := lcs.Changes()
	for i := range len(a) {
		if !ch[0].Get(i) {
			t.Fatalf("expected all changed in first at index %d", i)
		}
	}
	for i := range len(b) {
		if !ch[1].Get(i) {
			t.Fatalf("expected all changed in second at index %d", i)
		}
	}
}

func TestMyersLCS_InsertMiddle(t *testing.T) {
	a := []int{1, 3}
	b := []int{1, 2, 3}
	lcs := NewMyersLCS(a, b)
	lcs.Execute()
	ch := lcs.Changes()
	// 1 and 3 should be unchanged in both
	if ch[0].Get(0) || ch[0].Get(1) {
		t.Fatal("expected first array fully in LCS")
	}
	// b[1]=2 should be a change (insertion)
	if !ch[1].Get(1) {
		t.Fatal("expected b[1] to be changed (inserted)")
	}
	if ch[1].Get(0) || ch[1].Get(2) {
		t.Fatal("expected b[0] and b[2] to be unchanged")
	}
}

func TestBuildChanges_SimpleInsert(t *testing.T) {
	a := []int{1, 3}
	b := []int{1, 2, 3}
	ch, err := BuildChanges(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if ch == nil {
		t.Fatal("expected a change")
	}
	if ch.Line0 != 1 || ch.Line1 != 1 || ch.Deleted != 0 || ch.Inserted != 1 {
		t.Fatalf("unexpected change: line0=%d line1=%d del=%d ins=%d", ch.Line0, ch.Line1, ch.Deleted, ch.Inserted)
	}
	if ch.Link != nil {
		t.Fatal("expected single change")
	}
}

func TestBuildChanges_SimpleDelete(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{1, 3}
	ch, err := BuildChanges(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if ch == nil {
		t.Fatal("expected a change")
	}
	if ch.Line0 != 1 || ch.Line1 != 1 || ch.Deleted != 1 || ch.Inserted != 0 {
		t.Fatalf("unexpected change: line0=%d line1=%d del=%d ins=%d", ch.Line0, ch.Line1, ch.Deleted, ch.Inserted)
	}
}

func TestBuildChanges_Identical(t *testing.T) {
	a := []int{1, 2, 3}
	ch, err := BuildChanges(a, a)
	if err != nil {
		t.Fatal(err)
	}
	if ch != nil {
		t.Fatal("expected no changes for identical arrays")
	}
}

func TestBuildChanges_Modification(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{1, 4, 3}
	ch, err := BuildChanges(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if ch == nil {
		t.Fatal("expected a change")
	}
	if ch.Deleted != 1 || ch.Inserted != 1 {
		t.Fatalf("expected 1 deleted + 1 inserted, got del=%d ins=%d", ch.Deleted, ch.Inserted)
	}
}
