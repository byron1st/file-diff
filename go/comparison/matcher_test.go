package comparison

import (
	"testing"

	"github.com/byron1st/file-diff/go/util"
)

// dummyIterable is a minimal FairDiffIterable for testing the interface.
type dummyIterable struct {
	length1   int
	length2   int
	changes   []util.Range
	unchanged []util.Range
}

func (d *dummyIterable) Length1() int            { return d.length1 }
func (d *dummyIterable) Length2() int            { return d.length2 }
func (d *dummyIterable) Changes() []util.Range   { return d.changes }
func (d *dummyIterable) Unchanged() []util.Range { return d.unchanged }

// dummyMatcher is a no-op LineMatcher for testing the interface contract.
type dummyMatcher struct{}

func (m *dummyMatcher) Match(left, right []string, policy ComparisonPolicy) FairDiffIterable {
	return &dummyIterable{
		length1:   len(left),
		length2:   len(right),
		changes:   nil,
		unchanged: nil,
	}
}

func TestLineMatcher_DummyImplementation(t *testing.T) {
	var matcher LineMatcher = &dummyMatcher{}
	result := matcher.Match([]string{"a", "b"}, []string{"a", "c"}, PolicyDefault)

	if result.Length1() != 2 {
		t.Fatalf("expected length1=2, got %d", result.Length1())
	}
	if result.Length2() != 2 {
		t.Fatalf("expected length2=2, got %d", result.Length2())
	}
}
