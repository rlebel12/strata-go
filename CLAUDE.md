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

Single-package library (`package strata`) that transforms a CSS directory structure into cascade-layered CSS output.

**Public API:**
- `Build(fsys fs.FS, dir string) (string, error)` - Walks filesystem, returns layered CSS
- `BuildWithHash(fsys fs.FS, dir string) (css, hash string, err error)` - Same as Build, plus SHA-256 content hash (16 hex chars)

**Layer derivation:**
- Root files → layer name from filename (e.g., `css/reset.css` → `reset`)
- Nested dirs → dot-separated layer name (e.g., `css/base/elements/` → `base.elements`)
- Ordering: depth-first (shallow before deep), then alphabetical

**Constraints:** Standard library only, no external dependencies.
