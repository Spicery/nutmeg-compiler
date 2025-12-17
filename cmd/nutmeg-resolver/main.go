package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	pflag "github.com/spf13/pflag"

	"github.com/spicery/nutmeg-compiler/pkg/common"
	"github.com/spicery/nutmeg-compiler/pkg/resolver"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const usage = `nutmeg-resolver - identifier resolution for the Nutmeg programming language

This tool reads a Nutmeg AST (unit node) in JSON format and annotates
identifiers with resolution information:
  - Unique identifier IDs
  - Scope information (global, outer, inner)
  - Whether the identifier is a definition or use

Usage:
  nutmeg-resolver [options]

Options:
`

const DEFAULT_FORMAT = "JSON"

func main() {
	var showHelp, showVersion, noSpans bool
	var inputFile, outputFile, format string
	var trim int

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		pflag.PrintDefaults()
	}

	pflag.BoolVarP(&showHelp, "help", "h", false, "Show help")
	pflag.BoolVar(&showVersion, "version", false, "Show version")
	pflag.StringVarP(&inputFile, "input", "i", "", "Input file (defaults to stdin)")
	pflag.StringVarP(&outputFile, "output", "o", "", "Output file (defaults to stdout)")
	pflag.StringVarP(&format, "format", "f", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")
	pflag.IntVar(&trim, "trim", 0, "Trim names for display purposes")
	pflag.BoolVar(&noSpans, "no-spans", false, "Suppress span information in output")

	pflag.Parse()

	if showHelp {
		pflag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("nutmeg-resolver version %s\n", Version)
		os.Exit(0)
	}

	// Reject any positional arguments.
	if len(pflag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Error: Unexpected positional arguments. Use --input and --output flags instead.\n\n")
		pflag.Usage()
		os.Exit(1)
	}

	// Determine input source.
	var input io.Reader = os.Stdin
	if inputFile != "" {
		file, err := os.Open(inputFile) // #nosec G304 - CLI tool reads user-specified input files
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
	}

	// Read input JSON.
	inputBytes, err := io.ReadAll(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON into Node structure.
	var tree common.Node
	if err := json.Unmarshal(inputBytes, &tree); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Perform resolution.
	r := resolver.NewResolver()
	if err := r.Resolve(&tree); err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving identifiers: %v\n", err)
		os.Exit(1)
	}

	// Determine output format.
	printFunc := common.PickPrintFunc(format)

	// Determine output destination.
	var output io.Writer = os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile) // #nosec G304 - CLI tool writes to user-specified output files
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		output = file
	}

	// Output the result.
	printFunc(&tree, "  ", output, &common.PrintOptions{
		TrimTokenOnOutput: trim,
		IncludeSpans:      !noSpans,
	})
}
