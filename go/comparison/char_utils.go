// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import "unicode"

// isSpaceEnterOrTab returns true if c is a space, newline, or tab.
func isSpaceEnterOrTab(c rune) bool {
	return c == ' ' || c == '\n' || c == '\t'
}

// isWhiteSpaceCodePoint returns true if c is an ASCII whitespace character.
func isWhiteSpaceCodePoint(c rune) bool {
	return c < 128 && isSpaceEnterOrTab(c)
}

// isPunctuation returns true if c is an ASCII punctuation character (excluding '_').
func isPunctuation(c rune) bool {
	b := int(c)
	if b == 95 { // exclude '_'
		return false
	}
	return (b >= 33 && b <= 47) || // !"#$%&'()*+,-./
		(b >= 58 && b <= 64) || // :;<=>?@
		(b >= 91 && b <= 96) || // [\]^_`
		(b >= 123 && b <= 126) // {|}~
}

// isAlpha returns true if c is neither whitespace nor punctuation.
func isAlpha(c rune) bool {
	if isWhiteSpaceCodePoint(c) {
		return false
	}
	return !isPunctuation(c)
}

// isContinuousScript checks if a rune belongs to a script where
// word boundaries cannot be determined by spaces (e.g. CJK, Thai).
// Characters from these scripts are treated as single-character words.
func isContinuousScript(c rune) bool {
	if c < 128 {
		return false
	}
	if unicode.IsDigit(c) {
		return false
	}
	if c >= 0x10000 { // non-BMP
		return true
	}
	if unicode.Is(unicode.Han, c) {
		return true
	}
	if !unicode.IsLetter(c) {
		return true
	}
	return unicode.Is(unicode.Hiragana, c) ||
		unicode.Is(unicode.Katakana, c) ||
		unicode.Is(unicode.Thai, c) ||
		unicode.Is(unicode.Javanese, c)
}
