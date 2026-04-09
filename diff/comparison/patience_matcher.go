package comparison

import (
	"strings"
	"unicode"

	"github.com/byron1st/file-diff/diff/patience"
)

// PatienceMatcher implements LineMatcher using the Patience Diff algorithm.
// It anchors on unique common lines via LIS, then falls back to Myers
// for regions without unique lines.
type PatienceMatcher struct{}

func (m *PatienceMatcher) Match(left, right []string, policy ComparisonPolicy) FairDiffIterable {
	normLeft := normalizeLines(left, policy)
	normRight := normalizeLines(right, policy)

	matches := patience.Diff(normLeft, normRight, func(l, r []string) []patience.Anchor {
		return myersFallback(l, r)
	})

	builder := NewChangeBuilder(len(left), len(right))
	for _, match := range matches {
		builder.MarkEqual(match.LeftIdx, match.RightIdx)
	}
	return Fair(builder.Finish())
}

// normalizeLines returns line contents normalized by policy for comparison.
func normalizeLines(lines []string, policy ComparisonPolicy) []string {
	if policy == PolicyDefault {
		return lines
	}
	result := make([]string, len(lines))
	for i, l := range lines {
		result[i] = normalizedContent(l, policy)
	}
	return result
}

func normalizedContent(s string, policy ComparisonPolicy) string {
	switch policy {
	case PolicyTrimWhitespaces:
		return strings.TrimSpace(s)
	case PolicyIgnoreWhitespaces:
		var b strings.Builder
		for _, c := range s {
			if !unicode.IsSpace(c) {
				b.WriteRune(c)
			}
		}
		return b.String()
	default:
		return s
	}
}

// myersFallback uses CompareLines (Myers) to produce matched pairs.
func myersFallback(left, right []string) []patience.Anchor {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}
	result := CompareLines(left, right, PolicyDefault)
	var matches []patience.Anchor
	for _, u := range result.Unchanged() {
		count := u.End1 - u.Start1
		for i := range count {
			matches = append(matches, patience.Anchor{LeftIdx: u.Start1 + i, RightIdx: u.Start2 + i})
		}
	}
	return matches
}
