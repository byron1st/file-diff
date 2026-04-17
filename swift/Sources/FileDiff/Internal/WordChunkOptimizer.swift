// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Adjusts word chunk boundaries to prefer whitespace-separated positions.
func optimizeWordChunks(
    _ bytes1: [UInt8], _ bytes2: [UInt8],
    _ words1: [InlineChunk], _ words2: [InlineChunk],
    _ iterable: FairDiffIterable
) -> FairDiffIterable {
    var ranges: [DiffRange] = []
    for range in iterable.unchanged {
        ranges.append(range)
        processLastWordRanges(bytes1: bytes1, bytes2: bytes2, words1: words1, words2: words2, ranges: &ranges)
    }
    return createUnchanged(ranges, length1: words1.count, length2: words2.count)
}

// swiftlint:disable function_body_length
private func processLastWordRanges(
    bytes1: [UInt8],
    bytes2: [UInt8],
    words1: [InlineChunk],
    words2: [InlineChunk],
    ranges: inout [DiffRange]
) {
    // swiftlint:enable function_body_length
    if ranges.count < 2 {
        return
    }
    let r1 = ranges[ranges.count - 2]
    let r2 = ranges[ranges.count - 1]

    if r1.left.upperBound != r2.left.lowerBound, r1.right.upperBound != r2.right.lowerBound {
        return
    }

    let count1 = r1.left.count
    let count2 = r2.left.count

    let eqFwd = expandForwardChunks(
        words1, words2,
        r1.left.upperBound, r1.right.upperBound,
        r1.left.upperBound + count2, r1.right.upperBound + count2
    )
    let eqBwd = expandBackwardChunks(
        words1, words2,
        r2.left.lowerBound - count1, r2.right.lowerBound - count1,
        r2.left.lowerBound, r2.right.lowerBound
    )

    if eqFwd == 0, eqBwd == 0 {
        return
    }

    if eqFwd == count2 {
        ranges.removeLast(2)
        ranges.append(DiffRange(
            start1: r1.left.lowerBound, end1: r1.left.upperBound + count2,
            start2: r1.right.lowerBound, end2: r1.right.upperBound + count2
        ))
        processLastWordRanges(bytes1: bytes1, bytes2: bytes2, words1: words1, words2: words2, ranges: &ranges)
        return
    }
    if eqBwd == count1 {
        ranges.removeLast(2)
        ranges.append(DiffRange(
            start1: r2.left.lowerBound - count1, end1: r2.left.upperBound,
            start2: r2.right.lowerBound - count1, end2: r2.right.upperBound
        ))
        processLastWordRanges(bytes1: bytes1, bytes2: bytes2, words1: words1, words2: words2, ranges: &ranges)
        return
    }

    let touchSideIsLeft = r1.left.upperBound == r2.left.lowerBound
    let shift = getWordShift(
        bytes1: bytes1, bytes2: bytes2,
        words1: words1, words2: words2,
        touchSideIsLeft: touchSideIsLeft,
        eqFwd: eqFwd, eqBwd: eqBwd,
        r1: r1, r2: r2
    )
    if shift != 0 {
        ranges.removeLast(2)
        ranges.append(DiffRange(
            start1: r1.left.lowerBound, end1: r1.left.upperBound + shift,
            start2: r1.right.lowerBound, end2: r1.right.upperBound + shift
        ))
        ranges.append(DiffRange(
            start1: r2.left.lowerBound + shift, end1: r2.left.upperBound,
            start2: r2.right.lowerBound + shift, end2: r2.right.upperBound
        ))
    }
}

// swiftlint:disable function_parameter_count
private func getWordShift(
    bytes1: [UInt8],
    bytes2: [UInt8],
    words1: [InlineChunk],
    words2: [InlineChunk],
    touchSideIsLeft: Bool,
    eqFwd: Int,
    eqBwd: Int,
    r1: DiffRange,
    r2: DiffRange
) -> Int {
    // swiftlint:enable function_parameter_count
    let touchWords: [InlineChunk]
    let touchBytes: [UInt8]
    let touchStart: Int
    if touchSideIsLeft {
        touchWords = words1
        touchBytes = bytes1
        touchStart = r2.left.lowerBound
    } else {
        touchWords = words2
        touchBytes = bytes2
        touchStart = r2.right.lowerBound
    }

    if isSeparatedWithWhitespace(touchBytes, touchWords[touchStart - 1], touchWords[touchStart]) {
        return 0
    }

    if let shift = findSequenceEdgeShift(touchBytes, touchWords, touchStart, eqFwd, leftToRight: true), shift > 0 {
        return shift
    }
    if let shift = findSequenceEdgeShift(touchBytes, touchWords, touchStart - 1, eqBwd, leftToRight: false), shift > 0 {
        return -shift
    }
    return 0
}

private func findSequenceEdgeShift(
    _ bytes: [UInt8],
    _ words: [InlineChunk],
    _ offset: Int,
    _ count: Int,
    leftToRight: Bool
) -> Int? {
    for i in 0..<count {
        let w1: InlineChunk
        let w2: InlineChunk
        if leftToRight {
            w1 = words[offset + i]
            w2 = words[offset + i + 1]
        } else {
            w1 = words[offset - i - 1]
            w2 = words[offset - i]
        }
        if isSeparatedWithWhitespace(bytes, w1, w2) {
            return i + 1
        }
    }
    return nil
}

private func isSeparatedWithWhitespace(_ bytes: [UInt8], _ w1: InlineChunk, _ w2: InlineChunk) -> Bool {
    if w1.isNewline || w2.isNewline {
        return true
    }
    for i in w1.offset2..<w2.offset1 where TrimUtils.isSpaceEnterOrTab(bytes[i]) {
        return true
    }
    return false
}

// swiftlint:disable:next function_parameter_count
func expandForwardChunks(
    _ words1: [InlineChunk], _ words2: [InlineChunk],
    _ start1: Int, _ start2: Int, _ end1: Int, _ end2: Int
) -> Int {
    var s1 = start1
    var s2 = start2
    while s1 < end1, s2 < end2, words1[s1].key == words2[s2].key {
        s1 += 1
        s2 += 1
    }
    return s1 - start1
}

// swiftlint:disable:next function_parameter_count
func expandBackwardChunks(
    _ words1: [InlineChunk], _ words2: [InlineChunk],
    _ start1: Int, _ start2: Int, _ end1: Int, _ end2: Int
) -> Int {
    var e1 = end1
    var e2 = end2
    while start1 < e1, start2 < e2, words1[e1 - 1].key == words2[e2 - 1].key {
        e1 -= 1
        e2 -= 1
    }
    return end1 - e1
}
