// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import (
	"unicode/utf8"

	"github.com/byron1st/file-diff/diff/fragment"
	"github.com/byron1st/file-diff/diff/util"
)

// InlineChunk represents a tokenized piece of text (word or newline).
type InlineChunk interface {
	Offset1() int // start byte offset
	Offset2() int // end byte offset
}

// WordChunk is a word token within a text.
type WordChunk struct {
	text    string
	offset1 int
	offset2 int
	hash    int
}

func newWordChunk(text string, start, end, hash int) *WordChunk {
	return &WordChunk{text: text, offset1: start, offset2: end, hash: hash}
}

func (w *WordChunk) Offset1() int { return w.offset1 }
func (w *WordChunk) Offset2() int { return w.offset2 }

func (w *WordChunk) content() string {
	return w.text[w.offset1:w.offset2]
}

// chunkKey returns a comparable key for enumeration-based diffing.
func (w *WordChunk) chunkKey() chunkKey {
	return chunkKey{isWord: true, content: w.content(), hash: w.hash}
}

// NewlineChunk represents a newline character in the text.
type NewlineChunk struct {
	offset int
}

func newNewlineChunk(offset int) *NewlineChunk {
	return &NewlineChunk{offset: offset}
}

func (n *NewlineChunk) Offset1() int { return n.offset }
func (n *NewlineChunk) Offset2() int { return n.offset + 1 }

func (n *NewlineChunk) chunkKey() chunkKey {
	return chunkKey{isWord: false, content: "\n", hash: 0}
}

// chunkKey is a comparable key for InlineChunk enumeration.
type chunkKey struct {
	isWord  bool
	content string
	hash    int
}

// GetInlineChunks tokenizes text into word and newline chunks.
// Words are sequences of non-whitespace, non-punctuation, non-continuous-script characters.
// Continuous script characters are treated as individual single-character words.
func GetInlineChunks(text string) []InlineChunk {
	var chunks []InlineChunk

	wordStart := -1
	wordHash := 0
	offset := 0

	for offset < len(text) {
		r, size := utf8.DecodeRuneInString(text[offset:])

		isA := isAlpha(r)
		isWordPart := isA && !isContinuousScript(r)

		if isWordPart {
			if wordStart == -1 {
				wordStart = offset
				wordHash = 0
			}
			wordHash = wordHash*31 + int(r)
		} else {
			if wordStart != -1 {
				chunks = append(chunks, newWordChunk(text, wordStart, offset, wordHash))
				wordStart = -1
			}

			if isA { // continuous script character: single-char word
				chunks = append(chunks, newWordChunk(text, offset, offset+size, int(r)))
			} else if r == '\n' {
				chunks = append(chunks, newNewlineChunk(offset))
			}
		}

		offset += size
	}

	if wordStart != -1 {
		chunks = append(chunks, newWordChunk(text, wordStart, len(text), wordHash))
	}

	return chunks
}

// CompareWords compares two texts at the word level and returns DiffFragments
// representing the changed regions.
func CompareWords(text1, text2 string, policy ComparisonPolicy) ([]fragment.DiffFragment, error) {
	words1 := GetInlineChunks(text1)
	words2 := GetInlineChunks(text2)

	wordChanges, err := diffInlineChunks(words1, words2)
	if err != nil {
		return nil, err
	}

	wordChanges = optimizeWordChunks(text1, text2, words1, words2, wordChanges)

	delimitersIterable, err := matchAdjustmentDelimiters(text1, text2, words1, words2, wordChanges)
	if err != nil {
		return nil, err
	}

	iterable := matchAdjustmentWhitespaces(text1, text2, delimitersIterable, policy)

	return convertIntoDiffFragments(iterable), nil
}

// diffInlineChunks compares two slices of InlineChunks using Myers diff.
func diffInlineChunks(chunks1, chunks2 []InlineChunk) (FairDiffIterable, error) {
	keys1 := make([]chunkKey, len(chunks1))
	for i, c := range chunks1 {
		keys1[i] = getChunkKey(c)
	}
	keys2 := make([]chunkKey, len(chunks2))
	for i, c := range chunks2 {
		keys2[i] = getChunkKey(c)
	}
	return DiffObjects(keys1, keys2)
}

func getChunkKey(c InlineChunk) chunkKey {
	switch v := c.(type) {
	case *WordChunk:
		return v.chunkKey()
	case *NewlineChunk:
		return v.chunkKey()
	default:
		return chunkKey{}
	}
}

