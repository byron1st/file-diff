package comparison

import "testing"

func TestChangeBuilder_IndexTracking(t *testing.T) {
	builder := NewChangeBuilder(5, 6)
	if builder.Index1() != 0 || builder.Index2() != 0 {
		t.Fatalf("expected zero indexes, got (%d, %d)", builder.Index1(), builder.Index2())
	}

	builder.MarkEqualRange(1, 2, 3, 4)
	if builder.Index1() != 3 || builder.Index2() != 4 {
		t.Fatalf("expected indexes to advance to (3, 4), got (%d, %d)", builder.Index1(), builder.Index2())
	}

	changes := builder.Finish().Changes()
	if len(changes) != 2 {
		t.Fatalf("expected 2 change ranges, got %d", len(changes))
	}
	if changes[0].Start1 != 0 || changes[0].End1 != 1 || changes[0].Start2 != 0 || changes[0].End2 != 2 {
		t.Fatalf("unexpected first change: %v", changes[0])
	}
	if changes[1].Start1 != 3 || changes[1].End1 != 5 || changes[1].Start2 != 4 || changes[1].End2 != 6 {
		t.Fatalf("unexpected second change: %v", changes[1])
	}
}

func TestExpandChangeBuilder_ShrinksEqualBorders(t *testing.T) {
	left := []lineEquatable{
		NewLine("same", PolicyDefault),
		NewLine("left-only", PolicyDefault),
		NewLine("tail", PolicyDefault),
	}
	right := []lineEquatable{
		NewLine("same", PolicyDefault),
		NewLine("right-only", PolicyDefault),
		NewLine("tail", PolicyDefault),
	}

	builder := NewExpandChangeBuilder(left, right)
	changes := builder.Finish().Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 shrunk change, got %d", len(changes))
	}

	ch := changes[0]
	if ch.Start1 != 1 || ch.End1 != 2 || ch.Start2 != 1 || ch.End2 != 2 {
		t.Fatalf("unexpected shrunk change: %v", ch)
	}
}

func TestFair_WrapsPlainDiffIterable(t *testing.T) {
	iterable := CreateFromRanges(nil, 2, 3)
	fair := Fair(iterable)
	if fair.Length1() != 2 || fair.Length2() != 3 {
		t.Fatalf("unexpected lengths after Fair wrap: (%d, %d)", fair.Length1(), fair.Length2())
	}
}
