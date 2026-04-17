// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

import Foundation

/// Compares two texts at the Unicode code-point level and returns a diff
/// expressed in UTF-8 byte offsets.
public func compareChars(_ text1: String, _ text2: String) -> FairDiffIterable {
    let bytes1 = Array(text1.utf8)
    let bytes2 = Array(text2.utf8)
    let cp1 = getAllCodePoints(text1)
    let cp2 = getAllCodePoints(text2)

    let iterable = diff(cp1.codePoints, cp2.codePoints)
    let builder = ChangeBuilder(length1: bytes1.count, length2: bytes2.count)

    var offset1 = 0
    var offset2 = 0
    for entry in iterateAll(iterable) {
        let end1 = offset1 + countCharBytes(cp1, entry.range.left.lowerBound, entry.range.left.upperBound)
        let end2 = offset2 + countCharBytes(cp2, entry.range.right.lowerBound, entry.range.right.upperBound)
        if entry.equal {
            builder.markEqualRange(offset1, offset2, end1, end2)
        }
        offset1 = end1
        offset2 = end2
    }

    return fair(builder.finish())
}

/// Two-step character comparison: matches non-space characters first, then fills gaps.
public func compareCharsTwoStep(_ text1: String, _ text2: String) -> FairDiffIterable {
    let cp1 = getNonSpaceCodePoints(text1)
    let cp2 = getNonSpaceCodePoints(text2)
    let nonSpaceChanges = diff(cp1.codePoints, cp2.codePoints)
    return matchAdjustmentSpaces(cp1: cp1, cp2: cp2, text1: text1, text2: text2, changes: nonSpaceChanges)
}

/// Character-level comparison that trims whitespace from change boundaries.
public func compareCharsTrimWhitespaces(_ text1: String, _ text2: String) -> DiffIterable {
    let iterable = compareCharsTwoStep(text1, text2)
    let bytes1 = Array(text1.utf8)
    let bytes2 = Array(text2.utf8)
    return matchAdjustmentWhitespaces(
        bytes1: bytes1, bytes2: bytes2,
        iterable: iterable, policy: .trimWhitespaces
    )
}

/// Character-level comparison that ignores whitespace differences.
public func compareCharsIgnoreWhitespaces(_ text1: String, _ text2: String) -> DiffIterable {
    let bytes1 = Array(text1.utf8)
    let bytes2 = Array(text2.utf8)
    let cp1 = getNonSpaceCodePoints(text1)
    let cp2 = getNonSpaceCodePoints(text2)
    let changes = diff(cp1.codePoints, cp2.codePoints)
    return matchAdjustmentSpacesIW(
        cp1: cp1, cp2: cp2,
        bytes1: bytes1, bytes2: bytes2,
        changes: changes
    )
}

func comparePunctuation(_ text1: String, _ text2: String) -> FairDiffIterable {
    let bytes1 = Array(text1.utf8)
    let bytes2 = Array(text2.utf8)
    let chars1 = getPunctuationChars(text1)
    let chars2 = getPunctuationChars(text2)
    let nonSpaceChanges = diff(chars1.codePoints, chars2.codePoints)
    return transferPunctuation(
        chars1: chars1, chars2: chars2,
        bytes1: bytes1, bytes2: bytes2,
        changes: nonSpaceChanges
    )
}

// MARK: - Code point extraction

struct CodePointsOffsets {
    let codePoints: [Int]
    let offsets: [Int]
}

func charOffset(_ cp: CodePointsOffsets, _ index: Int) -> Int {
    cp.offsets[index]
}

func charOffsetAfter(_ cp: CodePointsOffsets, _ index: Int) -> Int {
    cp.offsets[index] + utf8LengthForCodePoint(cp.codePoints[index])
}

func utf8LengthForCodePoint(_ value: Int) -> Int {
    if value < 0x80 {
        return 1
    }
    if value < 0x800 {
        return 2
    }
    if value < 0x10000 {
        return 3
    }
    return 4
}

