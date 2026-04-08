// Copyright 2000-2021 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import (
	"github.com/byron1st/file-diff/diff/myers"
	"github.com/byron1st/file-diff/diff/util"
)

// optimizeLineChunks adjusts chunk boundaries to prefer empty/unimportant lines
// as natural boundaries for insertions and deletions.
func optimizeLineChunks(lines1, lines2 []*Line, iterable FairDiffIterable) FairDiffIterable {
	ranges := buildOptimizedRanges(lines1, lines2, iterable)
	return Fair(CreateUnchanged(ranges, len(lines1), len(lines2)))
}

func buildOptimizedRanges(lines1, lines2 []*Line, iterable FairDiffIterable) []util.Range {
	var ranges []util.Range

	for _, r := range iterable.Unchanged() {
		ranges = append(ranges, r)
		processLastRanges(lines1, lines2, &ranges)
	}
	return ranges
}

func processLastRanges(lines1, lines2 []*Line, ranges *[]util.Range) {
	if len(*ranges) < 2 {
		return
	}

	r1 := (*ranges)[len(*ranges)-2]
	r2 := (*ranges)[len(*ranges)-1]

	if r1.End1 != r2.Start1 && r1.End2 != r2.Start2 {
		return
	}

	count1 := r1.End1 - r1.Start1
	count2 := r2.End1 - r2.Start1

	eqFwd := expandForwardLines(lines1, lines2, r1.End1, r1.End2, r1.End1+count2, r1.End2+count2)
	eqBwd := expandBackwardLines(lines1, lines2, r2.Start1-count1, r2.Start2-count1, r2.Start1, r2.Start2)

	if eqFwd == 0 && eqBwd == 0 {
		return
	}

	// merge left
	if eqFwd == count2 {
		*ranges = (*ranges)[:len(*ranges)-2]
		*ranges = append(*ranges, util.NewRange(r1.Start1, r1.End1+count2, r1.Start2, r1.End2+count2))
		processLastRanges(lines1, lines2, ranges)
		return
	}

	// merge right
	if eqBwd == count1 {
		*ranges = (*ranges)[:len(*ranges)-2]
		*ranges = append(*ranges, util.NewRange(r2.Start1-count1, r2.End1, r2.Start2-count1, r2.End2))
		processLastRanges(lines1, lines2, ranges)
		return
	}

	touchSideIsLeft := r1.End1 == r2.Start1
	shift := getLineShift(lines1, lines2, touchSideIsLeft, eqFwd, eqBwd, r1, r2)
	if shift != 0 {
		*ranges = (*ranges)[:len(*ranges)-2]
		*ranges = append(*ranges,
			util.NewRange(r1.Start1, r1.End1+shift, r1.Start2, r1.End2+shift),
			util.NewRange(r2.Start1+shift, r2.End1, r2.Start2+shift, r2.End2),
		)
	}
}

func getLineShift(lines1, lines2 []*Line, touchSideIsLeft bool, eqFwd, eqBwd int, r1, r2 util.Range) int {
	threshold := myers.UnimportantLineCharCount()

	var touchLines []*Line
	var touchStart int
	if touchSideIsLeft {
		touchLines = lines1
		touchStart = r2.Start1
	} else {
		touchLines = lines2
		touchStart = r2.Start2
	}

	// Try unchanged boundary shift
	if s := findBoundaryShift(touchLines, touchStart, eqFwd, eqBwd, 0); s != nil {
		return *s
	}

	// Try changed boundary shift
	var nonTouchLines []*Line
	var changeStart, changeEnd int
	if touchSideIsLeft {
		nonTouchLines = lines2
		changeStart = r1.End2
		changeEnd = r2.Start2
	} else {
		nonTouchLines = lines1
		changeStart = r1.End1
		changeEnd = r2.Start1
	}
	if s := findBoundaryShift(nonTouchLines, changeStart, eqFwd, eqBwd, 0); s != nil {
		return *s
	}

	// Try with threshold
	if s := findBoundaryShift(touchLines, touchStart, eqFwd, eqBwd, threshold); s != nil {
		return *s
	}
	if s := findBoundaryShift(nonTouchLines, changeStart, eqFwd, eqBwd, threshold); s != nil {
		_ = changeEnd // suppress unused
		return *s
	}

	return 0
}

func findBoundaryShift(lines []*Line, offset, eqFwd, eqBwd, threshold int) *int {
	fwd := findNextUnimportant(lines, offset, eqFwd+1, threshold)
	bwd := findPrevUnimportant(lines, offset-1, eqBwd+1, threshold)

	if fwd == -1 && bwd == -1 {
		return nil
	}
	if fwd == 0 || bwd == 0 {
		zero := 0
		return &zero
	}
	if fwd != -1 {
		return &fwd
	}
	neg := -bwd
	return &neg
}

func findNextUnimportant(lines []*Line, offset, count, threshold int) int {
	for i := range count {
		if lines[offset+i].nonSpaceChars <= threshold {
			return i
		}
	}
	return -1
}

func findPrevUnimportant(lines []*Line, offset, count, threshold int) int {
	for i := range count {
		if lines[offset-i].nonSpaceChars <= threshold {
			return i
		}
	}
	return -1
}

func expandForwardLines(lines1, lines2 []*Line, start1, start2, end1, end2 int) int {
	s1, s2 := start1, start2
	for s1 < end1 && s2 < end2 && lines1[s1].equals(lines2[s2]) {
		s1++
		s2++
	}
	return s1 - start1
}

func expandBackwardLines(lines1, lines2 []*Line, start1, start2, end1, end2 int) int {
	e1, e2 := end1, end2
	for start1 < e1 && start2 < e2 && lines1[e1-1].equals(lines2[e2-1]) {
		e1--
		e2--
	}
	return end1 - e1
}
