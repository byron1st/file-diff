# Enhanced Diff Engine

Enhanced Diff Engine is a port of the JetBrains IDE's multi-stage diff engine (Apache 2.0) combined with the Histogram Diff algorithm for the line matching phase. The repository is a multi-language monorepo hosting the canonical Go implementation under `go/` and a Swift Package Manager port under `swift/` (module name `FileDiff`). It is designed to provide highly readable and intuitive diff results for both macro-level structural changes (like method additions/moves) and micro-level detailed modifications (like word or character changes).

## Key requirements

- **High Readability**: Clear distinction of structural code changes and in-line token modifications to improve the code-review experience.
- **Pluggable Architecture**: The line matching algorithms (Myers, Patience, Histogram) must be abstracted through an interface and easily swappable.
- **Minimal dependencies**: Prefer to use the standard library.
- **Cross-language parity**: The Go and Swift implementations must share the same API shape and byte-offset (UTF-8) semantics so their outputs stay aligned.

## Repository structure

```text
go/                   # Go implementation — see go/CLAUDE.md
swift/                # Swift Package Manager port — see swift/CLAUDE.md
docs/                 # User-facing guides (go-guide.md, swift-guide.md)
references/           # JetBrains IntelliJ diff source (Kotlin) — read-only reference
```

## Language-specific instructions

Per-language tech stack, package layout, testing conventions, and validation commands live in:

- [go/CLAUDE.md](go/CLAUDE.md)
- [swift/CLAUDE.md](swift/CLAUDE.md)

When working inside one of those subtrees, follow the instructions in that subtree's `CLAUDE.md` in addition to this file.

## Testing (shared expectations)

- Test each matching stage (Line-level, Word-level, Character-level) independently in both languages.
- Validate matching outcomes against standard testing datasets and JSON schema evaluations shared across implementations.
- New features or components must not degrade performance compared to the default Myers fallback algorithm.
