import Testing
@testable import FileDiff

@Suite("DiffIterable")
struct DiffIterableTests {
    @Test("computeUnchanged fills gaps between changes")
    func computeUnchangedFills() {
        let changes = [
            DiffRange(start1: 2, end1: 3, start2: 2, end2: 3),
        ]
        let iterable = createFromRanges(changes, length1: 5, length2: 5)
        let unchanged = iterable.unchanged
        #expect(unchanged.count == 2)
        #expect(unchanged[0] == DiffRange(start1: 0, end1: 2, start2: 0, end2: 2))
        #expect(unchanged[1] == DiffRange(start1: 3, end1: 5, start2: 3, end2: 5))
    }

    @Test("invert swaps changes and unchanged")
    func invertSwaps() {
        let changes = [DiffRange(start1: 0, end1: 1, start2: 0, end2: 0)]
        let iterable = createFromRanges(changes, length1: 2, length2: 2)
        let inverted = invert(iterable)
        #expect(inverted.changes == iterable.unchanged)
        #expect(inverted.unchanged == iterable.changes)
    }
}

@Suite("ChangeBuilder")
struct ChangeBuilderTests {
    @Test("markEqual produces a single change for the remaining delta")
    func markEqual() {
        let builder = ChangeBuilder(length1: 3, length2: 3)
        builder.markEqual(0, 0)
        builder.markEqual(2, 2)
        let iterable = builder.finish()
        #expect(iterable.changes == [DiffRange(start1: 1, end1: 2, start2: 1, end2: 2)])
    }

    @Test("finish flushes trailing gap")
    func finishFlushes() {
        let builder = ChangeBuilder(length1: 3, length2: 4)
        builder.markEqual(0, 0)
        let iterable = builder.finish()
        #expect(iterable.changes == [DiffRange(start1: 1, end1: 3, start2: 1, end2: 4)])
    }
}
