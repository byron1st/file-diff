// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Describes computed differences between two sequences.
///
/// All ranges reported by `changes` are non-empty on at least one side and
/// do not overlap with each other.
public protocol DiffIterable {
    /// Length of the left (original) sequence.
    var length1: Int { get }
    /// Length of the right (new) sequence.
    var length2: Int { get }
    /// All changed ranges in order.
    var changes: [DiffRange] { get }
    /// All unchanged ranges in order.
    var unchanged: [DiffRange] { get }
}

/// A `DiffIterable` whose unchanged ranges are guaranteed to have equal deltas on both sides,
/// i.e. `unchanged[i].left.count == unchanged[i].right.count`.
public protocol FairDiffIterable: DiffIterable {}

struct RangesDiffIterable: FairDiffIterable {
    let changes: [DiffRange]
    let length1: Int
    let length2: Int

    var unchanged: [DiffRange] {
        computeUnchanged(from: changes, length1: length1, length2: length2)
    }
}

struct InvertedDiffIterable: DiffIterable {
    let inner: DiffIterable
    var length1: Int {
        inner.length1
    }

    var length2: Int {
        inner.length2
    }

    var changes: [DiffRange] {
        inner.unchanged
    }

    var unchanged: [DiffRange] {
        inner.changes
    }
}

struct FairDiffIterableWrapper: FairDiffIterable {
    let inner: DiffIterable
    var length1: Int {
        inner.length1
    }

    var length2: Int {
        inner.length2
    }

    var changes: [DiffRange] {
        inner.changes
    }

    var unchanged: [DiffRange] {
        inner.unchanged
    }
}

/// Derives the list of unchanged ranges from a list of changed ranges.
func computeUnchanged(from changes: [DiffRange], length1: Int, length2: Int) -> [DiffRange] {
    var result: [DiffRange] = []
    var last1 = 0
    var last2 = 0
    for change in changes {
        if change.left.lowerBound > last1 || change.right.lowerBound > last2 {
            result.append(DiffRange(
                start1: last1, end1: change.left.lowerBound,
                start2: last2, end2: change.right.lowerBound
            ))
        }
        last1 = change.left.upperBound
        last2 = change.right.upperBound
    }
    if last1 < length1 || last2 < length2 {
        result.append(DiffRange(start1: last1, end1: length1, start2: last2, end2: length2))
    }
    return result
}
