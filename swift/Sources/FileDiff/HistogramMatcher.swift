// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Line matcher backed by the Histogram Diff algorithm.
///
/// Extends Patience Diff by using line occurrence frequency to select
/// anchors, which handles repetitive structures (JSON, YAML, boilerplate)
/// where Patience would fall back to Myers.
public struct HistogramMatcher: LineMatcher {
    public init() {}

    public func match(_ left: [String], _ right: [String], policy: ComparisonPolicy) -> FairDiffIterable {
        let normLeft = normalizeLines(left, policy: policy)
        let normRight = normalizeLines(right, policy: policy)

        let matches = histogramDiff(normLeft, normRight) { subLeft, subRight in
            histogramMyersFallback(subLeft, subRight)
        }

        let builder = ChangeBuilder(length1: left.count, length2: right.count)
        for match in matches {
            builder.markEqual(match.leftIndex, match.rightIndex)
        }
        return fair(builder.finish())
    }
}

func histogramMyersFallback(_ left: [String], _ right: [String]) -> [HistogramAnchor] {
    if left.isEmpty || right.isEmpty {
        return []
    }
    let result = compareLines(left, right, policy: .default)
    var matches: [HistogramAnchor] = []
    for unchanged in result.unchanged {
        let count = unchanged.left.count
        for i in 0..<count {
            matches.append(HistogramAnchor(
                leftIndex: unchanged.left.lowerBound + i,
                rightIndex: unchanged.right.lowerBound + i
            ))
        }
    }
    return matches
}
