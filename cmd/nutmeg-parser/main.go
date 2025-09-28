package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spicery/nutmeg-parser/pkg/parser"
)

func main() {
	// Define command line flags according to CLAUDE.md specifications.
	var format = flag.String("f", "XML", "Output format (JSON, XML, etc.)")
	var formatLong = flag.String("format", "XML", "Output format (JSON, XML, etc.)")
	var srcPath = flag.String("src-path", "", "Source path to annotate the unit with origin")
	var trim = flag.Int("trim", 0, "Trim names for display purposes")
	var noSpans = flag.Bool("no-spans", false, "Suppress span information in output")

	flag.Parse()

	// Use the long form if provided, otherwise use the short form.
	selectedFormat := *format
	if *formatLong != "XML" {
		selectedFormat = *formatLong
	}

	p := parser.NewParser(os.Stdin, false)
	var err error
	var node *parser.Node

	// Select the appropriate print function based on format
	printFunc := pickPrintFunc(selectedFormat)

	tree := &parser.Node{
		Name:     "unit",
		Options:  map[string]string{},
		Children: []*parser.Node{},
	}
	if srcPath != nil && *srcPath != "" {
		tree.Options["src"] = *srcPath
	}
	for node, err = p.TryReadExpr(); node != nil; node, err = p.TryReadExpr() {
		if len(tree.Children) == 0 {
			tree.Span = node.Span
		} else {
			tree.Span = *tree.Span.ToSpan(&node.Span)
		}
		tree.Children = append(tree.Children, node)
	}
	printFunc(tree, "  ", os.Stdout, &parser.ConfigurableOptions{
		TrimTokenOnOutput: *trim,
		IncludeSpans:      !*noSpans,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	token, err := p.GetToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	if token != nil {
		fmt.Fprintf(os.Stderr, "parsing incomplete, next token: '%s' of type %s at line %d, char %d\n", token.Text, token.Type, token.Span.StartLine, token.Span.StartColumn)
		os.Exit(1)
	}
}

func pickPrintFunc(format string) func(*parser.Node, string, io.Writer, *parser.ConfigurableOptions) {
	switch strings.ToUpper(format) {
	case "JSON":
		return parser.PrintASTJSON
	case "XML":
		return parser.PrintASTXML
	case "YAML":
		return parser.PrintASTYAML
	case "MERMAID":
		return parser.PrintASTMermaid
	case "ASCIITREE":
		return parser.PrintASTAsciiTree
	case "DOT":
		return parser.PrintASTDOT
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", format)
		os.Exit(1)
		return nil
	}
}
