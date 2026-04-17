import Testing
@testable import FileDiff

@Suite("Sanity")
struct SanityTests {
    @Test("module exposes a version string")
    func versionIsNonEmpty() {
        #expect(!FileDiff.version.isEmpty)
    }
}
