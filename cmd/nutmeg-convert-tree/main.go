package main

import (
	"fmt"
	"os"

	pflag "github.com/spf13/pflag"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const DEFAULT_FORMAT = "XML"

func main() {
	// Define command line flags.
	var format = pflag.StringP("format", "f", DEFAULT_FORMAT, "Output format (JSON, XML, YAML, MERMAID, ASCIITREE, DOT)")
	var indent = pflag.Int("indent", 2, "Indentation level for display purposes")
	var trim = pflag.Int("trim", 0, "Trim names for display purposes")
	var noSpans = pflag.Bool("no-spans", false, "Suppress span information in output")
	var version = pflag.Bool("version", false, "Print version and exit")
	var help = pflag.BoolP("help", "h", false, "Print help message and exit")

	// Custom usage message.
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nConverts a parse tree from JSON format to various output formats.\n")
		fmt.Fprintf(os.Stderr, "Reads JSON from stdin and writes the converted tree to stdout.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		pflag.PrintDefaults()
	}

	pflag.Parse()

	// Handle version flag.
	if *version {
		fmt.Printf("nutmeg-convert-tree version %s\n", Version)
		os.Exit(0)
	}

	// Handle help flag.
	if *help {
		pflag.Usage()
		os.Exit(0)
	}

	// Read JSON from stdin.
	tree, err := common.ReadASTJSON(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON input: %v\n", err)
		os.Exit(1)
	}

	// Select the appropriate print function based on format.
	printFunc := common.PickPrintFunc(*format)

	// Create indent string based on indent level.
	indentStr := ""
	for i := 0; i < *indent; i++ {
		indentStr += " "
	}

	// Print the tree in the selected format.
	printFunc(tree, indentStr, os.Stdout, &common.PrintOptions{
		Format:            *format,
		Indent:            *indent,
		TrimTokenOnOutput: *trim,
		IncludeSpans:      !*noSpans,
	})
}
