// Package strata provides CSS layer ordering from directory structure.
package strata

import (
	"path"
	"strings"
)

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
