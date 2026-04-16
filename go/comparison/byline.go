// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import (
	"unicode"

	"github.com/byron1st/file-diff/go/myers"
	"github.com/byron1st/file-diff/go/util"
)

// Line wraps a line of text with its comparison policy, cached hash, and
// non-space character count.
type Line struct {
	Content       string
	Policy        ComparisonPolicy
	hash          int
	nonSpaceChars int
}

func NewLine(content string, policy ComparisonPolicy) *Line {
	return &Line{
		Content:       content,
		Policy:        policy,
		hash:          HashCode(content, policy),
		nonSpaceChars: countNonSpaceChars(content),
	}
}

func (l *Line) equals(other *Line) bool {
	if l == other {
		return true
	}
	if l.hash != other.hash {
		return false
	}
	return IsEqual(l.Content, other.Content, l.Policy)
}

// Equals implements lineEquatable for ExpandChangeBuilder.
func (l *Line) Equals(other lineEquatable) bool {
	if o, ok := other.(*Line); ok {
		return l.equals(o)
	}
	return false
}

func countNonSpaceChars(s string) int {
	n := 0
	for _, c := range s {
		if !unicode.IsSpace(c) {
			n++
		}
	}
	return n
}

// CompareLines is the main entry point for line-level comparison.
// It implements the ByLineRt.compare algorithm.
func CompareLines(lines1, lines2 []string, policy ComparisonPolicy) FairDiffIterable {
	l1 := toLines(lines1, policy)
	l2 := toLines(lines2, policy)
	return doCompareLines(l1, l2, policy)
}

func doCompareLines(lines1, lines2 []*Line, policy ComparisonPolicy) FairDiffIterable {
	if policy == PolicyIgnoreWhitespaces {
		changes := compareSmart(lines1, lines2)
		changes = optimizeLineChunks(lines1, lines2, changes)
		return expandRanges(lines1, lines2, changes)
	}

	iwLines1 := convertMode(lines1, PolicyIgnoreWhitespaces)
	iwLines2 := convertMode(lines2, PolicyIgnoreWhitespaces)

	iwChanges := compareSmart(iwLines1, iwLines2)
	iwChanges = optimizeLineChunks(lines1, lines2, iwChanges)
	return correctChangesSecondStep(lines1, lines2, iwChanges)
}

// compareSmart compares in two steps:
// 1. Compare ignoring "unimportant" (short) lines
// 2. Fill gaps using full comparison
func compareSmart(lines1, lines2 []*Line) FairDiffIterable {
	threshold := myers.UnimportantLineCharCount()
	if threshold == 0 {
		result, _ := diffLineSlice(lines1, lines2)
		return result
	}

	bigLines1, idx1 := getBigLines(lines1, threshold)
	bigLines2, idx2 := getBigLines(lines2, threshold)

	changes, err := diffLineSlice(bigLines1, bigLines2)
	if err != nil {
		// fallback to full diff
		result, _ := diffLineSlice(lines1, lines2)
		return result
	}

	return newSmartLineChangeCorrector(idx1, idx2, lines1, lines2, changes).build()
}

func getBigLines(lines []*Line, threshold int) ([]*Line, []int) {
	var big []*Line
	var indexes []int
	for i, l := range lines {
		if l.nonSpaceChars > threshold {
			big = append(big, l)
			indexes = append(indexes, i)
		}
	}
	return big, indexes
}

// correctChangesSecondStep adjusts IW-matched lines to prefer exact matches.
func correctChangesSecondStep(lines1, lines2 []*Line, changes FairDiffIterable) FairDiffIterable {
	eqLines1 := make([]lineEquatable, len(lines1))
	for i, l := range lines1 {
		eqLines1[i] = l
	}
	eqLines2 := make([]lineEquatable, len(lines2))
	for i, l := range lines2 {
		eqLines2[i] = l
	}
	builder := NewExpandChangeBuilder(eqLines1, eqLines2)

	var sample string
	hasSample := false
	last1, last2 := 0, 0

	for _, r := range changes.Unchanged() {
		count := r.End1 - r.Start1
		for i := range count {
			idx1 := r.Start1 + i
			idx2 := r.Start2 + i
			l1 := lines1[idx1]
			l2 := lines2[idx2]

			if !hasSample || !IsEqual(sample, l1.Content, PolicyIgnoreWhitespaces) {
				if l1.equals(l2) {
					flushSecondStep(builder, lines1, lines2, sample, hasSample, last1, last2, idx1, idx2)
					hasSample = false
					builder.MarkEqual(idx1, idx2)
				} else {
					flushSecondStep(builder, lines1, lines2, sample, hasSample, last1, last2, idx1, idx2)
					sample = l1.Content
					hasSample = true
				}
			}
			last1 = idx1 + 1
			last2 = idx2 + 1
		}
	}
	flushSecondStep(builder, lines1, lines2, sample, hasSample, last1, last2, changes.Length1(), changes.Length2())

	return Fair(builder.Finish())
}

