// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import (
	"github.com/byron1st/file-diff/diff/myers"
	"github.com/byron1st/file-diff/diff/util"
)

// CreateFromRanges creates a DiffIterable from a list of changed ranges.
func CreateFromRanges(changes []util.Range, length1, length2 int) DiffIterable {
	return &rangesDiffIterable{changes: changes, length1: length1, length2: length2}
}

// CreateFromChanges creates a DiffIterable from a linked list of Change records.
func CreateFromChanges(change *myers.Change, length1, length2 int) DiffIterable {
	var ranges []util.Range
	for ch := change; ch != nil; ch = ch.Link {
		ranges = append(ranges, util.NewRange(ch.Line0, ch.Line0+ch.Deleted, ch.Line1, ch.Line1+ch.Inserted))
	}
	return CreateFromRanges(ranges, length1, length2)
}

// CreateUnchanged creates a DiffIterable from a list of unchanged ranges.
func CreateUnchanged(unchanged []util.Range, length1, length2 int) DiffIterable {
	return Invert(CreateFromRanges(unchanged, length1, length2))
}

// Invert swaps the changed/unchanged interpretation.
func Invert(iterable DiffIterable) DiffIterable {
	return &invertedDiffIterable{inner: iterable}
}

// Fair wraps a DiffIterable as a FairDiffIterable.
func Fair(iterable DiffIterable) FairDiffIterable {
	if f, ok := iterable.(FairDiffIterable); ok {
		return f
	}
	return &fairDiffIterableWrapper{iterable}
}

// Diff compares two equal-hash integer arrays using Myers and returns a FairDiffIterable.
func Diff(data1, data2 []int) (FairDiffIterable, error) {
	change, err := myers.BuildChanges(data1, data2)
	if err != nil {
		return nil, err
	}
	return Fair(CreateFromChanges(change, len(data1), len(data2))), nil
}

// DiffObjects compares two slices of objects (using a hash function for equality)
// and returns a FairDiffIterable.
func DiffObjects[T comparable](data1, data2 []T) (FairDiffIterable, error) {
	// Enumerate objects to int IDs
	m := make(map[T]int)
	nextID := 1
	enumerate := func(obj T) int {
		if id, ok := m[obj]; ok {
			return id
		}
		id := nextID
		nextID++
		m[obj] = id
		return id
	}

	ints1 := make([]int, len(data1))
	for i, v := range data1 {
		ints1[i] = enumerate(v)
	}
	ints2 := make([]int, len(data2))
	for i, v := range data2 {
		ints2[i] = enumerate(v)
	}
	return Diff(ints1, ints2)
}

// ChangeBuilder collects markEqual calls and produces a DiffIterable.
type ChangeBuilder struct {
	length1 int
	length2 int
	index1  int
	index2  int
	changes []util.Range
}

func NewChangeBuilder(length1, length2 int) *ChangeBuilder {
	return &ChangeBuilder{length1: length1, length2: length2}
}

func (b *ChangeBuilder) Index1() int { return b.index1 }
func (b *ChangeBuilder) Index2() int { return b.index2 }

func (b *ChangeBuilder) MarkEqual(index1, index2 int) {
	b.MarkEqualRange(index1, index2, index1+1, index2+1)
}

func (b *ChangeBuilder) MarkEqualCount(index1, index2, count int) {
	b.MarkEqualRange(index1, index2, index1+count, index2+count)
}

func (b *ChangeBuilder) MarkEqualRange(index1, index2, end1, end2 int) {
	if index1 == end1 && index2 == end2 {
		return
	}
	if b.index1 != index1 || b.index2 != index2 {
		b.addChange(b.index1, b.index2, index1, index2)
	}
	b.index1 = end1
	b.index2 = end2
}

func (b *ChangeBuilder) Finish() DiffIterable {
	if b.length1 != b.index1 || b.length2 != b.index2 {
		b.addChange(b.index1, b.index2, b.length1, b.length2)
		b.index1 = b.length1
		b.index2 = b.length2
	}
	return CreateFromRanges(b.changes, b.length1, b.length2)
}

func (b *ChangeBuilder) addChange(start1, start2, end1, end2 int) {
	b.changes = append(b.changes, util.NewRange(start1, end1, start2, end2))
}

// ExpandChangeBuilder extends ChangeBuilder by expanding equal ranges to
// consume adjacent equal elements (using list equality).
type ExpandChangeBuilder struct {
	*ChangeBuilder
	objects1 []lineEquatable
	objects2 []lineEquatable
}

type lineEquatable interface {
	Equals(other lineEquatable) bool
}

func NewExpandChangeBuilder(objects1, objects2 []lineEquatable) *ExpandChangeBuilder {
	return &ExpandChangeBuilder{
		ChangeBuilder: NewChangeBuilder(len(objects1), len(objects2)),
		objects1:      objects1,
		objects2:      objects2,
	}
}

func (b *ExpandChangeBuilder) addChange(start1, start2, end1, end2 int) {
	r := expandRange(b.objects1, b.objects2, start1, start2, end1, end2)
	if !r.IsEmpty() {
		b.changes = append(b.changes, r)
	}
}

func (b *ExpandChangeBuilder) MarkEqual(index1, index2 int) {
	b.MarkEqualRange(index1, index2, index1+1, index2+1)
}

func (b *ExpandChangeBuilder) MarkEqualRange(index1, index2, end1, end2 int) {
	if index1 == end1 && index2 == end2 {
		return
	}
	if b.index1 != index1 || b.index2 != index2 {
		b.addChange(b.index1, b.index2, index1, index2)
	}
	b.index1 = end1
	b.index2 = end2
}

func (b *ExpandChangeBuilder) Finish() DiffIterable {
	if b.length1 != b.index1 || b.length2 != b.index2 {
		b.addChange(b.index1, b.index2, b.length1, b.length2)
		b.index1 = b.length1
		b.index2 = b.length2
	}
	return CreateFromRanges(b.changes, b.length1, b.length2)
}

// expandRange shrinks a change range by consuming equal elements at the borders.
func expandRange(objects1, objects2 []lineEquatable, start1, start2, end1, end2 int) util.Range {
	// Expand forward
	s1, s2 := start1, start2
	for s1 < end1 && s2 < end2 && objects1[s1].Equals(objects2[s2]) {
		s1++
		s2++
	}
	// Expand backward
	e1, e2 := end1, end2
	for s1 < e1 && s2 < e2 && objects1[e1-1].Equals(objects2[e2-1]) {
		e1--
		e2--
	}
	return util.NewRange(s1, e1, s2, e2)
}
