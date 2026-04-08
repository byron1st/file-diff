// Copyright 2000-2022 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package comparison

import "github.com/byron1st/file-diff/diff/util"

// DiffIterable represents computed differences between two sequences.
//
// All Ranges are non-empty (have at least one element in one of the sides).
// Ranges do not overlap.
type DiffIterable interface {
	// Length1 returns the length of the first sequence.
	Length1() int
	// Length2 returns the length of the second sequence.
	Length2() int
	// Changes returns all changed ranges.
	Changes() []util.Range
	// Unchanged returns all unchanged ranges.
	Unchanged() []util.Range
}

// FairDiffIterable is a DiffIterable where elements are compared one-by-one.
//
// If range [a, b) is equal to [a', b'), then element(a+i) equals element(a'+i)
// for all i in [0, b-a). Therefore, unchanged ranges are guaranteed to have
// equal deltas (End1-Start1 == End2-Start2).
type FairDiffIterable interface {
	DiffIterable
}
