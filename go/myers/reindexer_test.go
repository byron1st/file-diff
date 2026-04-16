package myers

import "testing"

type recordedOp struct {
	kind   string
	first  int
	second int
}

type recordingBuilder struct {
	ops []recordedOp
}

func (b *recordingBuilder) AddEqual(length int) {
	b.ops = append(b.ops, recordedOp{kind: "equal", first: length})
}

func (b *recordingBuilder) AddChange(first, second int) {
	b.ops = append(b.ops, recordedOp{kind: "change", first: first, second: second})
}

func TestReindexer_DiscardUniqueKeepsCommonValues(t *testing.T) {
	r := &Reindexer{}
	discarded := r.DiscardUnique([]int{1, 2, 3, 4}, []int{2, 4, 5})

	if len(discarded[0]) != 2 || discarded[0][0] != 2 || discarded[0][1] != 4 {
		t.Fatalf("unexpected discarded left slice: %v", discarded[0])
	}
	if len(discarded[1]) != 2 || discarded[1][0] != 2 || discarded[1][1] != 4 {
		t.Fatalf("unexpected discarded right slice: %v", discarded[1])
	}
}

func TestReindexer_ReindexRestoresIntermediateAndTrailingGaps(t *testing.T) {
	r := &Reindexer{}
	discarded := r.DiscardUnique([]int{1, 2, 3, 4}, []int{1, 3, 4, 5})
	if len(discarded[0]) != 3 || len(discarded[1]) != 3 {
		t.Fatalf("expected 3 retained elements on each side, got %v", discarded)
	}

	builder := &recordingBuilder{}
	r.Reindex([2]*BitSet{NewBitSet(len(discarded[0])), NewBitSet(len(discarded[1]))}, builder)

	if len(builder.ops) != 4 {
		t.Fatalf("expected 4 operations, got %d: %v", len(builder.ops), builder.ops)
	}
	if builder.ops[0] != (recordedOp{kind: "equal", first: 1}) {
		t.Fatalf("unexpected first op: %v", builder.ops[0])
	}
	if builder.ops[1] != (recordedOp{kind: "change", first: 1, second: 0}) {
		t.Fatalf("unexpected second op: %v", builder.ops[1])
	}
	if builder.ops[2] != (recordedOp{kind: "equal", first: 2}) {
		t.Fatalf("unexpected third op: %v", builder.ops[2])
	}
	if builder.ops[3] != (recordedOp{kind: "change", first: 0, second: 1}) {
		t.Fatalf("unexpected fourth op: %v", builder.ops[3])
	}
}

func TestReindexer_ReindexMarksEverythingChangedWhenNothingIsRetained(t *testing.T) {
	r := &Reindexer{}
	discarded := r.DiscardUnique([]int{1, 2}, []int{3, 4})
	if len(discarded[0]) != 0 || len(discarded[1]) != 0 {
		t.Fatalf("expected both discarded slices to be empty, got %v", discarded)
	}

	builder := &recordingBuilder{}
	r.Reindex([2]*BitSet{NewBitSet(0), NewBitSet(0)}, builder)

	if len(builder.ops) != 1 {
		t.Fatalf("expected 1 operation, got %d: %v", len(builder.ops), builder.ops)
	}
	if builder.ops[0] != (recordedOp{kind: "change", first: 2, second: 2}) {
		t.Fatalf("unexpected operation for fully changed sequences: %v", builder.ops[0])
	}
}
