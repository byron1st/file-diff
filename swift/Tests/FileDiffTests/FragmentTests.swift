import Testing
@testable import FileDiff

@Suite("DiffFragment")
struct DiffFragmentTests {
    @Test("stores the two offset ranges")
    func storesBounds() {
        let fragment = DiffFragment(startOffset1: 0, endOffset1: 5, startOffset2: 0, endOffset2: 3)
        #expect(fragment.left == 0..<5)
        #expect(fragment.right == 0..<3)
    }

    @Test("description matches Go formatting")
    func descriptionFormat() {
        let fragment = DiffFragment(startOffset1: 0, endOffset1: 5, startOffset2: 0, endOffset2: 3)
        #expect(String(describing: fragment) == "[0, 5) - [0, 3)")
    }
}

@Suite("LineFragment")
struct LineFragmentTests {
    @Test("keeps partial inner fragments")
    func keepsPartialInner() {
        let inner = [DiffFragment(startOffset1: 2, endOffset1: 5, startOffset2: 2, endOffset2: 4)]
        let fragment = LineFragment(
            leftLines: 0..<1,
            rightLines: 0..<1,
            leftOffsets: 0..<10,
            rightOffsets: 0..<8,
            inner: inner
        )
        #expect(fragment.inner.count == 1)
    }

    @Test("drops a single inner fragment spanning the whole range")
    func dropsWholeInner() {
        let inner = [DiffFragment(startOffset1: 0, endOffset1: 10, startOffset2: 0, endOffset2: 8)]
        let fragment = LineFragment(
            leftLines: 0..<1,
            rightLines: 0..<1,
            leftOffsets: 0..<10,
            rightOffsets: 0..<8,
            inner: inner
        )
        #expect(fragment.inner.isEmpty)
    }

    @Test("keeps multiple inner fragments")
    func keepsMultipleInner() {
        let inner = [
            DiffFragment(startOffset1: 0, endOffset1: 3, startOffset2: 0, endOffset2: 3),
            DiffFragment(startOffset1: 5, endOffset1: 8, startOffset2: 5, endOffset2: 7),
        ]
        let fragment = LineFragment(
            leftLines: 0..<1,
            rightLines: 0..<1,
            leftOffsets: 0..<10,
            rightOffsets: 0..<8,
            inner: inner
        )
        #expect(fragment.inner.count == 2)
    }

    @Test("represents an insertion (empty left side)")
    func insertionOnly() {
        let fragment = LineFragment(
            leftLines: 0..<0,
            rightLines: 0..<2,
            leftOffsets: 0..<0,
            rightOffsets: 0..<16
        )
        #expect(fragment.leftLines.isEmpty)
        #expect(fragment.rightLines == 0..<2)
    }

    @Test("represents a deletion (empty right side)")
    func deletionOnly() {
        let fragment = LineFragment(
            leftLines: 0..<2,
            rightLines: 0..<0,
            leftOffsets: 0..<16,
            rightOffsets: 0..<0
        )
        #expect(fragment.leftLines == 0..<2)
        #expect(fragment.rightLines.isEmpty)
    }

    @Test("description reflects the inner fragment count")
    func descriptionFormat() {
        let inner = [DiffFragment(startOffset1: 2, endOffset1: 5, startOffset2: 2, endOffset2: 4)]
        let fragment = LineFragment(
            leftLines: 0..<1,
            rightLines: 0..<1,
            leftOffsets: 0..<10,
            rightOffsets: 0..<8,
            inner: inner
        )
        #expect(
            String(describing: fragment)
                == "Lines [0, 1) - [0, 1); Offsets [0, 10) - [0, 8); Inner 1"
        )
    }
}
