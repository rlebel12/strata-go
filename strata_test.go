package strata

import (
	"errors"
	"io/fs"
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

			got, err := Build(tt.giveFS, tt.giveDir)
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

	got, err := Build(testFS, "css")
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

	got, err := Build(testFS, "css")
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

	_, err := Build(brokenFS{}, "css")
	if err == nil {
		t.Fatal("Build() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "walk") {
		t.Errorf("Build() error = %q, want error containing 'walk'", err.Error())
	}
}
