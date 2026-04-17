import Testing
@testable import FileDiff

@Suite("histogramDiff")
struct HistogramDiffTests {
    private static let simpleFallback: HistogramFallback = { left, right in
        var result: [HistogramAnchor] = []
        var used = Array(repeating: false, count: right.count)
        for (leftIdx, leftLine) in left.enumerated() {
            for (rightIdx, rightLine) in right.enumerated() where !used[rightIdx] && leftLine == rightLine {
                result.append(HistogramAnchor(leftIndex: leftIdx, rightIndex: rightIdx))
                used[rightIdx] = true
                break
            }
        }
        return result
    }

    @Test("identical sequences match every line")
    func identical() {
        let lines = ["a", "b", "c"]
        let result = histogramDiff(lines, lines, fallback: Self.simpleFallback)
        #expect(result.count == 3)
        for (index, anchor) in result.enumerated() {
            #expect(anchor.leftIndex == index)
            #expect(anchor.rightIndex == index)
        }
    }

    @Test("empty inputs return no matches")
    func empty() {
        #expect(histogramDiff([], [], fallback: Self.simpleFallback).isEmpty)
        #expect(histogramDiff(["a"], [], fallback: Self.simpleFallback).isEmpty)
    }

    @Test("no common lines yield no anchors")
    func noCommon() {
        #expect(histogramDiff(["a", "b"], ["x", "y"], fallback: Self.simpleFallback).isEmpty)
    }

    @Test("single insert preserves anchors")
    func singleInsert() {
        let result = histogramDiff(["a", "c"], ["a", "b", "c"], fallback: Self.simpleFallback)
        #expect(result.count == 2)
        #expect(result[0] == HistogramAnchor(leftIndex: 0, rightIndex: 0))
        #expect(result[1] == HistogramAnchor(leftIndex: 1, rightIndex: 2))
    }

    @Test("repetitive lines anchor on unique content")
    func repetitiveLines() {
        let left = ["{", "  a: 1", "}", "{", "  b: 2", "}"]
        let right = ["{", "  a: 1", "}", "{", "  c: 3", "}", "{", "  b: 2", "}"]
        let result = histogramDiff(left, right, fallback: Self.simpleFallback)
        let matchedLines = Set(result.map { left[$0.leftIndex] })
        #expect(matchedLines.contains("  a: 1"))
        #expect(matchedLines.contains("  b: 2"))
    }

    @Test("JSON-like structure anchors on unique keys")
    func jsonLike() {
        let left = [
            "{",
            #"  "name": "alice""#,
            #"  "age": 30"#,
            "}",
        ]
        let right = [
            "{",
            #"  "name": "alice""#,
            #"  "email": "alice@example.com""#,
            #"  "age": 30"#,
            "}",
        ]
        let result = histogramDiff(left, right, fallback: Self.simpleFallback)
        let matchedLines = Set(result.map { left[$0.leftIndex] })
        #expect(matchedLines.contains(#"  "name": "alice""#))
        #expect(matchedLines.contains(#"  "age": 30"#))
    }

    @Test("all-repeated lines still find lowest-frequency anchor")
    func allRepeated() {
        let left = ["x", "x", "y", "x"]
        let right = ["x", "y", "x", "x"]
        let result = histogramDiff(left, right, fallback: Self.simpleFallback)
        let yMatch = result.first { left[$0.leftIndex] == "y" }
        #expect(yMatch != nil)
        #expect(yMatch?.rightIndex == 1)
    }
}

@Suite("HistogramMatcher")
struct HistogramMatcherTests {
    @Test("identical input has no changes")
    func identical() {
        let lines = ["a", "b", "c"]
        let result = HistogramMatcher().match(lines, lines, policy: .default)
        #expect(result.changes.isEmpty)
    }

    @Test("repetitive insertion matched via histogram anchors")
    func repetitiveInsertion() {
        let left = ["{", "  a: 1", "}", "{", "  b: 2", "}"]
        let right = ["{", "  a: 1", "}", "{", "  c: 3", "}", "{", "  b: 2", "}"]
        let result = HistogramMatcher().match(left, right, policy: .default)
        var matched = 0
        for range in result.unchanged {
            matched += range.left.count
        }
        #expect(matched >= 2)
    }

    @Test("ignoreWhitespaces normalizes before matching")
    func ignorePolicy() {
        let left = ["a  b", "c"]
        let right = ["a b", "c"]
        let result = HistogramMatcher().match(left, right, policy: .ignoreWhitespaces)
        var matched = 0
        for range in result.unchanged {
            matched += range.left.count
        }
        #expect(matched == 2)
    }

    @Test("empty inputs produce no ranges")
    func empty() {
        let result = HistogramMatcher().match([], [], policy: .default)
        #expect(result.unchanged.isEmpty)
        #expect(result.changes.isEmpty)
    }

    @Test("one side empty yields a single change spanning the other side")
    func oneSideEmpty() {
        let result = HistogramMatcher().match(["a", "b", "c"], [], policy: .default)
        #expect(result.changes.count == 1)
        #expect(result.changes[0].left == 0..<3)
        #expect(result.changes[0].right == 0..<0)
    }

    @Test("trimWhitespaces matches lines differing only by edge whitespace")
    func trimPolicy() {
        let result = HistogramMatcher().match(["  hello  ", "world"], ["hello", "world"], policy: .trimWhitespaces)
        var matched = 0
        for range in result.unchanged {
            matched += range.left.count
        }
        #expect(matched == 2)
    }

    @Test("conforms to LineMatcher")
    func conformance() {
        let matcher: any LineMatcher = HistogramMatcher()
        let result = matcher.match(["a", "b"], ["a", "c"], policy: .default)
        #expect(result.changes.count == 1)
    }

    @Test("duplicate-only sub-range exercises Myers fallback")
    func myersFallback() {
        // Sandwich a duplicate-only sub-range between unique anchors so
        // histogramDiff recurses into the fallback closure for the middle.
        let left = ["start", "x", "x", "y", "y", "end"]
        let right = ["start", "x", "x", "y", "y", "end"]
        let result = HistogramMatcher().match(left, right, policy: .default)
        #expect(result.changes.isEmpty)
        var matched = 0
        for range in result.unchanged {
            matched += range.left.count
        }
        #expect(matched == 6)
    }
}

@Suite("histogramMyersFallback")
struct HistogramMyersFallbackTests {
    @Test("empty left returns no anchors")
    func emptyLeft() {
        #expect(histogramMyersFallback([], ["a"]).isEmpty)
    }

    @Test("empty right returns no anchors")
    func emptyRight() {
        #expect(histogramMyersFallback(["a"], []).isEmpty)
    }

    @Test("identical sub-range maps every line to its own index")
    func identical() {
        let anchors = histogramMyersFallback(["a", "b", "c"], ["a", "b", "c"])
        #expect(anchors == [
            HistogramAnchor(leftIndex: 0, rightIndex: 0),
            HistogramAnchor(leftIndex: 1, rightIndex: 1),
            HistogramAnchor(leftIndex: 2, rightIndex: 2),
        ])
    }

    @Test("disjoint sub-ranges produce no anchors")
    func disjoint() {
        #expect(histogramMyersFallback(["a"], ["b"]).isEmpty)
    }

    @Test("partial overlap returns matching positions only")
    func partialOverlap() {
        let anchors = histogramMyersFallback(["a", "x", "c"], ["a", "y", "c"])
        #expect(anchors.contains(HistogramAnchor(leftIndex: 0, rightIndex: 0)))
        #expect(anchors.contains(HistogramAnchor(leftIndex: 2, rightIndex: 2)))
    }
}
