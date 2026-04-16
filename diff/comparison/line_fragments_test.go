package comparison

import (
	"errors"
	"testing"

	"github.com/byron1st/file-diff/diff/util"
)

type stubLineMatcher struct {
	changes []util.Range
}

func (m *stubLineMatcher) Match(left, right []string, policy ComparisonPolicy) FairDiffIterable {
	return Fair(CreateFromRanges(m.changes, len(left), len(right)))
}

func TestCompareLineFragments_Identical(t *testing.T) {
	lines := splitLines("alpha\nbeta")

	fragments, err := CompareLineFragments(lines, lines, &HistogramMatcher{}, PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 0 {
		t.Fatalf("expected no fragments, got %v", fragments)
	}
}

func TestCompareLineFragments_SingleModifiedLine(t *testing.T) {
	left := splitLines("alpha quick omega")
	right := splitLines("alpha slow omega")

	fragments, err := CompareLineFragments(left, right, &MyersMatcher{}, PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(fragments))
	}

	f := fragments[0]
	if f.StartLine1 != 0 || f.EndLine1 != 1 || f.StartLine2 != 0 || f.EndLine2 != 1 {
		t.Fatalf("unexpected line range: %v", f)
	}
	if f.StartOffset1 != 0 || f.EndOffset1 != len(left[0]) || f.StartOffset2 != 0 || f.EndOffset2 != len(right[0]) {
		t.Fatalf("unexpected offsets: %v", f)
	}
	if len(f.InnerFragments) == 0 {
		t.Fatalf("expected inner fragments, got %v", f.InnerFragments)
	}

	inner := f.InnerFragments[0]
	if got := left[0][inner.StartOffset1:inner.EndOffset1]; got != "quick" {
		t.Fatalf("unexpected left inner change: %q", got)
	}
	if got := right[0][inner.StartOffset2:inner.EndOffset2]; got != "slow" {
		t.Fatalf("unexpected right inner change: %q", got)
	}
}

func TestCompareLineFragments_InsertionOnly(t *testing.T) {
	left := splitLines("alpha\nomega")
	right := splitLines("alpha\nbeta value\nomega")

	fragments, err := CompareLineFragments(left, right, &MyersMatcher{}, PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(fragments))
	}

	f := fragments[0]
	if f.StartLine1 != 1 || f.EndLine1 != 1 || f.StartLine2 != 1 || f.EndLine2 != 2 {
		t.Fatalf("unexpected line range: %v", f)
	}
	if f.StartOffset1 != 0 || f.EndOffset1 != 0 {
		t.Fatalf("unexpected left offsets: %v", f)
	}
	if f.StartOffset2 != 0 || f.EndOffset2 != len("beta value") {
		t.Fatalf("unexpected right offsets: %v", f)
	}
	if f.InnerFragments != nil {
		t.Fatalf("expected nil inner fragments for whole-block insertion, got %v", f.InnerFragments)
	}
}

func TestCompareLineFragments_DeletionOnly(t *testing.T) {
	left := splitLines("alpha\nbeta value\nomega")
	right := splitLines("alpha\nomega")

	fragments, err := CompareLineFragments(left, right, &MyersMatcher{}, PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(fragments))
	}

	f := fragments[0]
	if f.StartLine1 != 1 || f.EndLine1 != 2 || f.StartLine2 != 1 || f.EndLine2 != 1 {
		t.Fatalf("unexpected line range: %v", f)
	}
	if f.StartOffset1 != 0 || f.EndOffset1 != len("beta value") {
		t.Fatalf("unexpected left offsets: %v", f)
	}
	if f.StartOffset2 != 0 || f.EndOffset2 != 0 {
		t.Fatalf("unexpected right offsets: %v", f)
	}
	if f.InnerFragments != nil {
		t.Fatalf("expected nil inner fragments for whole-block deletion, got %v", f.InnerFragments)
	}
}

