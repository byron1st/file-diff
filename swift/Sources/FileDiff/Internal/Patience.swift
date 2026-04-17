// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// A matched pair of line indices produced by the Patience diff algorithm.
struct PatienceAnchor: Equatable {
    let leftIndex: Int
    let rightIndex: Int
}

/// Fallback matcher used for sub-regions that contain no unique common lines.
typealias PatienceFallback = @Sendable (_ left: [String], _ right: [String]) -> [PatienceAnchor]

/// Runs Patience Diff on two line sequences.
///
/// Anchors are unique common lines whose right-side positions form a longest
/// increasing subsequence. Sub-regions between anchors are recursively diffed,
/// and regions without unique anchors delegate to `fallback`.
func patienceDiff(
    _ left: [String], _ right: [String], fallback: PatienceFallback
) -> [PatienceAnchor] {
    var result: [PatienceAnchor] = []
    patienceDiffRecursive(left, right, offsetL: 0, offsetR: 0, fallback: fallback, result: &result)
    return result
}

// swiftlint:disable:next function_parameter_count
private func patienceDiffRecursive(
    _ left: [String], _ right: [String],
    offsetL: Int, offsetR: Int,
    fallback: PatienceFallback,
    result: inout [PatienceAnchor]
) {
    if left.isEmpty || right.isEmpty {
        return
    }
    let anchors = findAnchors(left, right)

    if anchors.isEmpty {
        let matches = fallback(left, right)
        for match in matches {
            result.append(PatienceAnchor(
                leftIndex: match.leftIndex + offsetL,
                rightIndex: match.rightIndex + offsetR
            ))
        }
        return
    }

    var prevL = 0
    var prevR = 0
    for anchor in anchors {
        if anchor.leftIndex > prevL || anchor.rightIndex > prevR {
            patienceDiffRecursive(
                Array(left[prevL..<anchor.leftIndex]),
                Array(right[prevR..<anchor.rightIndex]),
                offsetL: offsetL + prevL, offsetR: offsetR + prevR,
                fallback: fallback, result: &result
            )
        }
        result.append(PatienceAnchor(
            leftIndex: anchor.leftIndex + offsetL,
            rightIndex: anchor.rightIndex + offsetR
        ))
        prevL = anchor.leftIndex + 1
        prevR = anchor.rightIndex + 1
    }

    if prevL < left.count || prevR < right.count {
        patienceDiffRecursive(
            Array(left[prevL...]),
            Array(right[prevR...]),
            offsetL: offsetL + prevL, offsetR: offsetR + prevR,
            fallback: fallback, result: &result
        )
    }
}

private func findAnchors(_ left: [String], _ right: [String]) -> [PatienceAnchor] {
    var leftCount: [String: Int] = [:]
    for line in left {
        leftCount[line, default: 0] += 1
    }
    var rightCount: [String: Int] = [:]
    for line in right {
        rightCount[line, default: 0] += 1
    }
    var rightIndex: [String: Int] = [:]
    for (index, line) in right.enumerated() where rightCount[line] == 1 {
        rightIndex[line] = index
    }

    var pairs: [PatienceAnchor] = []
    for (index, line) in left.enumerated() where leftCount[line] == 1 {
        if rightCount[line] == 1, let rIdx = rightIndex[line] {
            pairs.append(PatienceAnchor(leftIndex: index, rightIndex: rIdx))
        }
    }
    if pairs.isEmpty {
        return []
    }
    return longestIncreasingSubsequence(pairs)
}

/// Longest increasing subsequence of `pairs` by `rightIndex`.
///
/// Input must be sorted by `leftIndex`. Uses patience sorting (`O(n log n)`).
func longestIncreasingSubsequence(_ pairs: [PatienceAnchor]) -> [PatienceAnchor] {
    let count = pairs.count
    if count == 0 {
        return []
    }

    var tails: [Int] = []
    tails.reserveCapacity(count)
    var prev = Array(repeating: -1, count: count)

    for (index, pair) in pairs.enumerated() {
        var lo = 0
        var hi = tails.count
        while lo < hi {
            let mid = (lo + hi) / 2
            if pairs[tails[mid]].rightIndex < pair.rightIndex {
                lo = mid + 1
            } else {
                hi = mid
            }
        }
        if lo == tails.count {
            tails.append(index)
        } else {
            tails[lo] = index
        }
        if lo > 0 {
            prev[index] = tails[lo - 1]
        }
    }

    var result = Array(repeating: PatienceAnchor(leftIndex: 0, rightIndex: 0), count: tails.count)
    var idx = tails[tails.count - 1]
    for position in (0..<tails.count).reversed() {
        result[position] = pairs[idx]
        idx = prev[idx]
    }
    return result
}
