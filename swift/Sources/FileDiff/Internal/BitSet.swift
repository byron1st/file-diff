// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// A simple, grow-on-demand bit set used internally by the Myers LCS algorithm.
final class BitSet {
    private static let elementSize = 64

    private var bits: [UInt64]
    private(set) var size: Int

    init(size: Int) {
        let slots = max(1, (size + Self.elementSize - 1) / Self.elementSize)
        bits = Array(repeating: 0, count: slots)
        self.size = size
    }

    func get(_ index: Int) -> Bool {
        guard index >= 0, index < size else {
            return false
        }
        let slot = index / Self.elementSize
        let mask = UInt64(1) << (index % Self.elementSize)
        return (bits[slot] & mask) != 0
    }

    func set(_ index: Int, _ value: Bool) {
        ensureCapacity(for: index)
        let slot = index / Self.elementSize
        let mask = UInt64(1) << (index % Self.elementSize)
        if value {
            bits[slot] |= mask
        } else {
            bits[slot] &= ~mask
        }
    }

    /// Sets bits in `[start, end)` to `value`.
    func setRange(_ start: Int, _ end: Int, _ value: Bool) {
        for index in start..<end {
            set(index, value)
        }
    }

    private func ensureCapacity(for index: Int) {
        if index >= size {
            size = index + 1
        }
        let needed = (index / Self.elementSize) + 1
        if needed > bits.count {
            bits.append(contentsOf: Array(repeating: 0, count: needed - bits.count))
        }
    }
}
