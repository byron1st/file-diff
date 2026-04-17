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
}
