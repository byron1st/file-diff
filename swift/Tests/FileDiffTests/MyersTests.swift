import Testing
@testable import FileDiff

@Suite("BitSet")
struct BitSetTests {
    @Test("grows on write beyond initial size")
    func grows() {
        let set = BitSet(size: 2)
        set.set(100, true)
        #expect(set.get(100))
        #expect(!set.get(99))
    }

    @Test("clearing a bit removes it")
    func setFalse() {
        let set = BitSet(size: 10)
        set.set(3, true)
        #expect(set.get(3))
        set.set(3, false)
        #expect(!set.get(3))
    }

    @Test("out-of-range reads return false")
    func outOfRange() {
        let set = BitSet(size: 5)
        #expect(!set.get(-1))
        #expect(!set.get(10))
    }
}

@Suite("MyersLCS")
struct MyersLCSTests {
    @Test("identical arrays produce no changes")
    func identical() {
        let lcs = MyersLCS([1, 2, 3], [1, 2, 3])
        lcs.execute()
        for index in 0..<3 {
            #expect(!lcs.changes1.get(index))
            #expect(!lcs.changes2.get(index))
        }
    }

    @Test("completely different arrays mark everything changed")
    func allDifferent() {
        let lcs = MyersLCS([1, 2, 3], [4, 5, 6])
        lcs.execute()
        for index in 0..<3 {
            #expect(lcs.changes1.get(index))
            #expect(lcs.changes2.get(index))
        }
    }

    @Test("insertion marks only the added element")
    func insertMiddle() {
        let lcs = MyersLCS([1, 3], [1, 2, 3])
        lcs.execute()
        #expect(!lcs.changes1.get(0))
        #expect(!lcs.changes1.get(1))
        #expect(!lcs.changes2.get(0))
        #expect(lcs.changes2.get(1))
        #expect(!lcs.changes2.get(2))
    }

    @Test("empty first sequence marks all second as inserted")
    func emptyFirst() {
        let lcs = MyersLCS([], [1, 2])
        lcs.execute()
        #expect(lcs.changes2.get(0))
        #expect(lcs.changes2.get(1))
    }

    @Test("executeWithThreshold succeeds on small inputs")
    func thresholdSmallInput() throws {
        let lcs = MyersLCS([1, 2, 3], [1, 4, 3])
        try lcs.executeWithThreshold()
        #expect(lcs.changes1.get(1))
        #expect(lcs.changes2.get(1))
        #expect(!lcs.changes1.get(0))
        #expect(!lcs.changes1.get(2))
    }
}

@Suite("buildChanges")
struct BuildChangesTests {
    @Test("identical int arrays return nil")
    func identicalInts() {
        #expect(buildChanges([1, 2, 3], [1, 2, 3]) == nil)
    }

    @Test("both empty int arrays return nil")
    func bothEmpty() {
        #expect(buildChanges([], []) == nil)
    }

    @Test("insertion in middle is a single 0-deleted / 1-inserted change")
    func simpleInsert() throws {
        let node = try #require(buildChanges([1, 3], [1, 2, 3]))
        #expect(node.line0 == 1)
        #expect(node.line1 == 1)
        #expect(node.deleted == 0)
        #expect(node.inserted == 1)
        #expect(node.link == nil)
    }

    @Test("deletion is a single 1-deleted / 0-inserted change")
    func simpleDelete() throws {
        let node = try #require(buildChanges([1, 2, 3], [1, 3]))
        #expect(node.line0 == 1)
        #expect(node.deleted == 1)
        #expect(node.inserted == 0)
    }

    @Test("modification is 1-deleted + 1-inserted at the changed position")
    func modification() throws {
        let node = try #require(buildChanges([1, 2, 3], [1, 4, 3]))
        #expect(node.line0 == 1)
        #expect(node.deleted == 1)
        #expect(node.inserted == 1)
    }

    @Test("multiple changes are linked together")
    func multipleChanges() throws {
        let node = try #require(buildChanges([1, 2, 3, 4, 5], [1, 9, 3, 8, 5]))
        #expect(node.link != nil)
    }

    @Test("one empty side yields a pure insertion change")
    func oneEmpty() throws {
        let node = try #require(buildChanges([], [1, 2, 3]))
        #expect(node.deleted == 0)
        #expect(node.inserted == 3)
    }

    @Test("identical string arrays return nil")
    func identicalStrings() {
        #expect(buildChanges(fromObjects: ["foo", "bar", "baz"], ["foo", "bar", "baz"]) == nil)
    }

    @Test("string insertion produces a 0-deleted / 1-inserted change")
    func stringInsert() throws {
        let node = try #require(buildChanges(fromObjects: ["foo", "baz"], ["foo", "bar", "baz"]))
        #expect(node.deleted == 0)
        #expect(node.inserted == 1)
    }

    @Test("string modification produces a 1-deleted / 1-inserted change")
    func stringModification() throws {
        let node = try #require(
            buildChanges(fromObjects: ["foo", "bar", "baz"], ["foo", "qux", "baz"])
        )
        #expect(node.deleted == 1)
        #expect(node.inserted == 1)
    }
}

private enum RecordedOperation: Equatable {
    case equal(Int)
    case change(Int, Int)
}

private final class RecordingBuilder: LCSBuilder {
    var ops: [RecordedOperation] = []
    func addEqual(_ length: Int) {
        ops.append(.equal(length))
    }

    func addChange(_ deleted: Int, _ inserted: Int) {
        ops.append(.change(deleted, inserted))
    }
}

@Suite("Reindexer")
struct ReindexerTests {
    @Test("discardUnique keeps common values")
    func discardKeepsCommon() {
        let reindexer = Reindexer()
        let discarded = reindexer.discardUnique([1, 2, 3, 4], [2, 4, 5])
        #expect(discarded[0] == [2, 4])
        #expect(discarded[1] == [2, 4])
    }

    @Test("reindex restores intermediate and trailing gaps")
    func reindexGaps() {
        let reindexer = Reindexer()
        let discarded = reindexer.discardUnique([1, 2, 3, 4], [1, 3, 4, 5])
        #expect(discarded[0].count == 3)
        #expect(discarded[1].count == 3)

        let builder = RecordingBuilder()
        reindexer.reindex(
            (BitSet(size: discarded[0].count), BitSet(size: discarded[1].count)),
            builder: builder
        )
        #expect(
            builder.ops == [
                .equal(1),
                .change(1, 0),
                .equal(2),
                .change(0, 1),
            ]
        )
    }

    @Test("fully disjoint sequences produce a single full-change op")
    func fullyDisjoint() {
        let reindexer = Reindexer()
        let discarded = reindexer.discardUnique([1, 2], [3, 4])
        #expect(discarded[0].isEmpty)
        #expect(discarded[1].isEmpty)

        let builder = RecordingBuilder()
        reindexer.reindex((BitSet(size: 0), BitSet(size: 0)), builder: builder)
        #expect(builder.ops == [.change(2, 2)])
    }
}
