package common

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Node struct {
	Name     string            // The name of the node
	Span     Span              // The span of the node in the source
	Options  map[string]string // Attributes (name-value pairs)
	Children []*Node           // Child nodes
}

const NameForm = "form"
const NamePart = "part"
const NameUnit = "unit"
const NameApply = "apply"
const NameArguments = "arguments"
const NameDelimited = "delimited"
const NameGet = "get"
const NameIdentifier = "id"
const NameInvoke = "invoke"
const NameNumber = "number"
const NameOperator = "operator"
const NameString = "string"
const NameJoin = "join"
const NameJoinLines = "joinlines"
const NameInterpolate = "interpolation"
const NameElement = "element"
const NameElementAttributes = "attributes"
const NameElementChildren = "children"
const NameTag = "tag"

const OptionValue = "value"
const OptionsDecimalValue = "decimal"
const OptionName = "name"
const OptionKind = "kind"
const OptionSeparator = "separator"
const OptionKeyword = "keyword"
const OptionSpan = "span"
const OptionSyntax = "syntax"
const OptionQuote = "quote"
const OptionSpecifier = "specifier"
const OptionSrc = "src"

const ValueInfix = "infix"
const ValuePrefix = "prefix"
const ValuePostfix = "postfix"
const ValueSurround = "surround"
const ValueComma = "comma"
const ValueSemicolon = "semicolon"
const ValueUndefined = "undefined"
const ValueNewline = "newline"
const ValueChevron = "chevron"
const ValueRegex = "regex"
const ValueBlank = ""

// TrimValue trims a value if it's a token value and trimming is enabled
func TrimValue(key, value string, trimLength int) string {
	if key == OptionValue && trimLength > 0 && len(value) > trimLength {
		// Reserve space for Unicode ellipsis (1 character: "…")
		if trimLength >= 2 {
			return value[:trimLength-1] + "…"
		} else if trimLength >= 1 {
			// If trim length is too small for ellipsis, just truncate
			return value[:trimLength]
		}
	}
	return value
}

func PickPrintFunc(format string) func(*Node, string, io.Writer, *PrintOptions) {
	switch strings.ToUpper(format) {
	case "JSON":
		return PrintASTJSON
	case "XML":
		return PrintASTXML
	case "YAML":
		return PrintASTYAML
	case "MERMAID":
		return PrintASTMermaid
	case "ASCIITREE":
		return PrintASTAsciiTree
	case "DOT":
		return PrintASTDOT
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", format)
		os.Exit(1)
		return nil
	}
}
