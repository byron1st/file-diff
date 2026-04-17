// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Objects that participate in `ExpandChangeBuilder` border expansion must
/// support equality comparison through this protocol.
protocol LineEquatable: AnyObject {
    func equals(_ other: LineEquatable) -> Bool
}

/// Creates a `FairDiffIterable` from a list of changed ranges.
func createFromRanges(_ changes: [DiffRange], length1: Int, length2: Int) -> FairDiffIterable {
    RangesDiffIterable(changes: changes, length1: length1, length2: length2)
}

/// Creates a `FairDiffIterable` from a `ChangeNode` linked list.
func createFromChanges(_ change: ChangeNode?, length1: Int, length2: Int) -> FairDiffIterable {
    var ranges: [DiffRange] = []
    var node = change
    while let current = node {
        ranges.append(DiffRange(
            start1: current.line0, end1: current.line0 + current.deleted,
            start2: current.line1, end2: current.line1 + current.inserted
        ))
        node = current.link
    }
    return createFromRanges(ranges, length1: length1, length2: length2)
}

/// Swaps the changed/unchanged interpretation of an iterable.
func invert(_ iterable: DiffIterable) -> DiffIterable {
    InvertedDiffIterable(inner: iterable)
}

/// Creates a `FairDiffIterable` from a list of unchanged ranges.
func createUnchanged(_ unchanged: [DiffRange], length1: Int, length2: Int) -> FairDiffIterable {
    fair(invert(createFromRanges(unchanged, length1: length1, length2: length2)))
}

/// Wraps an arbitrary `DiffIterable` as a `FairDiffIterable`.
func fair(_ iterable: DiffIterable) -> FairDiffIterable {
    if let already = iterable as? FairDiffIterable {
        return already
    }
    return FairDiffIterableWrapper(inner: iterable)
}

/// Runs Myers on two equal-hash integer arrays and returns the resulting diff.
func diff(_ data1: [Int], _ data2: [Int]) -> FairDiffIterable {
    let node = buildChanges(data1, data2)
    return createFromChanges(node, length1: data1.count, length2: data2.count)
}

/// Enumerates `Hashable` values to integer IDs and runs Myers on the result.
func diffObjects<T: Hashable>(_ data1: [T], _ data2: [T]) -> FairDiffIterable {
    var table: [T: Int] = [:]
    table.reserveCapacity(data1.count + data2.count)
    var nextID = 1

    func intern(_ value: T) -> Int {
        if let id = table[value] {
            return id
        }
        let id = nextID
        nextID += 1
        table[value] = id
        return id
    }

    let ints1 = data1.map(intern)
    let ints2 = data2.map(intern)
    return diff(ints1, ints2)
}

/// Collects `markEqual` calls and produces a `FairDiffIterable`.
class ChangeBuilder {
    let length1: Int
    let length2: Int
    private(set) var index1: Int = 0
    private(set) var index2: Int = 0
    var changes: [DiffRange] = []

    init(length1: Int, length2: Int) {
        self.length1 = length1
        self.length2 = length2
    }

    func markEqual(_ idx1: Int, _ idx2: Int) {
        markEqualRange(idx1, idx2, idx1 + 1, idx2 + 1)
    }

    func markEqualCount(_ idx1: Int, _ idx2: Int, _ count: Int) {
        markEqualRange(idx1, idx2, idx1 + count, idx2 + count)
    }

    func markEqualRange(_ idx1: Int, _ idx2: Int, _ end1: Int, _ end2: Int) {
        if idx1 == end1, idx2 == end2 {
            return
        }
        if index1 != idx1 || index2 != idx2 {
            addChange(index1, index2, idx1, idx2)
        }
        index1 = end1
        index2 = end2
    }

    func finish() -> FairDiffIterable {
        if length1 != index1 || length2 != index2 {
            addChange(index1, index2, length1, length2)
            index1 = length1
            index2 = length2
        }
        return createFromRanges(changes, length1: length1, length2: length2)
    }

    /// Adds a `[start1, end1) x [start2, end2)` change range.
    /// Subclasses may override to adjust (e.g. border expansion).
    func addChange(_ start1: Int, _ start2: Int, _ end1: Int, _ end2: Int) {
        changes.append(DiffRange(start1: start1, end1: end1, start2: start2, end2: end2))
    }
}

/// A `ChangeBuilder` that shrinks each recorded change by consuming equal
/// elements on its borders.
final class ExpandChangeBuilder: ChangeBuilder {
    private let objects1: [LineEquatable]
    private let objects2: [LineEquatable]

    init(objects1: [LineEquatable], objects2: [LineEquatable]) {
        self.objects1 = objects1
        self.objects2 = objects2
        super.init(length1: objects1.count, length2: objects2.count)
    }

    override func addChange(_ start1: Int, _ start2: Int, _ end1: Int, _ end2: Int) {
        var s1 = start1
        var s2 = start2
        while s1 < end1, s2 < end2, objects1[s1].equals(objects2[s2]) {
            s1 += 1
            s2 += 1
        }
        var e1 = end1
        var e2 = end2
        while s1 < e1, s2 < e2, objects1[e1 - 1].equals(objects2[e2 - 1]) {
            e1 -= 1
            e2 -= 1
        }
        let range = DiffRange(start1: s1, end1: e1, start2: s2, end2: e2)
        if !range.isEmpty {
            changes.append(range)
        }
    }
}
