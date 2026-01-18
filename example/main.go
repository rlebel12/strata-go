// Example demonstrates strata-go usage with a real directory structure.
//
// Run from the example directory:
//
//	go run main.go
package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"

	strata "github.com/rlebel12/strata-go"
)

func main() {
	// Create a sub-filesystem rooted at the css/ directory
	cssFS, err := fs.Sub(os.DirFS("."), "css")
	if err != nil {
		log.Fatal(err)
	}

	// Build CSS from the filesystem
	css, hash, err := strata.BuildWithHash(strata.Source{FS: cssFS})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated CSS (hash: %s):\n\n", hash)
	fmt.Println(css)

	// Example: write to a hashed filename for cache busting
	filename := fmt.Sprintf("styles.%s.css", hash)
	fmt.Printf("Cache-busted filename: %s\n", filename)
}
