package main

import (
	"encoding/json"
	"fmt"
	"os"

	pflag "github.com/spf13/pflag"

	"github.com/spicery/nutmeg-compiler/pkg/codegen"
	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const DEFAULT_FORMAT = "JSON"

func main() {
	// Define command line flags according to specifications.
	var format = pflag.StringP("format", "f", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")
	var srcPath = pflag.String("src-path", "", "Source path to annotate the unit with origin")
	var trim = pflag.Int("trim", 0, "Trim names for display purposes")
	var noSpans = pflag.Bool("no-spans", false, "Suppress span information in output")
	var version = pflag.Bool("version", false, "Print version and exit")

	pflag.Parse()

	// Handle version flag.
	if *version {
		fmt.Printf("nutmeg-codegen version %s\n", Version)
		os.Exit(0)
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
	printFunc := common.PickPrintFunc(*format)
	printFunc(&root, "  ", os.Stdout, &common.PrintOptions{
		TrimTokenOnOutput: *trim,
		IncludeSpans:      !*noSpans,
	})
}