func getAllCodePoints(_ text: String) -> CodePointsOffsets {
    var points: [Int] = []
    var offsets: [Int] = []
    var offset = 0
    for scalar in text.unicodeScalars {
        points.append(Int(scalar.value))
        offsets.append(offset)
        offset += utf8Length(scalar)
    }
    return CodePointsOffsets(codePoints: points, offsets: offsets)
}

func getNonSpaceCodePoints(_ text: String) -> CodePointsOffsets {
    var points: [Int] = []
    var offsets: [Int] = []
    var offset = 0
    for scalar in text.unicodeScalars {
        let size = utf8Length(scalar)
        if !CharacterUtils.isWhiteSpaceCodePoint(scalar) {
            points.append(Int(scalar.value))
            offsets.append(offset)
        }
        offset += size
    }
    return CodePointsOffsets(codePoints: points, offsets: offsets)
}

func getPunctuationChars(_ text: String) -> CodePointsOffsets {
    var points: [Int] = []
    var offsets: [Int] = []
    var offset = 0
    for scalar in text.unicodeScalars {
        let size = utf8Length(scalar)
        if CharacterUtils.isPunctuation(scalar) {
            points.append(Int(scalar.value))
            offsets.append(offset)
        }
        offset += size
    }
    return CodePointsOffsets(codePoints: points, offsets: offsets)
}

func countCharBytes(_ cp: CodePointsOffsets, _ start: Int, _ end: Int) -> Int {
    var total = 0
    for i in start..<end {
        total += utf8LengthForCodePoint(cp.codePoints[i])
    }
    return total
}

// MARK: - Iterate all ranges

struct IterateAllEntry {
    let range: DiffRange
    let equal: Bool
}

func iterateAll(_ iterable: DiffIterable) -> [IterateAllEntry] {
    let changes = iterable.changes
    let unchanged = iterable.unchanged
    var result: [IterateAllEntry] = []
    var ci = 0
    var ui = 0
    while ci < changes.count || ui < unchanged.count {
        let takeChange: Bool
        if ci >= changes.count {
            takeChange = false
        } else if ui >= unchanged.count {
            takeChange = true
        } else {
            let change = changes[ci]
            let uc = unchanged[ui]
            if change.left.lowerBound < uc.left.lowerBound {
                takeChange = true
            } else if change.left.lowerBound == uc.left.lowerBound, change.right.lowerBound < uc.right.lowerBound {
                takeChange = true
            } else {
                takeChange = false
            }
        }
        if takeChange {
            result.append(IterateAllEntry(range: changes[ci], equal: false))
            ci += 1
        } else {
            result.append(IterateAllEntry(range: unchanged[ui], equal: true))
            ui += 1
        }
    }
    return result
}

// MARK: - Adjustment

func matchAdjustmentSpaces(
    cp1: CodePointsOffsets, cp2: CodePointsOffsets,
    text1: String, text2: String,
    changes: FairDiffIterable
) -> FairDiffIterable {
    DefaultCharChangeCorrector(
        cp1: cp1, cp2: cp2,
        text1: text1, text2: text2,
        changes: changes
    ).build()
}

func matchAdjustmentSpacesIW(
    cp1: CodePointsOffsets, cp2: CodePointsOffsets,
    bytes1: [UInt8], bytes2: [UInt8],
    changes: FairDiffIterable
) -> DiffIterable {
    var ranges: [DiffRange] = []
    for change in changes.changes {
        let startOffset1: Int
        let endOffset1: Int
        if change.left.lowerBound == change.left.upperBound {
            let value = expandForwardW(cp1: cp1, cp2: cp2, bytes1: bytes1, bytes2: bytes2, change: change, left: true)
            startOffset1 = value
            endOffset1 = value
        } else {
            startOffset1 = charOffset(cp1, change.left.lowerBound)
            endOffset1 = charOffsetAfter(cp1, change.left.upperBound - 1)
        }

        let startOffset2: Int
        let endOffset2: Int
        if change.right.lowerBound == change.right.upperBound {
            let value = expandForwardW(cp1: cp1, cp2: cp2, bytes1: bytes1, bytes2: bytes2, change: change, left: false)
            startOffset2 = value
            endOffset2 = value
        } else {
            startOffset2 = charOffset(cp2, change.right.lowerBound)
            endOffset2 = charOffsetAfter(cp2, change.right.upperBound - 1)
        }

        ranges.append(DiffRange(
            start1: startOffset1, end1: endOffset1,
            start2: startOffset2, end2: endOffset2
        ))
    }
    return createFromRanges(ranges, length1: bytes1.count, length2: bytes2.count)
}

