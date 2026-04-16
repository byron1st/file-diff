// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package myers

import "sort"

// LCSBuilder receives the results of an LCS computation as a stream of
// equal-length runs and changed blocks.
type LCSBuilder interface {
	AddEqual(length int)
	AddChange(first, second int)
}

// Reindexer discards elements that appear in only one of the two sequences
// (since they can never be part of the LCS), then maps the results back
// to original indices.
type Reindexer struct {
	oldIndices      [2][]int
	originalLengths [2]int
	discardedLens   [2]int
}

// DiscardUnique removes elements unique to each side and returns the
// filtered arrays. The reindexer remembers the mapping for Reindex.
func (r *Reindexer) DiscardUnique(ints1, ints2 []int) [2][]int {
	discarded1 := r.discard(ints2, ints1, 0)
	discarded2 := r.discard(discarded1, ints2, 1)
	return [2][]int{discarded1, discarded2}
}

// Reindex maps changes from the discarded space back to the original space
// and streams the result to the builder.
func (r *Reindexer) Reindex(changes [2]*BitSet, builder LCSBuilder) {
	var changes1, changes2 *BitSet

	if r.discardedLens[0] == r.originalLengths[0] && r.discardedLens[1] == r.originalLengths[1] {
		changes1 = changes[0]
		changes2 = changes[1]
	} else {
		changes1 = NewBitSet(r.originalLengths[0])
		changes2 = NewBitSet(r.originalLengths[1])

		x, y := 0, 0
		for x < r.discardedLens[0] || y < r.discardedLens[1] {
			if x < r.discardedLens[0] && y < r.discardedLens[1] && !changes[0].Get(x) && !changes[1].Get(y) {
				x = r.increment(r.oldIndices[0], x, changes1, r.originalLengths[0])
				y = r.increment(r.oldIndices[1], y, changes2, r.originalLengths[1])
			} else if x < r.discardedLens[0] && changes[0].Get(x) {
				changes1.Set(r.oldIndices[0][x], true)
				x = r.increment(r.oldIndices[0], x, changes1, r.originalLengths[0])
			} else if y < r.discardedLens[1] && changes[1].Get(y) {
				changes2.Set(r.oldIndices[1][y], true)
				y = r.increment(r.oldIndices[1], y, changes2, r.originalLengths[1])
			}
		}

		if r.discardedLens[0] == 0 {
			changes1.SetRange(0, r.originalLengths[0], true)
		} else {
			changes1.SetRange(0, r.oldIndices[0][0], true)
		}
		if r.discardedLens[1] == 0 {
			changes2.SetRange(0, r.originalLengths[1], true)
		} else {
			changes2.SetRange(0, r.oldIndices[1][0], true)
		}
	}

	x, y := 0, 0
	for x < r.originalLengths[0] && y < r.originalLengths[1] {
		startX := x
		for x < r.originalLengths[0] && y < r.originalLengths[1] && !changes1.Get(x) && !changes2.Get(y) {
			x++
			y++
		}
		if x > startX {
			builder.AddEqual(x - startX)
		}
		dx, dy := 0, 0
		for x < r.originalLengths[0] && changes1.Get(x) {
			dx++
			x++
		}
		for y < r.originalLengths[1] && changes2.Get(y) {
			dy++
			y++
		}
		if dx != 0 || dy != 0 {
			builder.AddChange(dx, dy)
		}
	}
	if x != r.originalLengths[0] || y != r.originalLengths[1] {
		builder.AddChange(r.originalLengths[0]-x, r.originalLengths[1]-y)
	}
}

func (r *Reindexer) discard(needed, toDiscard []int, arrayIdx int) []int {
	r.originalLengths[arrayIdx] = len(toDiscard)
	sorted := make([]int, len(needed))
	copy(sorted, needed)
	sort.Ints(sorted)

	var discarded []int
	var oldIdx []int
	for i, v := range toDiscard {
		if sort.SearchInts(sorted, v) < len(sorted) && sorted[sort.SearchInts(sorted, v)] == v {
			discarded = append(discarded, v)
			oldIdx = append(oldIdx, i)
		}
	}
	r.oldIndices[arrayIdx] = oldIdx
	r.discardedLens[arrayIdx] = len(discarded)
	if discarded == nil {
		return []int{}
	}
	return discarded
}

func (r *Reindexer) increment(indices []int, i int, set *BitSet, length int) int {
	if i+1 < len(indices) {
		set.SetRange(indices[i]+1, indices[i+1], true)
	} else {
		set.SetRange(indices[i]+1, length, true)
	}
	return i + 1
}
