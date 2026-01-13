# Phase 02: Build Function

**Depends on:** Phase 01
**Status:** Pending

---

## RED: Write Tests

**Objective:** Test `Build` function that walks fs.FS and generates layered CSS output.

**Files:**

- `strata_test.go`

**Test Data Structures:**

```go
// single_root_file
singleRootFS := fstest.MapFS{
    "css/reset.css": {Data: []byte("* { margin: 0; }")},
}

// multiple_root_files
multipleRootFS := fstest.MapFS{
    "css/reset.css":  {Data: []byte("a")},
    "css/tokens.css": {Data: []byte("b")},
}

// single_nested_dir
singleNestedFS := fstest.MapFS{
    "css/base/typography.css": {Data: []byte("h1 {}")},
    "css/base/links.css":      {Data: []byte("a {}")},
}

// mixed_depths
mixedDepthsFS := fstest.MapFS{
    "css/reset.css":              {Data: []byte("x")},
    "css/base/file.css":          {Data: []byte("y")},
    "css/base/elements/btn.css":  {Data: []byte("z")},
}

// deeply_nested
deeplyNestedFS := fstest.MapFS{
    "css/a/b/c/d.css": {Data: []byte("x")},
}

// ignores_non_css
ignoresNonCSSFS := fstest.MapFS{
    "css/readme.md":  {Data: []byte("# hi")},
    "css/reset.css":  {Data: []byte("x")},
}
```

**Test Cases (table-driven):**

| Case | giveFS | giveDir | wantLayerDecl | wantLayerCount | Notes |
|------|--------|---------|---------------|----------------|-------|
| `single_root_file` | `singleRootFS` | `css` | `@layer reset;` | 1 | Single layer |
| `multiple_root_files` | `multipleRootFS` | `css` | `@layer reset, tokens;` | 2 | Alphabetical |
| `single_nested_dir` | `singleNestedFS` | `css` | `@layer base;` | 1 | Files concat alphabetically |
| `mixed_depths` | `mixedDepthsFS` | `css` | `@layer reset, base, base.elements;` | 3 | Depth-first then alpha |
| `deeply_nested` | `deeplyNestedFS` | `css` | `@layer a.b.c;` | 1 | Deep nesting |
| `empty_result` | `fstest.MapFS{}` | `css` | `""` | 0 | No CSS files found |
| `ignores_non_css` | `ignoresNonCSSFS` | `css` | `@layer reset;` | 1 | Non-.css ignored |

**Discrete Tests:**

- **Test fs.WalkDir error propagation**: Use broken fs.FS that returns error from Open, verify error is wrapped and returned

**Assertions:**

- Layer declaration header lists layers in correct order (depth-first, then alphabetical)
- Each layer block contains concatenated file contents
- Files within same layer concatenated alphabetically (links.css before typography.css)
- Empty directories produce no empty layers
- Non-.css files ignored

**Edge Cases:**

- Empty filesystem returns empty string (not error)
- CSS files with no content still create layer

### Gate: RED

- [ ] Test file created with table-driven test cases using fstest.MapFS
- [ ] All tests FAIL (Build function does not exist)
- [ ] Test coverage includes layer ordering, concatenation, and edge cases

---

## GREEN: Implement

**Objective:** Implement `Build` to walk filesystem and generate layered CSS.

**Files:**

- `strata.go`

**Implementation Guidance:**

```go
// layer represents a CSS cascade layer being built.
type layer struct {
    name    string
    depth   int
    content *bytes.Buffer
}

// Build walks the filesystem and returns CSS with @layer declarations.
//
// Implementation approach:
// 1. Create map[string]*layer to collect layers by name
// 2. Walk fs.FS using fs.WalkDir
// 3. For each .css file:
//    a. Read file contents
//    b. Derive layer name using pathToLayerName
//    c. Create layer if not exists, append content
// 4. Sort layers: depth ascending, then name alphabetically
// 5. Generate output:
//    a. @layer declaration header (comma-separated names)
//    b. Each layer block: @layer name { content }
//
// File concatenation within layers:
// - Read each file's content
// - Append to layer's bytes.Buffer with trailing newline
// - Sort files alphabetically before processing
//
// Error handling:
// - fs.WalkDir errors -> return "", fmt.Errorf("walk filesystem: %w", err)
// - fs.ReadFile errors -> return "", fmt.Errorf("read %s: %w", path, err)
// - Empty fs (no .css files found) -> return "", nil (not an error condition)
//
// Detection: Track whether any .css files were encountered during walk.
// If map[string]*layer is empty after walk completes → empty result.
func Build(fsys fs.FS, dir string) (string, error)
```

**Key Details:**

- Use `path` package (not `filepath`) for consistent "/" separators
- Layer depth = `strings.Count(name, ".")`
- Sort with `sort.Slice` using depth then name comparison
- Collect all file paths first, sort, then process for deterministic concatenation order

### Gate: GREEN

- [ ] All tests from RED phase now PASS
- [ ] Test command: `go test -run ^TestBuild$ -v`
- [ ] Implementation handles empty fs gracefully

---

## REFACTOR: Quality

**Focus:** Code quality improvements, not new functionality.

**Review Areas:**

- **Duplication**: Extract layer sorting into helper if reused
- **Naming**: Ensure `layer` struct fields are clear
- **Buffer handling**: Pre-allocate buffer sizes if beneficial
- **Error messages**: Use `fmt.Errorf("context: %w", err)` pattern

### Gate: REFACTOR

- [ ] No code duplication in layer handling
- [ ] Error messages provide context about what operation failed
- [ ] Function documented with godoc comment and example

---

## Phase Complete

When all gates pass:

1. Update this file's status to **Complete**
2. Update index.md status table
3. Proceed to Phase 03

---

**Previous:** [Phase 01: Path to Layer](01-path-to-layer.md)
**Next:** [Phase 03: Hash](03-hash.md)
