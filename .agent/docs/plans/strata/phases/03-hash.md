# Phase 03: BuildWithHash Function

**Depends on:** Phase 02
**Status:** Complete

---

## RED: Write Tests

**Objective:** Test `BuildWithHash` function that returns CSS with content hash for cache busting.

**Files:**

- `strata_test.go`

**Test Data Structures:**

```go
// hash_basic
hashBasicFS := fstest.MapFS{
    "css/reset.css": {Data: []byte("* { margin: 0; }")},
}

// hash_different_content
hashDifferentFS := fstest.MapFS{
    "css/reset.css": {Data: []byte("* { margin: 1px; }")},
}
```

**Test Cases (table-driven):**

| Case | giveFS | giveDir | wantHashLen | wantCSSNonEmpty | Notes |
|------|--------|---------|-------------|-----------------|-------|
| `returns_hash` | `hashBasicFS` | `css` | 16 | true | Hash is 16 hex chars |
| `empty_fs_empty_hash` | `fstest.MapFS{}` | `css` | 0 | false | Empty returns empty |

**Discrete Tests:**

- **Test hash stability**: Call `BuildWithHash` twice with same input, verify hashes match
- **Test hash uniqueness**: Call with `hashBasicFS` and `hashDifferentFS`, verify hashes differ
- **Test hash is lowercase hex**: Verify hash matches regex `^[0-9a-f]{16}$`
- **Test CSS matches Build**: Call both `Build` and `BuildWithHash`, verify CSS output identical
- **Test error propagation**: Build errors bubble up through BuildWithHash

**Assertions:**

- Hash length is exactly 16 characters (or 0 for empty)
- Hash matches regex pattern `^[0-9a-f]{16}$` (lowercase hex only)
- CSS output matches `Build()` output exactly (call both and compare)
- Same input → same hash (call twice, compare hashes)
- Different input → different hash (use different fs, verify hash differs)

**Edge Cases:**

- Empty filesystem returns empty CSS and empty hash (not hash of empty string)
- Single byte difference produces different hash

### Gate: RED

- [x] Test file created with hash validation tests
- [x] All tests FAIL (BuildWithHash does not exist)
- [x] Tests verify hash format, stability, and uniqueness

---

## GREEN: Implement

**Objective:** Implement `BuildWithHash` to wrap Build with SHA-256 hashing.

**Files:**

- `strata.go`

**Implementation Guidance:**

```go
// BuildWithHash returns the built CSS and a content hash for cache busting.
//
// Implementation approach:
// 1. Call Build(fsys, dir) to get CSS string
// 2. If error -> return "", "", wrapped error
// 3. If empty CSS -> return "", "", nil
// 4. Compute SHA-256 of CSS bytes
// 5. Take first 8 bytes of hash
// 6. Encode as 16 hex characters (lowercase)
// 7. Return css, hash, nil
//
// The hash enables immutable caching:
//   <link rel="stylesheet" href="/static/styles.{hash}.css">
func BuildWithHash(fsys fs.FS, dir string) (css string, hash string, err error)
```

**Key Details:**

- Use `crypto/sha256` for hashing
- Use `encoding/hex.EncodeToString()` for hex encoding
- Take `[:8]` of hash bytes → 16 hex chars
- Empty CSS should return empty hash, not hash of empty string

### Gate: GREEN

- [x] All tests from RED phase now PASS
- [x] Test command: `go test -run ^TestBuildWithHash$ -v`
- [x] Hash format matches spec (16 lowercase hex chars)

---

## REFACTOR: Quality

**Focus:** Code quality improvements, not new functionality.

**Review Areas:**

- **Simplification**: Ensure hash computation is straightforward
- **Documentation**: Add usage example in godoc
- **Consistency**: Verify error handling matches Build function

### Gate: REFACTOR

- [x] Function documented with godoc comment and usage example
- [x] Error handling consistent with Build function
- [x] No unnecessary allocations in hash computation

---

## Phase Complete

When all gates pass:

1. Update this file's status to **Complete**
2. Update index.md status table
3. Run `go test ./...` to verify all tests pass
4. Implementation complete - ready for plan cleanup

---

**Previous:** [Phase 02: Build](02-build.md)
**Next:** Final phase
