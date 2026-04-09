// Patience Diff algorithm implementation.
//
// The algorithm finds lines that appear exactly once in each side,
// computes the Longest Increasing Subsequence (LIS) of their positions
// to establish anchors, then recursively diffs regions between anchors.
// Regions with no unique lines fall back to the provided fallback differ.

package patience

// Anchor represents a unique line matched between left and right.
type Anchor struct {
	LeftIdx  int
	RightIdx int
}

// DiffFunc computes diff ranges for sub-regions. Returns a list of
// (LeftIdx, RightIdx) pairs representing equal lines.
type DiffFunc func(left, right []string) []Anchor

// Diff runs the Patience Diff algorithm on left and right line slices.
// fallback is used for regions with no unique common lines.
// Returns a list of matched (equal) line pairs as anchors.
func Diff(left, right []string, fallback DiffFunc) []Anchor {
	var result []Anchor
	diffRecursive(left, right, 0, 0, fallback, &result)
	return result
}

func diffRecursive(left, right []string, offsetL, offsetR int, fallback DiffFunc, result *[]Anchor) {
	if len(left) == 0 || len(right) == 0 {
		return
	}

	anchors := findAnchors(left, right)

	if len(anchors) == 0 {
		// No unique common lines — use fallback (Myers)
		matches := fallback(left, right)
		for _, m := range matches {
			*result = append(*result, Anchor{LeftIdx: m.LeftIdx + offsetL, RightIdx: m.RightIdx + offsetR})
		}
		return
	}

	// Process regions between anchors
	prevL, prevR := 0, 0
	for _, a := range anchors {
		if a.LeftIdx > prevL || a.RightIdx > prevR {
			diffRecursive(
				left[prevL:a.LeftIdx], right[prevR:a.RightIdx],
				offsetL+prevL, offsetR+prevR,
				fallback, result,
			)
		}
		*result = append(*result, Anchor{LeftIdx: a.LeftIdx + offsetL, RightIdx: a.RightIdx + offsetR})
		prevL = a.LeftIdx + 1
		prevR = a.RightIdx + 1
	}

	// Diff the region after the last anchor
	if prevL < len(left) || prevR < len(right) {
		diffRecursive(
			left[prevL:], right[prevR:],
			offsetL+prevL, offsetR+prevR,
			fallback, result,
		)
	}
}

// findAnchors finds lines unique in both sides, matches them, and returns
// the LIS of their right-side indices as anchors.
func findAnchors(left, right []string) []Anchor {
	leftCount := make(map[string]int)
	for _, l := range left {
		leftCount[l]++
	}

	rightCount := make(map[string]int)
	for _, r := range right {
		rightCount[r]++
	}

	// Build index of unique lines in right
	rightIndex := make(map[string]int)
	for i, r := range right {
		if rightCount[r] == 1 {
			rightIndex[r] = i
		}
	}

	// Collect matching pairs: lines unique in both sides
	var pairs []Anchor
	for i, l := range left {
		if leftCount[l] == 1 {
			if rIdx, ok := rightIndex[l]; ok && rightCount[l] == 1 {
				pairs = append(pairs, Anchor{LeftIdx: i, RightIdx: rIdx})
			}
		}
	}

	if len(pairs) == 0 {
		return nil
	}

	return lis(pairs)
}

// lis computes the Longest Increasing Subsequence of anchors by RightIdx.
// Input must be sorted by LeftIdx. Uses patience sorting (O(n log n)).
func lis(pairs []Anchor) []Anchor {
	n := len(pairs)
	if n == 0 {
		return nil
	}

	// tails[i] = index into pairs of the smallest RightIdx ending an IS of length i+1
	tails := make([]int, 0, n)
	// prev[i] = index of previous element in the LIS ending at pairs[i]
	prev := make([]int, n)
	for i := range prev {
		prev[i] = -1
	}

	for i, p := range pairs {
		lo, hi := 0, len(tails)
		for lo < hi {
			mid := (lo + hi) / 2
			if pairs[tails[mid]].RightIdx < p.RightIdx {
				lo = mid + 1
			} else {
				hi = mid
			}
		}

		if lo == len(tails) {
			tails = append(tails, i)
		} else {
			tails[lo] = i
		}

		if lo > 0 {
			prev[i] = tails[lo-1]
		}
	}

	length := len(tails)
	result := make([]Anchor, length)
	idx := tails[length-1]
	for i := length - 1; i >= 0; i-- {
		result[i] = pairs[idx]
		idx = prev[idx]
	}
	return result
}
