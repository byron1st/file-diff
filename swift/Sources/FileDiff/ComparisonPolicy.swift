// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Controls how whitespace differences are treated during comparison.
public enum ComparisonPolicy: Sendable {
    /// Compare as-is.
    case `default`
    /// Treat leading/trailing whitespace on each line as insignificant.
    case trimWhitespaces
    /// Ignore all whitespace differences.
    case ignoreWhitespaces
}
