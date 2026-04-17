// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

import Foundation

/// A word or newline token produced by tokenizing text.
///
/// Offsets are byte positions into the UTF-8 representation of the original text.
struct InlineChunk: Equatable {
    let offset1: Int
    let offset2: Int
    let isNewline: Bool
    let hash: Int
    let content: String

    static func word(start: Int, end: Int, hash: Int, content: String) -> InlineChunk {
        InlineChunk(offset1: start, offset2: end, isNewline: false, hash: hash, content: content)
    }

    static func newline(offset: Int) -> InlineChunk {
        InlineChunk(offset1: offset, offset2: offset + 1, isNewline: true, hash: 0, content: "\n")
    }

    var key: ChunkKey {
        ChunkKey(isWord: !isNewline, content: content, hash: hash)
    }
}

/// Comparable identity of an `InlineChunk` used for Myers-based equality.
struct ChunkKey: Hashable {
    let isWord: Bool
    let content: String
    let hash: Int
}

/// Tokenizes UTF-8 encoded `text` into word and newline chunks.
///
/// Words are maximal sequences of alphabetic (non-whitespace, non-punctuation, non-continuous-script)
/// scalars. Continuous-script scalars (CJK, Thai, ...) each become their own single-scalar word.
func getInlineChunks(_ text: String) -> [InlineChunk] {
    var chunks: [InlineChunk] = []
    var wordStart = -1
    var wordHash = 0
    var offset = 0

    for scalar in text.unicodeScalars {
        let size = utf8Length(scalar)
        let isA = CharacterUtils.isAlpha(scalar)
        let isWordPart = isA && !CharacterUtils.isContinuousScript(scalar)

        if isWordPart {
            if wordStart == -1 {
                wordStart = offset
                wordHash = 0
            }
            wordHash = 31 &* wordHash &+ Int(scalar.value)
        } else {
            if wordStart != -1 {
                let content = substring(of: text, byteStart: wordStart, byteEnd: offset)
                chunks.append(.word(start: wordStart, end: offset, hash: wordHash, content: content))
                wordStart = -1
            }
            if isA {
                // Continuous-script scalar: single-character word.
                let content = String(scalar)
                chunks.append(.word(start: offset, end: offset + size, hash: Int(scalar.value), content: content))
            } else if scalar == "\n" {
                chunks.append(.newline(offset: offset))
            }
        }

        offset += size
    }

    if wordStart != -1 {
        let content = substring(of: text, byteStart: wordStart, byteEnd: offset)
        chunks.append(.word(start: wordStart, end: offset, hash: wordHash, content: content))
    }

    return chunks
}

func utf8Length(_ scalar: Unicode.Scalar) -> Int {
    let value = scalar.value
    if value < 0x80 {
        return 1
    }
    if value < 0x800 {
        return 2
    }
    if value < 0x10000 {
        return 3
    }
    return 4
}

private func substring(of text: String, byteStart: Int, byteEnd: Int) -> String {
    let utf8 = text.utf8
    let start = utf8.index(utf8.startIndex, offsetBy: byteStart)
    let end = utf8.index(utf8.startIndex, offsetBy: byteEnd)
    return String(bytes: Array(utf8[start..<end]), encoding: .utf8) ?? ""
}
