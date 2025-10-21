package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/common"
	"github.com/spicery/nutmeg-compiler/pkg/parser"
	"github.com/spicery/nutmeg-compiler/pkg/rewriter"
	"github.com/spicery/nutmeg-compiler/pkg/tokenizer"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const usage = `nutmeg-common - integrated tokenizer, parser, and rewriter for Nutmeg

This command pipes together tokenization, parsing, and rewriting in memory,
providing a single integrated tool for processing Nutmeg source code.

Usage:
  nutmeg-common [options] < input.nutmeg

Options:
`

const DEFAULT_FORMAT = "JSON"

func main() {
	var showHelp, showVersion, noSpans bool
	var inputFile, outputFile, tokenRulesFile, rewriteRulesFile, format, srcPath string
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
	flag.StringVar(&srcPath, "src-path", "", "Source path to annotate the unit node")
	flag.StringVar(&tokenRulesFile, "token-rules", "", "YAML file containing tokenizer rules (optional)")
	flag.StringVar(&rewriteRulesFile, "rewrite-rules", "", "YAML file containing rewrite rules (optional)")
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
		fmt.Printf("nutmeg-common version %s\n", Version)
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
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
	}

	// Read input into string.
	inputBytes, err := io.ReadAll(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	inputString := string(inputBytes)

	// Phase 1: Tokenization
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

	// Phase 2: Parsing - use the tokens directly without conversion.
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

	// Phase 3: Rewriting (optional)
	var rewriteConfig *rewriter.RewriteConfig
	if rewriteRulesFile != "" {
		rewriteConfig, err = rewriter.LoadRewriteConfig(rewriteRulesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading rewrite configuration file: %v\n", err)
			os.Exit(1)
		}

		r, err := rewriter.NewRewriter(rewriteConfig)
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

		r, err := rewriter.NewRewriter(rewriteConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating rewriter: %v\n", err)
			os.Exit(1)
		}

		tree, _ = r.Rewrite(tree)
	}

	// Determine output format.
	printFunc := common.PickPrintFunc(format)

	// Determine output destination.
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

	// Output the result.
	printFunc(tree, "  ", output, &common.PrintOptions{
		TrimTokenOnOutput: trim,
		IncludeSpans:      !noSpans,
	})
}
