// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import (
	"unicode/utf8"

	"github.com/byron1st/file-diff/go/util"
)

// codePointsOffsets stores non-space (or punctuation) code points along with
// their byte offsets in the original string.
type codePointsOffsets struct {
	codePoints []int
	offsets    []int // byte offset of each code point in the original text
}

// charOffset returns the byte offset of the code point at index.
func (c *codePointsOffsets) charOffset(index int) int {
	return c.offsets[index]
}

// charOffsetAfter returns the byte offset after the code point at index.
func (c *codePointsOffsets) charOffsetAfter(index int) int {
	return c.offsets[index] + utf8.RuneLen(rune(c.codePoints[index]))
}

// CompareChars compares two strings at the character (code point) level.
func CompareChars(text1, text2 string) (FairDiffIterable, error) {
	cp1 := getAllCodePoints(text1)
	cp2 := getAllCodePoints(text2)

	iterable, err := Diff(cp1.codePoints, cp2.codePoints)
	if err != nil {
		return nil, err
	}

	builder := NewChangeBuilder(len(text1), len(text2))
	offset1, offset2 := 0, 0

	for _, r := range iterateAll(iterable) {
		end1 := offset1 + countCharBytes(cp1, r.rng.Start1, r.rng.End1)
		end2 := offset2 + countCharBytes(cp2, r.rng.Start2, r.rng.End2)
		if r.equal {
			builder.MarkEqualRange(offset1, offset2, end1, end2)
		}
		offset1 = end1
		offset2 = end2
	}

	return Fair(builder.Finish()), nil
}

// CompareCharsTwoStep compares two strings at the character level using a
// two-step approach: first match non-space characters, then fill gaps.
func CompareCharsTwoStep(text1, text2 string) (FairDiffIterable, error) {
	cp1 := getNonSpaceCodePoints(text1)
	cp2 := getNonSpaceCodePoints(text2)

	nonSpaceChanges, err := Diff(cp1.codePoints, cp2.codePoints)
	if err != nil {
		return nil, err
	}

	return matchAdjustmentSpaces(cp1, cp2, text1, text2, nonSpaceChanges)
}

// CompareCharsTrimWhitespaces compares two strings at the character level,
// trimming whitespace from change boundaries.
func CompareCharsTrimWhitespaces(text1, text2 string) (DiffIterable, error) {
	iterable, err := CompareCharsTwoStep(text1, text2)
	if err != nil {
		return nil, err
	}
	return newTrimSpacesCorrector(iterable, text1, text2).build(), nil
}

// CompareCharsIgnoreWhitespaces compares two strings ignoring whitespace.
func CompareCharsIgnoreWhitespaces(text1, text2 string) (DiffIterable, error) {
	cp1 := getNonSpaceCodePoints(text1)
	cp2 := getNonSpaceCodePoints(text2)

	changes, err := Diff(cp1.codePoints, cp2.codePoints)
	if err != nil {
		return nil, err
	}

	return matchAdjustmentSpacesIW(cp1, cp2, text1, text2, changes), nil
}

// ComparePunctuation compares only punctuation characters between two texts.
// All other characters are left unmatched.
func ComparePunctuation(text1, text2 string) (FairDiffIterable, error) {
	chars1 := getPunctuationChars(text1)
	chars2 := getPunctuationChars(text2)

	nonSpaceChanges, err := Diff(chars1.codePoints, chars2.codePoints)
	if err != nil {
		return nil, err
	}

	return transferPunctuation(chars1, chars2, text1, text2, nonSpaceChanges), nil
}

// matchAdjustmentSpaces converts a diff on non-space characters into a diff
// on the original texts by running a fair diff on gaps between matched characters.
func matchAdjustmentSpaces(
	cp1, cp2 *codePointsOffsets,
	text1, text2 string,
	changes FairDiffIterable,
) (FairDiffIterable, error) {
	return newDefaultCharChangeCorrector(cp1, cp2, text1, text2, changes).build()
}

// matchAdjustmentSpacesIW converts a diff on non-whitespace characters into a diff
// on the original texts. Matched characters include matched non-space characters
// plus all adjacent whitespace.
func matchAdjustmentSpacesIW(
	cp1, cp2 *codePointsOffsets,
	text1, text2 string,
	changes FairDiffIterable,
) DiffIterable {
	var ranges []util.Range

	for _, ch := range changes.Changes() {
		var startOffset1, endOffset1 int
		if ch.Start1 == ch.End1 {
			v := expandForwardW(cp1, cp2, text1, text2, ch, true)
			startOffset1, endOffset1 = v, v
		} else {
			startOffset1 = cp1.charOffset(ch.Start1)
			endOffset1 = cp1.charOffsetAfter(ch.End1 - 1)
		}

		var startOffset2, endOffset2 int
		if ch.Start2 == ch.End2 {
			v := expandForwardW(cp1, cp2, text1, text2, ch, false)
			startOffset2, endOffset2 = v, v
		} else {
			startOffset2 = cp2.charOffset(ch.Start2)
			endOffset2 = cp2.charOffsetAfter(ch.End2 - 1)
		}

		ranges = append(ranges, util.NewRange(startOffset1, endOffset1, startOffset2, endOffset2))
	}

	return CreateFromRanges(ranges, len(text1), len(text2))
}

