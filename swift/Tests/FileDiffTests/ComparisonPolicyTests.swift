import Testing
@testable import FileDiff

@Suite("TextUtils.isEqual")
struct TextUtilsIsEqualTests {
    @Test("default compares verbatim")
    func defaultPolicy() {
        #expect(TextUtils.isEqual("abc", "abc", policy: .default))
        #expect(!TextUtils.isEqual("abc", "ab c", policy: .default))
    }

    @Test("trimWhitespaces ignores leading/trailing spaces")
    func trim() {
        #expect(TextUtils.isEqual("  abc  ", "abc", policy: .trimWhitespaces))
        #expect(!TextUtils.isEqual("a b c", "abc", policy: .trimWhitespaces))
    }

    @Test("ignoreWhitespaces ignores every space")
    func ignore() {
        #expect(TextUtils.isEqual("a b c", "abc", policy: .ignoreWhitespaces))
        #expect(!TextUtils.isEqual("abc", "abd", policy: .ignoreWhitespaces))
    }
}

@Suite("TextUtils.hashCode")
struct TextUtilsHashCodeTests {
    @Test("trim and default agree on already-trimmed input")
    func trimAgrees() {
        let defaultHash = TextUtils.hashCode("abc", policy: .default)
        let trimHash = TextUtils.hashCode("  abc  ", policy: .trimWhitespaces)
        #expect(defaultHash == trimHash)
    }

    @Test("ignoreWhitespaces collapses spaces")
    func ignoreSpaces() {
        #expect(TextUtils.hashCode("a b c", policy: .ignoreWhitespaces)
            == TextUtils.hashCode("abc", policy: .ignoreWhitespaces))
    }
}

@Suite("CharacterUtils")
struct CharacterUtilsTests {
    @Test("classifies ASCII whitespace")
    func whitespace() {
        #expect(CharacterUtils.isSpaceEnterOrTab(" "))
        #expect(CharacterUtils.isSpaceEnterOrTab("\n"))
        #expect(CharacterUtils.isSpaceEnterOrTab("\t"))
        #expect(!CharacterUtils.isSpaceEnterOrTab("a"))
    }

    @Test("punctuation excludes underscore")
    func punctuation() {
        #expect(CharacterUtils.isPunctuation("."))
        #expect(!CharacterUtils.isPunctuation("_"))
        #expect(!CharacterUtils.isPunctuation("a"))
    }

    @Test("CJK characters are continuous script")
    func cjk() throws {
        #expect(try CharacterUtils.isContinuousScript(#require(Unicode.Scalar(0x4E2D))))
        #expect(!CharacterUtils.isContinuousScript("a"))
    }
}
