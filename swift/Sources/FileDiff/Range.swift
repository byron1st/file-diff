// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// A pair of half-open intervals over two sequences.
///
/// `left` describes a range in the first sequence and `right` describes the corresponding
/// range in the second sequence. Ranges are half-open (`lowerBound ..< upperBound`).
public struct DiffRange: Hashable, Sendable, CustomStringConvertible {
    public let left: Range<Int>
    public let right: Range<Int>

    public init(left: Range<Int>, right: Range<Int>) {
        self.left = left
        self.right = right
    }

    /// Convenience initializer mirroring the Go API (`start1, end1, start2, end2`).
    public init(start1: Int, end1: Int, start2: Int, end2: Int) {
        precondition(
            start1 <= end1 && start2 <= end2,
            "invalid range: [\(start1), \(end1)) - [\(start2), \(end2))"
        )
        left = start1..<end1
        right = start2..<end2
    }

    public var isEmpty: Bool {
        left.isEmpty && right.isEmpty
    }

    public var description: String {
        "[\(left.lowerBound), \(left.upperBound)) - [\(right.lowerBound), \(right.upperBound))"
    }
}
