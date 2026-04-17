// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Refines word-level change ranges according to the supplied whitespace policy.
func matchAdjustmentWhitespaces(
    bytes1: [UInt8], bytes2: [UInt8],
    iterable: FairDiffIterable,
    policy: ComparisonPolicy
) -> DiffIterable {
    switch policy {
    case .trimWhitespaces:
        let defaulted = buildDefaultCorrector(iterable, bytes1: bytes1, bytes2: bytes2)
        return buildTrimCorrector(defaulted, bytes1: bytes1, bytes2: bytes2)
    case .ignoreWhitespaces:
        return buildIgnoreCorrector(iterable, bytes1: bytes1, bytes2: bytes2)
    case .default:
        return buildDefaultCorrector(iterable, bytes1: bytes1, bytes2: bytes2)
    }
}

private func buildDefaultCorrector(
    _ iterable: DiffIterable, bytes1: [UInt8], bytes2: [UInt8]
) -> DiffIterable {
    var changes: [DiffRange] = []
    for range in iterable.changes {
        let endCut = TrimUtils.expandWhitespacesBackward(
            bytes1, bytes2,
            range.left.lowerBound, range.right.lowerBound,
            range.left.upperBound, range.right.upperBound
        )
        let startCut = TrimUtils.expandWhitespacesForward(
            bytes1, bytes2,
            range.left.lowerBound, range.right.lowerBound,
            range.left.upperBound - endCut, range.right.upperBound - endCut
        )

        let expanded = DiffRange(
            start1: range.left.lowerBound + startCut,
            end1: range.left.upperBound - endCut,
            start2: range.right.lowerBound + startCut,
            end2: range.right.upperBound - endCut
        )
        if !expanded.isEmpty {
            changes.append(expanded)
        }
    }
    return createFromRanges(changes, length1: bytes1.count, length2: bytes2.count)
}

private func buildTrimCorrector(
    _ iterable: DiffIterable, bytes1: [UInt8], bytes2: [UInt8]
) -> DiffIterable {
    var changes: [DiffRange] = []
    for range in iterable.changes {
        var start1 = range.left.lowerBound
        var start2 = range.right.lowerBound
        var end1 = range.left.upperBound
        var end2 = range.right.upperBound

        if TrimUtils.isLeadingTrailingSpace(bytes1, start1) {
            start1 = TrimUtils.trimStartText(bytes1, start1, end1)
        }
        if TrimUtils.isLeadingTrailingSpace(bytes1, end1 - 1) {
            end1 = TrimUtils.trimEndText(bytes1, start1, end1)
        }
        if TrimUtils.isLeadingTrailingSpace(bytes2, start2) {
            start2 = TrimUtils.trimStartText(bytes2, start2, end2)
        }
        if TrimUtils.isLeadingTrailingSpace(bytes2, end2 - 1) {
            end2 = TrimUtils.trimEndText(bytes2, start2, end2)
        }

        let trimmed = DiffRange(start1: start1, end1: end1, start2: start2, end2: end2)
        if !trimmed.isEmpty, !TrimUtils.isEqualTextRange(bytes1, bytes2, trimmed) {
            changes.append(trimmed)
        }
    }
    return createFromRanges(changes, length1: bytes1.count, length2: bytes2.count)
}

private func buildIgnoreCorrector(
    _ iterable: DiffIterable, bytes1: [UInt8], bytes2: [UInt8]
) -> DiffIterable {
    var changes: [DiffRange] = []
    for range in iterable.changes {
        let expanded = TrimUtils.expandWhitespacesRange(bytes1, bytes2, range)
        let trimmed = TrimUtils.trimTextRange(
            bytes1, bytes2,
            expanded.left.lowerBound, expanded.right.lowerBound,
            expanded.left.upperBound, expanded.right.upperBound
        )
        if !trimmed.isEmpty, !TrimUtils.isEqualTextRangeIgnoreWhitespaces(bytes1, bytes2, trimmed) {
            changes.append(trimmed)
        }
    }
    return createFromRanges(changes, length1: bytes1.count, length2: bytes2.count)
}
