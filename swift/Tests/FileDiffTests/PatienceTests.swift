import Testing
@testable import FileDiff

@Suite("patienceDiff")
struct PatienceDiffTests {
    private static let noFallback: PatienceFallback = { _, _ in [] }

    @Test("identical sequences match every line")
    func identical() {
        let lines = ["a", "b", "c"]
        let matches = patienceDiff(lines, lines, fallback: Self.noFallback)
        #expect(matches.count == 3)
        for (index, match) in matches.enumerated() {
            #expect(match.leftIndex == index)
            #expect(match.rightIndex == index)
        }
    }

    @Test("single insertion keeps unique anchors")
    func singleInsert() {
        let matches = patienceDiff(["a", "c"], ["a", "b", "c"], fallback: Self.noFallback)
        #expect(matches.count == 2)
        #expect(matches[0] == PatienceAnchor(leftIndex: 0, rightIndex: 0))
        #expect(matches[1] == PatienceAnchor(leftIndex: 1, rightIndex: 2))
    }

    @Test("single deletion keeps unique anchors")
    func singleDelete() {
        let matches = patienceDiff(["a", "b", "c"], ["a", "c"], fallback: Self.noFallback)
        #expect(matches.count == 2)
        #expect(matches[0] == PatienceAnchor(leftIndex: 0, rightIndex: 0))
        #expect(matches[1] == PatienceAnchor(leftIndex: 2, rightIndex: 1))
    }

    @Test("empty inputs produce no matches")
    func emptyInputs() {
        #expect(patienceDiff([], [], fallback: Self.noFallback).isEmpty)
        #expect(patienceDiff(["a"], [], fallback: Self.noFallback).isEmpty)
    }

    @Test("duplicate lines fall back to external matcher")
    func duplicateLinesUseFallback() {
        let left = ["{", "x", "{", "x", "}"]
        let right = ["{", "y", "{", "y", "}"]
        let flag = CallFlag()
        let fallback: PatienceFallback = { subLeft, subRight in
            flag.value = true
            return [
                PatienceAnchor(leftIndex: 0, rightIndex: 0),
                PatienceAnchor(leftIndex: subLeft.count - 1, rightIndex: subRight.count - 1),
            ]
        }
        let matches = patienceDiff(left, right, fallback: fallback)
        #expect(flag.value)
        #expect(matches.count >= 2)
    }

    @Test("function move scenario anchors on unique signatures")
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
        let matches = patienceDiff(left, right, fallback: Self.noFallback)
        #expect(matches.count >= 2)
    }
}

final class CallFlag: @unchecked Sendable {
    var value = false
}

@Suite("longestIncreasingSubsequence")
struct LongestIncreasingSubsequenceTests {
    @Test("picks length-2 increasing chain")
    func basic() {
        let pairs = [
            PatienceAnchor(leftIndex: 0, rightIndex: 3),
            PatienceAnchor(leftIndex: 1, rightIndex: 1),
            PatienceAnchor(leftIndex: 2, rightIndex: 4),
            PatienceAnchor(leftIndex: 3, rightIndex: 2),
        ]
        let result = longestIncreasingSubsequence(pairs)
        #expect(result.count == 2)
        for i in 1..<result.count {
            #expect(result[i].rightIndex > result[i - 1].rightIndex)
        }
    }

    @Test("already sorted keeps everything")
    func alreadySorted() {
        let pairs = (0..<3).map { PatienceAnchor(leftIndex: $0, rightIndex: $0) }
        #expect(longestIncreasingSubsequence(pairs).count == 3)
    }

    @Test("reversed keeps one element")
    func reversed() {
        let pairs = [
            PatienceAnchor(leftIndex: 0, rightIndex: 2),
            PatienceAnchor(leftIndex: 1, rightIndex: 1),
            PatienceAnchor(leftIndex: 2, rightIndex: 0),
        ]
        #expect(longestIncreasingSubsequence(pairs).count == 1)
    }

    @Test("empty input returns empty")
    func empty() {
        #expect(longestIncreasingSubsequence([]).isEmpty)
    }
}
