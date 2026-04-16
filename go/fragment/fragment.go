// Copyright 2000-2021 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package fragment

import "fmt"

// DiffFragment represents a modified part of text at the character/word offset level.
// Offsets are byte offsets within the compared text segments.
type DiffFragment struct {
	StartOffset1 int
	EndOffset1   int
	StartOffset2 int
	EndOffset2   int
}

func NewDiffFragment(startOffset1, endOffset1, startOffset2, endOffset2 int) DiffFragment {
	if startOffset1 == endOffset1 && startOffset2 == endOffset2 {
		panic(fmt.Sprintf("DiffFragment should not be empty: [%d, %d) - [%d, %d)", startOffset1, endOffset1, startOffset2, endOffset2))
	}
	if startOffset1 > endOffset1 || startOffset2 > endOffset2 {
		panic(fmt.Sprintf("DiffFragment is invalid: [%d, %d) - [%d, %d)", startOffset1, endOffset1, startOffset2, endOffset2))
	}
	return DiffFragment{
		StartOffset1: startOffset1,
		EndOffset1:   endOffset1,
		StartOffset2: startOffset2,
		EndOffset2:   endOffset2,
	}
}

func (f DiffFragment) String() string {
	return fmt.Sprintf("[%d, %d) - [%d, %d)", f.StartOffset1, f.EndOffset1, f.StartOffset2, f.EndOffset2)
}

// LineFragment represents a modified part of text at the line level.
// It includes both line ranges and byte offset ranges.
// InnerFragments contains optional word/char-level detail within this line change.
type LineFragment struct {
	StartLine1 int
	EndLine1   int
	StartLine2 int
	EndLine2   int

	StartOffset1 int
	EndOffset1   int
	StartOffset2 int
	EndOffset2   int

	// InnerFragments holds high-granularity changes inside this line fragment
	// (e.g., detected by word-level diff). nil means no inner similarities were found.
	// Offsets of inner fragments are relative to the start of this LineFragment.
	InnerFragments []DiffFragment
}

func NewLineFragment(
	startLine1, endLine1, startLine2, endLine2 int,
	startOffset1, endOffset1, startOffset2, endOffset2 int,
	innerFragments []DiffFragment,
) LineFragment {
	if startLine1 == endLine1 && startLine2 == endLine2 {
		panic(fmt.Sprintf("LineFragment should not be empty: lines [%d, %d) - [%d, %d)", startLine1, endLine1, startLine2, endLine2))
	}
	if startLine1 > endLine1 || startLine2 > endLine2 || startOffset1 > endOffset1 || startOffset2 > endOffset2 {
		panic(fmt.Sprintf("LineFragment is invalid: lines [%d, %d) - [%d, %d); offsets [%d, %d) - [%d, %d)",
			startLine1, endLine1, startLine2, endLine2,
			startOffset1, endOffset1, startOffset2, endOffset2))
	}
	innerFragments = dropWholeChangedFragments(innerFragments, endOffset1-startOffset1, endOffset2-startOffset2)
	return LineFragment{
		StartLine1:     startLine1,
		EndLine1:       endLine1,
		StartLine2:     startLine2,
		EndLine2:       endLine2,
		StartOffset1:   startOffset1,
		EndOffset1:     endOffset1,
		StartOffset2:   startOffset2,
		EndOffset2:     endOffset2,
		InnerFragments: innerFragments,
	}
}

func (f LineFragment) String() string {
	innerLen := -1
	if f.InnerFragments != nil {
		innerLen = len(f.InnerFragments)
	}
	return fmt.Sprintf("Lines [%d, %d) - [%d, %d); Offsets [%d, %d) - [%d, %d); Inner %d",
		f.StartLine1, f.EndLine1, f.StartLine2, f.EndLine2,
		f.StartOffset1, f.EndOffset1, f.StartOffset2, f.EndOffset2,
		innerLen)
}

// dropWholeChangedFragments removes a single inner fragment that spans the
// entire line fragment, since it adds no information.
func dropWholeChangedFragments(fragments []DiffFragment, length1, length2 int) []DiffFragment {
	if len(fragments) == 1 {
		f := fragments[0]
		if f.StartOffset1 == 0 && f.StartOffset2 == 0 && f.EndOffset1 == length1 && f.EndOffset2 == length2 {
			return nil
		}
	}
	return fragments
}