// optimizeWordChunks adjusts word chunk boundaries to prefer whitespace-separated positions.
func optimizeWordChunks(
	text1, text2 string,
	words1, words2 []InlineChunk,
	iterable FairDiffIterable,
) FairDiffIterable {
	var ranges []util.Range
	for _, r := range iterable.Unchanged() {
		ranges = append(ranges, r)
		processLastWordRanges(text1, text2, words1, words2, &ranges)
	}
	return Fair(CreateUnchanged(ranges, len(words1), len(words2)))
}

func processLastWordRanges(text1, text2 string, words1, words2 []InlineChunk, ranges *[]util.Range) {
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

	eqFwd := expandForwardChunks(words1, words2, r1.End1, r1.End2, r1.End1+count2, r1.End2+count2)
	eqBwd := expandBackwardChunks(words1, words2, r2.Start1-count1, r2.Start2-count1, r2.Start1, r2.Start2)

	if eqFwd == 0 && eqBwd == 0 {
		return
	}

	if eqFwd == count2 {
		*ranges = (*ranges)[:len(*ranges)-2]
		*ranges = append(*ranges, util.NewRange(r1.Start1, r1.End1+count2, r1.Start2, r1.End2+count2))
		processLastWordRanges(text1, text2, words1, words2, ranges)
		return
	}

	if eqBwd == count1 {
		*ranges = (*ranges)[:len(*ranges)-2]
		*ranges = append(*ranges, util.NewRange(r2.Start1-count1, r2.End1, r2.Start2-count1, r2.End2))
		processLastWordRanges(text1, text2, words1, words2, ranges)
		return
	}

	touchSideIsLeft := r1.End1 == r2.Start1
	shift := getWordShift(text1, text2, words1, words2, touchSideIsLeft, eqFwd, eqBwd, r1, r2)
	if shift != 0 {
		*ranges = (*ranges)[:len(*ranges)-2]
		*ranges = append(*ranges,
			util.NewRange(r1.Start1, r1.End1+shift, r1.Start2, r1.End2+shift),
			util.NewRange(r2.Start1+shift, r2.End1, r2.Start2+shift, r2.End2),
		)
	}
}

func getWordShift(text1, text2 string, words1, words2 []InlineChunk, touchSideIsLeft bool, eqFwd, eqBwd int, r1, r2 util.Range) int {
	var touchWords []InlineChunk
	var touchText string
	var touchStart int
	if touchSideIsLeft {
		touchWords = words1
		touchText = text1
		touchStart = r2.Start1
	} else {
		touchWords = words2
		touchText = text2
		touchStart = r2.Start2
	}

	// check if already separated by whitespace
	if isSeparatedWithWhitespace(touchText, touchWords[touchStart-1], touchWords[touchStart]) {
		return 0
	}

	// shift forward
	if s := findSequenceEdgeShift(touchText, touchWords, touchStart, eqFwd, true); s > 0 {
		return s
	}

	// shift backward
	if s := findSequenceEdgeShift(touchText, touchWords, touchStart-1, eqBwd, false); s > 0 {
		return -s
	}

	return 0
}

func findSequenceEdgeShift(text string, words []InlineChunk, offset, count int, leftToRight bool) int {
	for i := range count {
		var w1, w2 InlineChunk
		if leftToRight {
			w1 = words[offset+i]
			w2 = words[offset+i+1]
		} else {
			w1 = words[offset-i-1]
			w2 = words[offset-i]
		}
		if isSeparatedWithWhitespace(text, w1, w2) {
			return i + 1
		}
	}
	return -1
}

func isSeparatedWithWhitespace(text string, w1, w2 InlineChunk) bool {
	if _, ok := w1.(*NewlineChunk); ok {
		return true
	}
	if _, ok := w2.(*NewlineChunk); ok {
		return true
	}
	for i := w1.Offset2(); i < w2.Offset1(); i++ {
		if isSpaceEnterOrTab(rune(text[i])) {
			return true
		}
	}
	return false
}

func expandForwardChunks(words1, words2 []InlineChunk, start1, start2, end1, end2 int) int {
	s1, s2 := start1, start2
	for s1 < end1 && s2 < end2 && getChunkKey(words1[s1]) == getChunkKey(words2[s2]) {
		s1++
		s2++
	}
	return s1 - start1
}

func expandBackwardChunks(words1, words2 []InlineChunk, start1, start2, end1, end2 int) int {
	e1, e2 := end1, end2
	for start1 < e1 && start2 < e2 && getChunkKey(words1[e1-1]) == getChunkKey(words2[e2-1]) {
		e1--
		e2--
	}
	return end1 - e1
}

//
// Punctuation matching
//

