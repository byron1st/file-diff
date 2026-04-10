package comparison

import (
	"github.com/byron1st/file-diff/diff/histogram"
)

// HistogramMatcher implements LineMatcher using the Histogram Diff algorithm.
// It extends Patience Diff by using line occurrence frequency to select
// anchors, handling repetitive structures (JSON, YAML, boilerplate) where
// Patience falls back to Myers due to lack of unique lines.
type HistogramMatcher struct{}

func (m *HistogramMatcher) Match(left, right []string, policy ComparisonPolicy) FairDiffIterable {
	normLeft := normalizeLines(left, policy)
	normRight := normalizeLines(right, policy)

	matches := histogram.Diff(normLeft, normRight, func(l, r []string) []histogram.Anchor {
		return histogramMyersFallback(l, r)
	})

	builder := NewChangeBuilder(len(left), len(right))
	for _, match := range matches {
		builder.MarkEqual(match.LeftIdx, match.RightIdx)
	}
	return Fair(builder.Finish())
}

// histogramMyersFallback uses CompareLines (Myers) to produce matched pairs.
func histogramMyersFallback(left, right []string) []histogram.Anchor {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}
	result := CompareLines(left, right, PolicyDefault)
	var matches []histogram.Anchor
	for _, u := range result.Unchanged() {
		count := u.End1 - u.Start1
		for i := range count {
			matches = append(matches, histogram.Anchor{LeftIdx: u.Start1 + i, RightIdx: u.Start2 + i})
		}
	}
	return matches
}
