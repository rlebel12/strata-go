# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go test ./...                           # Run all tests
go test -run TestBuild ./...            # Run specific test
go test -v ./...                        # Verbose test output
go vet ./...                            # Static analysis
go build -o /dev/null ./...             # Verify compilation
```

Pre-commit hooks (lefthook) run go-vet, go-build, and go-test automatically on commit.

## Architecture

Single-package library (`package strata`) that transforms CSS directory structures into cascade-layered CSS output.

**Public API:**
- `Build(sources ...Source) (string, error)` - Builds CSS from one or more sources
- `BuildWithHash(sources ...Source) (css, hash string, err error)` - Same as Build, plus SHA-256 content hash (16 hex chars)
- `Source` struct:
  - `FS fs.FS` - Filesystem to read from (use `fs.Sub()` to create sub-filesystems)
  - `Prefix string` - Optional namespace prefix for layer names

**Layer derivation:**
- Root files → layer name from filename (e.g., `reset.css` → `reset`)
- Nested dirs → dot-separated layer name (e.g., `base/elements/` → `base.elements`)
- Prefix → prepended to layer name (e.g., Prefix: "comp", `button.css` → `comp.button`)
- Ordering: Sources processed in slice order; within each source, depth-first (shallow before deep), then alphabetical

**Multi-directory usage:** Use `fs.Sub()` to create sub-filesystems, then pass multiple `Source` structs (e.g., styles/, components/, routes/)

**Constraints:** Standard library only, no external dependencies.