// matchAdjustmentDelimiters matches punctuation characters in the gaps between matched words.
func matchAdjustmentDelimiters(
	text1, text2 string,
	words1, words2 []InlineChunk,
	changes FairDiffIterable,
) (FairDiffIterable, error) {
	m := &adjustmentPunctuationMatcher{
		text1: text1, text2: text2,
		words1: words1, words2: words2,
		changes: changes,
		builder: NewChangeBuilder(len(text1), len(text2)),
	}
	return m.build()
}

type adjustmentPunctuationMatcher struct {
	text1, text2   string
	words1, words2 []InlineChunk
	changes        FairDiffIterable
	builder        *ChangeBuilder

	lastStart1, lastStart2, lastEnd1, lastEnd2 int
}

func (m *adjustmentPunctuationMatcher) build() (FairDiffIterable, error) {
	m.clearLastRange()

	m.matchForwardIdx(-1, -1)

	for _, ch := range m.changes.Unchanged() {
		count := ch.End1 - ch.Start1
		for i := range count {
			idx1 := ch.Start1 + i
			idx2 := ch.Start2 + i

			start1 := m.words1[idx1].Offset1()
			start2 := m.words2[idx2].Offset1()
			end1 := m.words1[idx1].Offset2()
			end2 := m.words2[idx2].Offset2()

			if err := m.matchBackwardIdx(idx1, idx2); err != nil {
				return nil, err
			}

			m.builder.MarkEqualRange(start1, start2, end1, end2)

			m.matchForwardIdx(idx1, idx2)
		}
	}

	if err := m.matchBackwardIdx(len(m.words1), len(m.words2)); err != nil {
		return nil, err
	}

	return Fair(m.builder.Finish()), nil
}

func (m *adjustmentPunctuationMatcher) clearLastRange() {
	m.lastStart1 = -1
	m.lastStart2 = -1
	m.lastEnd1 = -1
	m.lastEnd2 = -1
}

func (m *adjustmentPunctuationMatcher) matchForwardIdx(idx1, idx2 int) {
	var start1, start2, end1, end2 int
	if idx1 == -1 {
		start1 = 0
	} else {
		start1 = m.words1[idx1].Offset2()
	}
	if idx2 == -1 {
		start2 = 0
	} else {
		start2 = m.words2[idx2].Offset2()
	}
	if idx1+1 == len(m.words1) {
		end1 = len(m.text1)
	} else {
		end1 = m.words1[idx1+1].Offset1()
	}
	if idx2+1 == len(m.words2) {
		end2 = len(m.text2)
	} else {
		end2 = m.words2[idx2+1].Offset1()
	}

	m.lastStart1 = start1
	m.lastStart2 = start2
	m.lastEnd1 = end1
	m.lastEnd2 = end2
}

func (m *adjustmentPunctuationMatcher) matchBackwardIdx(idx1, idx2 int) error {
	var start1, start2, end1, end2 int
	if idx1 == 0 {
		start1 = 0
	} else {
		start1 = m.words1[idx1-1].Offset2()
	}
	if idx2 == 0 {
		start2 = 0
	} else {
		start2 = m.words2[idx2-1].Offset2()
	}
	if idx1 == len(m.words1) {
		end1 = len(m.text1)
	} else {
		end1 = m.words1[idx1].Offset1()
	}
	if idx2 == len(m.words2) {
		end2 = len(m.text2)
	} else {
		end2 = m.words2[idx2].Offset1()
	}

	if err := m.matchBackwardRange(start1, start2, end1, end2); err != nil {
		return err
	}
	m.clearLastRange()
	return nil
}

func (m *adjustmentPunctuationMatcher) matchBackwardRange(start1, start2, end1, end2 int) error {
	if m.lastStart1 == start1 && m.lastStart2 == start2 {
		// same gap - match it
		return m.matchRange(start1, start2, end1, end2)
	}
	if m.lastStart1 < start1 && m.lastStart2 < start2 {
		// two separate gaps
		if err := m.matchRange(m.lastStart1, m.lastStart2, m.lastEnd1, m.lastEnd2); err != nil {
			return err
		}
		return m.matchRange(start1, start2, end1, end2)
	}
	// complex: one side has an adjacent matched word, the other has unmatched words
	return m.matchRange(min(m.lastStart1, start1), min(m.lastStart2, start2),
		max(m.lastEnd1, end1), max(m.lastEnd2, end2))
}

func (m *adjustmentPunctuationMatcher) matchRange(start1, start2, end1, end2 int) error {
	if start1 == end1 && start2 == end2 {
		return nil
	}

	seq1 := m.text1[start1:end1]
	seq2 := m.text2[start2:end2]

	changes, err := ComparePunctuation(seq1, seq2)
	if err != nil {
		return err
	}

	for _, ch := range changes.Unchanged() {
		m.builder.MarkEqualRange(start1+ch.Start1, start2+ch.Start2, start1+ch.End1, start2+ch.End2)
	}
	return nil
}

