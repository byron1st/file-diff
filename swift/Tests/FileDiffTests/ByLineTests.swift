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

@Suite("compareLines second-step alignment")
struct CompareLinesSecondStepTests {
    // Default policy treats internal-whitespace differences as real changes,
    // but the first ignoreWhitespaces pass groups them. The second-step
    // refinement then collects "sample-equal" lines and decides whether
    // changed-region lines should be re-aligned.

    @Test("sample-equal line sitting inside a change is re-evaluated")
    func sampleAcrossChangeGap() {
        // The middle "  a" on the right side falls inside an inserted block
        // under the ignoreWhitespaces pass. The second-step inspects it and
        // sees that it is sample-equal to the left-side "  a", but since the
        // exact match still differs, it leaves the change in place.
        let left = split("X\n  a\nY")
        let right = split("X\n a\n  a\nY")
        let result = compareLines(left, right, policy: .default)
        #expect(!result.changes.isEmpty)
        let totalChange = result.changes.reduce(0) { $0 + $1.left.count + $1.right.count }
        #expect(totalChange >= 1)
    }

    @Test("symmetric sample blocks with equal counts trigger fast alignment")
    func sampleEqualCounts() {
        // Two ws-different "a" lines on each side under default policy.
        // The unchanged ignoreWS pairing is symmetric, so flushSecondStep
        // walks the sub1.count == sub2.count branch (skipAligning fast path).
        let left = split("X\n  a\nY\n  a\nZ")
        let right = split("X\n a\nY\n a\nZ")
        let result = compareLines(left, right, policy: .default)
        // Anchors X / Y / Z stay aligned; "a" lines remain as changes.
        #expect(result.changes.count >= 1)
    }

    @Test("trimWhitespaces resolves all changes for repeated whitespace-only lines")
    func trimResolvesRepeatedSamples() {
        let left = split(" same\nfoo\n same\n")
        let right = split("same\nfoo\nsame\n")
        let result = compareLines(left, right, policy: .trimWhitespaces)
        #expect(result.changes.isEmpty)
    }
}

@Suite("convertMode")
struct ConvertModeTests {
    @Test("same policy keeps existing line instances")
    func samePolicy() {
        let lines = toLines(["a", "b"], policy: .default)
        let converted = convertMode(lines, policy: .default)
        #expect(converted[0] === lines[0])
        #expect(converted[1] === lines[1])
    }

    @Test("different policy creates new line instances")
    func differentPolicy() {
        let lines = toLines(["  a", "b  "], policy: .default)
        let converted = convertMode(lines, policy: .ignoreWhitespaces)
        #expect(converted[0] !== lines[0])
        #expect(converted[1] !== lines[1])
        #expect(converted[0].policy == .ignoreWhitespaces)
        #expect(converted[1].policy == .ignoreWhitespaces)
        #expect(converted[0].content == "  a")
    }

    @Test("empty input returns empty array")
    func empty() {
        #expect(convertMode([], policy: .default).isEmpty)
    }
}
