// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Computes the diff between two integer arrays and returns a linked list of
/// `ChangeNode`. Returns `nil` when the two inputs are identical.
func buildChanges(_ ints1: [Int], _ ints2: [Int]) -> ChangeNode? {
    let startShift = commonPrefix(ints1, ints2)
    let endCut = commonSuffix(ints1, ints2, startShift: startShift)

    if let fast = buildChangesFast(
        length1: ints1.count,
        length2: ints2.count,
        startShift: startShift,
        endCut: endCut
    ) {
        return fast
    }

    let trimmed1 = Array(ints1[startShift..<(ints1.count - endCut)])
    let trimmed2 = Array(ints2[startShift..<(ints2.count - endCut)])
    return doBuildChanges(trimmed1, trimmed2, startShift: startShift)
}

/// Computes the diff between two string arrays by first enumerating them.
func buildChanges(fromObjects objects1: [String], _ objects2: [String]) -> ChangeNode? {
    let startShift = commonPrefix(objects1, objects2)
    let endCut = commonSuffix(objects1, objects2, startShift: startShift)

    if let fast = buildChangesFast(
        length1: objects1.count,
        length2: objects2.count,
        startShift: startShift,
        endCut: endCut
    ) {
        return fast
    }

    let trimmedLength = objects1.count + objects2.count - 2 * startShift - 2 * endCut
    var enumerator = Enumerator(expectedCapacity: trimmedLength)
    let ints1 = enumerator.enumerate(Array(objects1[startShift..<(objects1.count - endCut)]))
    let ints2 = enumerator.enumerate(Array(objects2[startShift..<(objects2.count - endCut)]))
    return doBuildChanges(ints1, ints2, startShift: startShift)
}

private func buildChangesFast(
    length1: Int,
    length2: Int,
    startShift: Int,
    endCut: Int
) -> ChangeNode?? {
    let trimmed1 = length1 - startShift - endCut
    let trimmed2 = length2 - startShift - endCut
    if trimmed1 != 0, trimmed2 != 0 {
        return nil
    }
    if trimmed1 == 0, trimmed2 == 0 {
        return .some(nil)
    }
    return .some(ChangeNode(
        line0: startShift,
        line1: startShift,
        deleted: trimmed1,
        inserted: trimmed2
    ))
}

private func doBuildChanges(_ ints1: [Int], _ ints2: [Int], startShift: Int) -> ChangeNode? {
    let reindexer = Reindexer()
    let discarded = reindexer.discardUnique(ints1, ints2)
    let builder = ChangeNodeBuilder(startShift: startShift)

    if discarded[0].isEmpty, discarded[1].isEmpty {
        builder.addChange(ints1.count, ints2.count)
        return builder.first
    }

    let lcs = MyersLCS(discarded[0], discarded[1])
    do {
        try lcs.executeWithThreshold()
    } catch {
        builder.addChange(ints1.count, ints2.count)
        return builder.first
    }

    reindexer.reindex((lcs.changes1, lcs.changes2), builder: builder)
    return builder.first
}

private func commonPrefix<T: Equatable>(_ lhs: [T], _ rhs: [T]) -> Int {
    let length = min(lhs.count, rhs.count)
    for index in 0..<length where lhs[index] != rhs[index] {
        return index
    }
    return length
}

private func commonSuffix<T: Equatable>(_ lhs: [T], _ rhs: [T], startShift: Int) -> Int {
    let length = min(lhs.count, rhs.count) - startShift
    for index in 0..<length where lhs[lhs.count - 1 - index] != rhs[rhs.count - 1 - index] {
        return index
    }
    return length
}
