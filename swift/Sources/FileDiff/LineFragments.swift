// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Compares two line sequences using `matcher` and refines each changed range
/// with word-level inner fragments.
///
/// The returned `LineFragment.leftOffsets` / `rightOffsets` use UTF-8 byte offsets
/// of the joined (newline-delimited) text for each side, and `inner` offsets are
/// relative to those ranges.
public func compareLineFragments(
    _ lines1: [String], _ lines2: [String],
    matcher: LineMatcher,
    policy: ComparisonPolicy
) -> [LineFragment] {
    let lineDiff = matcher.match(lines1, lines2, policy: policy)
    var result: [LineFragment] = []
    result.reserveCapacity(lineDiff.changes.count)

    for range in lineDiff.changes {
        let left = lines1[range.left].joined(separator: "\n")
        let right = lines2[range.right].joined(separator: "\n")
        let inner = compareWords(left, right, policy: policy)

        let leftByteCount = left.utf8.count
        let rightByteCount = right.utf8.count
        result.append(LineFragment(
            leftLines: range.left,
            rightLines: range.right,
            leftOffsets: 0..<leftByteCount,
            rightOffsets: 0..<rightByteCount,
            inner: inner
        ))
    }
    return result
}
