// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

import Foundation

enum TextUtils {
    /// Compares two strings according to `policy`.
    static func isEqual(_ lhs: String, _ rhs: String, policy: ComparisonPolicy) -> Bool {
        switch policy {
        case .default:
            lhs == rhs
        case .trimWhitespaces:
            lhs.trimmingCharacters(in: .whitespacesAndNewlines)
                == rhs.trimmingCharacters(in: .whitespacesAndNewlines)
        case .ignoreWhitespaces:
            equalsIgnoreWhitespaces(lhs, rhs)
        }
    }

    /// Computes a Java-compatible `String.hashCode()` under `policy`.
    static func hashCode(_ value: String, policy: ComparisonPolicy) -> Int {
        switch policy {
        case .default:
            stringHashCode(value)
        case .trimWhitespaces:
            stringHashCode(value.trimmingCharacters(in: .whitespacesAndNewlines))
        case .ignoreWhitespaces:
            stringHashCodeIgnoreWhitespaces(value)
        }
    }

    static func countNonSpaceChars(_ value: String) -> Int {
        var count = 0
        for scalar in value.unicodeScalars where !CharacterSet.whitespacesAndNewlines.contains(scalar) {
            count += 1
        }
        return count
    }

    private static func equalsIgnoreWhitespaces(_ lhs: String, _ rhs: String) -> Bool {
        let lhsScalars = Array(lhs.unicodeScalars)
        let rhsScalars = Array(rhs.unicodeScalars)
        var i = 0
        var j = 0
        while i < lhsScalars.count, j < rhsScalars.count {
            if CharacterSet.whitespacesAndNewlines.contains(lhsScalars[i]) {
                i += 1
                continue
            }
            if CharacterSet.whitespacesAndNewlines.contains(rhsScalars[j]) {
                j += 1
                continue
            }
            if lhsScalars[i] != rhsScalars[j] {
                return false
            }
            i += 1
            j += 1
        }
        while i < lhsScalars.count {
            if !CharacterSet.whitespacesAndNewlines.contains(lhsScalars[i]) {
                return false
            }
            i += 1
        }
        while j < rhsScalars.count {
            if !CharacterSet.whitespacesAndNewlines.contains(rhsScalars[j]) {
                return false
            }
            j += 1
        }
        return true
    }

    private static func stringHashCode(_ value: String) -> Int {
        var hash = 0
        for scalar in value.unicodeScalars {
            hash = 31 &* hash &+ Int(scalar.value)
        }
        return hash
    }

    private static func stringHashCodeIgnoreWhitespaces(_ value: String) -> Int {
        var hash = 0
        for scalar in value.unicodeScalars where !CharacterSet.whitespacesAndNewlines.contains(scalar) {
            hash = 31 &* hash &+ Int(scalar.value)
        }
        return hash
    }
}
