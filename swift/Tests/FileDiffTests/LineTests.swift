import Testing
@testable import FileDiff

private final class OtherEquatable: LineEquatable {
    func equals(_: LineEquatable) -> Bool {
        false
    }
}

@Suite("Line.equals")
struct LineEqualsTests {
    @Test("identical instances are equal")
    func identity() {
        let line = Line(content: "hello", policy: .default)
        #expect(line.equals(line))
    }

    @Test("same content with same policy is equal")
    func sameContent() {
        let lhs = Line(content: "hello", policy: .default)
        let rhs = Line(content: "hello", policy: .default)
        #expect(lhs.equals(rhs))
    }

    @Test("different content with same hash slot still falls through to content check")
    func differentContent() {
        let lhs = Line(content: "hello", policy: .default)
        let rhs = Line(content: "world", policy: .default)
        #expect(!lhs.equals(rhs))
    }

    @Test("trimWhitespaces policy ignores edge whitespace")
    func trimPolicy() {
        let lhs = Line(content: "  hello  ", policy: .trimWhitespaces)
        let rhs = Line(content: "hello", policy: .trimWhitespaces)
        #expect(lhs.equals(rhs))
    }

    @Test("equals(LineEquatable) returns false for non-Line conformer")
    func nonLineConformer() {
        let line = Line(content: "hello", policy: .default)
        let other: LineEquatable = OtherEquatable()
        #expect(!line.equals(other))
    }

    @Test("equals(LineEquatable) dispatches to Line overload when given a Line")
    func lineConformer() {
        let lhs: LineEquatable = Line(content: "hello", policy: .default)
        let rhs: LineEquatable = Line(content: "hello", policy: .default)
        #expect(lhs.equals(rhs))
    }
}

@Suite("toLines")
struct ToLinesTests {
    @Test("each input line becomes a Line tagged with the supplied policy")
    func tagsPolicy() {
        let lines = toLines(["a", "b", "c"], policy: .ignoreWhitespaces)
        #expect(lines.count == 3)
        for line in lines {
            #expect(line.policy == .ignoreWhitespaces)
        }
        #expect(lines.map(\.content) == ["a", "b", "c"])
    }

    @Test("empty input produces no lines")
    func empty() {
        #expect(toLines([], policy: .default).isEmpty)
    }
}
