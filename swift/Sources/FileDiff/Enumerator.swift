// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Assigns unique integer IDs to strings.
///
/// Equal strings always receive the same ID. IDs start from `1` so that `0` can be used
/// as a sentinel by the calling code.
public struct Enumerator: Sendable {
    private var numbers: [String: Int]
    private var nextNumber: Int

    public init(expectedCapacity: Int = 0) {
        var dict: [String: Int] = [:]
        if expectedCapacity > 0 {
            dict.reserveCapacity(expectedCapacity)
        }
        numbers = dict
        nextNumber = 1
    }

    /// Returns an array of IDs, one per input string.
    public mutating func enumerate(_ strings: [String]) -> [Int] {
        var result: [Int] = []
        result.reserveCapacity(strings.count)
        for str in strings {
            if let existing = numbers[str] {
                result.append(existing)
                continue
            }
            let id = nextNumber
            nextNumber += 1
            numbers[str] = id
            result.append(id)
        }
        return result
    }
}
