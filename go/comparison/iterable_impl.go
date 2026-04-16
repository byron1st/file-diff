// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import "github.com/byron1st/file-diff/go/util"

// rangesDiffIterable is backed by a list of changed ranges.
type rangesDiffIterable struct {
	changes []util.Range
	length1 int
	length2 int
}

func (r *rangesDiffIterable) Length1() int          { return r.length1 }
func (r *rangesDiffIterable) Length2() int          { return r.length2 }
func (r *rangesDiffIterable) Changes() []util.Range { return r.changes }

func (r *rangesDiffIterable) Unchanged() []util.Range {
	return computeUnchanged(r.changes, r.length1, r.length2)
}

// invertedDiffIterable swaps changes and unchanged ranges.
type invertedDiffIterable struct {
	inner DiffIterable
}

func (inv *invertedDiffIterable) Length1() int            { return inv.inner.Length1() }
func (inv *invertedDiffIterable) Length2() int            { return inv.inner.Length2() }
func (inv *invertedDiffIterable) Changes() []util.Range   { return inv.inner.Unchanged() }
func (inv *invertedDiffIterable) Unchanged() []util.Range { return inv.inner.Changes() }

// fairDiffIterableWrapper wraps a DiffIterable as a FairDiffIterable.
type fairDiffIterableWrapper struct {
	DiffIterable
}

// computeUnchanged derives unchanged ranges from a list of changed ranges.
func computeUnchanged(changes []util.Range, length1, length2 int) []util.Range {
	var unchanged []util.Range
	last1, last2 := 0, 0

	for _, ch := range changes {
		if ch.Start1 > last1 || ch.Start2 > last2 {
			unchanged = append(unchanged, util.NewRange(last1, ch.Start1, last2, ch.Start2))
		}
		last1 = ch.End1
		last2 = ch.End2
	}
	if last1 < length1 || last2 < length2 {
		unchanged = append(unchanged, util.NewRange(last1, length1, last2, length2))
	}
	return unchanged
}
