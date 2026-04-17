# Go Implementation

Canonical Go port of the JetBrains multi-stage diff engine. Public API lives under `go/comparison`.

## Technology Stack

- Go 1.22+
- `golangci-lint` for linting
- Standard library only — do not add third-party dependencies.

## Package layout

```text
go/
├── comparison/       # Core comparison logic: policies, matchers, iterables, ByLine, ByWord, ByChar, text utilities
├── fragment/         # Diff result types: DiffFragment (word/char), LineFragment (line)
├── histogram/        # Histogram Diff algorithm
├── myers/            # Myers O(ND) LCS algorithm, BitSet, Reindexer, Change builder
├── patience/         # Patience Diff algorithm (LIS-based)
└── util/             # Shared types: Range, LineOffsets, Enumerator
```

## Testing

- Use the standard `testing` package.
- Test each matching stage (Line-level, Word-level, Character-level) independently.
- Validate matching outcomes against the shared testing datasets / JSON schema evaluations.

## Code validation

ALWAYS run the following before declaring work complete:

- `make go-check` — lint/format
- `make go-test` and `make go-race` — functional and race validation

New features or refactors must not degrade performance compared to the default Myers fallback algorithm.
