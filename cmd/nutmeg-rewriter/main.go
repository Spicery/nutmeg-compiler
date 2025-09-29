package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/common"
	"github.com/spicery/nutmeg-compiler/pkg/rewriter"
)

const (
	version = "0.1.0"
	usage   = `nutmeg-rewrite - a tree rewriter for the Nutmeg programming language`
)

func main() {
	var showHelp, showVersion bool
	var inputFile, outputFile, configFile string

	// Set up custom usage function that includes the description and flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\nUsage:\n", usage)
		flag.PrintDefaults()
	}

	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.StringVar(&inputFile, "input", "", "Input file (defaults to stdin)")
	flag.StringVar(&outputFile, "output", "", "Output file (defaults to stdout)")
	flag.StringVar(&configFile, "rewrite", "", "YAML file containing rewrite rules")

	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("nutmeg-rewrite version %s\n", version)
		os.Exit(0)
	}

	// Reject any positional arguments
	if len(flag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Error: Unexpected positional arguments. Use --input and --output flags instead.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	var rewriteConfig *rewriter.RewriteConfig
	var err error
	if configFile != "" {
		rewriteConfig, err = rewriter.LoadRewriteConfig(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading rewrite configuration file: %v\n", err)
			os.Exit(1)
		}
	}
	var r *rewriter.Rewriter
	if rewriteConfig != nil {
		r = rewriter.NewRewriter(rewriteConfig)
	}

	// Determine input source
	var input io.Reader = os.Stdin
	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
	}

	// Determine output destination
	var output io.Writer = os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		output = file
	}

	// Read JSON from input and decode into a Node
	var node *common.Node
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(&node); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding JSON: %v\n", err)
		os.Exit(1)
	}

	if r != nil {
		node = r.Rewrite(node)
	}

	// For now, just output the node back as JSON (placeholder for rewriting logic)
	encoder := json.NewEncoder(output)
	if err := encoder.Encode(&node); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
