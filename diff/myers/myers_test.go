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
	if ch.Line0 != 1 || ch.Line1 != 1 {
		t.Fatalf("expected change at position 1, got line0=%d line1=%d", ch.Line0, ch.Line1)
	}
}

func TestUnimportantLineCharCount(t *testing.T) {
	if UnimportantLineCharCount() != 3 {
		t.Fatalf("expected 3, got %d", UnimportantLineCharCount())
	}
}

func TestBuildChangesFromObjects_Identical(t *testing.T) {
	a := []string{"foo", "bar", "baz"}
	ch, err := BuildChangesFromObjects(a, a)
	if err != nil {
		t.Fatal(err)
	}
	if ch != nil {
		t.Fatal("expected no changes for identical arrays")
	}
}

func TestBuildChangesFromObjects_Insert(t *testing.T) {
	a := []string{"foo", "baz"}
	b := []string{"foo", "bar", "baz"}
	ch, err := BuildChangesFromObjects(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if ch == nil {
		t.Fatal("expected a change")
	}
	if ch.Deleted != 0 || ch.Inserted != 1 {
		t.Fatalf("expected 0 deleted + 1 inserted, got del=%d ins=%d", ch.Deleted, ch.Inserted)
	}
}

func TestBuildChangesFromObjects_Delete(t *testing.T) {
	a := []string{"foo", "bar", "baz"}
	b := []string{"foo", "baz"}
	ch, err := BuildChangesFromObjects(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if ch == nil {
		t.Fatal("expected a change")
	}
	if ch.Deleted != 1 || ch.Inserted != 0 {
		t.Fatalf("expected 1 deleted + 0 inserted, got del=%d ins=%d", ch.Deleted, ch.Inserted)
	}
}

func TestBuildChangesFromObjects_Modification(t *testing.T) {
	a := []string{"foo", "bar", "baz"}
	b := []string{"foo", "qux", "baz"}
	ch, err := BuildChangesFromObjects(a, b)
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

func TestExecuteWithThreshold_SmallInput(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{1, 4, 3}
	lcs := NewMyersLCS(a, b)
	if err := lcs.ExecuteWithThreshold(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ch := lcs.Changes()
	// index 1 should be changed in both
	if !ch[0].Get(1) || !ch[1].Get(1) {
		t.Fatal("expected index 1 to be changed")
	}
	// index 0, 2 should be unchanged
	if ch[0].Get(0) || ch[0].Get(2) || ch[1].Get(0) || ch[1].Get(2) {
		t.Fatal("expected indices 0 and 2 to be unchanged")
	}
}

func TestMyersLCS_EmptyInput(t *testing.T) {
	lcs := NewMyersLCS([]int{}, []int{1, 2})
	lcs.Execute()
	ch := lcs.Changes()
	// empty first: all second should be changed
	for i := range 2 {
		if !ch[1].Get(i) {
			t.Fatalf("expected second[%d] to be changed", i)
		}
	}
}

func TestBuildChanges_BothEmpty(t *testing.T) {
	ch, err := BuildChanges([]int{}, []int{})
	if err != nil {
		t.Fatal(err)
	}
	if ch != nil {
		t.Fatal("expected no changes for empty arrays")
	}
}

func TestBuildChanges_OneEmpty(t *testing.T) {
	ch, err := BuildChanges([]int{}, []int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if ch == nil {
		t.Fatal("expected a change")
	}
	if ch.Deleted != 0 || ch.Inserted != 3 {
		t.Fatalf("expected 0 deleted + 3 inserted, got del=%d ins=%d", ch.Deleted, ch.Inserted)
	}
}

func TestBitSet_EnsureCapacity(t *testing.T) {
	bs := NewBitSet(2)
	// Set beyond initial capacity
	bs.Set(100, true)
	if !bs.Get(100) {
		t.Fatal("expected bit 100 to be set")
	}
	if bs.Get(99) {
		t.Fatal("expected bit 99 to be unset")
	}
}

func TestBitSet_GetOutOfRange(t *testing.T) {
	bs := NewBitSet(5)
	// Negative index
	if bs.Get(-1) {
		t.Fatal("expected false for negative index")
	}
	// Beyond size
	if bs.Get(10) {
		t.Fatal("expected false for index beyond size")
	}
}

func TestBitSet_SetFalse(t *testing.T) {
	bs := NewBitSet(10)
	bs.Set(3, true)
	if !bs.Get(3) {
		t.Fatal("expected bit 3 to be set")
	}
	bs.Set(3, false)
	if bs.Get(3) {
		t.Fatal("expected bit 3 to be unset after clearing")
	}
}

func TestBuildChanges_MultipleChanges(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}
	b := []int{1, 9, 3, 8, 5}
	ch, err := BuildChanges(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if ch == nil {
		t.Fatal("expected changes")
	}
	// Should have two changes linked together
	if ch.Link == nil {
		t.Fatal("expected linked changes")
	}
}
