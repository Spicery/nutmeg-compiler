package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/bundler"
	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const usage = `nutmeg-bundler - creates a SQLITE bundle for the Nutmeg runtime`

func main() {
	var showHelp, showVersion, migrate bool
	var bundleFile, inputFile, srcPath string
	var trim int

	// Set up custom usage function.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\nUsage:\n", usage)
		flag.PrintDefaults()
	}

	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&migrate, "migrate", false, "Perform database migration")
	flag.StringVar(&bundleFile, "bundle", "", "Bundle file path (required)")
	flag.StringVar(&inputFile, "input", "", "Input file (defaults to stdin)")
	flag.StringVar(&srcPath, "src-path", "", "Source path to annotate the unit with origin")
	flag.IntVar(&trim, "trim", 0, "Trim names for display purposes (not used)")

	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("nutmeg-bundler version %s\n", Version)
		os.Exit(0)
	}

	// Bundle file is mandatory.
	if bundleFile == "" {
		fmt.Fprintf(os.Stderr, "Error: --bundle flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check if the bundle file exists.
	_, err := os.Stat(bundleFile)
	fileExists := err == nil

	// Create bundler.
	b, err := bundler.NewBundler(bundleFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create bundler: %v\n", err)
		os.Exit(1)
	}
	defer b.Close()

	// Check if migration is needed.
	upToDate, err := b.CheckMigration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to check migration status: %v\n", err)
		os.Exit(1)
	}

	if !upToDate {
		// If the file didn't exist before, auto-migrate.
		// If it existed but schema is out of date, require --migrate flag.
		if !fileExists {
			// Fresh database - auto-migrate.
			if err := b.Migrate(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to migrate database: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Database initialized successfully.\n")
		} else {
			// Existing database needs migration.
			if !migrate {
				fmt.Fprintf(os.Stderr, "Error: database schema is not up to date. Use --migrate to update.\n")
				os.Exit(1)
			}

			// Perform migration.
			if err := b.Migrate(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to migrate database: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Database migration completed successfully.\n")
		}
	}

	// Open input.
	var input io.Reader
	if inputFile == "" {
		input = os.Stdin
	} else {
		f, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open input file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		input = f
	}

	// Read and parse JSON input.
	decoder := json.NewDecoder(input)
	for {
		var node common.Node
		if err := decoder.Decode(&node); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error: failed to decode JSON: %v\n", err)
			os.Exit(1)
		}

		// Process the unit node.
		if err := b.ProcessUnit(&node); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to process unit: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "Bundling completed successfully.\n")
}
