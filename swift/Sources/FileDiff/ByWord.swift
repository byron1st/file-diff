// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Compares two texts at the word level.
///
/// The returned fragments use UTF-8 byte offsets into `text1` and `text2`.
public func compareWords(_ text1: String, _ text2: String, policy: ComparisonPolicy) -> [DiffFragment] {
    let bytes1 = Array(text1.utf8)
    let bytes2 = Array(text2.utf8)
    let words1 = getInlineChunks(text1)
    let words2 = getInlineChunks(text2)

    var wordChanges = diffInlineChunks(words1, words2)
    wordChanges = optimizeWordChunks(bytes1, bytes2, words1, words2, wordChanges)

    let delimiters = matchAdjustmentDelimiters(
        bytes1: bytes1, bytes2: bytes2,
        words1: words1, words2: words2,
        changes: wordChanges
    )
    let iterable = matchAdjustmentWhitespaces(
        bytes1: bytes1, bytes2: bytes2,
        iterable: delimiters,
        policy: policy
    )
    return convertIntoDiffFragments(iterable)
}

func diffInlineChunks(_ chunks1: [InlineChunk], _ chunks2: [InlineChunk]) -> FairDiffIterable {
    diffObjects(chunks1.map(\.key), chunks2.map(\.key))
}

func convertIntoDiffFragments(_ iterable: DiffIterable) -> [DiffFragment] {
    iterable.changes.map { range in
        DiffFragment(
            startOffset1: range.left.lowerBound, endOffset1: range.left.upperBound,
            startOffset2: range.right.lowerBound, endOffset2: range.right.upperBound
        )
    }
}
