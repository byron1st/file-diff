import Testing
@testable import FileDiff

private func bytes(_ string: String) -> [UInt8] {
    Array(string.utf8)
}

@Suite("TrimUtils.isSpaceEnterOrTab")
struct TrimUtilsIsSpaceEnterOrTabTests {
    @Test("recognizes space, newline, and tab")
    func recognized() {
        #expect(TrimUtils.isSpaceEnterOrTab(0x20))
        #expect(TrimUtils.isSpaceEnterOrTab(0x0A))
        #expect(TrimUtils.isSpaceEnterOrTab(0x09))
    }

    @Test("rejects other bytes")
    func rejected() {
        #expect(!TrimUtils.isSpaceEnterOrTab(0x00))
        #expect(!TrimUtils.isSpaceEnterOrTab(0x41))
        #expect(!TrimUtils.isSpaceEnterOrTab(0x0D))
    }
}

@Suite("TrimUtils.trimStartText / trimEndText")
struct TrimUtilsTrimTests {
    @Test("trimStartText skips leading whitespace")
    func trimStart() {
        let input = bytes("   hello")
        #expect(TrimUtils.trimStartText(input, 0, input.count) == 3)
    }

    @Test("trimStartText returns end when range is all whitespace")
    func trimStartAllWhitespace() {
        let input = bytes("   ")
        #expect(TrimUtils.trimStartText(input, 0, input.count) == input.count)
    }

    @Test("trimStartText is a no-op when start is non-whitespace")
    func trimStartNoop() {
        let input = bytes("hello")
        #expect(TrimUtils.trimStartText(input, 0, input.count) == 0)
    }

    @Test("trimEndText skips trailing whitespace")
    func trimEnd() {
        let input = bytes("hello   ")
        #expect(TrimUtils.trimEndText(input, 0, input.count) == 5)
    }

    @Test("trimEndText returns start when range is empty")
    func trimEndEmpty() {
        let input = bytes("hello")
        #expect(TrimUtils.trimEndText(input, 2, 2) == 2)
    }

    @Test("trimTextRange trims both sides")
    func trimRange() {
        let b1 = bytes("  abc  ")
        let b2 = bytes(" def ")
        let range = TrimUtils.trimTextRange(b1, b2, 0, 0, b1.count, b2.count)
        #expect(range.left == 2..<5)
        #expect(range.right == 1..<4)
    }
}

@Suite("TrimUtils.expandWhitespaces")
struct TrimUtilsExpandTests {
    @Test("forward expansion stops at first non-whitespace")
    func forwardStopsAtNonWhitespace() {
        let b1 = bytes("  ab")
        let b2 = bytes("  cd")
        #expect(TrimUtils.expandWhitespacesForward(b1, b2, 0, 0, b1.count, b2.count) == 2)
    }

    @Test("forward expansion stops at differing whitespace bytes")
    func forwardStopsOnMismatch() {
        let b1 = bytes(" \t")
        let b2 = bytes("  ")
        #expect(TrimUtils.expandWhitespacesForward(b1, b2, 0, 0, b1.count, b2.count) == 1)
    }

    @Test("backward expansion stops at first non-whitespace from the right")
    func backwardStopsAtNonWhitespace() {
        let b1 = bytes("ab  ")
        let b2 = bytes("cd  ")
        #expect(TrimUtils.expandWhitespacesBackward(b1, b2, 0, 0, b1.count, b2.count) == 2)
    }

    @Test("expandWhitespacesRange returns the inner range")
    func rangeShrinks() {
        let b1 = bytes("  ab  ")
        let b2 = bytes("  cd  ")
        let range = DiffRange(start1: 0, end1: b1.count, start2: 0, end2: b2.count)
        let expanded = TrimUtils.expandWhitespacesRange(b1, b2, range)
        #expect(expanded.left == 2..<4)
        #expect(expanded.right == 2..<4)
    }
}

@Suite("TrimUtils.isEqualTextRange")
struct TrimUtilsIsEqualTextRangeTests {
    @Test("identical bytes are equal")
    func identical() {
        let buf = bytes("abc")
        let range = DiffRange(start1: 0, end1: buf.count, start2: 0, end2: buf.count)
        #expect(TrimUtils.isEqualTextRange(buf, buf, range))
    }

    @Test("different lengths are not equal")
    func differentLengths() {
        let b1 = bytes("abc")
        let b2 = bytes("abcd")
        let range = DiffRange(start1: 0, end1: b1.count, start2: 0, end2: b2.count)
        #expect(!TrimUtils.isEqualTextRange(b1, b2, range))
    }

