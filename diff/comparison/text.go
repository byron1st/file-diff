// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import (
	"strings"
	"unicode"
)

// IsEqual compares two strings according to the given ComparisonPolicy.
func IsEqual(s1, s2 string, policy ComparisonPolicy) bool {
	switch policy {
	case PolicyTrimWhitespaces:
		return equalsTrimWhitespaces(s1, s2)
	case PolicyIgnoreWhitespaces:
		return equalsIgnoreWhitespaces(s1, s2)
	default:
		return s1 == s2
	}
}

// HashCode computes a hash for a string according to the given ComparisonPolicy.
func HashCode(s string, policy ComparisonPolicy) int {
	switch policy {
	case PolicyTrimWhitespaces:
		return stringHashCode(strings.TrimSpace(s))
	case PolicyIgnoreWhitespaces:
		return stringHashCodeIgnoreWhitespaces(s)
	default:
		return stringHashCode(s)
	}
}

func equalsTrimWhitespaces(s1, s2 string) bool {
	return strings.TrimSpace(s1) == strings.TrimSpace(s2)
}

func equalsIgnoreWhitespaces(s1, s2 string) bool {
	i, j := 0, 0
	for i < len(s1) && j < len(s2) {
		c1 := rune(s1[i])
		c2 := rune(s2[j])
		if unicode.IsSpace(c1) {
			i++
			continue
		}
		if unicode.IsSpace(c2) {
			j++
			continue
		}
		if c1 != c2 {
			return false
		}
		i++
		j++
	}
	for i < len(s1) {
		if !unicode.IsSpace(rune(s1[i])) {
			return false
		}
		i++
	}
	for j < len(s2) {
		if !unicode.IsSpace(rune(s2[j])) {
			return false
		}
		j++
	}
	return true
}

// stringHashCode computes a simple hash (Java-compatible algorithm) for a string.
func stringHashCode(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	return h
}

// stringHashCodeIgnoreWhitespaces computes a hash ignoring whitespace characters.
func stringHashCodeIgnoreWhitespaces(s string) int {
	h := 0
	for _, c := range s {
		if !unicode.IsSpace(c) {
			h = 31*h + int(c)
		}
	}
	return h
}
