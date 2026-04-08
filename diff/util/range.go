// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package util

import "fmt"

// Range stores half-open intervals [start, end).
// Start1/End1 represent the range in the first sequence,
// Start2/End2 represent the range in the second sequence.
type Range struct {
	Start1 int
	End1   int
	Start2 int
	End2   int
}

func NewRange(start1, end1, start2, end2 int) Range {
	if start1 > end1 || start2 > end2 {
		panic(fmt.Sprintf("invalid range: [%d, %d, %d, %d]", start1, end1, start2, end2))
	}
	return Range{Start1: start1, End1: end1, Start2: start2, End2: end2}
}

func (r Range) IsEmpty() bool {
	return r.Start1 == r.End1 && r.Start2 == r.End2
}

func (r Range) String() string {
	return fmt.Sprintf("[%d, %d) - [%d, %d)", r.Start1, r.End1, r.Start2, r.End2)
}
