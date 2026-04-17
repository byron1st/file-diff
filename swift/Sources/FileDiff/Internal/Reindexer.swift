// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Removes elements unique to one side before running LCS, then maps results
/// back to the original index space.
final class Reindexer {
    private var oldIndices: [[Int]] = [[], []]
    private var originalLengths: [Int] = [0, 0]
    private var discardedLengths: [Int] = [0, 0]

    /// Discards elements appearing only in one sequence and returns the filtered pair.
    @discardableResult
    func discardUnique(_ ints1: [Int], _ ints2: [Int]) -> [[Int]] {
        let discarded1 = discard(needed: ints2, toDiscard: ints1, arrayIndex: 0)
        let discarded2 = discard(needed: discarded1, toDiscard: ints2, arrayIndex: 1)
        return [discarded1, discarded2]
    }

    // Maps the BitSet changes from the discarded index space back to the original
    // index space and streams alternating equal/change blocks to `builder`.
    // swiftlint:disable function_body_length cyclomatic_complexity
    func reindex(_ changes: (BitSet, BitSet), builder: LCSBuilder) {
        // swiftlint:enable function_body_length cyclomatic_complexity
        let changes1: BitSet
        let changes2: BitSet

        if discardedLengths[0] == originalLengths[0], discardedLengths[1] == originalLengths[1] {
            changes1 = changes.0
            changes2 = changes.1
        } else {
            changes1 = BitSet(size: originalLengths[0])
            changes2 = BitSet(size: originalLengths[1])

            var x = 0
            var y = 0
            while x < discardedLengths[0] || y < discardedLengths[1] {
                if x < discardedLengths[0], y < discardedLengths[1],
                   !changes.0.get(x), !changes.1.get(y)
                {
                    x = increment(indices: oldIndices[0], at: x, set: changes1, length: originalLengths[0])
                    y = increment(indices: oldIndices[1], at: y, set: changes2, length: originalLengths[1])
                } else if x < discardedLengths[0], changes.0.get(x) {
                    changes1.set(oldIndices[0][x], true)
                    x = increment(indices: oldIndices[0], at: x, set: changes1, length: originalLengths[0])
                } else if y < discardedLengths[1], changes.1.get(y) {
                    changes2.set(oldIndices[1][y], true)
                    y = increment(indices: oldIndices[1], at: y, set: changes2, length: originalLengths[1])
                }
            }

            if discardedLengths[0] == 0 {
                changes1.setRange(0, originalLengths[0], true)
            } else {
                changes1.setRange(0, oldIndices[0][0], true)
            }
            if discardedLengths[1] == 0 {
                changes2.setRange(0, originalLengths[1], true)
            } else {
                changes2.setRange(0, oldIndices[1][0], true)
            }
        }

        var x = 0
        var y = 0
        while x < originalLengths[0], y < originalLengths[1] {
            let startX = x
            while x < originalLengths[0], y < originalLengths[1],
                  !changes1.get(x), !changes2.get(y)
            {
                x += 1
                y += 1
            }
            if x > startX {
                builder.addEqual(x - startX)
            }
            var dx = 0
            var dy = 0
            while x < originalLengths[0], changes1.get(x) {
                dx += 1
                x += 1
            }
            while y < originalLengths[1], changes2.get(y) {
                dy += 1
                y += 1
            }
            if dx != 0 || dy != 0 {
                builder.addChange(dx, dy)
            }
        }
        if x != originalLengths[0] || y != originalLengths[1] {
            builder.addChange(originalLengths[0] - x, originalLengths[1] - y)
        }
    }

    private func discard(needed: [Int], toDiscard: [Int], arrayIndex: Int) -> [Int] {
        originalLengths[arrayIndex] = toDiscard.count
        let sorted = needed.sorted()
        var discarded: [Int] = []
        var old: [Int] = []
        for (index, value) in toDiscard.enumerated() where Self.contains(sorted, value) {
            discarded.append(value)
            old.append(index)
        }
        oldIndices[arrayIndex] = old
        discardedLengths[arrayIndex] = discarded.count
        return discarded
    }

    private static func contains(_ sorted: [Int], _ target: Int) -> Bool {
        var lo = 0
        var hi = sorted.count
        while lo < hi {
            let mid = (lo + hi) / 2
            if sorted[mid] < target {
                lo = mid + 1
            } else {
                hi = mid
            }
        }
        return lo < sorted.count && sorted[lo] == target
    }

    private func increment(indices: [Int], at index: Int, set: BitSet, length: Int) -> Int {
        if index + 1 < indices.count {
            set.setRange(indices[index] + 1, indices[index + 1], true)
        } else {
            set.setRange(indices[index] + 1, length, true)
        }
        return index + 1
    }
}
