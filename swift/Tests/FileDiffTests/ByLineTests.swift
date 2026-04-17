import Testing
@testable import FileDiff

private func split(_ text: String) -> [String] {
    text.isEmpty ? [] : text.components(separatedBy: "\n")
}

@Suite("compareLines")
struct CompareLinesTests {
    @Test("identical input produces no changes")
    func identical() {
        let lines = ["aaa", "bbb", "ccc"]
        let result = compareLines(lines, lines, policy: .default)
        #expect(result.changes.isEmpty)
    }

    @Test("single insertion in the middle")
    func singleInsert() {
        let result = compareLines(split("aaa\nccc"), split("aaa\nbbb\nccc"), policy: .default)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left == 1..<1)
        #expect(change.right == 1..<2)
    }

    @Test("single deletion in the middle")
    func singleDelete() {
        let result = compareLines(split("aaa\nbbb\nccc"), split("aaa\nccc"), policy: .default)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left == 1..<2)
        #expect(change.right == 1..<1)
    }

    @Test("single-line modification")
    func modification() {
        let result = compareLines(split("aaa\nbbb\nccc"), split("aaa\nxxx\nccc"), policy: .default)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left == 1..<2)
        #expect(change.right == 1..<2)
    }

    @Test("multiple changes kept separate")
    func multiple() {
        let result = compareLines(split("a\nb\nc\nd\ne"), split("a\nB\nc\nD\ne"), policy: .default)
        #expect(result.changes.count == 2)
    }

    @Test("empty inputs produce no changes")
    func empty() {
        let result = compareLines([], [], policy: .default)
        #expect(result.changes.isEmpty)
    }

    @Test("one side empty produces a single deletion")
    func oneEmpty() {
        let result = compareLines(split("aaa\nbbb"), [], policy: .default)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left == 0..<2)
        #expect(change.right == 0..<0)
    }

    @Test("trimWhitespaces ignores padding")
    func trimWhitespaces() {
        let result = compareLines(split("  aaa  \n  bbb  "), split("aaa\nbbb"), policy: .trimWhitespaces)
        #expect(result.changes.isEmpty)
    }

    @Test("ignoreWhitespaces ignores internal spaces")
    func ignoreWhitespaces() {
        let result = compareLines(split("a b c\nd e f"), split("abc\ndef"), policy: .ignoreWhitespaces)
        #expect(result.changes.isEmpty)
    }

    @Test("trimWhitespaces finds real changes")
    func trimWhitespacesWithChange() {
        let result = compareLines(
            split("  aaa  \n  bbb  \n  ccc  "),
            split("aaa\nxxx\nccc"),
            policy: .trimWhitespaces
        )
        #expect(result.changes.count == 1)
    }

    @Test("small-line second-step path")
    func smallLinesSecondStep() {
        let result = compareLines(split("a\n{\nb\n}\nc"), split("a\n{\nx\n}\nc"), policy: .default)
        #expect(result.changes.count == 1)
    }

    @Test("large input with a single mid-file change")
    func largeFile() {
        let count = 200
        var left = [String](repeating: "", count: count)
        var right = [String](repeating: "", count: count)
        for i in 0..<count {
            let line = String(repeating: "x", count: i + 1)
            left[i] = line
            right[i] = line
        }
        right[100] = "CHANGED_LINE"
        let result = compareLines(left, right, policy: .default)
        #expect(result.changes.count == 1)
        #expect(result.changes[0].left == 100..<101)
    }

    @Test("repeated insertion collapses to a single change")
    func repeatedInsertion() {
        let left = ["alpha", "beta", "gamma"]
        let right = ["alpha", "beta", "beta", "gamma"]
        let result = compareLines(left, right, policy: .default)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left == 2..<2)
        #expect(change.right == 2..<3)
    }

    @Test("trimWhitespaces prefers exact alignment on repeated lines")
    func repeatedExactAlignment() {
        let left = ["same", "foo", " foo", "bar"]
        let right = ["same", " foo", "foo", "foo", "bar"]
        let result = compareLines(left, right, policy: .trimWhitespaces)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left == 3..<3)
        #expect(change.right == 3..<4)
    }
}

@Suite("MyersMatcher")
struct MyersMatcherTests {
    @Test("conforms to LineMatcher")
    func conformance() {
        let matcher: any LineMatcher = MyersMatcher()
        let result = matcher.match(["a", "b"], ["a", "c"], policy: .default)
        #expect(result.changes.count == 1)
    }
}
