// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Compares two sequences of text lines at the line level.
///
/// This is the main entry point of the ByLine comparison stage.
public func compareLines(_ lines1: [String], _ lines2: [String], policy: ComparisonPolicy) -> FairDiffIterable {
    let first = toLines(lines1, policy: policy)
    let second = toLines(lines2, policy: policy)
    return doCompareLines(first, second, policy: policy)
}

/// Line matcher backed by the Myers O(ND) algorithm with JetBrains' two-step refinement.
public struct MyersMatcher: LineMatcher {
    public init() {}

    public func match(_ left: [String], _ right: [String], policy: ComparisonPolicy) -> FairDiffIterable {
        compareLines(left, right, policy: policy)
    }
}

func doCompareLines(_ lines1: [Line], _ lines2: [Line], policy: ComparisonPolicy) -> FairDiffIterable {
    if policy == .ignoreWhitespaces {
        var changes = compareSmart(lines1, lines2)
        changes = optimizeLineChunks(lines1, lines2, changes)
        return expandRanges(lines1, lines2, changes)
    }

    let iwLines1 = convertMode(lines1, policy: .ignoreWhitespaces)
    let iwLines2 = convertMode(lines2, policy: .ignoreWhitespaces)

    var iwChanges = compareSmart(iwLines1, iwLines2)
    iwChanges = optimizeLineChunks(lines1, lines2, iwChanges)
    return correctChangesSecondStep(lines1, lines2, iwChanges)
}

private func compareSmart(_ lines1: [Line], _ lines2: [Line]) -> FairDiffIterable {
    let threshold = unimportantLineCharCount
    if threshold == 0 {
        return diffLineSlice(lines1, lines2)
    }
    let (bigLines1, indexes1) = getBigLines(lines1, threshold: threshold)
    let (bigLines2, indexes2) = getBigLines(lines2, threshold: threshold)
    let changes = diffLineSlice(bigLines1, bigLines2)
    return SmartLineChangeCorrector(
        indexes1: indexes1, indexes2: indexes2,
        lines1: lines1, lines2: lines2,
        changes: changes
    ).build()
}

private func getBigLines(_ lines: [Line], threshold: Int) -> ([Line], [Int]) {
    var big: [Line] = []
    var indexes: [Int] = []
    for (index, line) in lines.enumerated() where line.nonSpaceChars > threshold {
        big.append(line)
        indexes.append(index)
    }
    return (big, indexes)
}

private func expandRanges(_ lines1: [Line], _ lines2: [Line], _ iterable: FairDiffIterable) -> FairDiffIterable {
    var changes: [DiffRange] = []
    for change in iterable.changes {
        let expanded = expandRangeLines(
            lines1, lines2,
            change.left.lowerBound, change.right.lowerBound,
            change.left.upperBound, change.right.upperBound
        )
        if !expanded.isEmpty {
            changes.append(expanded)
        }
    }
    return createFromRanges(changes, length1: lines1.count, length2: lines2.count)
}

// swiftlint:disable function_body_length
private func correctChangesSecondStep(
    _ lines1: [Line],
    _ lines2: [Line],
    _ changes: FairDiffIterable
) -> FairDiffIterable {
    // swiftlint:enable function_body_length
    let eqLines1 = lines1.map { $0 as LineEquatable }
    let eqLines2 = lines2.map { $0 as LineEquatable }
    let builder = ExpandChangeBuilder(objects1: eqLines1, objects2: eqLines2)

    var sample: String?
    var last1 = 0
    var last2 = 0

    for unchanged in changes.unchanged {
        let count = unchanged.left.count
        for i in 0..<count {
            let idx1 = unchanged.left.lowerBound + i
            let idx2 = unchanged.right.lowerBound + i
            let line1 = lines1[idx1]
            let line2 = lines2[idx2]

            if sample == nil || !TextUtils.isEqual(sample ?? "", line1.content, policy: .ignoreWhitespaces) {
                if line1.equals(line2) {
                    flushSecondStep(
                        builder: builder,
                        lines1: lines1,
                        lines2: lines2,
                        sample: sample,
                        last1: last1,
                        last2: last2,
                        line1: idx1,
                        line2: idx2
                    )
                    sample = nil
                    builder.markEqual(idx1, idx2)
                } else {
                    flushSecondStep(
                        builder: builder,
                        lines1: lines1,
                        lines2: lines2,
                        sample: sample,
                        last1: last1,
                        last2: last2,
                        line1: idx1,
                        line2: idx2
                    )
                    sample = line1.content
                }
            }
            last1 = idx1 + 1
            last2 = idx2 + 1
        }
    }
    flushSecondStep(
        builder: builder, lines1: lines1, lines2: lines2,
        sample: sample, last1: last1, last2: last2,
        line1: changes.length1, line2: changes.length2
    )
    return fair(builder.finish())
}

