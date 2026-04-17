import Foundation
import Testing
@testable import FileDiff

@Suite("optimizeWordChunks via compareWords")
struct OptimizeWordChunksTests {
    private func chunkText(_ source: String, _ fragment: DiffFragment, side: Side) -> String {
        let bytes = Array(source.utf8)
        let range = side == .left ? fragment.left : fragment.right
        return String(bytes: Array(bytes[range]), encoding: .utf8) ?? ""
    }

    private enum Side { case left, right }

    @Test("change boundaries snap to whitespace separators")
    func snapToWhitespace() {
        // "bar" vs "qux" — the fragment should not bleed into surrounding spaces.
        let left = "foo bar baz"
        let right = "foo qux baz"
        let fragments = compareWords(left, right, policy: .default)
        #expect(fragments.count == 1)
        #expect(chunkText(left, fragments[0], side: .left) == "bar")
        #expect(chunkText(right, fragments[0], side: .right) == "qux")
    }

    @Test("insertion at the start of a line collapses to the inserted word only")
    func insertionAtLineStart() {
        // The boundary shift prefers the newline separator on the left side,
        // so the inserted word lands cleanly at the head of the new line.
        let left = "alpha\nbeta"
        let right = "alpha\ngamma beta"
        let fragments = compareWords(left, right, policy: .default)
        #expect(!fragments.isEmpty)
        let combined = fragments.map { chunkText(right, $0, side: .right) }.joined()
        #expect(combined.contains("gamma"))
    }

    @Test("trailing word insertion stays at the right boundary")
    func trailingInsertion() {
        let left = "alpha beta"
        let right = "alpha beta gamma"
        let fragments = compareWords(left, right, policy: .default)
        #expect(!fragments.isEmpty)
        let combined = fragments.map { chunkText(right, $0, side: .right) }.joined()
        #expect(combined.contains("gamma"))
    }

    @Test("repeated word run absorbed forward into a single change")
    func repeatedRunAbsorbed() {
        // The right side has an extra "x" before the trailing tail. The
        // optimizer should keep the suffix anchors aligned and emit one
        // tight insertion.
        let left = "a x x b"
        let right = "a x x x b"
        let fragments = compareWords(left, right, policy: .default)
        let total = fragments.reduce(0) { $0 + $1.right.count }
        // "x" plus an adjacent space — bounded by a single word's worth of bytes.
        #expect(total <= 4)
    }

    @Test("repeated word run absorbed backward into a single change")
    func repeatedRunAbsorbedBackward() {
        let left = "a x x b"
        let right = "x a x x b"
        let fragments = compareWords(left, right, policy: .default)
        let total = fragments.reduce(0) { $0 + $1.right.count }
        #expect(total <= 4)
    }

    @Test("change touching newline boundary stays separated")
    func newlineSeparator() {
        let left = "foo\nbar"
        let right = "FOO\nbar"
        let fragments = compareWords(left, right, policy: .default)
        #expect(fragments.count == 1)
        #expect(chunkText(left, fragments[0], side: .left) == "foo")
    }
}

@Suite("expandForwardChunks")
struct ExpandForwardChunksTests {
    @Test("matches all chunks when both ends align")
    func fullMatch() {
        let chunks = getInlineChunks("a b c")
        let result = expandForwardChunks(chunks, chunks, 0, 0, chunks.count, chunks.count)
        #expect(result == chunks.count)
    }

    @Test("stops at the first key mismatch")
    func mismatch() {
        let chunks1 = getInlineChunks("a b c")
        let chunks2 = getInlineChunks("a x c")
        let result = expandForwardChunks(chunks1, chunks2, 0, 0, chunks1.count, chunks2.count)
        #expect(result == 1)
    }

    @Test("returns zero when start equals end")
    func emptyRange() {
        let chunks = getInlineChunks("a b c")
        #expect(expandForwardChunks(chunks, chunks, 1, 1, 1, 1) == 0)
    }
}

@Suite("expandBackwardChunks")
struct ExpandBackwardChunksTests {
    @Test("matches all chunks when both ends align")
    func fullMatch() {
        let chunks = getInlineChunks("a b c")
        let result = expandBackwardChunks(chunks, chunks, 0, 0, chunks.count, chunks.count)
        #expect(result == chunks.count)
    }

    @Test("stops at the first key mismatch from the back")
    func mismatch() {
        let chunks1 = getInlineChunks("a b c")
        let chunks2 = getInlineChunks("a b X")
        let result = expandBackwardChunks(chunks1, chunks2, 0, 0, chunks1.count, chunks2.count)
        #expect(result == 0)
    }

    @Test("returns zero when start equals end")
    func emptyRange() {
        let chunks = getInlineChunks("a b c")
        #expect(expandBackwardChunks(chunks, chunks, 2, 2, 2, 2) == 0)
    }
}
