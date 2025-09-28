package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	version = "0.1.0"
	usage   = `nutmeg-rewrite - a tree rewriter for the Nutmeg programming language`
)

func main() {
	var showHelp, showVersion bool
	var inputFile, outputFile string

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

}
