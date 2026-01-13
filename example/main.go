// Example demonstrates strata-go usage with a real directory structure.
//
// Run from the example directory:
//
//	go run main.go
package main

import (
	"fmt"
	"log"
	"os"

	strata "github.com/rlebel12/strata-go"
)

func main() {
	// Build CSS from the css/ directory
	css, hash, err := strata.BuildWithHash(os.DirFS("."), "css")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated CSS (hash: %s):\n\n", hash)
	fmt.Println(css)

	// Example: write to a hashed filename for cache busting
	filename := fmt.Sprintf("styles.%s.css", hash)
	fmt.Printf("Cache-busted filename: %s\n", filename)
}
