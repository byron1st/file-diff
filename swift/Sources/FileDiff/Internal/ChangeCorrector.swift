// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

// Expands a change range by consuming matching lines on both ends.
// swiftlint:disable function_parameter_count
func expandRangeLines(
    _ lines1: [Line],
    _ lines2: [Line],
    _ start1: Int,
    _ start2: Int,
    _ end1: Int,
    _ end2: Int
) -> DiffRange {
    // swiftlint:enable function_parameter_count
    var s1 = start1
    var s2 = start2
    while s1 < end1, s2 < end2, lines1[s1].equals(lines2[s2]) {
        s1 += 1
        s2 += 1
    }
    var e1 = end1
    var e2 = end2
    while s1 < e1, s2 < e2, lines1[e1 - 1].equals(lines2[e2 - 1]) {
        e1 -= 1
        e2 -= 1
    }
    return DiffRange(start1: s1, end1: e1, start2: s2, end2: e2)
}

/// Diffs two slices of `Line` using their cached hashes.
func diffLineSlice(_ lines1: [Line], _ lines2: [Line]) -> FairDiffIterable {
    let ints1 = lines1.map(\.hash)
    let ints2 = lines2.map(\.hash)
    return diff(ints1, ints2)
}

/// Fills gaps in a "big lines only" diff by running a full diff inside each gap.
final class SmartLineChangeCorrector {
    private let indexes1: [Int]
    private let indexes2: [Int]
    private let lines1: [Line]
    private let lines2: [Line]
    private let changes: FairDiffIterable

    init(indexes1: [Int], indexes2: [Int], lines1: [Line], lines2: [Line], changes: FairDiffIterable) {
        self.indexes1 = indexes1
        self.indexes2 = indexes2
        self.lines1 = lines1
        self.lines2 = lines2
        self.changes = changes
    }

    func build() -> FairDiffIterable {
        let builder = ChangeBuilder(length1: lines1.count, length2: lines2.count)
        var last1 = 0
        var last2 = 0
        for unchanged in changes.unchanged {
            let count = unchanged.left.count
            for i in 0..<count {
                let orig1 = indexes1[unchanged.left.lowerBound + i]
                let orig2 = indexes2[unchanged.right.lowerBound + i]
                matchGap(builder: builder, start1: last1, end1: orig1, start2: last2, end2: orig2)
                builder.markEqualRange(orig1, orig2, orig1 + 1, orig2 + 1)
                last1 = orig1 + 1
                last2 = orig2 + 1
            }
        }
        matchGap(builder: builder, start1: last1, end1: lines1.count, start2: last2, end2: lines2.count)
        return fair(builder.finish())
    }

    private func matchGap(builder: ChangeBuilder, start1: Int, end1: Int, start2: Int, end2: Int) {
        let expanded = expandRangeLines(lines1, lines2, start1, start2, end1, end2)
        let inner1 = Array(lines1[expanded.left])
        let inner2 = Array(lines2[expanded.right])
        let innerChanges = diffLineSlice(inner1, inner2)

        builder.markEqualRange(start1, start2, expanded.left.lowerBound, expanded.right.lowerBound)
        for change in innerChanges.unchanged {
            builder.markEqualCount(
                expanded.left.lowerBound + change.left.lowerBound,
                expanded.right.lowerBound + change.right.lowerBound,
                change.left.count
            )
        }
        builder.markEqualRange(expanded.left.upperBound, expanded.right.upperBound, end1, end2)
    }
}
