# Phase 01: Path to Layer Name Conversion

**Depends on:** None
**Status:** Pending

---

## RED: Write Tests

**Objective:** Test `pathToLayerName` function that converts file paths to CSS layer names.

**Files:**

- `strata_test.go`

**Test Cases (table-driven):**

| Case | givePath | giveDir | wantLayerName | Notes |
|------|----------|---------|---------------|-------|
| `root_file` | `css/reset.css` | `css` | `reset` | Root file becomes own layer |
| `root_file_tokens` | `css/tokens.css` | `css` | `tokens` | Another root file |
| `nested_single` | `css/base/typography.css` | `css` | `base` | Single folder depth |
| `nested_sibling` | `css/base/links.css` | `css` | `base` | Same layer as typography |
| `deeply_nested` | `css/base/elements/buttons.css` | `css` | `base.elements` | Two-level nesting |
| `very_deep` | `css/a/b/c/d.css` | `css` | `a.b.c` | Deep nesting |
| `different_root` | `styles/main.css` | `styles` | `main` | Non-css root dir |
| `different_root_nested` | `assets/css/base/file.css` | `assets/css` | `base` | Nested root dir |

**Assertions:**

- Layer name derived from directory path, not filename
- Root files use filename (sans extension) as layer name
- Nested paths use dots as separators
- Dir prefix stripped correctly regardless of depth

**Edge Cases:**

- Single character names: `css/a/b.css` → `a`
- Hyphens in names: `css/my-layer/file.css` → `my-layer`

### Gate: RED

- [ ] Test file created with table-driven test cases
- [ ] All tests FAIL (pathToLayerName does not exist)
- [ ] Test coverage includes root files and nested directories

---

## GREEN: Implement

**Objective:** Implement `pathToLayerName` to convert paths to layer names.

**Files:**

- `strata.go`

**Implementation Guidance:**

```go
// pathToLayerName converts a file path to its CSS layer name.
//
// Implementation approach:
// 1. Strip dir prefix from path (e.g., "css/" from "css/base/file.css")
// 2. Get directory portion of relative path
// 3. If dir is "." (root file) -> return filename without extension
// 4. Otherwise -> replace "/" with "." to form layer name
//
// Use path package (not filepath) to ensure consistent "/" separators
// regardless of platform.
func pathToLayerName(filePath, dir string) string
```

### Gate: GREEN

- [ ] All tests from RED phase now PASS
- [ ] Test command: `go test -run ^TestPathToLayerName$ -v`
- [ ] Implementation uses `path` package for separator consistency

---

## REFACTOR: Quality

**Focus:** Code quality improvements, not new functionality.

**Review Areas:**

- **Naming**: Ensure function and variable names convey intent
- **Simplification**: Look for unnecessary string operations
- **Edge cases**: Verify trailing slash handling in dir parameter

### Gate: REFACTOR

- [ ] Variable names are clear and descriptive
- [ ] No unnecessary allocations or string operations
- [ ] Function documented with godoc comment

---

## Phase Complete

When all gates pass:

1. Update this file's status to **Complete**
2. Update index.md status table
3. Proceed to Phase 02

---

**Previous:** First phase
**Next:** [Phase 02: Build](02-build.md)
