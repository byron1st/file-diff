// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// A modified region at the word or character level.
///
/// `left` and `right` are UTF-8 byte-offset ranges within their respective texts.
public struct DiffFragment: Hashable, Sendable, CustomStringConvertible {
    public let left: Range<Int>
    public let right: Range<Int>

    public init(left: Range<Int>, right: Range<Int>) {
        precondition(!(left.isEmpty && right.isEmpty), "DiffFragment should not be empty: \(left) - \(right)")
        self.left = left
        self.right = right
    }

    public init(startOffset1: Int, endOffset1: Int, startOffset2: Int, endOffset2: Int) {
        precondition(
            startOffset1 <= endOffset1 && startOffset2 <= endOffset2,
            "DiffFragment is invalid: [\(startOffset1), \(endOffset1)) - [\(startOffset2), \(endOffset2))"
        )
        self.init(left: startOffset1..<endOffset1, right: startOffset2..<endOffset2)
    }

    public var description: String {
        "[\(left.lowerBound), \(left.upperBound)) - [\(right.lowerBound), \(right.upperBound))"
    }
}

/// A modified region at the line level, optionally refined by inner word/character fragments.
///
/// - `lines` / `offsets` describe the change in the first (`left`) and second (`right`)
///   texts as line indices and UTF-8 byte offsets respectively.
/// - `inner` holds sub-line fragments whose offsets are **relative to the start of this
///   line fragment**. An empty `inner` means either that no inner detail was computed
///   or that the whole fragment changed (see the initializer).
public struct LineFragment: Hashable, Sendable, CustomStringConvertible {
    public let leftLines: Range<Int>
    public let rightLines: Range<Int>
    public let leftOffsets: Range<Int>
    public let rightOffsets: Range<Int>
    public let inner: [DiffFragment]

    public init(
        leftLines: Range<Int>,
        rightLines: Range<Int>,
        leftOffsets: Range<Int>,
        rightOffsets: Range<Int>,
        inner: [DiffFragment] = []
    ) {
        precondition(
            !(leftLines.isEmpty && rightLines.isEmpty),
            "LineFragment should not be empty: lines \(leftLines) - \(rightLines)"
        )
        self.leftLines = leftLines
        self.rightLines = rightLines
        self.leftOffsets = leftOffsets
        self.rightOffsets = rightOffsets
        self.inner = Self.stripWholeChange(
            inner,
            leftLength: leftOffsets.count,
            rightLength: rightOffsets.count
        )
    }

    public var description: String {
        "Lines [\(leftLines.lowerBound), \(leftLines.upperBound)) - "
            + "[\(rightLines.lowerBound), \(rightLines.upperBound)); "
            + "Offsets [\(leftOffsets.lowerBound), \(leftOffsets.upperBound)) - "
            + "[\(rightOffsets.lowerBound), \(rightOffsets.upperBound)); "
            + "Inner \(inner.count)"
    }

    /// Drops a single inner fragment that spans the entire line fragment (it conveys no
    /// additional information beyond the line-level change itself).
    private static func stripWholeChange(
        _ fragments: [DiffFragment],
        leftLength: Int,
        rightLength: Int
    ) -> [DiffFragment] {
        guard fragments.count == 1 else {
            return fragments
        }
        let only = fragments[0]
        let coversLeft = only.left.lowerBound == 0 && only.left.upperBound == leftLength
        let coversRight = only.right.lowerBound == 0 && only.right.upperBound == rightLength
        return (coversLeft && coversRight) ? [] : fragments
    }
}
