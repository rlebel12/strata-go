# Strata: Cascade Layer CSS Framework

## Problem Statement

CSS cascade management is error-prone. Developers either:
- Use specificity hacks (`!important`, deep nesting, ID selectors)
- Manually maintain `@layer` declarations that drift from actual file structure
- Adopt heavyweight solutions (CSS-in-JS, Tailwind) with significant tooling overhead

The CSS `@layer` feature solves cascade ordering but requires manual declaration ordering that must stay synchronized with actual stylesheet organization.

## Proposed Solution

A Go library that derives `@layer` cascade priority from folder hierarchy. Convention replaces configuration:

- **Folder depth = cascade priority** (deeper wins)
- **Folder name = layer name** (dot-separated for nesting)
- **Sibling files = same layer** (orthogonal concerns, no conflict)

Input structure:
```
css/
  reset.css           → @layer reset
  base/
    typography.css    → @layer base
    elements/
      buttons.css     → @layer base.elements
```

## Input/Output Example

**Input files:**

```
css/reset.css
css/base/typography.css
css/base/links.css
css/base/elements/buttons.css
```

```css
/* css/reset.css */
* { margin: 0; box-sizing: border-box; }

/* css/base/typography.css */
h1 { font-size: 2rem; }

/* css/base/links.css */
a { color: blue; }

/* css/base/elements/buttons.css */
button { padding: 0.5rem 1rem; }
```

**Output (single string):**

```css
@layer reset, base, base.elements;

@layer reset {
* { margin: 0; box-sizing: border-box; }
}

@layer base {
h1 { font-size: 2rem; }
a { color: blue; }
}

@layer base.elements {
button { padding: 0.5rem 1rem; }
}
```

**Layer ordering:** `reset` < `base` < `base.elements` (deeper = higher priority, wins conflicts)

**File concatenation:** Within `base`, `links.css` and `typography.css` are concatenated alphabetically.

## Design Decisions

### D1: Use `fs.FS` interface for input

**Decision:** Accept `fs.FS` rather than filesystem paths.

**Rationale:**
- Enables `embed.FS` for compiled-in CSS (zero-runtime-IO serving)
- Enables `fstest.MapFS` for testing without fixtures
- Enables `os.DirFS` for traditional filesystem access
- Standard library interface, no custom abstractions

**Tradeoffs:**
- Users must wrap paths with `os.DirFS()` for filesystem access
- Slightly more verbose API for simple cases

### D2: Pass CSS through unmodified

**Decision:** Treat CSS content as opaque bytes. No parsing, no validation.

**Rationale:**
- Avoids CSS parser dependency and complexity
- Browsers provide superior error messages with source context
- Future CSS features work automatically
- Faster build times

**Tradeoffs:**
- Syntax errors only surface at runtime in browser
- Cannot detect conflicting selectors between sibling files

### D3: Depth-first, then alphabetical ordering

**Decision:** Sort layers by depth ascending, then alphabetically within same depth.

**Rationale:**
- Depth ordering ensures nested layers override parents (CSS `@layer` semantics)
- Alphabetical provides deterministic output for same-depth layers
- No configuration needed - structure is the configuration

**Tradeoffs:**
- Layer order can't be customized without renaming folders
- Users must understand depth = priority mental model

### D4: Root files become own layers

**Decision:** Files directly in the CSS root (e.g., `css/reset.css`) become their own layer named after the file.

**Rationale:**
- Handles common patterns: reset, tokens, variables
- No special directory needed for single-file layers
- Filename naturally describes the layer

**Tradeoffs:**
- Many root files create many top-level layers
- Could encourage flat structures over intentional hierarchy

### D5: Library only, no CLI

**Decision:** Provide Go library. No standalone CLI tool.

**Rationale:**
- Primary use case is embedded serving (single binary deployment)
- CLI would duplicate what `go run` with a trivial main achieves
- Keeps scope minimal

**Tradeoffs:**
- Non-Go projects can't easily use the tool
- No watch mode out of the box

## Scope

### In Scope

- Walk `fs.FS` and collect `.css` files
- Derive layer names from directory paths
- Sort layers by depth, then alphabetically
- Generate `@layer` declaration header
- Wrap each layer's content in `@layer name { ... }`
- Concatenate files within same layer (alphabetically)
- Generate content hash for cache busting
- Return combined CSS as string

### Out of Scope

- CSS parsing or validation
- Source maps
- Minification
- Autoprefixing
- Watch mode / hot reload (user responsibility)
- HTTP serving (user responsibility)
- CLI tool

## Interface Definitions

```go
package strata

import (
    "io/fs"
)

// Build walks the filesystem rooted at dir and returns CSS with @layer declarations.
// Files are grouped into layers based on their directory path.
// Layers are sorted by depth (ascending) then alphabetically.
//
// Example:
//   css, err := strata.Build(cssFS, "css")
func Build(fsys fs.FS, dir string) (string, error)

// BuildWithHash returns the built CSS and a content hash suitable for cache busting.
// The hash is the first 16 hex characters of the SHA-256 digest.
//
// Example:
//   css, hash, err := strata.BuildWithHash(cssFS, "css")
//   // Use hash in URL: /static/styles.{hash}.css
func BuildWithHash(fsys fs.FS, dir string) (css string, hash string, err error)
```

### Internal Types (unexported)

```go
// layer represents a CSS cascade layer being built.
type layer struct {
    name    string        // e.g., "base.elements"
    depth   int           // number of dots in name (0 for root)
    content *bytes.Buffer // concatenated CSS from all files in layer
}
```

### Path-to-Layer Mapping

| Path | Layer Name |
|------|------------|
| `css/reset.css` | `reset` |
| `css/base/typography.css` | `base` |
| `css/base/links.css` | `base` |
| `css/base/elements/buttons.css` | `base.elements` |

## Dependencies

### Required
- Standard library only (`io/fs`, `bytes`, `crypto/sha256`, `encoding/hex`, `path/filepath`, `sort`, `strings`)

### Development
- `testing/fstest` for in-memory filesystem tests

## Testing Strategy

### Approach: Filesystem-based with `fstest.MapFS`

Tests create in-memory filesystems representing CSS directory structures, call `Build()`, and assert on the complete output string.

### Test Cases

1. **Single root file** - `css/reset.css` → `@layer reset`
2. **Multiple root files** - Alphabetical ordering of top-level layers
3. **Single nested directory** - `css/base/*.css` → `@layer base`
4. **Multiple files same layer** - Concatenation order is alphabetical
5. **Deeply nested** - `css/a/b/c/d.css` → `@layer a.b.c`
6. **Mixed depths** - Verify depth-first sorting
7. **Empty directory** - Ignored (no empty layers)
8. **Non-CSS files** - Ignored
9. **Hash stability** - Same input produces same hash
10. **Hash uniqueness** - Different input produces different hash

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Layer name collisions (file and folder with same name at same level) | Low | Medium | Document: avoid `css/base.css` alongside `css/base/` |
| Large CSS causing memory pressure | Low | Low | Streaming output in future version if needed |
| Platform path separator issues | Medium | High | Use `path` (not `filepath`) for layer name construction |
| Users expect CSS validation | Medium | Low | Document pass-through behavior clearly |

## File Structure

```
strata-go/
├── go.mod
├── strata.go           # Build, BuildWithHash
├── strata_test.go      # Table-driven tests with fstest.MapFS
└── .agent/
    └── docs/
        └── plans/
            └── strata/
                └── README.md  # This file
```