// expandForwardW adjusts insertion/deletion placement to prefer matching whitespace.
func expandForwardW(
	cp1, cp2 *codePointsOffsets,
	text1, text2 string,
	ch util.Range,
	left bool,
) int {
	var offset1, offset2 int
	if ch.Start1 == 0 {
		offset1 = 0
	} else {
		offset1 = cp1.charOffsetAfter(ch.Start1 - 1)
	}
	if ch.Start2 == 0 {
		offset2 = 0
	} else {
		offset2 = cp2.charOffsetAfter(ch.Start2 - 1)
	}

	start := offset1
	if !left {
		start = offset2
	}

	return start + expandWhitespacesForward(text1, text2, offset1, offset2, len(text1), len(text2))
}

func transferPunctuation(
	chars1, chars2 *codePointsOffsets,
	text1, text2 string,
	changes FairDiffIterable,
) FairDiffIterable {
	builder := NewChangeBuilder(len(text1), len(text2))

	for _, r := range changes.Unchanged() {
		count := r.End1 - r.Start1
		for i := range count {
			// punctuation code points are always 1 byte in ASCII
			o1 := chars1.offsets[r.Start1+i]
			o2 := chars2.offsets[r.Start2+i]
			builder.MarkEqual(o1, o2)
		}
	}

	return Fair(builder.Finish())
}

// defaultCharChangeCorrector fills gaps between matched non-space code points
// by running a full character comparison on each gap.
type defaultCharChangeCorrector struct {
	cp1, cp2     *codePointsOffsets
	text1, text2 string
	changes      FairDiffIterable
}

func newDefaultCharChangeCorrector(
	cp1, cp2 *codePointsOffsets,
	text1, text2 string,
	changes FairDiffIterable,
) *defaultCharChangeCorrector {
	return &defaultCharChangeCorrector{cp1: cp1, cp2: cp2, text1: text1, text2: text2, changes: changes}
}

func (c *defaultCharChangeCorrector) build() (FairDiffIterable, error) {
	builder := NewChangeBuilder(len(c.text1), len(c.text2))
	last1, last2 := 0, 0

	for _, ch := range c.changes.Unchanged() {
		count := ch.End1 - ch.Start1
		for i := range count {
			start1 := c.cp1.charOffset(ch.Start1 + i)
			start2 := c.cp2.charOffset(ch.Start2 + i)
			end1 := c.cp1.charOffsetAfter(ch.Start1 + i)
			end2 := c.cp2.charOffsetAfter(ch.Start2 + i)

			if err := c.matchGap(builder, last1, start1, last2, start2); err != nil {
				return nil, err
			}
			builder.MarkEqualRange(start1, start2, end1, end2)

			last1 = end1
			last2 = end2
		}
	}
	if err := c.matchGap(builder, last1, len(c.text1), last2, len(c.text2)); err != nil {
		return nil, err
	}

	return Fair(builder.Finish()), nil
}

func (c *defaultCharChangeCorrector) matchGap(builder *ChangeBuilder, start1, end1, start2, end2 int) error {
	inner1 := c.text1[start1:end1]
	inner2 := c.text2[start2:end2]
	if len(inner1) == 0 && len(inner2) == 0 {
		return nil
	}

	innerChanges, err := CompareChars(inner1, inner2)
	if err != nil {
		return err
	}

	for _, ch := range innerChanges.Unchanged() {
		builder.MarkEqualCount(start1+ch.Start1, start2+ch.Start2, ch.End1-ch.Start1)
	}
	return nil
}

// iterateAllEntry represents a range that is either equal or changed.
type iterateAllEntry struct {
	rng   util.Range
	equal bool
}

// iterateAll yields all ranges (both changed and unchanged) in order.
func iterateAll(iterable DiffIterable) []iterateAllEntry {
	var result []iterateAllEntry
	changes := iterable.Changes()
	unchanged := iterable.Unchanged()

	ci, ui := 0, 0
	for ci < len(changes) || ui < len(unchanged) {
		takeChange := ci < len(changes) && (ui >= len(unchanged) ||
			changes[ci].Start1 < unchanged[ui].Start1 ||
			(changes[ci].Start1 == unchanged[ui].Start1 && changes[ci].Start2 < unchanged[ui].Start2))
		if takeChange {
			result = append(result, iterateAllEntry{rng: changes[ci], equal: false})
			ci++
		} else {
			result = append(result, iterateAllEntry{rng: unchanged[ui], equal: true})
			ui++
		}
	}
	return result
}

// getAllCodePoints extracts all code points from text along with byte offsets.
func getAllCodePoints(text string) *codePointsOffsets {
	var cps []int
	var offsets []int
	for offset, r := range text {
		cps = append(cps, int(r))
		offsets = append(offsets, offset)
	}
	return &codePointsOffsets{codePoints: cps, offsets: offsets}
}

// getNonSpaceCodePoints extracts non-whitespace code points with byte offsets.
func getNonSpaceCodePoints(text string) *codePointsOffsets {
	var cps []int
	var offsets []int
	for offset, r := range text {
		if !isWhiteSpaceCodePoint(r) {
			cps = append(cps, int(r))
			offsets = append(offsets, offset)
		}
	}
	return &codePointsOffsets{codePoints: cps, offsets: offsets}
}

// getPunctuationChars extracts punctuation characters with byte offsets.
func getPunctuationChars(text string) *codePointsOffsets {
	var cps []int
	var offsets []int
	for i, r := range text {
		if isPunctuation(r) {
			cps = append(cps, int(r))
			offsets = append(offsets, i)
		}
	}
	return &codePointsOffsets{codePoints: cps, offsets: offsets}
}

// countCharBytes counts total byte length of code points in range [start, end).
func countCharBytes(cp *codePointsOffsets, start, end int) int {
	total := 0
	for i := start; i < end; i++ {
		total += utf8.RuneLen(rune(cp.codePoints[i]))
	}
	return total
}
