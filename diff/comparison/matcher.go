package comparison

// LineMatcher matches lines between two texts and produces a DiffIterable
// describing which lines are changed and which are unchanged.
//
// Implementations include Myers, Patience, and Histogram algorithms.
type LineMatcher interface {
	Match(left, right []string, policy ComparisonPolicy) FairDiffIterable
}
