# Enhanced Diff Engine

Enhanced Diff Engine is a port of the JetBrains IDE's multi-stage diff engine (Apache 2.0) combined with the Histogram Diff algorithm for the line matching phase. The repository is a multi-language monorepo hosting the canonical Go implementation under `go/` and a Swift Package Manager port under `swift/` (module name `FileDiff`, macOS 13+, Swift 6.0+). It is designed to provide highly readable and intuitive diff results for both macro-level structural changes (like method additions/moves) and micro-level detailed modifications (like word or character changes).

## Key requirements

- **High Readability**: Clear distinction of structural code changes and in-line token modifications to improve the code-review experience.
- **Pluggable Architecture**: The line matching algorithms (Myers, Patience, Histogram) must be abstracted through an interface and easily swappable.
- **Minimal dependencies**: Prefer to use standard library.

## Architecture

### Technology Stack

- Go 1.26+ with `golangci-lint` for the Go implementation
- Swift 6.0+ / macOS 13+ with SwiftFormat + SwiftLint for the Swift port
- Root `Package.swift` points at `swift/Sources/FileDiff` and `swift/Tests/FileDiffTests`

### Repository structure

```text
go/
├── comparison/       # Core comparison logic: policies, matchers, iterables, ByLine, ByWord, ByChar, text utilities
├── fragment/         # Diff result types: DiffFragment (word/char), LineFragment (line)
├── histogram/        # Histogram Diff algorithm
├── myers/            # Myers O(ND) LCS algorithm, BitSet, Reindexer, Change builder
├── patience/         # Patience Diff algorithm (LIS-based)
└── util/             # Shared types: Range, LineOffsets, Enumerator
swift/
├── Sources/FileDiff/           # Public API (ByLine, ByWord, ByChar, matchers) + Internal/ helpers
└── Tests/FileDiffTests/        # swift-testing suites mirroring the Go test coverage
references/           # JetBrains IntelliJ diff source (Kotlin) — read-only reference
```

## Testing

- Go: use the standard `testing` package.
- Swift: use `swift-testing` (`@Suite` / `@Test` / `#expect`); do not introduce XCTest.
- Test each matching stage (Line-level, Word-level, Character-level) independently in both languages.
- Validate matching outcomes utilizing standard testing datasets and JSON schema evaluations.

### Code validation

- Go: ALWAYS run `make check` to lint/format and `make test` / `make race` to validate.
- Swift: ALWAYS run `make swift-check` (wraps SwiftFormat + SwiftLint) and `make swift-test`.
- Assure new features or components don't degrade performance compared to the default Myers fallback algorithm.
