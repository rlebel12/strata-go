package main_test

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/rlebel12/strata-go"
)

func ExampleBuild() {
	// Create a sub-filesystem rooted at the css/ directory
	cssFS, err := fs.Sub(os.DirFS("."), "css")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	// Build CSS from the filesystem
	output, err := strata.Build(strata.Source{FS: cssFS})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(output)
	// Output:
	// @layer base, components, reset, tokens, base.elements;
	// @layer base {
	// body {
	// 	font-family: system-ui, sans-serif;
	// 	line-height: 1.5;
	// }
	//
	// h1 { font-size: 2rem; }
	// h2 { font-size: 1.5rem; }
	//
	// }
	// @layer components {
	// .card {
	// 	border: 1px solid var(--color-border);
	// 	border-radius: var(--radius-md);
	// 	padding: var(--spacing-md);
	// 	box-shadow: 0 2px 4px var(--color-shadow);
	//
	// 	.card-header {
	// 		font-weight: bold;
	// 		margin-bottom: var(--spacing-sm);
	// 	}
	//
	// 	.card-body {
	// 		line-height: 1.5;
	// 	}
	// }
	//
	// }
	// @layer reset {
	// *,
	// *::before,
	// *::after {
	// 	box-sizing: border-box;
	// 	margin: 0;
	// 	padding: 0;
	// }
	//
	// }
	// @layer tokens {
	// :root {
	// 	--color-primary: #3b82f6;
	// 	--color-primary-hover: #2563eb;
	// 	--color-border: #e5e7eb;
	// 	--color-shadow: rgba(0, 0, 0, 0.1);
	//
	// 	--radius-sm: 4px;
	// 	--radius-md: 8px;
	//
	// 	--spacing-sm: 0.5rem;
	// 	--spacing-md: 1rem;
	// }
	//
	// }
	// @layer base.elements {
	// button {
	// 	cursor: pointer;
	// 	border: 1px solid var(--color-border);
	// 	border-radius: var(--radius-sm);
	// 	padding: var(--spacing-sm) var(--spacing-md);
	// 	background: var(--color-primary);
	// 	color: white;
	//
	// 	&:hover {
	// 		background: var(--color-primary-hover);
	// 	}
	// }
	//
	// }
}
