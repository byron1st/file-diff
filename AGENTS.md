# Enhanced Diff Engine

Enhanced Diff Engine is a Go-based port of the JetBrains IDE's multi-stage diff engine (Apache 2.0) combined with the Histogram Diff algorithm for the line matching phase. It is designed to provide highly readable and intuitive diff results for both macro-level structural changes (like method additions/moves) and micro-level detailed modifications (like word or character changes).

## Key requirements

- **High Readability**: Clear distinction of structural code changes and in-line token modifications to improve the code-review experience.
- **Pluggable Architecture**: The line matching algorithms (Myers, Patience, Histogram) must be abstracted through an interface and easily swappable.

## Architecture

### Technology Stack

- Go 1.26+
- `golangci-lint` for linting

### Repository structure

```text
diff/
├── comparison/       # Core comparison logic: policies, matchers, iterables, ByLine, text utilities
├── fragment/         # Diff result types: DiffFragment (word/char), LineFragment (line)
├── myers/            # Myers O(ND) LCS algorithm, BitSet, Reindexer, Change builder
└── util/             # Shared types: Range, LineOffsets, Enumerator
references/           # JetBrains IntelliJ diff source (Kotlin) — read-only reference
```

## Testing

- Use Go's standard `testing` package.
- Test each matching stage (Line-level, Word-level, Character-level) independently.
- Validate matching outcomes utilizing standard testing datasets and JSON schema evaluations.

### Code validation

- ALWAYS run `make check` to verify code.
- ALWAYS run `make test` and `make race` to validate code.
- Assure new features or components don't degrade performance compared to the default Myers fallback algorithm.
