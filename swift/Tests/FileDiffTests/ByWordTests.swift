import Foundation
import Testing
@testable import FileDiff

@Suite("getInlineChunks")
struct GetInlineChunksTests {
    @Test("simple words produce two word chunks")
    func simpleWords() {
        let chunks = getInlineChunks("hello world")
        #expect(chunks.count == 2)
        #expect(!chunks[0].isNewline)
        #expect(chunks[0].content == "hello")
        #expect(chunks[1].content == "world")
    }

    @Test("newlines produce newline chunks")
    func newlines() {
        let chunks = getInlineChunks("a\nb")
        #expect(chunks.count == 3)
        #expect(chunks[1].isNewline)
    }

    @Test("punctuation is not a chunk")
    func punctuation() {
        let chunks = getInlineChunks("a.b")
        #expect(chunks.count == 2)
    }

    @Test("underscore is word-part")
    func underscore() {
        let chunks = getInlineChunks("my_var")
        #expect(chunks.count == 1)
        #expect(chunks[0].content == "my_var")
    }

    @Test("CJK characters are individual single-char words")
    func cjk() {
        let chunks = getInlineChunks("中文")
        #expect(chunks.count == 2)
    }

    @Test("empty input yields no chunks")
    func empty() {
        #expect(getInlineChunks("").isEmpty)
    }

    @Test("newline chunk offsets")
    func newlineOffsets() {
        let chunks = getInlineChunks("ab\ncd")
        let newline = chunks.first(where: \.isNewline)
        #expect(newline?.offset1 == 2)
        #expect(newline?.offset2 == 3)
    }
}

@Suite("compareWords")
struct CompareWordsTests {
    @Test("identical text has no fragments")
    func identical() {
        let fragments = compareWords("hello world", "hello world", policy: .default)
        #expect(fragments.isEmpty)
    }

    @Test("both empty has no fragments")
    func bothEmpty() {
        #expect(compareWords("", "", policy: .default).isEmpty)
    }

    @Test("empty vs non-empty has one fragment")
    func emptyVsNonEmpty() {
        let fragments = compareWords("", "hello", policy: .default)
        #expect(fragments.count == 1)
    }

    @Test("single word change covers changed region")
    func singleWord() {
        let fragments = compareWords("hello world", "hello earth", policy: .default)
        #expect(!fragments.isEmpty)
        let first = fragments[0]
        #expect(first.left.lowerBound >= 6)
        #expect(first.left.upperBound <= 11)
    }

    @Test("variable spelling change stays within word region")
    func variableSpelling() {
        let fragments = compareWords("int coutner = 0;", "int counter = 0;", policy: .default)
        #expect(fragments.contains { $0.left.lowerBound >= 4 && $0.left.upperBound <= 11 })
    }

    @Test("bracket differences keep unchanged content short")
    func brackets() {
        let fragments = compareWords("(a + b)", "[a + b]", policy: .default)
        #expect(!fragments.isEmpty)
        let total = fragments.reduce(0) { $0 + $1.left.count + $1.right.count }
        #expect(total <= 8)
    }

    @Test("default policy treats whitespace change as diff")
    func defaultWhitespace() {
        let fragments = compareWords("a  b", "a b", policy: .default)
        #expect(!fragments.isEmpty)
    }

    @Test("ignore whitespaces sees equal strings")
    func ignoreWhitespaces() {
        let fragments = compareWords("a  b", "a b", policy: .ignoreWhitespaces)
        #expect(fragments.isEmpty)
    }

    @Test("trim whitespaces sees equal strings")
    func trimWhitespaces() {
        let fragments = compareWords("  hello  ", "hello", policy: .trimWhitespaces)
        #expect(fragments.isEmpty)
    }

    @Test("word insertion produces fragments")
    func insertion() {
        #expect(!compareWords("a b", "a x b", policy: .default).isEmpty)
    }

    @Test("word deletion produces fragments")
    func deletion() {
        #expect(!compareWords("a b c", "a c", policy: .default).isEmpty)
    }

    @Test("multiple changes preserve matching anchors")
    func multiple() {
        let left = "foo bar baz"
        let fragments = compareWords(left, "foo qux baz", policy: .default)
        for fragment in fragments {
            let utf8 = Array(left.utf8)
            let chunk = String(bytes: Array(utf8[fragment.left]), encoding: .utf8) ?? ""
            #expect(chunk != "foo" && chunk != "baz")
        }
    }

    @Test("multi-line word change")
    func multiLine() {
        #expect(!compareWords("hello\nworld", "hello\nearth", policy: .default).isEmpty)
    }

    @Test("UTF-8 byte offsets are preserved")
    func utf8Offsets() {
        let left = "go 한글 test"
        let right = "go 한 test"
        let fragments = compareWords(left, right, policy: .default)
        #expect(fragments.count == 1)
        let fragment = fragments[0]
        let leftBytes = Array(left.utf8)
        let rightBytes = Array(right.utf8)
        #expect(String(bytes: Array(leftBytes[fragment.left]), encoding: .utf8) == "한글")
        #expect(String(bytes: Array(rightBytes[fragment.right]), encoding: .utf8) == "한")
    }

    @Test("repeated word with punctuation insertion")
    func repeatedWordPunctuation() {
        let right = "foo(bar)(bar) baz"
        let fragments = compareWords("foo(bar) baz", right, policy: .default)
        #expect(fragments.count == 1)
        let bytes = Array(right.utf8)
        #expect(String(bytes: Array(bytes[fragments[0].right]), encoding: .utf8) == "(bar)")
    }
}
