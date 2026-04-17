// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Matched line pair produced by the Histogram Diff algorithm.
struct HistogramAnchor: Equatable {
    let leftIndex: Int
    let rightIndex: Int
}

typealias HistogramFallback = @Sendable (_ left: [String], _ right: [String]) -> [HistogramAnchor]

private let histogramMaxChainLength = 64

/// Runs the Histogram Diff algorithm on two line sequences.
///
/// Partitions on the line with the lowest occurrence in `right` and recurses
/// on sub-regions. Handles repetitive structures (JSON/YAML) better than
/// Patience by tolerating duplicates.
func histogramDiff(
    _ left: [String], _ right: [String], fallback: HistogramFallback
) -> [HistogramAnchor] {
    var result: [HistogramAnchor] = []
    histogramDiffRecursive(
        left, right, offsetL: 0, offsetR: 0,
        fallback: fallback, result: &result, depth: 0
    )
    return result
}

// swiftlint:disable:next function_parameter_count function_body_length
private func histogramDiffRecursive(
    _ left: [String], _ right: [String],
    offsetL: Int, offsetR: Int,
    fallback: HistogramFallback,
    result: inout [HistogramAnchor],
    depth: Int
) {
    if left.isEmpty || right.isEmpty {
        return
    }

    if depth >= histogramMaxChainLength {
        appendFallback(left, right, offsetL: offsetL, offsetR: offsetR, fallback: fallback, result: &result)
        return
    }

    var rightCount: [String: Int] = [:]
    for line in right {
        rightCount[line, default: 0] += 1
    }

    var bestLine = ""
    var bestFreq = 0
    var bestLeftIdx = -1
    var found = false

    for (index, line) in left.enumerated() {
        guard let freq = rightCount[line] else {
            continue
        }
        if !found || freq < bestFreq {
            bestLine = line
            bestFreq = freq
            bestLeftIdx = index
            found = true
            if freq == 1 {
                break
            }
        }
    }

    if !found {
        appendFallback(left, right, offsetL: offsetL, offsetR: offsetR, fallback: fallback, result: &result)
        return
    }

    let bestRightIdx = findBestRightMatch(right, line: bestLine, leftIdx: bestLeftIdx, leftLen: left.count)

    if bestLeftIdx > 0 || bestRightIdx > 0 {
        histogramDiffRecursive(
            Array(left[..<bestLeftIdx]), Array(right[..<bestRightIdx]),
            offsetL: offsetL, offsetR: offsetR,
            fallback: fallback, result: &result, depth: depth + 1
        )
    }

    result.append(HistogramAnchor(leftIndex: bestLeftIdx + offsetL, rightIndex: bestRightIdx + offsetR))

    let afterL = bestLeftIdx + 1
    let afterR = bestRightIdx + 1
    if afterL < left.count || afterR < right.count {
        histogramDiffRecursive(
            Array(left[afterL...]), Array(right[afterR...]),
            offsetL: offsetL + afterL, offsetR: offsetR + afterR,
            fallback: fallback, result: &result, depth: depth + 1
        )
    }
}

// swiftlint:disable:next function_parameter_count
private func appendFallback(
    _ left: [String], _ right: [String],
    offsetL: Int, offsetR: Int,
    fallback: HistogramFallback,
    result: inout [HistogramAnchor]
) {
    for match in fallback(left, right) {
        result.append(HistogramAnchor(
            leftIndex: match.leftIndex + offsetL,
            rightIndex: match.rightIndex + offsetR
        ))
    }
}

private func findBestRightMatch(_ right: [String], line: String, leftIdx: Int, leftLen: Int) -> Int {
    var positions: [Int] = []
    for (index, value) in right.enumerated() where value == line {
        positions.append(index)
    }
    if positions.count == 1 {
        return positions[0]
    }

    let denom = Double(max(leftLen, 1))
    let target = Double(leftIdx) / denom * Double(right.count)

    var bestPos = positions[0]
    var bestDist = abs(Double(positions[0]) - target)
    for position in positions.dropFirst() {
        let dist = abs(Double(position) - target)
        if dist < bestDist {
            bestDist = dist
            bestPos = position
        }
    }
    return bestPos
}