func TestCompareLineFragments_MultipleSeparatedChangeChunks(t *testing.T) {
	left := []string{"keep", "alpha old value", "middle", "beta old value", "tail"}
	right := []string{"keep", "alpha new value", "middle", "beta new value", "tail"}
	matcher := &stubLineMatcher{
		changes: []util.Range{
			util.NewRange(1, 2, 1, 2),
			util.NewRange(3, 4, 3, 4),
		},
	}

	fragments, err := CompareLineFragments(left, right, matcher, PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 2 {
		t.Fatalf("expected 2 fragments, got %d", len(fragments))
	}
	if fragments[0].StartLine1 != 1 || fragments[0].StartLine2 != 1 {
		t.Fatalf("unexpected first fragment order: %v", fragments[0])
	}
	if fragments[1].StartLine1 != 3 || fragments[1].StartLine2 != 3 {
		t.Fatalf("unexpected second fragment order: %v", fragments[1])
	}
}

func TestCompareLineFragments_TrimWhitespaces(t *testing.T) {
	left := splitLines("  alpha  \n beta ")
	right := splitLines("alpha\nbeta")

	fragments, err := CompareLineFragments(left, right, &MyersMatcher{}, PolicyTrimWhitespaces)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 0 {
		t.Fatalf("expected no fragments, got %v", fragments)
	}
}

func TestCompareLineFragments_IgnoreWhitespaces(t *testing.T) {
	left := splitLines("a b c\nd e f")
	right := splitLines("abc\ndef")

	fragments, err := CompareLineFragments(left, right, &MyersMatcher{}, PolicyIgnoreWhitespaces)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 0 {
		t.Fatalf("expected no fragments, got %v", fragments)
	}
}

func TestCompareLineFragments_NilMatcher(t *testing.T) {
	_, err := CompareLineFragments(nil, nil, nil, PolicyDefault)
	if !errors.Is(err, ErrNilLineMatcher) {
		t.Fatalf("expected ErrNilLineMatcher, got %v", err)
	}
}

func TestCompareLineFragments_InnerFragmentOffsetsAreRelativeToJoinedText(t *testing.T) {
	left := splitLines("alpha quick omega")
	right := splitLines("alpha slow omega")

	fragments, err := CompareLineFragments(left, right, &MyersMatcher{}, PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(fragments))
	}
	if len(fragments[0].InnerFragments) == 0 {
		t.Fatalf("expected inner fragments, got %v", fragments[0].InnerFragments)
	}

	leftJoined := left[0]
	rightJoined := right[0]
	inner := fragments[0].InnerFragments[0]

	if got := leftJoined[inner.StartOffset1:inner.EndOffset1]; got != "quick" {
		t.Fatalf("unexpected left joined-text slice: %q", got)
	}
	if got := rightJoined[inner.StartOffset2:inner.EndOffset2]; got != "slow" {
		t.Fatalf("unexpected right joined-text slice: %q", got)
	}
}

func TestCompareLineFragments_MultiLineChangedBlock(t *testing.T) {
	left := splitLines("keep\nfoo quick\nbar end\ntail")
	right := splitLines("keep\nfoo slow\nbar finish\ntail")

	fragments, err := CompareLineFragments(left, right, &MyersMatcher{}, PolicyDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(fragments))
	}

	f := fragments[0]
	leftJoined := "foo quick\nbar end"
	rightJoined := "foo slow\nbar finish"

	if f.StartLine1 != 1 || f.EndLine1 != 3 || f.StartLine2 != 1 || f.EndLine2 != 3 {
		t.Fatalf("unexpected line range: %v", f)
	}
	if f.StartOffset1 != 0 || f.EndOffset1 != len(leftJoined) {
		t.Fatalf("unexpected left offsets: %v", f)
	}
	if f.StartOffset2 != 0 || f.EndOffset2 != len(rightJoined) {
		t.Fatalf("unexpected right offsets: %v", f)
	}
}
