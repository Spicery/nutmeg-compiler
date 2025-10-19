package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const DEFAULT_FORMAT = "XML"

func main() {
	// Define command line flags.
	var format = flag.String("f", DEFAULT_FORMAT, "Output format (JSON, XML, YAML, MERMAID, ASCIITREE, DOT)")
	var formatLong = flag.String("format", DEFAULT_FORMAT, "Output format (JSON, XML, YAML, MERMAID, ASCIITREE, DOT)")
	var indent = flag.Int("indent", 2, "Indentation level for display purposes")
	var trim = flag.Int("trim", 0, "Trim names for display purposes")
	var noSpans = flag.Bool("no-spans", false, "Suppress span information in output")
	var version = flag.Bool("version", false, "Print version and exit")
	var help = flag.Bool("help", false, "Print help message and exit")

	// Custom usage message.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nConverts a parse tree from JSON format to various output formats.\n")
		fmt.Fprintf(os.Stderr, "Reads JSON from stdin and writes the converted tree to stdout.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Handle version flag.
	if *version {
		fmt.Printf("nutmeg-convert-tree version %s\n", Version)
		os.Exit(0)
	}

	// Handle help flag.
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Use the long form if provided, otherwise use the short form.
	selectedFormat := *format
	if *formatLong != DEFAULT_FORMAT {
		selectedFormat = *formatLong
	}

	// Read JSON from stdin.
	tree, err := common.ReadASTJSON(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON input: %v\n", err)
		os.Exit(1)
	}

	// Select the appropriate print function based on format.
	printFunc := common.PickPrintFunc(selectedFormat)

	// Create indent string based on indent level.
	indentStr := ""
	for i := 0; i < *indent; i++ {
		indentStr += " "
	}

	// Print the tree in the selected format.
	printFunc(tree, indentStr, os.Stdout, &common.PrintOptions{
		Format:            selectedFormat,
		Indent:            *indent,
		TrimTokenOnOutput: *trim,
		IncludeSpans:      !*noSpans,
	})
}
