// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Adjusts chunk boundaries so that insertions and deletions are preferably
/// aligned with empty or otherwise "unimportant" lines.
func optimizeLineChunks(_ lines1: [Line], _ lines2: [Line], _ iterable: FairDiffIterable) -> FairDiffIterable {
    var ranges: [DiffRange] = []
    for range in iterable.unchanged {
        ranges.append(range)
        processLastRanges(lines1: lines1, lines2: lines2, ranges: &ranges)
    }
    return createUnchanged(ranges, length1: lines1.count, length2: lines2.count)
}

// swiftlint:disable function_body_length
private func processLastRanges(lines1: [Line], lines2: [Line], ranges: inout [DiffRange]) {
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

    let eqFwd = expandForwardLines(
        lines1,
        lines2,
        r1.left.upperBound,
        r1.right.upperBound,
        r1.left.upperBound + count2,
        r1.right.upperBound + count2
    )
    let eqBwd = expandBackwardLines(
        lines1,
        lines2,
        r2.left.lowerBound - count1,
        r2.right.lowerBound - count1,
        r2.left.lowerBound,
        r2.right.lowerBound
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
        processLastRanges(lines1: lines1, lines2: lines2, ranges: &ranges)
        return
    }
    if eqBwd == count1 {
        ranges.removeLast(2)
        ranges.append(DiffRange(
            start1: r2.left.lowerBound - count1, end1: r2.left.upperBound,
            start2: r2.right.lowerBound - count1, end2: r2.right.upperBound
        ))
        processLastRanges(lines1: lines1, lines2: lines2, ranges: &ranges)
        return
    }

    let touchSideIsLeft = r1.left.upperBound == r2.left.lowerBound
    let shift = getLineShift(
        lines1: lines1,
        lines2: lines2,
        touchSideIsLeft: touchSideIsLeft,
        eqFwd: eqFwd,
        eqBwd: eqBwd,
        r1: r1,
        r2: r2
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
private func getLineShift(
    lines1: [Line],
    lines2: [Line],
    touchSideIsLeft: Bool,
    eqFwd: Int,
    eqBwd: Int,
    r1: DiffRange,
    r2: DiffRange
) -> Int {
    // swiftlint:enable function_parameter_count
    let threshold = unimportantLineCharCount

    let touchLines: [Line]
    let touchStart: Int
    if touchSideIsLeft {
        touchLines = lines1
        touchStart = r2.left.lowerBound
    } else {
        touchLines = lines2
        touchStart = r2.right.lowerBound
    }

    if let shift = findBoundaryShift(touchLines, touchStart, eqFwd, eqBwd, 0) {
        return shift
    }

    let nonTouchLines: [Line]
    let changeStart: Int
    if touchSideIsLeft {
        nonTouchLines = lines2
        changeStart = r1.right.upperBound
    } else {
        nonTouchLines = lines1
        changeStart = r1.left.upperBound
    }
    if let shift = findBoundaryShift(nonTouchLines, changeStart, eqFwd, eqBwd, 0) {
        return shift
    }
    if let shift = findBoundaryShift(touchLines, touchStart, eqFwd, eqBwd, threshold) {
        return shift
    }
    if let shift = findBoundaryShift(nonTouchLines, changeStart, eqFwd, eqBwd, threshold) {
        return shift
    }
    return 0
}

private func findBoundaryShift(_ lines: [Line], _ offset: Int, _ eqFwd: Int, _ eqBwd: Int, _ threshold: Int) -> Int? {
    let fwd = findNextUnimportant(lines, offset, eqFwd + 1, threshold)
    let bwd = findPrevUnimportant(lines, offset - 1, eqBwd + 1, threshold)

    if fwd == -1, bwd == -1 {
        return nil
    }
    if fwd == 0 || bwd == 0 {
        return 0
    }
    if fwd != -1 {
        return fwd
    }
    return -bwd
}

private func findNextUnimportant(_ lines: [Line], _ offset: Int, _ count: Int, _ threshold: Int) -> Int {
    for i in 0..<count where lines[offset + i].nonSpaceChars <= threshold {
        return i
    }
    return -1
}

private func findPrevUnimportant(_ lines: [Line], _ offset: Int, _ count: Int, _ threshold: Int) -> Int {
    for i in 0..<count where lines[offset - i].nonSpaceChars <= threshold {
        return i
    }
    return -1
}

// swiftlint:disable function_parameter_count
private func expandForwardLines(
    _ lines1: [Line],
    _ lines2: [Line],
    _ start1: Int,
    _ start2: Int,
    _ end1: Int,
    _ end2: Int
) -> Int {
    // swiftlint:enable function_parameter_count
    var s1 = start1
    var s2 = start2
    while s1 < end1, s2 < end2, lines1[s1].equals(lines2[s2]) {
        s1 += 1
        s2 += 1
    }
    return s1 - start1
}

// swiftlint:disable function_parameter_count
private func expandBackwardLines(
    _ lines1: [Line],
    _ lines2: [Line],
    _ start1: Int,
    _ start2: Int,
    _ end1: Int,
    _ end2: Int
) -> Int {
    // swiftlint:enable function_parameter_count
    var e1 = end1
    var e2 = end2
    while start1 < e1, start2 < e2, lines1[e1 - 1].equals(lines2[e2 - 1]) {
        e1 -= 1
        e2 -= 1
    }
    return end1 - e1
}
