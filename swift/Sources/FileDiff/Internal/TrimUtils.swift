// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Whitespace and text-range utilities operating on UTF-8 byte arrays.
///
/// Offsets passed to these helpers are always byte offsets into the supplied
/// `bytes` array, matching the Go implementation's `[]byte` semantics.
enum TrimUtils {
    static func isSpaceEnterOrTab(_ byte: UInt8) -> Bool {
        byte == 0x20 || byte == 0x0A || byte == 0x09
    }

    static func trimStartText(_ bytes: [UInt8], _ start: Int, _ end: Int) -> Int {
        var cursor = start
        while cursor < end, isSpaceEnterOrTab(bytes[cursor]) {
            cursor += 1
        }
        return cursor
    }

    static func trimEndText(_ bytes: [UInt8], _ start: Int, _ end: Int) -> Int {
        var cursor = end
        while start < cursor, isSpaceEnterOrTab(bytes[cursor - 1]) {
            cursor -= 1
        }
        return cursor
    }

    // swiftlint:disable:next function_parameter_count
    static func trimTextRange(
        _ bytes1: [UInt8], _ bytes2: [UInt8],
        _ start1: Int, _ start2: Int, _ end1: Int, _ end2: Int
    ) -> DiffRange {
        let s1 = trimStartText(bytes1, start1, end1)
        let e1 = trimEndText(bytes1, s1, end1)
        let s2 = trimStartText(bytes2, start2, end2)
        let e2 = trimEndText(bytes2, s2, end2)
        return DiffRange(start1: s1, end1: e1, start2: s2, end2: e2)
    }

    // swiftlint:disable:next function_parameter_count
    static func expandWhitespacesForward(
        _ bytes1: [UInt8], _ bytes2: [UInt8],
        _ start1: Int, _ start2: Int, _ end1: Int, _ end2: Int
    ) -> Int {
        var s1 = start1
        var s2 = start2
        while s1 < end1, s2 < end2 {
            if bytes1[s1] != bytes2[s2] {
                break
            }
            if !isSpaceEnterOrTab(bytes1[s1]) {
                break
            }
            s1 += 1
            s2 += 1
        }
        return s1 - start1
    }

    // swiftlint:disable:next function_parameter_count
    static func expandWhitespacesBackward(
        _ bytes1: [UInt8], _ bytes2: [UInt8],
        _ start1: Int, _ start2: Int, _ end1: Int, _ end2: Int
    ) -> Int {
        var e1 = end1
        var e2 = end2
        while start1 < e1, start2 < e2 {
            if bytes1[e1 - 1] != bytes2[e2 - 1] {
                break
            }
            if !isSpaceEnterOrTab(bytes1[e1 - 1]) {
                break
            }
            e1 -= 1
            e2 -= 1
        }
        return end1 - e1
    }

    static func expandWhitespacesRange(
        _ bytes1: [UInt8], _ bytes2: [UInt8], _ range: DiffRange
    ) -> DiffRange {
        var s1 = range.left.lowerBound
        var s2 = range.right.lowerBound
        let e1 = range.left.upperBound
        let e2 = range.right.upperBound
        let fwd = expandWhitespacesForward(bytes1, bytes2, s1, s2, e1, e2)
        s1 += fwd
        s2 += fwd
        let bwd = expandWhitespacesBackward(bytes1, bytes2, s1, s2, e1, e2)
        return DiffRange(start1: s1, end1: e1 - bwd, start2: s2, end2: e2 - bwd)
    }

    static func isEqualTextRange(
        _ bytes1: [UInt8], _ bytes2: [UInt8], _ range: DiffRange
    ) -> Bool {
        let leftLen = range.left.count
        if leftLen != range.right.count {
            return false
        }
        for offset in 0..<leftLen
            where bytes1[range.left.lowerBound + offset] != bytes2[range.right.lowerBound + offset]
        {
            return false
        }
        return true
    }

    static func isEqualTextRangeIgnoreWhitespaces(
        _ bytes1: [UInt8], _ bytes2: [UInt8], _ range: DiffRange
    ) -> Bool {
        var i = range.left.lowerBound
        var j = range.right.lowerBound
        let endI = range.left.upperBound
        let endJ = range.right.upperBound
        while i < endI, j < endJ {
            if isSpaceEnterOrTab(bytes1[i]) {
                i += 1
                continue
            }
            if isSpaceEnterOrTab(bytes2[j]) {
                j += 1
                continue
            }
            if bytes1[i] != bytes2[j] {
                return false
            }
            i += 1
            j += 1
        }
        while i < endI {
            if !isSpaceEnterOrTab(bytes1[i]) {
                return false
            }
            i += 1
        }
        while j < endJ {
            if !isSpaceEnterOrTab(bytes2[j]) {
                return false
            }
            j += 1
        }
        return true
    }

    static func isLeadingTrailingSpace(_ bytes: [UInt8], _ pos: Int) -> Bool {
        isLeadingSpace(bytes, pos) || isTrailingSpace(bytes, pos)
    }

    private static func isLeadingSpace(_ bytes: [UInt8], _ pos: Int) -> Bool {
        if pos < 0 || pos >= bytes.count {
            return false
        }
        if !isSpaceEnterOrTab(bytes[pos]) {
            return false
        }
        var cursor = pos - 1
        while cursor >= 0 {
            let byte = bytes[cursor]
            if byte == 0x0A {
                return true
            }
            if !isSpaceEnterOrTab(byte) {
                return false
            }
            cursor -= 1
        }
        return true
    }

    private static func isTrailingSpace(_ bytes: [UInt8], _ pos: Int) -> Bool {
        if pos < 0 || pos >= bytes.count {
            return false
        }
        if !isSpaceEnterOrTab(bytes[pos]) {
            return false
        }
        var cursor = pos
        while cursor < bytes.count {
            let byte = bytes[cursor]
            if byte == 0x0A {
                return true
            }
            if !isSpaceEnterOrTab(byte) {
                return false
            }
            cursor += 1
        }
        return true
    }
}
