// Copyright 2000-2021 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package util

import (
	"fmt"
	"sort"
	"strings"
)

// LineOffsets provides mappings between line numbers and byte offsets in text.
type LineOffsets struct {
	lineEnds   []int // each entry is the offset of the last char in the line (before '\n')
	textLength int
}

// NewLineOffsets creates LineOffsets by scanning text for newline positions.
func NewLineOffsets(text string) LineOffsets {
	var ends []int
	idx := 0
	for {
		lineEnd := strings.IndexByte(text[idx:], '\n')
		if lineEnd != -1 {
			ends = append(ends, idx+lineEnd)
			idx = idx + lineEnd + 1
		} else {
			ends = append(ends, len(text))
			break
		}
	}
	return LineOffsets{lineEnds: ends, textLength: len(text)}
}

func (lo LineOffsets) LineCount() int {
	return len(lo.lineEnds)
}

func (lo LineOffsets) TextLength() int {
	return lo.textLength
}

// LineStart returns the byte offset where the given line begins.
func (lo LineOffsets) LineStart(line int) int {
	lo.checkLineIndex(line)
	if line == 0 {
		return 0
	}
	return lo.lineEnds[line-1] + 1
}

// LineEnd returns the byte offset of the last character in the line (excluding '\n').
func (lo LineOffsets) LineEnd(line int) int {
	lo.checkLineIndex(line)
	return lo.lineEnds[line]
}

// LineEndWithNewline returns the byte offset past the line content,
// optionally including the trailing '\n'.
func (lo LineOffsets) LineEndWithNewline(line int, includeNewline bool) int {
	lo.checkLineIndex(line)
	end := lo.lineEnds[line]
	if includeNewline && line != len(lo.lineEnds)-1 {
		end++
	}
	return end
}

// LineNumber returns the line number that contains the given byte offset.
func (lo LineOffsets) LineNumber(offset int) int {
	if offset < 0 || offset > lo.textLength {
		panic(fmt.Sprintf("wrong offset: %d, text length: %d", offset, lo.textLength))
	}
	if offset == 0 {
		return 0
	}
	if offset == lo.textLength {
		return lo.LineCount() - 1
	}
	idx := sort.SearchInts(lo.lineEnds, offset)
	return idx
}

func (lo LineOffsets) checkLineIndex(line int) {
	if line < 0 || line >= lo.LineCount() {
		panic(fmt.Sprintf("wrong line: %d, line count: %d", line, lo.LineCount()))
	}
}
