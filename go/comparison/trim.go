// Copyright 2000-2021 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import "github.com/byron1st/file-diff/go/util"

// trimTextRange trims whitespace from both ends of the ranges within text1 and text2.
func trimTextRange(text1, text2 string, start1, start2, end1, end2 int) util.Range {
	s1 := trimStartText(text1, start1, end1)
	e1 := trimEndText(text1, s1, end1)
	s2 := trimStartText(text2, start2, end2)
	e2 := trimEndText(text2, s2, end2)
	return util.NewRange(s1, e1, s2, e2)
}

// trimStartText advances start past whitespace characters.
func trimStartText(text string, start, end int) int {
	for start < end {
		if !isSpaceEnterOrTab(rune(text[start])) {
			break
		}
		start++
	}
	return start
}

// trimEndText retracts end past whitespace characters.
func trimEndText(text string, start, end int) int {
	for start < end {
		if !isSpaceEnterOrTab(rune(text[end-1])) {
			break
		}
		end--
	}
	return end
}

// expandWhitespacesForward counts how many leading characters are equal AND
// are whitespace between text1[start1..end1) and text2[start2..end2).
func expandWhitespacesForward(text1, text2 string, start1, start2, end1, end2 int) int {
	s1, s2 := start1, start2
	for s1 < end1 && s2 < end2 {
		if text1[s1] != text2[s2] {
			break
		}
		if !isSpaceEnterOrTab(rune(text1[s1])) {
			break
		}
		s1++
		s2++
	}
	return s1 - start1
}

// expandWhitespacesBackward counts how many trailing characters are equal AND
// are whitespace between text1[start1..end1) and text2[start2..end2).
func expandWhitespacesBackward(text1, text2 string, start1, start2, end1, end2 int) int {
	e1, e2 := end1, end2
	for start1 < e1 && start2 < e2 {
		if text1[e1-1] != text2[e2-1] {
			break
		}
		if !isSpaceEnterOrTab(rune(text1[e1-1])) {
			break
		}
		e1--
		e2--
	}
	return end1 - e1
}

// expandWhitespacesRange expands a change range by consuming matching whitespace at both ends.
func expandWhitespacesRange(text1, text2 string, r util.Range) util.Range {
	s1, s2 := r.Start1, r.Start2
	e1, e2 := r.End1, r.End2

	fwd := expandWhitespacesForward(text1, text2, s1, s2, e1, e2)
	s1 += fwd
	s2 += fwd

	bwd := expandWhitespacesBackward(text1, text2, s1, s2, e1, e2)
	e1 -= bwd
	e2 -= bwd

	return util.NewRange(s1, e1, s2, e2)
}

// isEqualTextRange checks if text1[r.Start1..r.End1) == text2[r.Start2..r.End2).
func isEqualTextRange(text1, text2 string, r util.Range) bool {
	s1 := text1[r.Start1:r.End1]
	s2 := text2[r.Start2:r.End2]
	return s1 == s2
}

// isEqualTextRangeIgnoreWhitespaces checks equality ignoring whitespace.
func isEqualTextRangeIgnoreWhitespaces(text1, text2 string, r util.Range) bool {
	s1 := text1[r.Start1:r.End1]
	s2 := text2[r.Start2:r.End2]
	return equalsIgnoreWhitespaces(s1, s2)
}

// isLeadingTrailingSpace returns true if the character at pos is a whitespace
// that is part of leading or trailing whitespace on its line.
func isLeadingTrailingSpace(text string, pos int) bool {
	return isLeadingSpace(text, pos) || isTrailingSpace(text, pos)
}

func isLeadingSpace(text string, pos int) bool {
	if pos < 0 || pos >= len(text) {
		return false
	}
	if !isSpaceEnterOrTab(rune(text[pos])) {
		return false
	}
	for i := pos - 1; i >= 0; i-- {
		c := rune(text[i])
		if c == '\n' {
			return true
		}
		if !isSpaceEnterOrTab(c) {
			return false
		}
	}
	return true
}

func isTrailingSpace(text string, pos int) bool {
	if pos < 0 || pos >= len(text) {
		return false
	}
	if !isSpaceEnterOrTab(rune(text[pos])) {
		return false
	}
	for i := pos; i < len(text); i++ {
		c := rune(text[i])
		if c == '\n' {
			return true
		}
		if !isSpaceEnterOrTab(c) {
			return false
		}
	}
	return true
}