// swiftlint:disable:next function_parameter_count
private func flushSecondStep(
    builder: ExpandChangeBuilder,
    lines1: [Line],
    lines2: [Line],
    sample: String?,
    last1: Int,
    last2: Int,
    line1: Int,
    line2: Int
) {
    guard let sample else {
        return
    }
    let start1 = max(last1, builder.index1)
    let start2 = max(last2, builder.index2)

    var sub1: [Int] = []
    var sub2: [Int] = []
    for i in start1..<line1 where TextUtils.isEqual(sample, lines1[i].content, policy: .ignoreWhitespaces) {
        sub1.append(i)
    }
    for i in start2..<line2 where TextUtils.isEqual(sample, lines2[i].content, policy: .ignoreWhitespaces) {
        sub2.append(i)
    }

    if sub1.isEmpty || sub2.isEmpty {
        return
    }
    alignExactMatching(builder: builder, lines1: lines1, lines2: lines2, sub1: sub1, sub2: sub2)
}

private func alignExactMatching(
    builder: ExpandChangeBuilder,
    lines1: [Line],
    lines2: [Line],
    sub1: [Int],
    sub2: [Int]
) {
    let size = max(sub1.count, sub2.count)
    let skipAligning = size > 10 || sub1.count == sub2.count

    if skipAligning {
        let count = min(sub1.count, sub2.count)
        for i in 0..<count where lines1[sub1[i]].equals(lines2[sub2[i]]) {
            builder.markEqual(sub1[i], sub2[i])
        }
        return
    }

    if sub1.count < sub2.count {
        let matching = getBestMatchingAlignment(shorter: sub1, longer: sub2, linesS: lines1, linesL: lines2)
        for i in 0..<sub1.count where lines1[sub1[i]].equals(lines2[sub2[matching[i]]]) {
            builder.markEqual(sub1[i], sub2[matching[i]])
        }
    } else {
        let matching = getBestMatchingAlignment(shorter: sub2, longer: sub1, linesS: lines2, linesL: lines1)
        for i in 0..<sub2.count where lines1[sub1[matching[i]]].equals(lines2[sub2[i]]) {
            builder.markEqual(sub1[matching[i]], sub2[i])
        }
    }
}

private func getBestMatchingAlignment(shorter: [Int], longer: [Int], linesS: [Line], linesL: [Line]) -> [Int] {
    let size = shorter.count
    var best = Array(0..<size)
    var comb = Array(repeating: 0, count: size)
    var bestWeight = 0

    func combinations(start: Int, end: Int, depth: Int) {
        if depth == size {
            var weight = 0
            for i in 0..<size where linesS[shorter[i]].equals(linesL[longer[comb[i]]]) {
                weight += 1
            }
            if weight > bestWeight {
                bestWeight = weight
                best = comb
            }
            return
        }
        var i = start
        while i <= end {
            comb[depth] = i
            combinations(start: i + 1, end: end, depth: depth + 1)
            i += 1
        }
    }
    combinations(start: 0, end: longer.count - 1, depth: 0)
    return best
}
