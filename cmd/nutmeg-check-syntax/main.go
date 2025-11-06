package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/checker"
	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const usage = `nutmeg-check-syntax - syntax validation for the Nutmeg programming language

This tool reads a Nutmeg AST (unit node) in JSON format and validates
that nodes conform to additional syntactic rules not enforced by the parser.
If validation fails, it exits with a non-zero status; otherwise it emits
the tree unchanged to stdout.

Usage:
  nutmeg-check-syntax [options]

Options:
`

const DEFAULT_FORMAT = "JSON"

func main() {
	var showHelp, showVersion, noSpans bool
	var inputFile, outputFile, format string
	var trim int

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		flag.PrintDefaults()
	}

	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.StringVar(&inputFile, "input", "", "Input file (defaults to stdin)")
	flag.StringVar(&outputFile, "output", "", "Output file (defaults to stdout)")
	flag.StringVar(&format, "f", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")
	flag.StringVar(&format, "format", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")
	flag.IntVar(&trim, "trim", 0, "Trim names for display purposes")
	flag.BoolVar(&noSpans, "no-spans", false, "Suppress span information in output")

	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("nutmeg-check-syntax version %s\n", Version)
		os.Exit(0)
	}

	// Reject any positional arguments.
	if len(flag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Error: Unexpected positional arguments. Use --input and --output flags instead.\n\n")
		flag.Usage()
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

	// Perform syntax checking.
	c := checker.NewChecker()
	if !c.Check(&tree) {
		c.ReportErrors()
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
