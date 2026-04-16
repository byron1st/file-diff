// Histogram Diff algorithm implementation.
//
// An extension of Patience Diff that uses line occurrence frequency
// (histogram) to select anchors. Instead of requiring lines to be
// unique in both sides, it selects the line with the lowest occurrence
// count as the partition point, then recursively diffs the regions
// before and after the anchor. This handles repetitive structures
// (JSON, YAML, boilerplate test code) better than Patience.

package histogram

// Anchor represents a matched line between left and right.
type Anchor struct {
	LeftIdx  int
	RightIdx int
}

// DiffFunc computes diff for sub-regions that cannot be partitioned further.
type DiffFunc func(left, right []string) []Anchor

// maxChainLength limits recursion depth to avoid degenerate cases.
const maxChainLength = 64

// Diff runs the Histogram Diff algorithm on left and right line slices.
// fallback is used for regions where no low-frequency anchor can be found.
func Diff(left, right []string, fallback DiffFunc) []Anchor {
	var result []Anchor
	diffRecursive(left, right, 0, 0, fallback, &result, 0)
	return result
}

func diffRecursive(left, right []string, offsetL, offsetR int, fallback DiffFunc, result *[]Anchor, depth int) {
	if len(left) == 0 || len(right) == 0 {
		return
	}

	if depth >= maxChainLength {
		matches := fallback(left, right)
		for _, m := range matches {
			*result = append(*result, Anchor{LeftIdx: m.LeftIdx + offsetL, RightIdx: m.RightIdx + offsetR})
		}
		return
	}

	// Build histogram of right-side lines
	rightCount := make(map[string]int)
	for _, r := range right {
		rightCount[r]++
	}

	// Scan left to find the line with the lowest occurrence in right.
	// Among lines with the same frequency, pick the first one found.
	bestLine := ""
	bestFreq := 0
	bestLeftIdx := -1
	found := false

	for i, l := range left {
		freq, exists := rightCount[l]
		if !exists {
			continue
		}
		if !found || freq < bestFreq {
			bestLine = l
			bestFreq = freq
			bestLeftIdx = i
			found = true
			if freq == 1 {
				break // can't do better than unique
			}
		}
	}

	if !found {
		// No common line at all — fallback
		matches := fallback(left, right)
		for _, m := range matches {
			*result = append(*result, Anchor{LeftIdx: m.LeftIdx + offsetL, RightIdx: m.RightIdx + offsetR})
		}
		return
	}

	// Find the best matching occurrence of bestLine in right.
	// Use the occurrence whose position best preserves ordering relative to bestLeftIdx.
	bestRightIdx := findBestRightMatch(right, bestLine, bestLeftIdx, len(left))

	// Recursively diff the region before the anchor
	if bestLeftIdx > 0 || bestRightIdx > 0 {
		diffRecursive(
			left[:bestLeftIdx], right[:bestRightIdx],
			offsetL, offsetR,
			fallback, result, depth+1,
		)
	}

	// Emit the anchor
	*result = append(*result, Anchor{LeftIdx: bestLeftIdx + offsetL, RightIdx: bestRightIdx + offsetR})

	// Recursively diff the region after the anchor
	afterL := bestLeftIdx + 1
	afterR := bestRightIdx + 1
	if afterL < len(left) || afterR < len(right) {
		diffRecursive(
			left[afterL:], right[afterR:],
			offsetL+afterL, offsetR+afterR,
			fallback, result, depth+1,
		)
	}
}

// findBestRightMatch finds the occurrence of line in right that best
// matches the proportional position of leftIdx within the left side.
func findBestRightMatch(right []string, line string, leftIdx, leftLen int) int {
	// Collect all positions of the line in right
	var positions []int
	for i, r := range right {
		if r == line {
			positions = append(positions, i)
		}
	}

	if len(positions) == 1 {
		return positions[0]
	}

	// Choose the position closest to the proportional location
	ratio := float64(leftIdx) / float64(max(leftLen, 1))
	target := ratio * float64(len(right))

	bestPos := positions[0]
	bestDist := abs(float64(positions[0]) - target)
	for _, p := range positions[1:] {
		d := abs(float64(p) - target)
		if d < bestDist {
			bestDist = d
			bestPos = p
		}
	}
	return bestPos
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
