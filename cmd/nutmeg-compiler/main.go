package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/bundler"
	"github.com/spicery/nutmeg-compiler/pkg/checker"
	"github.com/spicery/nutmeg-compiler/pkg/codegen"
	"github.com/spicery/nutmeg-compiler/pkg/common"
	"github.com/spicery/nutmeg-compiler/pkg/parser"
	"github.com/spicery/nutmeg-compiler/pkg/resolver"
	"github.com/spicery/nutmeg-compiler/pkg/rewriter"
	"github.com/spicery/nutmeg-compiler/pkg/tokenizer"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const usage = `nutmeg-compiler - integrated Nutmeg compiler toolchain

This command pipes together tokenization, parsing, syntax checking, rewriting,
resolution, code generation, and bundling in memory, providing a complete
compiler pipeline for Nutmeg source code.

Usage:
  nutmeg-compiler [options]

Options:
`

const DEFAULT_FORMAT = "JSON"

func main() {
	var showHelp, showVersion, debug, skipOptional bool
	var inputFile, bundleFile, tokenRulesFile, rewriteRulesFile, format string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		flag.PrintDefaults()
	}

	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&debug, "debug", false, "Enable debug output to stderr")
	flag.BoolVar(&skipOptional, "skip-optional", false, "Skip optional rewrite passes")
	flag.StringVar(&inputFile, "input", "", "Input file (does NOT default to stdin, used for srcPath)")
	flag.StringVar(&bundleFile, "bundle", "", "Bundle file path (required)")
	flag.StringVar(&tokenRulesFile, "token-rules", "", "YAML file containing tokenizer rules (optional)")
	flag.StringVar(&rewriteRulesFile, "rewrite-rules", "", "YAML file containing rewrite rules (optional)")
	flag.StringVar(&format, "f", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")
	flag.StringVar(&format, "format", DEFAULT_FORMAT, "Output format (JSON, XML, etc.)")

	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("nutmeg-compiler version %s\n", Version)
		os.Exit(0)
	}

	// Bundle file is mandatory.
	if bundleFile == "" {
		fmt.Fprintf(os.Stderr, "Error: --bundle flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Input file is mandatory.
	if inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: --input flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Reject any positional arguments.
	if len(flag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Error: Unexpected positional arguments. Use --input flag instead.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Open input file.
	file, err := os.Open(inputFile) // #nosec G304 - CLI tool reads user-specified input files
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	srcPath := inputFile

	// Read input into string.
	inputBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	inputString := string(inputBytes)

	// Phase 1: Tokenization.
	var t *tokenizer.Tokenizer
	if tokenRulesFile != "" {
		rules, err := tokenizer.LoadRulesFile(tokenRulesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading token rules file '%s': %v\n", tokenRulesFile, err)
			os.Exit(1)
		}

		tokenizerRules, err := tokenizer.ApplyRulesToDefaults(rules)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error applying token rules: %v\n", err)
			os.Exit(1)
		}
		t = tokenizer.NewTokenizerWithRules(inputString, tokenizerRules)
	} else {
		t = tokenizer.NewTokenizer(inputString)
	}

	tokens, err := t.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Tokenization error: %v\n", err)
		os.Exit(1)
	}

	// Phase 2: Parsing.
	p := parser.NewParserFromTokens(tokens, true)

	tree := &common.Node{
		Name:     "unit",
		Options:  map[string]string{},
		Children: []*common.Node{},
	}
	if srcPath != "" {
		tree.Options["src"] = srcPath
	}

	var node *common.Node
	for node, err = p.TryReadExpr(); node != nil; node, err = p.TryReadExpr() {
		if len(tree.Children) == 0 {
			tree.Span = node.Span
		} else {
			tree.Span = *tree.Span.ToSpan(&node.Span)
		}
		tree.Children = append(tree.Children, node)
		// Check for semicolon separator.
		isSemicolon := p.TryReadSemiColon()
		if !isSemicolon {
			if p.PeekToken() != nil {
				fmt.Fprintf(os.Stderr, "Unexpected token at end of expression: `%s`\n", p.PeekToken().Text)
				os.Exit(1)
			}
			break
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Phase 3: Syntax checking.
	c := checker.NewChecker()
	if !c.Check(tree) {
		c.ReportErrors()
		os.Exit(1)
	}

	// Phase 4: Rewriting.
	var rewriteConfig *rewriter.RewriteConfig
	if rewriteRulesFile != "" {
		rewriteConfig, err = rewriter.LoadRewriteConfig(rewriteRulesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading rewrite configuration file: %v\n", err)
			os.Exit(1)
		}

		r, err := rewriter.NewRewriterWithOptions(rewriteConfig, debug, skipOptional)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating rewriter: %v\n", err)
			os.Exit(1)
		}

		tree, _ = r.Rewrite(tree)
	} else {
		// Use default rewrite rules.
		rewriteConfig, err = rewriter.LoadRewriteConfigFromString(rewriter.DefaultRewriteRules)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading default rewrite rules: %v\n", err)
			os.Exit(1)
		}

		r, err := rewriter.NewRewriterWithOptions(rewriteConfig, debug, skipOptional)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating rewriter: %v\n", err)
			os.Exit(1)
		}

		tree, _ = r.Rewrite(tree)
	}

	// Phase 5: Resolution.
	res := resolver.NewResolver()
	if err := res.Resolve(tree); err != nil {
		fmt.Fprintf(os.Stderr, "Resolution error: %v\n", err)
		os.Exit(1)
	}

	// Phase 6: Code generation.
	cg := codegen.NewCodeGenerator()
	if err := cg.Generate(tree); err != nil {
		fmt.Fprintf(os.Stderr, "Code generation error: %v\n", err)
		os.Exit(1)
	}

	// Phase 7: Bundling.
	// Check if the bundle file exists.
	_, err = os.Stat(bundleFile)
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
		if !fileExists {
			// Fresh database - auto-migrate.
			if err := b.Migrate(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to migrate database: %v\n", err)
				os.Exit(1)
			}
			if debug {
				fmt.Fprintf(os.Stderr, "Database initialized successfully.\n")
			}
		} else {
			// Existing database needs migration - fail with error.
			fmt.Fprintf(os.Stderr, "Error: database schema is not up to date. Please run migration separately.\n")
			os.Exit(1)
		}
	}

	// Process the unit node.
	if err := b.ProcessUnit(tree); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to process unit: %v\n", err)
		os.Exit(1)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Compilation completed successfully.\n")
	}
}
