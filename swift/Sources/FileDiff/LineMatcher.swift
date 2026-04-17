// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Swift (via the Go port in this repository).

/// Matches lines between two texts and produces a `FairDiffIterable`
/// describing which lines are changed and which are unchanged.
///
/// Implementations include Myers, Patience, and Histogram matchers.
public protocol LineMatcher {
    func match(_ left: [String], _ right: [String], policy: ComparisonPolicy) -> FairDiffIterable
}
