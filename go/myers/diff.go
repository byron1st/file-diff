// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package myers

import "github.com/byron1st/file-diff/go/util"

const (
	deltaThresholdSize       = 20000
	unimportantLineCharCount = 3
)

// UnimportantLineCharCount is the threshold below which a line is considered
// "unimportant" (e.g., blank or very short).
func UnimportantLineCharCount() int {
	return unimportantLineCharCount
}

// Change represents a single edit operation: deleted lines from sequence 1
// starting at Line0, and inserted lines in sequence 2 starting at Line1.
type Change struct {
	Line0    int
	Line1    int
	Deleted  int
	Inserted int
	Link     *Change
}

// BuildChanges computes the diff between two integer arrays and returns
// a linked list of Change records.
func BuildChanges(ints1, ints2 []int) (*Change, error) {
	startShift := getStartShift(ints1, ints2)
	endCut := getEndCut(ints1, ints2, startShift)

	if ch := buildChangesFast(len(ints1), len(ints2), startShift, endCut); ch != nil {
		return *ch, nil
	}

	copyArr := startShift != 0 || endCut != 0
	a1 := ints1
	a2 := ints2
	if copyArr {
		a1 = ints1[startShift : len(ints1)-endCut]
		a2 = ints2[startShift : len(ints2)-endCut]
	}
	return doBuildChanges(a1, a2, startShift)
}

// BuildChangesFromObjects enumerates objects by equality (hash+equals) then
// runs the integer diff.
func BuildChangesFromObjects(objects1, objects2 []string) (*Change, error) {
	startShift := getStartShiftStr(objects1, objects2)
	endCut := getEndCutStr(objects1, objects2, startShift)

	if ch := buildChangesFast(len(objects1), len(objects2), startShift, endCut); ch != nil {
		return *ch, nil
	}

	trimmedLen := len(objects1) + len(objects2) - 2*startShift - 2*endCut
	e := util.NewEnumerator(trimmedLen)
	ints1 := e.Enumerate(objects1[startShift : len(objects1)-endCut])
	ints2 := e.Enumerate(objects2[startShift : len(objects2)-endCut])
	return doBuildChanges(ints1, ints2, startShift)
}

func buildChangesFast(len1, len2, startShift, endCut int) **Change {
	trimmed1 := len1 - startShift - endCut
	trimmed2 := len2 - startShift - endCut
	if trimmed1 != 0 && trimmed2 != 0 {
		return nil
	}
	var ch *Change
	if trimmed1 != 0 || trimmed2 != 0 {
		ch = &Change{Line0: startShift, Line1: startShift, Deleted: trimmed1, Inserted: trimmed2}
	}
	return &ch
}

func doBuildChanges(ints1, ints2 []int, startShift int) (*Change, error) {
	reindexer := &Reindexer{}
	discarded := reindexer.DiscardUnique(ints1, ints2)

	builder := &changeBuilder{startShift: startShift}

	if len(discarded[0]) == 0 && len(discarded[1]) == 0 {
		builder.AddChange(len(ints1), len(ints2))
		return builder.first, nil
	}

	lcs := NewMyersLCS(discarded[0], discarded[1])
	if err := lcs.ExecuteWithThreshold(); err != nil {
		// fallback: just mark everything as changed
		builder.AddChange(len(ints1), len(ints2))
		return builder.first, nil
	}

	changes := lcs.Changes()
	reindexer.Reindex(changes, builder)
	return builder.first, nil
}

// changeBuilder collects LCS results into a linked list of Changes.
type changeBuilder struct {
	index1     int
	index2     int
	first      *Change
	last       *Change
	startShift int
}

func (b *changeBuilder) AddChange(first, second int) {
	ch := &Change{Line0: b.startShift + b.index1, Line1: b.startShift + b.index2, Deleted: first, Inserted: second}
	if b.last != nil {
		b.last.Link = ch
	} else {
		b.first = ch
	}
	b.last = ch
	b.skip(first, second)
}

func (b *changeBuilder) AddEqual(length int) {
	b.skip(length, length)
}

func (b *changeBuilder) skip(first, second int) {
	b.index1 += first
	b.index2 += second
}

func getStartShift(a, b []int) int {
	n := min(len(a), len(b))
	for i := range n {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func getEndCut(a, b []int, startShift int) int {
	n := min(len(a), len(b)) - startShift
	for i := range n {
		if a[len(a)-1-i] != b[len(b)-1-i] {
			return i
		}
	}
	return n
}

func getStartShiftStr(a, b []string) int {
	n := min(len(a), len(b))
	for i := range n {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func getEndCutStr(a, b []string, startShift int) int {
	n := min(len(a), len(b)) - startShift
	for i := range n {
		if a[len(a)-1-i] != b[len(b)-1-i] {
			return i
		}
	}
	return n
}
