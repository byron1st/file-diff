import Testing
@testable import FileDiff

@Suite("DiffRange")
struct DiffRangeTests {
    @Test("stores half-open intervals for both sequences")
    func storesBounds() {
        let range = DiffRange(start1: 1, end1: 3, start2: 2, end2: 5)
        #expect(range.left == 1..<3)
        #expect(range.right == 2..<5)
    }

    @Test("empty range has matching empty halves")
    func emptyRange() {
        #expect(DiffRange(start1: 0, end1: 0, start2: 0, end2: 0).isEmpty)
        #expect(!DiffRange(start1: 0, end1: 1, start2: 0, end2: 0).isEmpty)
    }

    @Test("description matches Go formatting")
    func descriptionFormat() {
        let range = DiffRange(start1: 1, end1: 3, start2: 2, end2: 5)
        #expect(String(describing: range) == "[1, 3) - [2, 5)")
    }

    @Test("both initializers produce equivalent values")
    func initializerEquivalence() {
        let viaTuple = DiffRange(start1: 1, end1: 3, start2: 2, end2: 5)
        let viaRanges = DiffRange(left: 1..<3, right: 2..<5)
        #expect(viaTuple == viaRanges)
    }

    @Test("equal ranges have equal hashes")
    func hashable() {
        let lhs = DiffRange(start1: 0, end1: 1, start2: 2, end2: 3)
        let rhs = DiffRange(start1: 0, end1: 1, start2: 2, end2: 3)
        let set: Set<DiffRange> = [lhs, rhs]
        #expect(set.count == 1)
    }

    @Test("ranges with different bounds are not equal")
    func notEqual() {
        let lhs = DiffRange(start1: 0, end1: 1, start2: 0, end2: 1)
        let rhs = DiffRange(start1: 0, end1: 2, start2: 0, end2: 1)
        #expect(lhs != rhs)
    }
}