// swiftlint:disable function_parameter_count
private func expandForwardW(
    cp1: CodePointsOffsets, cp2: CodePointsOffsets,
    bytes1: [UInt8], bytes2: [UInt8],
    change: DiffRange,
    left: Bool
) -> Int {
    // swiftlint:enable function_parameter_count
    let offset1 = change.left.lowerBound == 0 ? 0 : charOffsetAfter(cp1, change.left.lowerBound - 1)
    let offset2 = change.right.lowerBound == 0 ? 0 : charOffsetAfter(cp2, change.right.lowerBound - 1)
    let start = left ? offset1 : offset2
    let shift = TrimUtils.expandWhitespacesForward(
        bytes1, bytes2,
        offset1, offset2,
        bytes1.count, bytes2.count
    )
    return start + shift
}

func transferPunctuation(
    chars1: CodePointsOffsets, chars2: CodePointsOffsets,
    bytes1: [UInt8], bytes2: [UInt8],
    changes: FairDiffIterable
) -> FairDiffIterable {
    let builder = ChangeBuilder(length1: bytes1.count, length2: bytes2.count)
    for unchanged in changes.unchanged {
        let count = unchanged.left.count
        for i in 0..<count {
            let o1 = chars1.offsets[unchanged.left.lowerBound + i]
            let o2 = chars2.offsets[unchanged.right.lowerBound + i]
            builder.markEqual(o1, o2)
        }
    }
    return fair(builder.finish())
}

private final class DefaultCharChangeCorrector {
    private let cp1: CodePointsOffsets
    private let cp2: CodePointsOffsets
    private let text1: String
    private let text2: String
    private let bytes1: [UInt8]
    private let bytes2: [UInt8]
    private let changes: FairDiffIterable

    init(
        cp1: CodePointsOffsets, cp2: CodePointsOffsets,
        text1: String, text2: String,
        changes: FairDiffIterable
    ) {
        self.cp1 = cp1
        self.cp2 = cp2
        self.text1 = text1
        self.text2 = text2
        bytes1 = Array(text1.utf8)
        bytes2 = Array(text2.utf8)
        self.changes = changes
    }

    func build() -> FairDiffIterable {
        let builder = ChangeBuilder(length1: bytes1.count, length2: bytes2.count)
        var last1 = 0
        var last2 = 0
        for change in changes.unchanged {
            let count = change.left.count
            for i in 0..<count {
                let start1 = charOffset(cp1, change.left.lowerBound + i)
                let start2 = charOffset(cp2, change.right.lowerBound + i)
                let end1 = charOffsetAfter(cp1, change.left.lowerBound + i)
                let end2 = charOffsetAfter(cp2, change.right.lowerBound + i)

                matchGap(builder: builder, start1: last1, end1: start1, start2: last2, end2: start2)
                builder.markEqualRange(start1, start2, end1, end2)
                last1 = end1
                last2 = end2
            }
        }
        matchGap(builder: builder, start1: last1, end1: bytes1.count, start2: last2, end2: bytes2.count)
        return fair(builder.finish())
    }

    private func matchGap(builder: ChangeBuilder, start1: Int, end1: Int, start2: Int, end2: Int) {
        if start1 == end1, start2 == end2 {
            return
        }
        let inner1 = String(bytes: Array(bytes1[start1..<end1]), encoding: .utf8) ?? ""
        let inner2 = String(bytes: Array(bytes2[start2..<end2]), encoding: .utf8) ?? ""
        let inner = compareChars(inner1, inner2)
        for change in inner.unchanged {
            builder.markEqualCount(
                start1 + change.left.lowerBound,
                start2 + change.right.lowerBound,
                change.left.count
            )
        }
    }
}
