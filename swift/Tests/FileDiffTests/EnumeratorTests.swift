import Testing
@testable import FileDiff

@Suite("Enumerator")
struct EnumeratorTests {
    @Test("equal strings receive the same ID")
    func sharesIdForEqualStrings() {
        var enumerator = Enumerator(expectedCapacity: 4)
        let ids = enumerator.enumerate(["a", "b", "c", "a"])
        #expect(ids[0] == ids[3])
        #expect(ids[0] != ids[1])
        #expect(ids[1] != ids[2])
    }

    @Test("IDs start from 1")
    func firstIdIsOne() {
        var enumerator = Enumerator()
        let ids = enumerator.enumerate(["x"])
        #expect(ids[0] == 1)
    }

    @Test("subsequent calls continue numbering from the previous state")
    func continuesNumberingAcrossCalls() {
        var enumerator = Enumerator()
        let first = enumerator.enumerate(["a", "b"])
        let second = enumerator.enumerate(["b", "c"])
        #expect(second[0] == first[1])
        #expect(second[1] != first[0] && second[1] != first[1])
    }
}
