// Copyright 2000-2021 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import "github.com/byron1st/file-diff/go/util"

// smartLineChangeCorrector corrects a "big lines only" diff by filling in
// the gaps with a full diff of all lines. This is the two-step approach
// from ByLineRt.compareSmart.
type smartLineChangeCorrector struct {
	indexes1 []int
	indexes2 []int
	lines1   []*Line
	lines2   []*Line
	changes  FairDiffIterable
}

func newSmartLineChangeCorrector(
	indexes1, indexes2 []int,
	lines1, lines2 []*Line,
	changes FairDiffIterable,
) *smartLineChangeCorrector {
	return &smartLineChangeCorrector{
		indexes1: indexes1,
		indexes2: indexes2,
		lines1:   lines1,
		lines2:   lines2,
		changes:  changes,
	}
}

func (c *smartLineChangeCorrector) build() FairDiffIterable {
	builder := NewChangeBuilder(len(c.lines1), len(c.lines2))
	last1, last2 := 0, 0

	for _, ch := range c.changes.Unchanged() {
		count := ch.End1 - ch.Start1
		for i := range count {
			origIdx1 := c.indexes1[ch.Start1+i]
			origIdx2 := c.indexes2[ch.Start2+i]

			c.matchGap(builder, last1, origIdx1, last2, origIdx2)
			builder.MarkEqualRange(origIdx1, origIdx2, origIdx1+1, origIdx2+1)

			last1 = origIdx1 + 1
			last2 = origIdx2 + 1
		}
	}
	c.matchGap(builder, last1, len(c.lines1), last2, len(c.lines2))
	return Fair(builder.Finish())
}

func (c *smartLineChangeCorrector) matchGap(builder *ChangeBuilder, start1, end1, start2, end2 int) {
	expanded := expandRangeLines(c.lines1, c.lines2, start1, start2, end1, end2)

	inner1 := c.lines1[expanded.Start1:expanded.End1]
	inner2 := c.lines2[expanded.Start2:expanded.End2]

	innerChanges, err := diffLineSlice(inner1, inner2)
	if err != nil {
		return
	}

	builder.MarkEqualRange(start1, start2, expanded.Start1, expanded.Start2)
	for _, ch := range innerChanges.Unchanged() {
		builder.MarkEqualCount(expanded.Start1+ch.Start1, expanded.Start2+ch.Start2, ch.End1-ch.Start1)
	}
	builder.MarkEqualRange(expanded.End1, expanded.End2, end1, end2)
}

func expandRangeLines(lines1, lines2 []*Line, start1, start2, end1, end2 int) util.Range {
	s1, s2 := start1, start2
	for s1 < end1 && s2 < end2 && lines1[s1].equals(lines2[s2]) {
		s1++
		s2++
	}
	e1, e2 := end1, end2
	for s1 < e1 && s2 < e2 && lines1[e1-1].equals(lines2[e2-1]) {
		e1--
		e2--
	}
	return util.NewRange(s1, e1, s2, e2)
}

// diffLineSlice runs a diff on two slices of *Line using their hash/equals.
func diffLineSlice(lines1, lines2 []*Line) (FairDiffIterable, error) {
	ints1 := make([]int, len(lines1))
	for i, l := range lines1 {
		ints1[i] = l.hash
	}
	ints2 := make([]int, len(lines2))
	for i, l := range lines2 {
		ints2[i] = l.hash
	}
	return Diff(ints1, ints2)
}
