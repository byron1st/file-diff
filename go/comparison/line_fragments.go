package comparison

import (
	"errors"
	"strings"

	"github.com/byron1st/file-diff/go/fragment"
)

// ErrNilLineMatcher is returned when CompareLineFragments is called with a nil matcher.
var ErrNilLineMatcher = errors.New("nil line matcher")

// CompareLineFragments compares two line slices and returns changed line fragments
// with word-level inner fragments for each changed range.
func CompareLineFragments(
	lines1, lines2 []string,
	matcher LineMatcher,
	policy ComparisonPolicy,
) ([]fragment.LineFragment, error) {
	if matcher == nil {
		return nil, ErrNilLineMatcher
	}

	lineDiff := matcher.Match(lines1, lines2, policy)
	changes := lineDiff.Changes()
	result := make([]fragment.LineFragment, 0, len(changes))

	for _, r := range changes {
		left := strings.Join(lines1[r.Start1:r.End1], "\n")
		right := strings.Join(lines2[r.Start2:r.End2], "\n")

		inner, err := CompareWords(left, right, policy)
		if err != nil {
			return nil, err
		}

		result = append(result, fragment.NewLineFragment(
			r.Start1, r.End1,
			r.Start2, r.End2,
			0, len(left),
			0, len(right),
			inner,
		))
	}

	return result, nil
}