//
// Whitespace matching
//

func matchAdjustmentWhitespaces(text1, text2 string, iterable FairDiffIterable, policy ComparisonPolicy) DiffIterable {
	switch policy {
	case PolicyTrimWhitespaces:
		defaultIterable := newDefaultCorrector(iterable, text1, text2).build()
		return newTrimSpacesCorrector(defaultIterable, text1, text2).build()
	case PolicyIgnoreWhitespaces:
		return newIgnoreSpacesCorrector(iterable, text1, text2).build()
	default: // PolicyDefault
		return newDefaultCorrector(iterable, text1, text2).build()
	}
}

// defaultCorrector expands change boundaries by consuming matching whitespace.
type defaultCorrector struct {
	iterable     DiffIterable
	text1, text2 string
}

func newDefaultCorrector(iterable DiffIterable, text1, text2 string) *defaultCorrector {
	return &defaultCorrector{iterable: iterable, text1: text1, text2: text2}
}

func (c *defaultCorrector) build() DiffIterable {
	var changes []util.Range
	for _, r := range c.iterable.Changes() {
		endCut := expandWhitespacesBackward(c.text1, c.text2, r.Start1, r.Start2, r.End1, r.End2)
		startCut := expandWhitespacesForward(c.text1, c.text2, r.Start1, r.Start2, r.End1-endCut, r.End2-endCut)

		expanded := util.NewRange(r.Start1+startCut, r.End1-endCut, r.Start2+startCut, r.End2-endCut)
		if !expanded.IsEmpty() {
			changes = append(changes, expanded)
		}
	}
	return CreateFromRanges(changes, len(c.text1), len(c.text2))
}

// trimSpacesCorrector trims leading/trailing whitespace from change boundaries.
type trimSpacesCorrector struct {
	iterable     DiffIterable
	text1, text2 string
}

func newTrimSpacesCorrector(iterable DiffIterable, text1, text2 string) *trimSpacesCorrector {
	return &trimSpacesCorrector{iterable: iterable, text1: text1, text2: text2}
}

func (c *trimSpacesCorrector) build() DiffIterable {
	var changes []util.Range
	for _, r := range c.iterable.Changes() {
		start1, start2, end1, end2 := r.Start1, r.Start2, r.End1, r.End2

		if isLeadingTrailingSpace(c.text1, start1) {
			start1 = trimStartText(c.text1, start1, end1)
		}
		if isLeadingTrailingSpace(c.text1, end1-1) {
			end1 = trimEndText(c.text1, start1, end1)
		}
		if isLeadingTrailingSpace(c.text2, start2) {
			start2 = trimStartText(c.text2, start2, end2)
		}
		if isLeadingTrailingSpace(c.text2, end2-1) {
			end2 = trimEndText(c.text2, start2, end2)
		}

		trimmed := util.NewRange(start1, end1, start2, end2)
		if !trimmed.IsEmpty() && !isEqualTextRange(c.text1, c.text2, trimmed) {
			changes = append(changes, trimmed)
		}
	}
	return CreateFromRanges(changes, len(c.text1), len(c.text2))
}

// ignoreSpacesCorrector ignores whitespace-only differences.
type ignoreSpacesCorrector struct {
	iterable     DiffIterable
	text1, text2 string
}

func newIgnoreSpacesCorrector(iterable DiffIterable, text1, text2 string) *ignoreSpacesCorrector {
	return &ignoreSpacesCorrector{iterable: iterable, text1: text1, text2: text2}
}

func (c *ignoreSpacesCorrector) build() DiffIterable {
	var changes []util.Range
	for _, r := range c.iterable.Changes() {
		expanded := expandWhitespacesRange(c.text1, c.text2, r)
		trimmed := trimTextRange(c.text1, c.text2, expanded.Start1, expanded.Start2, expanded.End1, expanded.End2)

		if !trimmed.IsEmpty() && !isEqualTextRangeIgnoreWhitespaces(c.text1, c.text2, trimmed) {
			changes = append(changes, trimmed)
		}
	}
	return CreateFromRanges(changes, len(c.text1), len(c.text2))
}

//
// Output conversion
//

func convertIntoDiffFragments(changes DiffIterable) []fragment.DiffFragment {
	var fragments []fragment.DiffFragment
	for _, ch := range changes.Changes() {
		fragments = append(fragments, fragment.NewDiffFragment(ch.Start1, ch.End1, ch.Start2, ch.End2))
	}
	return fragments
}
