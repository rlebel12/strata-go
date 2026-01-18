package strata

import (
	"errors"
	"io/fs"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

func TestPathToLayerName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		givePath      string
		giveDir       string
		wantLayerName string
	}{
		// Root files become own layer
		{
			name:          "root_file",
			givePath:      "css/reset.css",
			giveDir:       "css",
			wantLayerName: "reset",
		},
		{
			name:          "root_file_tokens",
			givePath:      "css/tokens.css",
			giveDir:       "css",
			wantLayerName: "tokens",
		},
		// Single folder depth
		{
			name:          "nested_single",
			givePath:      "css/base/typography.css",
			giveDir:       "css",
			wantLayerName: "base",
		},
		{
			name:          "nested_sibling",
			givePath:      "css/base/links.css",
			giveDir:       "css",
			wantLayerName: "base",
		},
		// Multi-level nesting uses dots
		{
			name:          "deeply_nested",
			givePath:      "css/base/elements/buttons.css",
			giveDir:       "css",
			wantLayerName: "base.elements",
		},
		{
			name:          "very_deep",
			givePath:      "css/a/b/c/d.css",
			giveDir:       "css",
			wantLayerName: "a.b.c",
		},
		// Different root directories
		{
			name:          "different_root",
			givePath:      "styles/main.css",
			giveDir:       "styles",
			wantLayerName: "main",
		},
		{
			name:          "different_root_nested",
			givePath:      "assets/css/base/file.css",
			giveDir:       "assets/css",
			wantLayerName: "base",
		},
		// Edge cases
		{
			name:          "single_char_name",
			givePath:      "css/a/b.css",
			giveDir:       "css",
			wantLayerName: "a",
		},
		{
			name:          "hyphen_in_name",
			givePath:      "css/my-layer/file.css",
			giveDir:       "css",
			wantLayerName: "my-layer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := pathToLayerName(tt.givePath, tt.giveDir)
			if got != tt.wantLayerName {
				t.Errorf("pathToLayerName(%q, %q) = %q, want %q",
					tt.givePath, tt.giveDir, got, tt.wantLayerName)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	t.Parallel()

	// Test data structures
	singleRootFS := fstest.MapFS{
		"css/reset.css": {Data: []byte("* { margin: 0; }")},
	}

	multipleRootFS := fstest.MapFS{
		"css/reset.css":  {Data: []byte("a")},
		"css/tokens.css": {Data: []byte("b")},
	}

	singleNestedFS := fstest.MapFS{
		"css/base/typography.css": {Data: []byte("h1 {}")},
		"css/base/links.css":      {Data: []byte("a {}")},
	}

	mixedDepthsFS := fstest.MapFS{
		"css/reset.css":             {Data: []byte("x")},
		"css/base/file.css":         {Data: []byte("y")},
		"css/base/elements/btn.css": {Data: []byte("z")},
	}

	deeplyNestedFS := fstest.MapFS{
		"css/a/b/c/d.css": {Data: []byte("x")},
	}

	ignoresNonCSSFS := fstest.MapFS{
		"css/readme.md": {Data: []byte("# hi")},
		"css/reset.css": {Data: []byte("x")},
	}

	tests := []struct {
		name           string
		giveFS         fs.FS
		giveDir        string
		wantLayerDecl  string
		wantLayerCount int
	}{
		{
			name:           "single_root_file",
			giveFS:         singleRootFS,
			giveDir:        "css",
			wantLayerDecl:  "@layer reset;",
			wantLayerCount: 1,
		},
		{
			name:           "multiple_root_files",
			giveFS:         multipleRootFS,
			giveDir:        "css",
			wantLayerDecl:  "@layer reset, tokens;",
			wantLayerCount: 2,
		},
		{
			name:           "single_nested_dir",
			giveFS:         singleNestedFS,
			giveDir:        "css",
			wantLayerDecl:  "@layer base;",
			wantLayerCount: 1,
		},
		{
			name:           "mixed_depths",
			giveFS:         mixedDepthsFS,
			giveDir:        "css",
			wantLayerDecl:  "@layer base, reset, base.elements;",
			wantLayerCount: 3,
		},
		{
			name:           "deeply_nested",
			giveFS:         deeplyNestedFS,
			giveDir:        "css",
			wantLayerDecl:  "@layer a.b.c;",
			wantLayerCount: 1,
		},
		{
			name:           "empty_result",
			giveFS:         fstest.MapFS{},
			giveDir:        "css",
			wantLayerDecl:  "",
			wantLayerCount: 0,
		},
		{
			name:           "ignores_non_css",
			giveFS:         ignoresNonCSSFS,
			giveDir:        "css",
			wantLayerDecl:  "@layer reset;",
			wantLayerCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Build(Source{FS: tt.giveFS, Dir: tt.giveDir})
			if err != nil {
				t.Fatalf("Build() error = %v, want nil", err)
			}

			// Check empty case
			if tt.wantLayerCount == 0 {
				if got != "" {
					t.Errorf("Build() = %q, want empty string", got)
				}
				return
			}

			// Check layer declaration header
			if !strings.HasPrefix(got, tt.wantLayerDecl) {
				t.Errorf("Build() layer declaration = %q, want prefix %q",
					strings.SplitN(got, "\n", 2)[0], tt.wantLayerDecl)
			}

			// Count layer blocks
			layerCount := strings.Count(got, "@layer ") - 1 // subtract header declaration
			if layerCount != tt.wantLayerCount {
				t.Errorf("Build() layer count = %d, want %d", layerCount, tt.wantLayerCount)
			}
		})
	}
}

func TestBuild_concatenation_order(t *testing.T) {
	t.Parallel()

	// Files within same layer should be concatenated alphabetically
	testFS := fstest.MapFS{
		"css/base/typography.css": {Data: []byte("h1 {}")},
		"css/base/links.css":      {Data: []byte("a {}")},
	}

	got, err := Build(Source{FS: testFS, Dir: "css"})
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	// links.css comes before typography.css alphabetically
	linksIdx := strings.Index(got, "a {}")
	typoIdx := strings.Index(got, "h1 {}")

	if linksIdx == -1 || typoIdx == -1 {
		t.Fatalf("Build() missing expected content, got: %s", got)
	}

	if linksIdx > typoIdx {
		t.Errorf("Build() links.css content should come before typography.css content")
	}
}

func TestBuild_empty_file_creates_layer(t *testing.T) {
	t.Parallel()

	testFS := fstest.MapFS{
		"css/empty.css": {Data: []byte("")},
	}

	got, err := Build(Source{FS: testFS, Dir: "css"})
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	if !strings.Contains(got, "@layer empty") {
		t.Errorf("Build() should create layer for empty file, got: %s", got)
	}
}

// brokenFS returns an error when walking.
type brokenFS struct{}

func (brokenFS) Open(name string) (fs.File, error) {
	return nil, errors.New("simulated fs error")
}

func TestBuild_walk_error_propagation(t *testing.T) {
	t.Parallel()

	_, err := Build(Source{FS: brokenFS{}, Dir: "css"})
	if err == nil {
		t.Fatal("Build() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "walk") {
		t.Errorf("Build() error = %q, want error containing 'walk'", err.Error())
	}
}

func TestBuildWithHash(t *testing.T) {
	t.Parallel()

	// Test data structures
	hashBasicFS := fstest.MapFS{
		"css/reset.css": {Data: []byte("* { margin: 0; }")},
	}

	tests := []struct {
		name            string
		giveFS          fs.FS
		giveDir         string
		wantHashLen     int
		wantCSSNonEmpty bool
	}{
		{
			name:            "returns_hash",
			giveFS:          hashBasicFS,
			giveDir:         "css",
			wantHashLen:     16,
			wantCSSNonEmpty: true,
		},
		{
			name:            "empty_fs_empty_hash",
			giveFS:          fstest.MapFS{},
			giveDir:         "css",
			wantHashLen:     0,
			wantCSSNonEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			css, hash, err := BuildWithHash(Source{FS: tt.giveFS, Dir: tt.giveDir})
			if err != nil {
				t.Fatalf("BuildWithHash() error = %v, want nil", err)
			}

			if len(hash) != tt.wantHashLen {
				t.Errorf("BuildWithHash() hash len = %d, want %d", len(hash), tt.wantHashLen)
			}

			gotCSSNonEmpty := css != ""
			if gotCSSNonEmpty != tt.wantCSSNonEmpty {
				t.Errorf("BuildWithHash() CSS non-empty = %v, want %v", gotCSSNonEmpty, tt.wantCSSNonEmpty)
			}
		})
	}
}

func TestBuildWithHash_stability(t *testing.T) {
	t.Parallel()

	testFS := fstest.MapFS{
		"css/reset.css": {Data: []byte("* { margin: 0; }")},
	}

	css1, hash1, err := BuildWithHash(Source{FS: testFS, Dir: "css"})
	if err != nil {
		t.Fatalf("BuildWithHash() first call error = %v", err)
	}

	css2, hash2, err := BuildWithHash(Source{FS: testFS, Dir: "css"})
	if err != nil {
		t.Fatalf("BuildWithHash() second call error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("BuildWithHash() hashes differ: %q vs %q, want identical", hash1, hash2)
	}

	if css1 != css2 {
		t.Errorf("BuildWithHash() CSS differs, want identical")
	}
}

func TestBuildWithHash_uniqueness(t *testing.T) {
	t.Parallel()

	hashBasicFS := fstest.MapFS{
		"css/reset.css": {Data: []byte("* { margin: 0; }")},
	}
	hashDifferentFS := fstest.MapFS{
		"css/reset.css": {Data: []byte("* { margin: 1px; }")},
	}

	_, hash1, err := BuildWithHash(Source{FS: hashBasicFS, Dir: "css"})
	if err != nil {
		t.Fatalf("BuildWithHash() first call error = %v", err)
	}

	_, hash2, err := BuildWithHash(Source{FS: hashDifferentFS, Dir: "css"})
	if err != nil {
		t.Fatalf("BuildWithHash() second call error = %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("BuildWithHash() hashes should differ for different content, got same: %q", hash1)
	}
}

func TestBuildWithHash_hex_format(t *testing.T) {
	t.Parallel()

	testFS := fstest.MapFS{
		"css/reset.css": {Data: []byte("* { margin: 0; }")},
	}

	_, hash, err := BuildWithHash(Source{FS: testFS, Dir: "css"})
	if err != nil {
		t.Fatalf("BuildWithHash() error = %v", err)
	}

	hexPattern := regexp.MustCompile(`^[0-9a-f]{16}$`)
	if !hexPattern.MatchString(hash) {
		t.Errorf("BuildWithHash() hash = %q, want lowercase hex matching ^[0-9a-f]{16}$", hash)
	}
}

func TestBuildWithHash_matches_build(t *testing.T) {
	t.Parallel()

	testFS := fstest.MapFS{
		"css/reset.css":     {Data: []byte("* { margin: 0; }")},
		"css/base/file.css": {Data: []byte("h1 {}")},
	}

	buildCSS, err := Build(Source{FS: testFS, Dir: "css"})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	hashCSS, _, err := BuildWithHash(Source{FS: testFS, Dir: "css"})
	if err != nil {
		t.Fatalf("BuildWithHash() error = %v", err)
	}

	if buildCSS != hashCSS {
		t.Errorf("BuildWithHash() CSS differs from Build() output")
	}
}

func TestBuildWithHash_error_propagation(t *testing.T) {
	t.Parallel()

	_, _, err := BuildWithHash(Source{FS: brokenFS{}, Dir: "css"})
	if err == nil {
		t.Fatal("BuildWithHash() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "walk") {
		t.Errorf("BuildWithHash() error = %q, want error containing 'walk'", err.Error())
	}
}

func TestBuild_multi_directory(t *testing.T) {
	t.Parallel()

	// Simulate three separate directories with different purposes
	stylesFS := fstest.MapFS{
		"styles/reset.css":  {Data: []byte("/* reset */")},
		"styles/tokens.css": {Data: []byte("/* tokens */")},
	}

	componentsFS := fstest.MapFS{
		"components/button/button.css": {Data: []byte("/* button */")},
		"components/card/card.css":     {Data: []byte("/* card */")},
	}

	routesFS := fstest.MapFS{
		"routes/auth/login.css": {Data: []byte("/* login */")},
		"routes/home.css":       {Data: []byte("/* home */")},
	}

	got, err := Build(
		Source{FS: stylesFS, Dir: "styles"},
		Source{FS: componentsFS, Dir: "components"},
		Source{FS: routesFS, Dir: "routes"},
	)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	// Verify layer declaration order: styles first, then components, then routes
	wantLayerDecl := "@layer reset, tokens, button, card, auth, home;"
	if !strings.HasPrefix(got, wantLayerDecl) {
		t.Errorf("Build() layer declaration = %q, want %q",
			strings.SplitN(got, "\n", 2)[0], wantLayerDecl)
	}

	// Verify content order: reset should come before button, button before login
	resetIdx := strings.Index(got, "/* reset */")
	buttonIdx := strings.Index(got, "/* button */")
	loginIdx := strings.Index(got, "/* login */")

	if resetIdx == -1 || buttonIdx == -1 || loginIdx == -1 {
		t.Fatalf("Build() missing expected content")
	}

	if resetIdx > buttonIdx {
		t.Errorf("Build() reset content should come before button content")
	}
	if buttonIdx > loginIdx {
		t.Errorf("Build() button content should come before login content")
	}
}

func TestBuild_multi_directory_with_nesting(t *testing.T) {
	t.Parallel()

	// First source has nested layers
	source1FS := fstest.MapFS{
		"styles/reset.css":             {Data: []byte("a")},
		"styles/base/elements/btn.css": {Data: []byte("b")},
	}

	// Second source has only root layers
	source2FS := fstest.MapFS{
		"components/button.css": {Data: []byte("c")},
		"components/card.css":   {Data: []byte("d")},
	}

	got, err := Build(
		Source{FS: source1FS, Dir: "styles"},
		Source{FS: source2FS, Dir: "components"},
	)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	// Within source1: reset (depth 0) comes before base.elements (depth 2)
	// Then all of source2: button, card
	wantLayerDecl := "@layer reset, base.elements, button, card;"
	if !strings.HasPrefix(got, wantLayerDecl) {
		t.Errorf("Build() layer declaration = %q, want %q",
			strings.SplitN(got, "\n", 2)[0], wantLayerDecl)
	}
}

func TestBuild_with_prefix(t *testing.T) {
	t.Parallel()

	testFS := fstest.MapFS{
		"components/button/button.css": {Data: []byte("/* button */")},
		"components/card.css":          {Data: []byte("/* card */")},
	}

	got, err := Build(Source{
		FS:     testFS,
		Dir:    "components",
		Prefix: "comp",
	})
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	// Layers should be prefixed with "comp."
	wantLayerDecl := "@layer comp.button, comp.card;"
	if !strings.HasPrefix(got, wantLayerDecl) {
		t.Errorf("Build() layer declaration = %q, want %q",
			strings.SplitN(got, "\n", 2)[0], wantLayerDecl)
	}

	// Layer blocks should also use prefixed names
	if !strings.Contains(got, "@layer comp.button {") {
		t.Errorf("Build() should contain '@layer comp.button {', got: %s", got)
	}
	if !strings.Contains(got, "@layer comp.card {") {
		t.Errorf("Build() should contain '@layer comp.card {', got: %s", got)
	}
}

func TestBuild_multi_directory_with_prefixes(t *testing.T) {
	t.Parallel()

	componentsFS := fstest.MapFS{
		"components/button.css": {Data: []byte("/* button */")},
	}

	routesFS := fstest.MapFS{
		"routes/home.css": {Data: []byte("/* home */")},
	}

	got, err := Build(
		Source{FS: componentsFS, Dir: "components", Prefix: "comp"},
		Source{FS: routesFS, Dir: "routes", Prefix: "page"},
	)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	wantLayerDecl := "@layer comp.button, page.home;"
	if !strings.HasPrefix(got, wantLayerDecl) {
		t.Errorf("Build() layer declaration = %q, want %q",
			strings.SplitN(got, "\n", 2)[0], wantLayerDecl)
	}
}

func TestBuild_empty_prefix_ignored(t *testing.T) {
	t.Parallel()

	testFS := fstest.MapFS{
		"css/reset.css": {Data: []byte("x")},
	}

	got, err := Build(Source{
		FS:     testFS,
		Dir:    "css",
		Prefix: "", // Empty prefix should be ignored
	})
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	// Should be "reset", not ".reset"
	wantLayerDecl := "@layer reset;"
	if !strings.HasPrefix(got, wantLayerDecl) {
		t.Errorf("Build() layer declaration = %q, want %q",
			strings.SplitN(got, "\n", 2)[0], wantLayerDecl)
	}
}

func TestBuild_mixed_prefix_and_no_prefix(t *testing.T) {
	t.Parallel()

	stylesFS := fstest.MapFS{
		"styles/reset.css": {Data: []byte("/* reset */")},
	}

	componentsFS := fstest.MapFS{
		"components/button.css": {Data: []byte("/* button */")},
	}

	got, err := Build(
		Source{FS: stylesFS, Dir: "styles"}, // No prefix
		Source{FS: componentsFS, Dir: "components", Prefix: "comp"},
	)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	wantLayerDecl := "@layer reset, comp.button;"
	if !strings.HasPrefix(got, wantLayerDecl) {
		t.Errorf("Build() layer declaration = %q, want %q",
			strings.SplitN(got, "\n", 2)[0], wantLayerDecl)
	}
}

func TestBuild_prefix_with_nested_layers(t *testing.T) {
	t.Parallel()

	testFS := fstest.MapFS{
		"components/base/button.css":     {Data: []byte("/* button */")},
		"components/base/card/card.css":  {Data: []byte("/* card */")},
		"components/other/dropdown.css":  {Data: []byte("/* dropdown */")},
	}

	got, err := Build(Source{
		FS:     testFS,
		Dir:    "components",
		Prefix: "comp",
	})
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	// Prefix should prepend to the full layer name
	// base (depth 0), other (depth 0), base.card (depth 1)
	wantLayerDecl := "@layer comp.base, comp.other, comp.base.card;"
	if !strings.HasPrefix(got, wantLayerDecl) {
		t.Errorf("Build() layer declaration = %q, want %q",
			strings.SplitN(got, "\n", 2)[0], wantLayerDecl)
	}
}