    @Test("single byte difference is detected")
    func differentBytes() {
        let b1 = bytes("abc")
        let b2 = bytes("axc")
        let range = DiffRange(start1: 0, end1: b1.count, start2: 0, end2: b2.count)
        #expect(!TrimUtils.isEqualTextRange(b1, b2, range))
    }
}

@Suite("TrimUtils.isEqualTextRangeIgnoreWhitespaces")
struct TrimUtilsIsEqualIgnoreWhitespacesTests {
    @Test("equal once whitespace is removed")
    func equalIgnoringWhitespace() {
        let b1 = bytes(" a b ")
        let b2 = bytes("ab")
        let range = DiffRange(start1: 0, end1: b1.count, start2: 0, end2: b2.count)
        #expect(TrimUtils.isEqualTextRangeIgnoreWhitespaces(b1, b2, range))
    }

    @Test("non-whitespace difference is detected")
    func differentNonWhitespace() {
        let b1 = bytes(" ab ")
        let b2 = bytes(" ax ")
        let range = DiffRange(start1: 0, end1: b1.count, start2: 0, end2: b2.count)
        #expect(!TrimUtils.isEqualTextRangeIgnoreWhitespaces(b1, b2, range))
    }

    @Test("trailing non-whitespace on the left side fails")
    func trailingExtraOnLeft() {
        let b1 = bytes("abc")
        let b2 = bytes("ab")
        let range = DiffRange(start1: 0, end1: b1.count, start2: 0, end2: b2.count)
        #expect(!TrimUtils.isEqualTextRangeIgnoreWhitespaces(b1, b2, range))
    }

    @Test("trailing non-whitespace on the right side fails")
    func trailingExtraOnRight() {
        let b1 = bytes("ab")
        let b2 = bytes("abc")
        let range = DiffRange(start1: 0, end1: b1.count, start2: 0, end2: b2.count)
        #expect(!TrimUtils.isEqualTextRangeIgnoreWhitespaces(b1, b2, range))
    }

    @Test("trailing whitespace on either side is permitted")
    func trailingWhitespaceAllowed() {
        let b1 = bytes("ab   ")
        let b2 = bytes("ab")
        let range = DiffRange(start1: 0, end1: b1.count, start2: 0, end2: b2.count)
        #expect(TrimUtils.isEqualTextRangeIgnoreWhitespaces(b1, b2, range))
    }
}

@Suite("TrimUtils.isLeadingTrailingSpace")
struct TrimUtilsIsLeadingTrailingSpaceTests {
    @Test("returns false for negative position")
    func negativePosition() {
        #expect(!TrimUtils.isLeadingTrailingSpace(bytes("abc"), -1))
    }

    @Test("returns false for out-of-range position")
    func outOfRange() {
        let buf = bytes("abc")
        #expect(!TrimUtils.isLeadingTrailingSpace(buf, buf.count))
    }

    @Test("returns false for non-whitespace position")
    func nonWhitespace() {
        let buf = bytes("a b")
        #expect(!TrimUtils.isLeadingTrailingSpace(buf, 0))
    }

    @Test("space at start of file is leading whitespace")
    func leadingAtStart() {
        let buf = bytes("  hello")
        #expect(TrimUtils.isLeadingTrailingSpace(buf, 0))
        #expect(TrimUtils.isLeadingTrailingSpace(buf, 1))
    }

    @Test("space after newline is leading whitespace")
    func leadingAfterNewline() {
        let buf = bytes("foo\n  bar")
        #expect(TrimUtils.isLeadingTrailingSpace(buf, 4))
    }

    @Test("space before newline is trailing whitespace")
    func trailingBeforeNewline() {
        let buf = bytes("foo  \nbar")
        #expect(TrimUtils.isLeadingTrailingSpace(buf, 3))
        #expect(TrimUtils.isLeadingTrailingSpace(buf, 4))
    }

    @Test("space at end of file is trailing whitespace")
    func trailingAtEnd() {
        let buf = bytes("hello  ")
        #expect(TrimUtils.isLeadingTrailingSpace(buf, buf.count - 1))
    }

    @Test("space surrounded by content on both sides is neither")
    func interiorSpace() {
        let buf = bytes("foo bar baz")
        #expect(!TrimUtils.isLeadingTrailingSpace(buf, 3))
    }
}
