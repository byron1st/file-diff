// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

import Foundation

/// Threshold below which a line is considered "unimportant" (e.g., blank or very short).
let unimportantLineCharCount = 3

private let deltaThresholdSize = 20000

enum MyersError: Error {
    case tooBig
}

/// Finds the longest common subsequence of two integer arrays using the
/// Myers O(ND) algorithm.
///
/// Based on E.W. Myers, "An O(ND) Difference Algorithm and Its Variations" (1986).
final class MyersLCS {
    private let first: [Int]
    private let second: [Int]
    private let start1: Int
    private let count1: Int
    private let start2: Int
    private let count2: Int

    let changes1: BitSet
    let changes2: BitSet

    private var vForward: [Int]
    private var vBackward: [Int]

    init(_ first: [Int], _ second: [Int]) {
        self.first = first
        self.second = second
        start1 = 0
        count1 = first.count
        start2 = 0
        count2 = second.count
        changes1 = BitSet(size: first.count)
        changes2 = BitSet(size: second.count)
        changes1.setRange(0, first.count, true)
        changes2.setRange(0, second.count, true)

        let total = count1 + count2
        vForward = Array(repeating: 0, count: total + 1)
        vBackward = Array(repeating: 0, count: total + 1)
    }

    /// Runs the full Myers algorithm without any threshold.
    func execute() {
        _ = try? execute(threshold: count1 + count2, throwing: false)
    }

    /// Runs with a size-based threshold; throws `MyersError.tooBig` if exceeded.
    func executeWithThreshold() throws {
        let heuristic = 20000 + 10 * Int(Double(count1 + count2).squareRoot())
        try execute(threshold: max(heuristic, deltaThresholdSize), throwing: true)
    }

    private func execute(threshold: Int, throwing: Bool) throws {
        if count1 == 0 || count2 == 0 {
            return
        }
        let capped = min(threshold, count1 + count2)
        try run(
            oldStart: 0, oldEnd: count1,
            newStart: 0, newEnd: count2,
            diffEstimate: capped,
            throwing: throwing
        )
    }

    // swiftlint:disable function_body_length cyclomatic_complexity function_parameter_count
    private func run(
        oldStart: Int, oldEnd: Int,
        newStart: Int, newEnd: Int,
        diffEstimate: Int,
        throwing: Bool
    ) throws {
        // swiftlint:enable function_body_length cyclomatic_complexity function_parameter_count
        if oldStart >= oldEnd || newStart >= newEnd {
            return
        }

        let oldLength = oldEnd - oldStart
        let newLength = newEnd - newStart
        vForward[newLength + 1] = 0
        vBackward[newLength + 1] = 0
        let halfD = (diffEstimate + 1) / 2

        var xx = -1
        var kk = -1
        var td = -1

        outer: for dValue in 0...halfD {
            let lowK = newLength + max(-dValue, -newLength + ((dValue ^ newLength) & 1))
            let highK = newLength + min(dValue, oldLength - ((dValue ^ oldLength) & 1))

            // Forward pass
            var k = lowK
            while k <= highK {
                var x: Int = if k == lowK || (k != highK && vForward[k - 1] < vForward[k + 1]) {
                    vForward[k + 1]
                } else {
                    vForward[k - 1] + 1
                }
                let y = x - k + newLength
                let cap = min(oldEnd - oldStart - x, newEnd - newStart - y)
                x += commonLenForward(
                    oldIndex: oldStart + x,
                    newIndex: newStart + y,
                    maxLen: cap
                )
                vForward[k] = x
                k += 2
            }

            if (oldLength - newLength) % 2 != 0 {
                var k2 = lowK
                while k2 <= highK {
                    if oldLength - (dValue - 1) <= k2, k2 <= oldLength + (dValue - 1),
                       vForward[k2] + vBackward[newLength + oldLength - k2] >= oldLength
                    {
                        xx = vForward[k2]
                        kk = k2
                        td = 2 * dValue - 1
                        break outer
                    }
                    k2 += 2
                }
            }

            // Backward pass
            var kb = lowK
            while kb <= highK {
                var x: Int = if kb == lowK || (kb != highK && vBackward[kb - 1] < vBackward[kb + 1]) {
                    vBackward[kb + 1]
                } else {
                    vBackward[kb - 1] + 1
                }
                let y = x - kb + newLength
                let cap = min(oldEnd - oldStart - x, newEnd - newStart - y)
                x += commonLenBackward(
                    oldIndex: oldEnd - 1 - x,
                    newIndex: newEnd - 1 - y,
                    maxLen: cap
                )
                vBackward[kb] = x
                kb += 2
            }

            if (oldLength - newLength) % 2 == 0 {
                var k2 = lowK
                while k2 <= highK {
                    if oldLength - dValue <= k2, k2 <= oldLength + dValue,
                       vForward[oldLength + newLength - k2] + vBackward[k2] >= oldLength
                    {
                        xx = oldLength - vBackward[k2]
                        kk = oldLength + newLength - k2
                        td = 2 * dValue
                        break outer
                    }
                    k2 += 2
                }
            }
        }

        if td > 1 {
            let yy = xx - kk + newLength
            let oldDiff = (td + 1) / 2
            if xx > 0, yy > 0 {
                try run(
                    oldStart: oldStart, oldEnd: oldStart + xx,
                    newStart: newStart, newEnd: newStart + yy,
                    diffEstimate: oldDiff,
                    throwing: throwing
                )
            }
            if oldStart + xx < oldEnd, newStart + yy < newEnd {
                try run(
                    oldStart: oldStart + xx, oldEnd: oldEnd,
                    newStart: newStart + yy, newEnd: newEnd,
                    diffEstimate: td - oldDiff,
                    throwing: throwing
                )
            }
        } else if td >= 0 {
            var x = oldStart
            var y = newStart
            while x < oldEnd, y < newEnd {
                let cap = min(oldEnd - x, newEnd - y)
                let common = commonLenForward(oldIndex: x, newIndex: y, maxLen: cap)
                if common > 0 {
                    addUnchanged(start1: x, start2: y, count: common)
                    x += common
                    y += common
                } else if oldEnd - oldStart > newEnd - newStart {
                    x += 1
                } else {
                    y += 1
                }
            }
        } else if throwing {
            throw MyersError.tooBig
        }
    }

    private func addUnchanged(start1: Int, start2: Int, count: Int) {
        changes1.setRange(self.start1 + start1, self.start1 + start1 + count, false)
        changes2.setRange(self.start2 + start2, self.start2 + start2 + count, false)
    }

    private func commonLenForward(oldIndex: Int, newIndex: Int, maxLen: Int) -> Int {
        let cap = min(maxLen, min(count1 - oldIndex, count2 - newIndex))
        var x = oldIndex
        var y = newIndex
        while x - oldIndex < cap, first[start1 + x] == second[start2 + y] {
            x += 1
            y += 1
        }
        return x - oldIndex
    }

    private func commonLenBackward(oldIndex: Int, newIndex: Int, maxLen: Int) -> Int {
        let cap = min(maxLen, min(oldIndex + 1, newIndex + 1))
        var x = oldIndex
        var y = newIndex
        while oldIndex - x < cap, first[start1 + x] == second[start2 + y] {
            x -= 1
            y -= 1
        }
        return oldIndex - x
    }
}
