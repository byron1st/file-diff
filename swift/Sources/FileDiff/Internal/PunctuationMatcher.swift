// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

import Foundation

/// Matches punctuation characters sitting in the gaps between already-matched words.
func matchAdjustmentDelimiters(
    bytes1: [UInt8], bytes2: [UInt8],
    words1: [InlineChunk], words2: [InlineChunk],
    changes: FairDiffIterable
) -> FairDiffIterable {
    AdjustmentPunctuationMatcher(
        bytes1: bytes1, bytes2: bytes2,
        words1: words1, words2: words2,
        changes: changes
    ).build()
}

private final class AdjustmentPunctuationMatcher {
    private let bytes1: [UInt8]
    private let bytes2: [UInt8]
    private let words1: [InlineChunk]
    private let words2: [InlineChunk]
    private let changes: FairDiffIterable
    private let builder: ChangeBuilder

    private var lastStart1 = -1
    private var lastStart2 = -1
    private var lastEnd1 = -1
    private var lastEnd2 = -1

    init(
        bytes1: [UInt8], bytes2: [UInt8],
        words1: [InlineChunk], words2: [InlineChunk],
        changes: FairDiffIterable
    ) {
        self.bytes1 = bytes1
        self.bytes2 = bytes2
        self.words1 = words1
        self.words2 = words2
        self.changes = changes
        builder = ChangeBuilder(length1: bytes1.count, length2: bytes2.count)
    }

    func build() -> FairDiffIterable {
        clearLastRange()
        matchForwardIdx(-1, -1)

        for unchanged in changes.unchanged {
            let count = unchanged.left.count
            for i in 0..<count {
                let idx1 = unchanged.left.lowerBound + i
                let idx2 = unchanged.right.lowerBound + i

                let start1 = words1[idx1].offset1
                let start2 = words2[idx2].offset1
                let end1 = words1[idx1].offset2
                let end2 = words2[idx2].offset2

                matchBackwardIdx(idx1, idx2)
                builder.markEqualRange(start1, start2, end1, end2)
                matchForwardIdx(idx1, idx2)
            }
        }
        matchBackwardIdx(words1.count, words2.count)
        return fair(builder.finish())
    }

    private func clearLastRange() {
        lastStart1 = -1
        lastStart2 = -1
        lastEnd1 = -1
        lastEnd2 = -1
    }

    private func matchForwardIdx(_ idx1: Int, _ idx2: Int) {
        let start1 = idx1 == -1 ? 0 : words1[idx1].offset2
        let start2 = idx2 == -1 ? 0 : words2[idx2].offset2
        let end1 = idx1 + 1 == words1.count ? bytes1.count : words1[idx1 + 1].offset1
        let end2 = idx2 + 1 == words2.count ? bytes2.count : words2[idx2 + 1].offset1

        lastStart1 = start1
        lastStart2 = start2
        lastEnd1 = end1
        lastEnd2 = end2
    }

    private func matchBackwardIdx(_ idx1: Int, _ idx2: Int) {
        let start1 = idx1 == 0 ? 0 : words1[idx1 - 1].offset2
        let start2 = idx2 == 0 ? 0 : words2[idx2 - 1].offset2
        let end1 = idx1 == words1.count ? bytes1.count : words1[idx1].offset1
        let end2 = idx2 == words2.count ? bytes2.count : words2[idx2].offset1

        matchBackwardRange(start1, start2, end1, end2)
        clearLastRange()
    }

    private func matchBackwardRange(_ start1: Int, _ start2: Int, _ end1: Int, _ end2: Int) {
        if lastStart1 == start1, lastStart2 == start2 {
            matchRange(start1, start2, end1, end2)
            return
        }
        if lastStart1 < start1, lastStart2 < start2 {
            matchRange(lastStart1, lastStart2, lastEnd1, lastEnd2)
            matchRange(start1, start2, end1, end2)
            return
        }
        matchRange(
            min(lastStart1, start1), min(lastStart2, start2),
            max(lastEnd1, end1), max(lastEnd2, end2)
        )
    }

    private func matchRange(_ start1: Int, _ start2: Int, _ end1: Int, _ end2: Int) {
        if start1 == end1, start2 == end2 {
            return
        }
        let seq1 = String(bytes: Array(bytes1[start1..<end1]), encoding: .utf8) ?? ""
        let seq2 = String(bytes: Array(bytes2[start2..<end2]), encoding: .utf8) ?? ""
        let changes = comparePunctuation(seq1, seq2)
        for ch in changes.unchanged {
            builder.markEqualRange(
                start1 + ch.left.lowerBound, start2 + ch.right.lowerBound,
                start1 + ch.left.upperBound, start2 + ch.right.upperBound
            )
        }
    }
}
