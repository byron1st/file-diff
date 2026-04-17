# Swift Implementation

Swift Package Manager port of the diff engine. Mirrors the Go API surface and behaviour (including UTF-8 byte-offset semantics).

## Technology Stack

- Swift 6.0+
- macOS 13+
- SwiftFormat + SwiftLint for lint/format
- Root `Package.swift` points at `swift/Sources/FileDiff` and `swift/Tests/FileDiffTests`.
- SPM product / module name: `FileDiff`.
- Standard library / Foundation only — do not add third-party dependencies.

## Package layout

```text
swift/
├── Sources/FileDiff/           # Public API (ByLine, ByWord, ByChar, matchers) + Internal/ helpers
└── Tests/FileDiffTests/        # swift-testing suites mirroring the Go test coverage
```

## Testing

- Use `swift-testing` (`@Suite` / `@Test` / `#expect`). **Do not** introduce XCTest.
- Test each matching stage (Line-level, Word-level, Character-level) independently.
- Mirror the Go test datasets so both ports stay in lock-step.

## Code validation

ALWAYS run the following before declaring work complete:

- `make swift-check` — wraps SwiftFormat + SwiftLint
- `make swift-test`

New features or refactors must not degrade performance compared to the default Myers fallback algorithm.
