import Testing
@testable import FileDiff

private func range(_ iterable: FairDiffIterable, at index: Int) -> DiffRange {
    iterable.changes[index]
}

@Suite("compareChars")
struct CompareCharsTests {
    @Test("identical strings yield no changes")
    func identical() {
        let result = compareChars("abc", "abc")
        #expect(result.changes.isEmpty)
    }

    @Test("single-char change")
    func singleCharChange() {
        let result = compareChars("abc", "axc")
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left == 1..<2)
        #expect(change.right == 1..<2)
    }

    @Test("insertion")
    func insertion() {
        let result = compareChars("ac", "abc")
        #expect(result.changes.count == 1)
        #expect(result.changes[0].left == 1..<1)
        #expect(result.changes[0].right == 1..<2)
    }

    @Test("deletion")
    func deletion() {
        let result = compareChars("abc", "ac")
        #expect(result.changes.count == 1)
        #expect(result.changes[0].left == 1..<2)
        #expect(result.changes[0].right == 1..<1)
    }

    @Test("completely different")
    func completelyDifferent() {
        let result = compareChars("abc", "xyz")
        #expect(result.changes.count == 1)
        #expect(result.changes[0].left == 0..<3)
        #expect(result.changes[0].right == 0..<3)
    }

    @Test("empty vs non-empty")
    func empty() {
        let result = compareChars("", "abc")
        #expect(result.changes.count == 1)
        #expect(result.changes[0].left == 0..<0)
        #expect(result.changes[0].right == 0..<3)
    }

    @Test("both empty")
    func bothEmpty() {
        #expect(compareChars("", "").changes.isEmpty)
    }

    @Test("multi-byte rune change uses UTF-8 byte offsets")
    func multiByte() {
        let result = compareChars("a한b", "a😀b")
        #expect(result.changes.count == 1)
        let expectedLeft = 1..<("a한".utf8.count)
        let expectedRight = 1..<("a😀".utf8.count)
        #expect(result.changes[0].left == expectedLeft)
        #expect(result.changes[0].right == expectedRight)
    }
}

@Suite("compareCharsTwoStep")
struct CompareCharsTwoStepTests {
    @Test("space difference leaves non-space chars matched")
    func spaceDifference() {
        let result = compareCharsTwoStep("a b", "a  b")
        #expect(result.unchanged.count >= 2)
    }
}

@Suite("compareCharsTrimWhitespaces")
struct CompareCharsTrimTests {
    @Test("surrounding whitespace is ignored")
    func surroundingWhitespace() {
        #expect(compareCharsTrimWhitespaces("  hello  ", "hello").changes.isEmpty)
    }

    @Test("real difference produces changes")
    func realDifference() {
        #expect(!compareCharsTrimWhitespaces("  abc  ", "  axc  ").changes.isEmpty)
    }

    @Test("UTF-8 leading/trailing spaces are trimmed")
    func utf8Trim() {
        #expect(compareCharsTrimWhitespaces("  한글  ", "한글").changes.isEmpty)
    }
}

@Suite("compareCharsIgnoreWhitespaces")
struct CompareCharsIgnoreTests {
    @Test("same content with different whitespace has no changes")
    func sameContent() {
        #expect(compareCharsIgnoreWhitespaces("a b c", "a  b  c").changes.isEmpty)
    }

    @Test("real char difference produces change")
    func realDifference() {
        #expect(compareCharsIgnoreWhitespaces("a x c", "a y c").changes.count == 1)
    }

    @Test("both empty has no changes")
    func bothEmpty() {
        #expect(compareCharsIgnoreWhitespaces("", "").changes.isEmpty)
    }

    @Test("only spaces has no changes")
    func onlySpaces() {
        #expect(compareCharsIgnoreWhitespaces("   ", "  ").changes.isEmpty)
    }

    @Test("insertion range reflects trimmed boundaries")
    func insertion() {
        let result = compareCharsIgnoreWhitespaces("ab", "a x b")
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left == 1..<1)
        #expect(change.right == 2..<3)
    }
}

@Suite("comparePunctuation")
struct ComparePunctuationTests {
    @Test("brackets differ but '+' matches")
    func brackets() {
        let result = comparePunctuation("(a + b)", "[a + b]")
        #expect(result.unchanged.contains { $0.left.count == 1 })
    }
}

@Suite("code point extraction")
struct CodePointTests {
    @Test("getAllCodePoints")
    func allCodePoints() {
        let cp = getAllCodePoints("hello")
        #expect(cp.codePoints.count == 5)
        #expect(cp.codePoints.first == Int(Character("h").asciiValue ?? 0))
    }

    @Test("getNonSpaceCodePoints")
    func nonSpace() {
        let cp = getNonSpaceCodePoints("a b c")
        #expect(cp.codePoints.count == 3)
        #expect(cp.offsets == [0, 2, 4])
    }

    @Test("getPunctuationChars")
    func punctuation() {
        let cp = getPunctuationChars("a.b;c")
        #expect(cp.codePoints.count == 2)
    }
}
