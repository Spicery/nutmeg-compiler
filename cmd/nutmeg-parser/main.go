package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/common"
	"github.com/spicery/nutmeg-compiler/pkg/parser"
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

	// Select the appropriate print function based on format
	printFunc := common.PickPrintFunc(selectedFormat)

	tree := &common.Node{
		Name:     "unit",
		Options:  map[string]string{},
		Children: []*common.Node{},
	}
	if srcPath != nil && *srcPath != "" {
		tree.Options["src"] = *srcPath
	}

	var node *common.Node
	var err error
	for node, err = p.TryReadExpr(); node != nil; node, err = p.TryReadExpr() {
		if len(tree.Children) == 0 {
			tree.Span = node.Span
		} else {
			tree.Span = *tree.Span.ToSpan(&node.Span)
		}
		tree.Children = append(tree.Children, node)
	}
	printFunc(tree, "  ", os.Stdout, &common.PrintOptions{
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
