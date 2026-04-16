// Copyright 2000-2021 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

// ComparisonPolicy defines how whitespace is handled during comparison.
type ComparisonPolicy int

const (
	PolicyDefault           ComparisonPolicy = iota // Compare as-is
	PolicyTrimWhitespaces                           // Trim leading/trailing whitespace per line
	PolicyIgnoreWhitespaces                         // Ignore all whitespace differences
)

func (p ComparisonPolicy) String() string {
	switch p {
	case PolicyDefault:
		return "DEFAULT"
	case PolicyTrimWhitespaces:
		return "TRIM_WHITESPACES"
	case PolicyIgnoreWhitespaces:
		return "IGNORE_WHITESPACES"
	default:
		return "UNKNOWN"
	}
}