func flushSecondStep(
	builder *ExpandChangeBuilder,
	lines1, lines2 []*Line,
	sample string, hasSample bool,
	last1, last2, line1, line2 int,
) {
	if !hasSample {
		return
	}

	start1 := max(last1, builder.Index1())
	start2 := max(last2, builder.Index2())

	var sub1, sub2 []int
	for i := start1; i < line1; i++ {
		if IsEqual(sample, lines1[i].Content, PolicyIgnoreWhitespaces) {
			sub1 = append(sub1, i)
		}
	}
	for i := start2; i < line2; i++ {
		if IsEqual(sample, lines2[i].Content, PolicyIgnoreWhitespaces) {
			sub2 = append(sub2, i)
		}
	}

	if len(sub1) == 0 || len(sub2) == 0 {
		return
	}

	alignExactMatching(builder, lines1, lines2, sub1, sub2)
}

func alignExactMatching(builder *ExpandChangeBuilder, lines1, lines2 []*Line, sub1, sub2 []int) {
	n := max(len(sub1), len(sub2))
	skipAligning := n > 10 || len(sub1) == len(sub2)

	if skipAligning {
		count := min(len(sub1), len(sub2))
		for i := range count {
			if lines1[sub1[i]].equals(lines2[sub2[i]]) {
				builder.MarkEqual(sub1[i], sub2[i])
			}
		}
		return
	}

	if len(sub1) < len(sub2) {
		matching := getBestMatchingAlignment(sub1, sub2, lines1, lines2)
		for i := range len(sub1) {
			if lines1[sub1[i]].equals(lines2[sub2[matching[i]]]) {
				builder.MarkEqual(sub1[i], sub2[matching[i]])
			}
		}
	} else {
		matching := getBestMatchingAlignment(sub2, sub1, lines2, lines1)
		for i := range len(sub2) {
			if lines1[sub1[matching[i]]].equals(lines2[sub2[i]]) {
				builder.MarkEqual(sub1[matching[i]], sub2[i])
			}
		}
	}
}

func getBestMatchingAlignment(shorter, longer []int, linesS, linesL []*Line) []int {
	size := len(shorter)
	best := make([]int, size)
	for i := range size {
		best[i] = i
	}

	comb := make([]int, size)
	bestWeight := 0

	var combinations func(start, n, k int)
	combinations = func(start, n, k int) {
		if k == size {
			weight := 0
			for i := range size {
				if linesS[shorter[i]].equals(linesL[longer[comb[i]]]) {
					weight++
				}
			}
			if weight > bestWeight {
				bestWeight = weight
				copy(best, comb)
			}
			return
		}
		for i := start; i <= n; i++ {
			comb[k] = i
			combinations(i+1, n, k+1)
		}
	}
	combinations(0, len(longer)-1, 0)
	return best
}

// expandRanges is used for IGNORE_WHITESPACES policy: expand change ranges
// by trimming equal borders.
func expandRanges(lines1, lines2 []*Line, iterable FairDiffIterable) FairDiffIterable {
	var changes []util.Range
	for _, ch := range iterable.Changes() {
		expanded := expandRangeLines(lines1, lines2, ch.Start1, ch.Start2, ch.End1, ch.End2)
		if !expanded.IsEmpty() {
			changes = append(changes, expanded)
		}
	}
	return Fair(CreateFromRanges(changes, len(lines1), len(lines2)))
}

func toLines(text []string, policy ComparisonPolicy) []*Line {
	result := make([]*Line, len(text))
	for i, s := range text {
		result[i] = NewLine(s, policy)
	}
	return result
}

func convertMode(lines []*Line, policy ComparisonPolicy) []*Line {
	result := make([]*Line, len(lines))
	for i, l := range lines {
		if l.Policy != policy {
			result[i] = NewLine(l.Content, policy)
		} else {
			result[i] = l
		}
	}
	return result
}

// MyersMatcher implements LineMatcher using the Myers algorithm with the
// JetBrains two-step comparison approach (smart line comparison + chunk optimization).
type MyersMatcher struct{}

func (m *MyersMatcher) Match(left, right []string, policy ComparisonPolicy) FairDiffIterable {
	return CompareLines(left, right, policy)
}
