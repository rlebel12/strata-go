package strata

import "testing"

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
