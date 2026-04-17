// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Line matcher backed by the Patience Diff algorithm.
///
/// Anchors on unique common lines via longest increasing subsequence, then
/// delegates regions without unique anchors to a Myers-based fallback.
public struct PatienceMatcher: LineMatcher {
    public init() {}

    public func match(_ left: [String], _ right: [String], policy: ComparisonPolicy) -> FairDiffIterable {
        let normLeft = normalizeLines(left, policy: policy)
        let normRight = normalizeLines(right, policy: policy)

        let matches = patienceDiff(normLeft, normRight) { subLeft, subRight in
            myersFallback(subLeft, subRight)
        }

        let builder = ChangeBuilder(length1: left.count, length2: right.count)
        for match in matches {
            builder.markEqual(match.leftIndex, match.rightIndex)
        }
        return fair(builder.finish())
    }
}

func normalizeLines(_ lines: [String], policy: ComparisonPolicy) -> [String] {
    if policy == .default {
        return lines
    }
    return lines.map { normalizedContent($0, policy: policy) }
}

private func normalizedContent(_ line: String, policy: ComparisonPolicy) -> String {
    switch policy {
    case .trimWhitespaces:
        line.trimmingCharacters(in: .whitespacesAndNewlines)
    case .ignoreWhitespaces:
        String(line.unicodeScalars.filter { !CharacterUtils.isWhiteSpaceCodePoint($0) })
    default:
        line
    }
}

func myersFallback(_ left: [String], _ right: [String]) -> [PatienceAnchor] {
    if left.isEmpty || right.isEmpty {
        return []
    }
    let result = compareLines(left, right, policy: .default)
    var matches: [PatienceAnchor] = []
    for unchanged in result.unchanged {
        let count = unchanged.left.count
        for i in 0..<count {
            matches.append(PatienceAnchor(
                leftIndex: unchanged.left.lowerBound + i,
                rightIndex: unchanged.right.lowerBound + i
            ))
        }
    }
    return matches
}
