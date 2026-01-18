// Package strata provides CSS layer ordering from directory structure.
package strata

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

const cssExtension = ".css"

// Source represents a CSS source directory to build from.
type Source struct {
	// FS is the filesystem to read from
	FS fs.FS

	// Dir is the directory path to walk within the filesystem
	Dir string

	// Prefix is an optional namespace to prepend to all layer names.
	// If set, layer names will be "prefix.layername" instead of "layername".
	Prefix string
}

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

// Build walks one or more source directories and returns CSS with @layer declarations.
//
// Sources are processed in slice order. Within each source, the directory structure
// determines layer hierarchy:
//   - Root files (e.g., css/reset.css) become individual layers
//   - Nested directories use dot notation (e.g., css/base/elements/ -> base.elements)
//   - Optional Prefix prepends a namespace (e.g., Prefix: "comp" -> comp.button)
//
// Output format:
//
//	@layer name1, name2, name3;
//	@layer name1 { ... content ... }
//	@layer name2 { ... content ... }
//
// Files within the same layer are concatenated in alphabetical order.
// Within each source, layers are ordered depth-first (shallow before deep), then alphabetically.
// Empty sources return an empty string (not an error).
func Build(sources ...Source) (string, error) {
	var allLayers []*layer

	// Process each source in order
	for _, src := range sources {
		layers := make(map[string]*layer)
		var filePaths []string

		// Collect all CSS file paths from this source
		err := fs.WalkDir(src.FS, ".", func(filePath string, d fs.DirEntry, err error) error {
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

		// Skip empty sources
		if len(filePaths) == 0 {
			continue
		}

		// Sort file paths for deterministic concatenation order
		sort.Strings(filePaths)

		// Process each CSS file
		for _, filePath := range filePaths {
			content, err := fs.ReadFile(src.FS, filePath)
			if err != nil {
				return "", fmt.Errorf("read %s: %w", filePath, err)
			}

			layerName := pathToLayerName(filePath, src.Dir)

			// Apply prefix if specified
			if src.Prefix != "" {
				layerName = src.Prefix + "." + layerName
			}

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

		// Append this source's layers to the final list
		allLayers = append(allLayers, sortedLayers...)
	}

	// Handle empty result
	if len(allLayers) == 0 {
		return "", nil
	}

	// Build output
	var out bytes.Buffer

	// Write layer declaration header
	out.WriteString("@layer ")
	for i, l := range allLayers {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(l.name)
	}
	out.WriteString(";\n")

	// Write each layer block
	for _, l := range allLayers {
		out.WriteString("@layer ")
		out.WriteString(l.name)
		out.WriteString(" {\n")
		out.Write(l.content.Bytes())
		out.WriteString("}\n")
	}

	return out.String(), nil
}

// BuildWithHash returns the built CSS and a content hash for cache busting.
//
// The hash is computed from the CSS output using SHA-256, truncated to 16
// lowercase hexadecimal characters (8 bytes). Empty CSS returns an empty hash.
//
// Example usage:
//
//	css, hash, err := strata.BuildWithHash(strata.Source{
//	    FS:  os.DirFS("."),
//	    Dir: "css",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Use hash in filename: styles.{hash}.css
//	fmt.Printf("<link rel=\"stylesheet\" href=\"/static/styles.%s.css\">\n", hash)
func BuildWithHash(sources ...Source) (css string, hash string, err error) {
	css, err = Build(sources...)
	if err != nil {
		return "", "", err
	}

	if css == "" {
		return "", "", nil
	}

	sum := sha256.Sum256([]byte(css))
	hash = hex.EncodeToString(sum[:8])

	return css, hash, nil
}
