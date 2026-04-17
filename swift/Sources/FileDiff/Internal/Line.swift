// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Wraps a line of text with its comparison policy, cached hash, and non-space character count.
final class Line: LineEquatable {
    let content: String
    let policy: ComparisonPolicy
    let hash: Int
    let nonSpaceChars: Int

    init(content: String, policy: ComparisonPolicy) {
        self.content = content
        self.policy = policy
        hash = TextUtils.hashCode(content, policy: policy)
        nonSpaceChars = TextUtils.countNonSpaceChars(content)
    }

    func equals(_ other: Line) -> Bool {
        if self === other {
            return true
        }
        if hash != other.hash {
            return false
        }
        return TextUtils.isEqual(content, other.content, policy: policy)
    }

    func equals(_ other: LineEquatable) -> Bool {
        guard let other = other as? Line else {
            return false
        }
        return equals(other)
    }
}

/// Converts raw text lines to `Line` objects tagged with the supplied policy.
func toLines(_ lines: [String], policy: ComparisonPolicy) -> [Line] {
    lines.map { Line(content: $0, policy: policy) }
}

/// Re-tags an array of lines with a new comparison policy, preserving unchanged entries.
func convertMode(_ lines: [Line], policy: ComparisonPolicy) -> [Line] {
    lines.map { line in
        line.policy == policy ? line : Line(content: line.content, policy: policy)
    }
}
