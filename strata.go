// Package strata provides CSS layer ordering from directory structure.
package strata

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

const cssExtension = ".css"

// pathToLayerName converts a file path to its CSS layer name.
//
// The layer name is derived from the directory structure relative to the
// root directory. Root files use the filename (without extension) as the
// layer name. Nested paths use dots as separators.
//
// Examples:
//   - pathToLayerName("css/reset.css", "css") -> "reset"
//   - pathToLayerName("css/base/file.css", "css") -> "base"
//   - pathToLayerName("css/base/elements/btn.css", "css") -> "base.elements"
func pathToLayerName(filePath, dir string) string {
	// Normalize dir to ensure no trailing slash
	dir = strings.TrimSuffix(dir, "/")

	// Strip dir prefix from path
	relPath := strings.TrimPrefix(filePath, dir+"/")

	// Get directory portion of relative path
	dirPart := path.Dir(relPath)

	// Root file: use filename without extension
	if dirPart == "." {
		return strings.TrimSuffix(path.Base(relPath), path.Ext(relPath))
	}

	// Nested path: replace "/" with "." to form layer name
	return strings.ReplaceAll(dirPart, "/", ".")
}

// layer represents a CSS cascade layer being built.
type layer struct {
	name    string
	depth   int
	content *bytes.Buffer
}

// Build walks the filesystem and returns CSS with @layer declarations.
//
// The directory structure determines layer hierarchy:
//   - Root files (e.g., css/reset.css) become individual layers
//   - Nested directories use dot notation (e.g., css/base/elements/ -> base.elements)
//
// Output format:
//
//	@layer name1, name2, name3;
//	@layer name1 { ... content ... }
//	@layer name2 { ... content ... }
//
// Files within the same layer are concatenated in alphabetical order.
// Layers are ordered by depth first (shallow before deep), then alphabetically.
// Empty filesystems return an empty string (not an error).
func Build(fsys fs.FS, dir string) (string, error) {
	layers := make(map[string]*layer)
	var filePaths []string

	// Collect all CSS file paths
	err := fs.WalkDir(fsys, ".", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(filePath, cssExtension) {
			return nil
		}
		filePaths = append(filePaths, filePath)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walk filesystem: %w", err)
	}

	// Empty filesystem
	if len(filePaths) == 0 {
		return "", nil
	}

	// Sort file paths for deterministic concatenation order
	sort.Strings(filePaths)

	// Process each CSS file
	for _, filePath := range filePaths {
		content, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", filePath, err)
		}

		layerName := pathToLayerName(filePath, dir)

		l, exists := layers[layerName]
		if !exists {
			l = &layer{
				name:    layerName,
				depth:   strings.Count(layerName, "."),
				content: &bytes.Buffer{},
			}
			layers[layerName] = l
		}

		l.content.Write(content)
		l.content.WriteByte('\n')
	}

	// Convert map to slice and sort by depth then name
	sortedLayers := make([]*layer, 0, len(layers))
	for _, l := range layers {
		sortedLayers = append(sortedLayers, l)
	}
	sort.Slice(sortedLayers, func(i, j int) bool {
		if sortedLayers[i].depth != sortedLayers[j].depth {
			return sortedLayers[i].depth < sortedLayers[j].depth
		}
		return sortedLayers[i].name < sortedLayers[j].name
	})

	// Build output
	var out bytes.Buffer

	// Write layer declaration header
	out.WriteString("@layer ")
	for i, l := range sortedLayers {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(l.name)
	}
	out.WriteString(";\n")

	// Write each layer block
	for _, l := range sortedLayers {
		out.WriteString("@layer ")
		out.WriteString(l.name)
		out.WriteString(" {\n")
		out.Write(l.content.Bytes())
		out.WriteString("}\n")
	}

	return out.String(), nil
}
