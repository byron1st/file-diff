// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Represents a single edit operation in a linked list of changes.
///
/// - `line0` / `line1` are the starting indices in the first and second sequences.
/// - `deleted` lines were removed from the first sequence starting at `line0`.
/// - `inserted` lines were added in the second sequence starting at `line1`.
final class ChangeNode {
    let line0: Int
    let line1: Int
    let deleted: Int
    let inserted: Int
    var link: ChangeNode?

    init(line0: Int, line1: Int, deleted: Int, inserted: Int, link: ChangeNode? = nil) {
        self.line0 = line0
        self.line1 = line1
        self.deleted = deleted
        self.inserted = inserted
        self.link = link
    }
}

/// Consumer for LCS results, streamed as alternating equal and changed blocks.
protocol LCSBuilder: AnyObject {
    func addEqual(_ length: Int)
    func addChange(_ deleted: Int, _ inserted: Int)
}

/// Collects LCS stream callbacks into a linked list of `ChangeNode`.
final class ChangeNodeBuilder: LCSBuilder {
    private(set) var first: ChangeNode?
    private var last: ChangeNode?
    private var index1 = 0
    private var index2 = 0
    let startShift: Int

    init(startShift: Int) {
        self.startShift = startShift
    }

    func addEqual(_ length: Int) {
        skip(length, length)
    }

    func addChange(_ deleted: Int, _ inserted: Int) {
        let node = ChangeNode(
            line0: startShift + index1,
            line1: startShift + index2,
            deleted: deleted,
            inserted: inserted
        )
        if let last {
            last.link = node
        } else {
            first = node
        }
        last = node
        skip(deleted, inserted)
    }

    private func skip(_ first: Int, _ second: Int) {
        index1 += first
        index2 += second
    }
}
