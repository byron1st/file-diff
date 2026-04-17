import Testing
@testable import FileDiff

@Suite("LineOffsets")
struct LineOffsetsTests {
    @Test("single line without trailing newline")
    func singleLine() {
        let offsets = LineOffsets(text: "hello")
        #expect(offsets.lineCount == 1)
        #expect(offsets.lineStart(at: 0) == 0)
        #expect(offsets.lineEnd(at: 0) == 5)
        #expect(offsets.textLength == 5)
    }

    @Test("multiple lines resolve to correct start/end")
    func multipleLines() {
        let offsets = LineOffsets(text: "aaa\nbbb\nccc")
        #expect(offsets.lineCount == 3)

        struct Expectation { let line: Int
            let start: Int
            let end: Int
        }
        let expectations = [
            Expectation(line: 0, start: 0, end: 3),
            Expectation(line: 1, start: 4, end: 7),
            Expectation(line: 2, start: 8, end: 11),
        ]
        for expectation in expectations {
            #expect(offsets.lineStart(at: expectation.line) == expectation.start)
            #expect(offsets.lineEnd(at: expectation.line) == expectation.end)
        }
    }

    @Test("lineNumber maps byte offsets to lines")
    func lineNumberLookup() {
        let offsets = LineOffsets(text: "ab\ncd\nef")
        let expectations: [(offset: Int, line: Int)] = [
            (0, 0), (1, 0),
            (3, 1), (4, 1),
            (6, 2), (8, 2), // last: textLength
        ]
        for expectation in expectations {
            #expect(offsets.lineNumber(at: expectation.offset) == expectation.line)
        }
    }

    @Test("lineEnd honors includeNewline flag")
    func lineEndWithNewline() {
        let offsets = LineOffsets(text: "ab\ncd\nef")
        #expect(offsets.lineEnd(at: 0, includeNewline: false) == 2)
        #expect(offsets.lineEnd(at: 0, includeNewline: true) == 3)
        // Last line has no trailing newline to include.
        #expect(offsets.lineEnd(at: 2, includeNewline: true) == 8)
    }

    @Test("empty text yields a single empty line")
    func emptyText() {
        let offsets = LineOffsets(text: "")
        #expect(offsets.textLength == 0)
        #expect(offsets.lineCount == 1)
        #expect(offsets.lineStart(at: 0) == 0)
        #expect(offsets.lineEnd(at: 0) == 0)
    }

    @Test("trailing newline creates a final empty line")
    func trailingNewlineAddsLine() {
        let offsets = LineOffsets(text: "a\nb\n")
        #expect(offsets.lineCount == 3)
        #expect(offsets.lineStart(at: 2) == 4)
        #expect(offsets.lineEnd(at: 2) == 4)
    }
}
