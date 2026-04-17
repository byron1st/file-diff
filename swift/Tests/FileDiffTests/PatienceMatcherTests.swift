import Testing
@testable import FileDiff

@Suite("PatienceMatcher")
struct PatienceMatcherTests {
    @Test("identical input produces no changes")
    func identical() {
        let lines = ["a", "b", "c"]
        let result = PatienceMatcher().match(lines, lines, policy: .default)
        #expect(result.changes.isEmpty)
        #expect(result.unchanged.count == 1)
    }

    @Test("single insertion yields unchanged anchors")
    func singleInsert() {
        let result = PatienceMatcher().match(["a", "c"], ["a", "b", "c"], policy: .default)
        var matched = 0
        for range in result.unchanged {
            matched += range.left.count
        }
        #expect(matched == 2)
    }

    @Test("function move anchors via unique signatures")
    func functionMove() {
        let left = [
            "func foo() {",
            "  return 1",
            "}",
            "func bar() {",
            "  return 2",
            "}",
        ]
        let right = [
            "func bar() {",
            "  return 2",
            "}",
            "func foo() {",
            "  return 1",
            "}",
        ]
        let result = PatienceMatcher().match(left, right, policy: .default)
        var matched = 0
        for range in result.unchanged {
            matched += range.left.count
        }
        #expect(matched >= 2)
    }

    @Test("duplicate-only lines fall back to Myers")
    func duplicateFallback() {
        let left = ["{", "x", "{", "x", "}"]
        let right = ["{", "y", "{", "y", "}"]
        let result = PatienceMatcher().match(left, right, policy: .default)
        var matched = 0
        for range in result.unchanged {
            matched += range.left.count
        }
        #expect(matched >= 1)
    }

    @Test("trimWhitespaces matches lines differing only by edge whitespace")
    func trimPolicy() {
        let left = ["  hello  ", "world"]
        let right = ["hello", "world"]
        let result = PatienceMatcher().match(left, right, policy: .trimWhitespaces)
        var matched = 0
        for range in result.unchanged {
            matched += range.left.count
        }
        #expect(matched == 2)
    }

    @Test("ignoreWhitespaces matches lines with internal whitespace differences")
    func ignorePolicy() {
        let left = ["a  b", "c"]
        let right = ["a b", "c"]
        let result = PatienceMatcher().match(left, right, policy: .ignoreWhitespaces)
        var matched = 0
        for range in result.unchanged {
            matched += range.left.count
        }
        #expect(matched == 2)
    }

    @Test("empty inputs return no unchanged ranges")
    func empty() {
        let result = PatienceMatcher().match([], [], policy: .default)
        #expect(result.unchanged.isEmpty)
        #expect(result.changes.isEmpty)
    }
}
