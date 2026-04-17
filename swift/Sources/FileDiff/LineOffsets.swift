// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Maps between line numbers and UTF-8 byte offsets within a text.
///
/// All offsets are UTF-8 byte offsets, matching the byte-oriented semantics used by the
/// companion Go port. Consumers that want to slice the original `String` by these offsets
/// should use `String.utf8` (for example, `String(text.utf8[start ..< end])`).
public struct LineOffsets: Sendable {
    /// Offset of the `\n` that ends each line, or the text length for the final line.
    private let lineEnds: [Int]
    public let textLength: Int

    public init(text: String) {
        var ends: [Int] = []
        var offset = 0
        for byte in text.utf8 {
            if byte == 0x0A { // '\n'
                ends.append(offset)
            }
            offset += 1
        }
        // Always append a terminal entry so the last line has an explicit end offset.
        ends.append(offset)
        lineEnds = ends
        textLength = offset
    }

    public var lineCount: Int {
        lineEnds.count
    }

    /// Byte offset where the given line starts.
    public func lineStart(at line: Int) -> Int {
        checkLineIndex(line)
        return line == 0 ? 0 : lineEnds[line - 1] + 1
    }

    /// Byte offset of the line terminator for `line` (i.e., the exclusive end of the content).
    public func lineEnd(at line: Int) -> Int {
        checkLineIndex(line)
        return lineEnds[line]
    }

    /// Byte offset past the line content.
    ///
    /// - Parameter includeNewline: when `true`, includes the trailing `\n`. Has no effect
    ///   for the last line, which has no trailing newline in this representation.
    public func lineEnd(at line: Int, includeNewline: Bool) -> Int {
        checkLineIndex(line)
        var end = lineEnds[line]
        if includeNewline, line != lineEnds.count - 1 {
            end += 1
        }
        return end
    }

    /// Line index that contains the given byte offset.
    public func lineNumber(at offset: Int) -> Int {
        precondition(
            offset >= 0 && offset <= textLength,
            "wrong offset: \(offset), text length: \(textLength)"
        )
        if offset == 0 {
            return 0
        }
        if offset == textLength {
            return lineCount - 1
        }
        return lowerBound(of: offset, in: lineEnds)
    }

    private func checkLineIndex(_ line: Int) {
        precondition(
            line >= 0 && line < lineCount,
            "wrong line: \(line), line count: \(lineCount)"
        )
    }

    /// Returns the first index `i` in `sorted` such that `sorted[i] >= target`.
    private func lowerBound(of target: Int, in sorted: [Int]) -> Int {
        var lo = 0
        var hi = sorted.count
        while lo < hi {
            let mid = (lo + hi) / 2
            if sorted[mid] < target {
                lo = mid + 1
            } else {
                hi = mid
            }
        }
        return lo
    }
}
