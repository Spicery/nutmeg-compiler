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

// Version is injected at build time via ldflags.
var Version = "dev"

const usage = `nutmeg-rewrite - a tree rewriter for the Nutmeg programming language`

const DEFAULT_FORMAT = "JSON"

func main() {
	var showHelp, showVersion, noSpans, makeRules, debug, skipOptional bool
	var inputFile, outputFile, configFile, format string
	var trim, maxRewrites int

	// Set up custom usage function that includes the description and flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\nUsage:\n", usage)
		flag.PrintDefaults()
	}

	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&makeRules, "make-rewrite-rules", false, "Generate default rewrite rules YAML")
	flag.BoolVar(&debug, "debug", false, "Enable debug output to stderr")
	flag.BoolVar(&skipOptional, "skip-optional", false, "Skip optional rewrite passes")
	flag.StringVar(&inputFile, "input", "", "Input file (defaults to stdin)")
	flag.StringVar(&outputFile, "output", "", "Output file (defaults to stdout)")
	flag.StringVar(&configFile, "rewrite-rules", "", "YAML file containing rewrite rules")
	flag.StringVar(&format, "f", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")
	flag.StringVar(&format, "format", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")
	flag.IntVar(&trim, "trim", 0, "Trim names for display purposes")
	flag.BoolVar(&noSpans, "no-spans", false, "Suppress span information in output")
	flag.IntVar(&maxRewrites, "max-rewrites", 0, "Maximum number of rewrite iterations (0 = unlimited)")

	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("nutmeg-rewrite version %s\n", Version)
		os.Exit(0)
	}

	if makeRules {
		fmt.Print(rewriter.DefaultRewriteRules)
		os.Exit(0)
	}

	// Reject any positional arguments
	if len(flag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Error: Unexpected positional arguments. Use --input and --output flags instead.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	printFunc := common.PickPrintFunc(format)

	var rewriteConfig *rewriter.RewriteConfig
	var err error
	if configFile != "" {
		rewriteConfig, err = rewriter.LoadRewriteConfig(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading rewrite configuration file: %v\n", err)
			os.Exit(1)
		}
	} else {
		rewriteConfig, err = rewriter.LoadRewriteConfigFromString(rewriter.DefaultRewriteRules)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading default rewrite rules: %v\n", err)
			os.Exit(1)
		}
	}
	var r *rewriter.Rewriter
	if rewriteConfig != nil {
		r, err = rewriter.NewRewriterWithOptions(rewriteConfig, debug, skipOptional)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating rewriter: %v\n", err)
			os.Exit(1)
		}
	}

	// Determine input source
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

	// Determine output destination
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

	// Read JSON from input and decode into a Node
	var node *common.Node
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(&node); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding JSON: %v\n", err)
		os.Exit(1)
	}

	if r != nil {
		// Repeat rewriting until no changes occur (fixed point).
		iteration := 0
		reachedFixedPoint := false

		for {
			iteration++

			// Check if we've hit the iteration limit.
			if maxRewrites > 0 && iteration > maxRewrites {
				fmt.Fprintf(os.Stderr, "=== Stopped: reached maximum iterations (%d) ===\n", maxRewrites)
				break
			}

			fmt.Fprintf(os.Stderr, "=== Rewrite iteration %d ===\n", iteration)

			var changed bool
			node, changed = r.Rewrite(node)

			if changed {
				fmt.Fprintln(os.Stderr, "Rewrite modified the tree - continuing")
			} else {
				fmt.Fprintln(os.Stderr, "Rewrite made no changes - fixed point reached")
				reachedFixedPoint = true
				break
			}
		}

		if reachedFixedPoint {
			fmt.Fprintf(os.Stderr, "=== Completed after %d iteration(s) ===\n", iteration)
		} else {
			fmt.Fprintf(os.Stderr, "=== Warning: Did not reach fixed point after %d iteration(s) ===\n", iteration-1)
		}
	}

	printFunc(node, "  ", output, &common.PrintOptions{
		TrimTokenOnOutput: trim,
		IncludeSpans:      !noSpans,
	})
}
