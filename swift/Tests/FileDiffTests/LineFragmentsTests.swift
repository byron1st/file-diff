import Testing
@testable import FileDiff

@Suite("compareLineFragments")
struct CompareLineFragmentsTests {
    @Test("identical input produces no fragments")
    func identical() {
        let fragments = compareLineFragments(["a", "b"], ["a", "b"], matcher: MyersMatcher(), policy: .default)
        #expect(fragments.isEmpty)
    }

    @Test("single line change has inner word fragments")
    func singleLineChange() {
        let fragments = compareLineFragments(
            ["hello world"], ["hello earth"],
            matcher: MyersMatcher(), policy: .default
        )
        #expect(fragments.count == 1)
        let fragment = fragments[0]
        #expect(fragment.leftLines == 0..<1)
        #expect(fragment.rightLines == 0..<1)
        #expect(!fragment.inner.isEmpty)
    }

    @Test("inserted line is reported without inner fragments stripping whole change")
    func insertion() {
        let fragments = compareLineFragments(
            ["a", "c"], ["a", "b", "c"],
            matcher: MyersMatcher(), policy: .default
        )
        #expect(fragments.count == 1)
        #expect(fragments[0].leftLines == 1..<1)
        #expect(fragments[0].rightLines == 1..<2)
        #expect(fragments[0].inner.isEmpty)
    }

    @Test("offsets are in UTF-8 bytes of joined text")
    func utf8Offsets() {
        let fragments = compareLineFragments(
            ["go 한글 test"], ["go 한 test"],
            matcher: MyersMatcher(), policy: .default
        )
        #expect(fragments.count == 1)
        #expect(fragments[0].leftOffsets.upperBound == "go 한글 test".utf8.count)
        #expect(fragments[0].rightOffsets.upperBound == "go 한 test".utf8.count)
        #expect(!fragments[0].inner.isEmpty)
    }
}
