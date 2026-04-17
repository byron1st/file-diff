// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

import Foundation

enum CharacterUtils {
    /// Returns true if `scalar` is a space, newline, or tab.
    static func isSpaceEnterOrTab(_ scalar: Unicode.Scalar) -> Bool {
        scalar == " " || scalar == "\n" || scalar == "\t"
    }

    /// Returns true if `scalar` is an ASCII whitespace character.
    static func isWhiteSpaceCodePoint(_ scalar: Unicode.Scalar) -> Bool {
        scalar.value < 128 && isSpaceEnterOrTab(scalar)
    }

    /// Returns true if `scalar` is an ASCII punctuation character (excluding `_`).
    static func isPunctuation(_ scalar: Unicode.Scalar) -> Bool {
        let value = scalar.value
        if value == 95 {
            return false
        }
        return (value >= 33 && value <= 47) ||
            (value >= 58 && value <= 64) ||
            (value >= 91 && value <= 96) ||
            (value >= 123 && value <= 126)
    }

    /// Returns true if `scalar` is neither whitespace nor punctuation.
    static func isAlpha(_ scalar: Unicode.Scalar) -> Bool {
        if isWhiteSpaceCodePoint(scalar) {
            return false
        }
        return !isPunctuation(scalar)
    }

    /// Returns true if `scalar` belongs to a script where word boundaries cannot be
    /// determined by spaces (CJK, Thai, etc.). Such characters are treated as single-character words.
    static func isContinuousScript(_ scalar: Unicode.Scalar) -> Bool {
        if scalar.value < 128 {
            return false
        }
        if CharacterSet.decimalDigits.contains(scalar) {
            return false
        }
        if scalar.value >= 0x10000 {
            return true
        }
        if Unicode.Scalar.Han.contains(scalar) {
            return true
        }
        if !CharacterSet.letters.contains(scalar) {
            return true
        }
        return Unicode.Scalar.Hiragana.contains(scalar)
            || Unicode.Scalar.Katakana.contains(scalar)
            || Unicode.Scalar.Thai.contains(scalar)
            || Unicode.Scalar.Javanese.contains(scalar)
    }
}

private extension Unicode.Scalar {
    static let Han: ClosedRange<UInt32> = 0x4E00...0x9FFF
    static let Hiragana: ClosedRange<UInt32> = 0x3040...0x309F
    static let Katakana: ClosedRange<UInt32> = 0x30A0...0x30FF
    static let Thai: ClosedRange<UInt32> = 0x0E00...0x0E7F
    static let Javanese: ClosedRange<UInt32> = 0xA980...0xA9DF
}

private extension ClosedRange where Bound == UInt32 {
    func contains(_ scalar: Unicode.Scalar) -> Bool {
        contains(scalar.value)
    }
}
