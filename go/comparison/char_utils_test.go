// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import "testing"

func TestIsPunctuation(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'{', true},
		{'}', true},
		{'(', true},
		{')', true},
		{'[', true},
		{']', true},
		{';', true},
		{'.', true},
		{',', true},
		{'!', true},
		{'@', true},
		{'#', true},
		{'$', true},
		{'%', true},
		{'+', true},
		{'-', true},
		{'/', true},
		{'*', true},
		{'=', true},
		{'<', true},
		{'>', true},
		{'~', true},
		// Not punctuation
		{'_', false}, // underscore is excluded
		{'a', false},
		{'Z', false},
		{'0', false},
		{' ', false},
		{'\n', false},
	}
	for _, tt := range tests {
		if got := isPunctuation(tt.r); got != tt.want {
			t.Errorf("isPunctuation(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestIsAlpha(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'_', true},
		// Not alpha
		{' ', false},
		{'\n', false},
		{'\t', false},
		{'{', false},
		{';', false},
		{'.', false},
	}
	for _, tt := range tests {
		if got := isAlpha(tt.r); got != tt.want {
			t.Errorf("isAlpha(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestIsContinuousScript(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'a', false},    // ASCII
		{'0', false},    // digit
		{' ', false},    // space
		{'中', true},     // Han
		{'の', true},     // Hiragana
		{'ア', true},     // Katakana
		{'é', false},    // Latin Extended (letter, not continuous)
		{'Ω', false},    // Greek (letter, not continuous)
		{0x0E01, true},  // Thai
		{0x10000, true}, // non-BMP
	}
	for _, tt := range tests {
		if got := isContinuousScript(tt.r); got != tt.want {
			t.Errorf("isContinuousScript(%U) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestIsWhiteSpaceCodePoint(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{' ', true},
		{'\n', true},
		{'\t', true},
		{'a', false},
		{0x00A0, false}, // non-breaking space is >= 128
	}
	for _, tt := range tests {
		if got := isWhiteSpaceCodePoint(tt.r); got != tt.want {
			t.Errorf("isWhiteSpaceCodePoint(%U) = %v, want %v", tt.r, got, tt.want)
		}
	}
}
