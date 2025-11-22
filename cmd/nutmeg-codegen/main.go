package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/codegen"
	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const DEFAULT_FORMAT = "JSON"

func main() {
	// Define command line flags according to specifications.
	var format = flag.String("f", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")
	var formatLong = flag.String("format", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")
	var srcPath = flag.String("src-path", "", "Source path to annotate the unit with origin")
	var trim = flag.Int("trim", 0, "Trim names for display purposes")
	var noSpans = flag.Bool("no-spans", false, "Suppress span information in output")
	var version = flag.Bool("version", false, "Print version and exit")

	flag.Parse()

	// Handle version flag.
	if *version {
		fmt.Printf("nutmeg-codegen version %s\n", Version)
		os.Exit(0)
	}

	// Use the long form if provided, otherwise use the short form.
	selectedFormat := *format
	if *formatLong != DEFAULT_FORMAT {
		selectedFormat = *formatLong
	}

	// Read input JSON node tree from stdin.
	decoder := json.NewDecoder(os.Stdin)
	var root common.Node
	if err := decoder.Decode(&root); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding input JSON: %v\n", err)
		os.Exit(1)
	}

	// Add source path if provided.
	if srcPath != nil && *srcPath != "" {
		if root.Options == nil {
			root.Options = make(map[string]string)
		}
		root.Options["src"] = *srcPath
	}

	// Create code generator and process the tree.
	cg := codegen.NewCodeGenerator()
	err := cg.Generate(&root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during code generation: %v\n", err)
		os.Exit(1)
	}

	// Select the appropriate print function based on format.
	printFunc := common.PickPrintFunc(selectedFormat)
	printFunc(&root, "  ", os.Stdout, &common.PrintOptions{
		TrimTokenOnOutput: *trim,
		IncludeSpans:      !*noSpans,
	})
}
