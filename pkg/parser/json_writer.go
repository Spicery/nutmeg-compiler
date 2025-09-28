package parser

import (
	"fmt"
	"io"
	"strings"
)

// escapeJSONString ensures all strings in JSON are properly escaped.
func escapeJSONString(value string) string {
	var sb strings.Builder

	for _, r := range value {
		switch r {
		case '"':
			sb.WriteString("\\\"") // Escape double quotes
		case '\\':
			sb.WriteString("\\\\") // Escape backslashes
		case '\b':
			sb.WriteString("\\b") // Escape backspace
		case '\f':
			sb.WriteString("\\f") // Escape form feed
		case '\n':
			sb.WriteString("\\n") // Escape newline
		case '\r':
			sb.WriteString("\\r") // Escape carriage return
		case '\t':
			sb.WriteString("\\t") // Escape tab
		default:
			if r >= 0x20 && r <= 0x7E {
				sb.WriteRune(r) // Printable ASCII characters
			} else {
				sb.WriteString(fmt.Sprintf("\\u%04X", r)) // Non-printable and Unicode characters
			}
		}
	}

	return sb.String()
}

func PrintASTJSON(root *Node, indentDelta string, output io.Writer, options *ConfigurableOptions) {
	// Print the root node (which is the "unit" node)
	printNodeJSON(root, "", indentDelta, output, options)
}

// printNodeJSON recursively prints a single node and its children in JSON format.
func printNodeJSON(node *Node, currentIndent string, indentDelta string, output io.Writer, options *ConfigurableOptions) {
	// Precompute the next level of indentation
	nextIndent := currentIndent + indentDelta

	// Open the object
	fmt.Fprintf(output, "%s{\n", currentIndent)

	// Include the `role` field
	fmt.Fprintf(output, "%s\"role\": \"%s\",\n", nextIndent, node.Name)

	// Flatten the options map directly into string-valued fields
	current := 0
	for key, value := range node.Options {
		current++
		trimmedValue := TrimValue(key, value, options.TrimTokenOnOutput)
		escapedValue := escapeJSONString(trimmedValue) // Escape the option value
		// This cannot be the last field because span always comes last.
		fmt.Fprintf(output, "%s\"%s\": \"%s\",\n", nextIndent, key, escapedValue)
	}

	// Add the span field, which is always the last field.
	value := node.Span.SpanString()
	escapedValue := escapeJSONString(value) // Escape the option value
	fmt.Fprintf(output, "%s\"%s\": \"%s\"\n", nextIndent, OptionSpan, escapedValue)

	// Print the children field
	if len(node.Children) > 0 {
		fmt.Fprintf(output, "%s\"children\": [\n", nextIndent)

		childIndent := nextIndent + indentDelta
		for i, child := range node.Children {
			printNodeJSON(child, childIndent, indentDelta, output, options)
			if i < len(node.Children)-1 {
				fmt.Fprintln(output, ",") // Add a comma for all but the last child
			} else {
				fmt.Fprintln(output)
			}
		}

		fmt.Fprintf(output, "%s]\n", nextIndent) // Close the JSON array for children
	}

	// Close the object
	fmt.Fprintf(output, "%s}", currentIndent)
}
