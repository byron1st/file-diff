import Testing
@testable import FileDiff

private func split(_ text: String) -> [String] {
    text.isEmpty ? [] : text.components(separatedBy: "\n")
}

@Suite("optimizeLineChunks via compareLines")
struct OptimizeLineChunksTests {
    @Test("insertion shifts to land after the empty line on the touch side")
    func shiftToEmptyLineOnTouchSide() {
        // Without optimization, Myers may anchor the inserted block before
        // the empty line. The optimizer prefers cutting at the empty line so
        // the diff stays visually aligned.
        let left = split("alpha\n\nbeta")
        let right = split("alpha\n\ninserted\nbeta")
        let result = compareLines(left, right, policy: .default)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        // The change should be a pure insertion.
        #expect(change.left.isEmpty)
        #expect(change.right.count == 1)
    }

    @Test("deletion shifts to land before the empty line on the touch side")
    func shiftBeforeEmptyLineOnTouchSide() {
        let left = split("alpha\nremoved\n\nbeta")
        let right = split("alpha\n\nbeta")
        let result = compareLines(left, right, policy: .default)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left.count == 1)
        #expect(change.right.isEmpty)
    }

    @Test("optimizer keeps boundary stable when touch side has no empty line")
    func noEmptyLineFallsThrough() {
        // Both sides have only "important" lines; the optimizer falls
        // through every tier and returns the original boundary.
        let left = split("alpha\nbeta\ngamma")
        let right = split("alpha\nbeta\nDELTA\ngamma")
        let result = compareLines(left, right, policy: .default)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left.isEmpty)
        #expect(change.right.count == 1)
    }

    @Test("optimizer prefers an unimportant short line on the non-touch side")
    func shortLineOnNonTouchSide() {
        // Left has a short ("}") line that the threshold tier can lock onto
        // when the touch side has nothing better.
        let left = split("alpha\nbeta\n}\ngamma\ndelta")
        let right = split("alpha\nbeta\n}\nINS\ngamma\ndelta")
        let result = compareLines(left, right, policy: .default)
        #expect(result.changes.count == 1)
    }

    @Test("multiple adjacent insertions share a single boundary")
    func adjacentInsertionsCollapse() {
        // Two insertions next to a touching anchor exercise the eqFwd /
        // eqBwd absorption recursion.
        let left = split("alpha\nbeta\ngamma")
        let right = split("alpha\nbeta\nbeta\nbeta\ngamma")
        let result = compareLines(left, right, policy: .default)
        #expect(result.changes.count == 1)
        let change = result.changes[0]
        #expect(change.left.isEmpty)
        #expect(change.right.count == 2)
    }

    @Test("trimWhitespaces still benefits from boundary shifting")
    func shiftUnderTrimPolicy() {
        let left = split("alpha\n\nbeta")
        let right = split("alpha\n\ninserted\nbeta")
        let result = compareLines(left, right, policy: .trimWhitespaces)
        #expect(result.changes.count == 1)
        #expect(result.changes[0].right.count == 1)
    }
}
