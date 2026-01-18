# strata

CSS cascade layer ordering from directory structure.

## Install

```bash
go get github.com/rlebel12/strata-go
```

## Usage

### Single Directory

```go
import "github.com/rlebel12/strata-go"

// Build CSS with layers from directory structure
css, err := strata.Build(strata.Source{
    FS:  os.DirFS("."),
    Dir: "css",
})

// Or with a content hash for cache busting
css, hash, err := strata.BuildWithHash(strata.Source{
    FS:  os.DirFS("."),
    Dir: "css",
})
// hash: 16 lowercase hex chars (e.g., "a1b2c3d4e5f67890")
```

### Multiple Directories (Co-located CSS)

For projects with CSS co-located alongside components and routes:

```go
css, err := strata.Build(
    strata.Source{FS: os.DirFS("."), Dir: "styles"},      // First: resets, tokens
    strata.Source{FS: os.DirFS("."), Dir: "components"},  // Second: components
    strata.Source{FS: os.DirFS("."), Dir: "routes"},      // Third: routes
)
```

### With Prefixes (Namespacing)

Use prefixes to namespace layers from different directories:

```go
css, err := strata.Build(
    strata.Source{FS: os.DirFS("."), Dir: "styles"},                   // Layers: reset, tokens
    strata.Source{FS: os.DirFS("."), Dir: "components", Prefix: "c"},  // Layers: c.button, c.card
    strata.Source{FS: os.DirFS("."), Dir: "routes", Prefix: "page"},   // Layers: page.auth, page.home
)
// Output: @layer reset, tokens, c.button, c.card, page.auth, page.home;
```

## Directory Structure

The directory hierarchy determines layer names and ordering:

```
css/
├── reset.css        → @layer reset
├── tokens.css       → @layer tokens
├── base/
│   ├── typography.css   → @layer base
│   └── links.css        → @layer base
└── components/
    └── buttons/
        └── btn.css      → @layer components.buttons
```

**Rules:**
- Root files become individual layers (filename without extension)
- Nested directories use dot notation for layer names
- Files in the same directory are concatenated alphabetically
- Layers are ordered by depth (shallow first), then alphabetically

## Output

```css
@layer reset, tokens, base, components.buttons;
@layer reset {
/* contents of reset.css */
}
@layer tokens {
/* contents of tokens.css */
}
@layer base {
/* contents of links.css */
/* contents of typography.css */
}
@layer components.buttons {
/* contents of btn.css */
}
```

## Requirements

Go 1.24+ (standard library only, no external dependencies)
