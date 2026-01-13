package strata_test

import (
	"fmt"
	"testing/fstest"

	"github.com/rlebel12/strata-go"
)

func ExampleBuild() {
	// Create an in-memory filesystem with CSS files
	fsys := fstest.MapFS{
		"css/reset.css":           {Data: []byte("* { margin: 0; }")},
		"css/base/typography.css": {Data: []byte("h1 { font-size: 2rem; }")},
	}

	output, err := strata.Build(fsys, "css")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(output)
	// Output:
	// @layer base, reset;
	// @layer base {
	// h1 { font-size: 2rem; }
	// }
	// @layer reset {
	// * { margin: 0; }
	// }
}
