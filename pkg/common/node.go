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
const NamePartApply = "partapply"
const NameString = "string"
const NameJoin = "join"
const NameJoinLines = "joinlines"
const NameInterpolate = "interpolation"
const NameElement = "element"
const NameElementAttributes = "attributes"
const NameElementChildren = "children"
const NameTag = "tag"
const NameBind = "bind"
const NameAssign = "assign"
const NameUpdate = "update"
const NameDef = "def"
const NameFn = "fn"
const NameLet = "let"
const NameIf = "if"
const NameFor = "for"
const NameSeq = "seq"
const NameSysCall = "syscall"
const NameSysCallCounted = "syscall.counted"
const NameSysFn = "sysfn"
const NamePopLocal = "pop.local"
const NamePushLocal = "push.local"
const NamePushGlobal = "push.global"
const NamePushInt = "push.int"
const NamePushString = "push.string"
const NameReturn = "return"
const NameStackLength = "stack.length"
const NameCallGlobalCounted = "call.global.counted"

const OptionConst = "const"
const OptionKeyword = "keyword"
const OptionKind = "kind"
const OptionName = "name"
const OptionQuote = "quote"
const OptionScope = "scope"
const OptionDecimal = "decimal"
const OptionSeparator = "separator"
const OptionSerialNo = "no"
const OptionSpan = "span"
const OptionSpecifier = "specifier"
const OptionSrc = "src"
const OptionSyntax = "syntax"
const OptionValue = "value"
const OptionVar = "var"
const OptionOffset = "offset"
const OptionBase = "base"
const OptionFraction = "fraction"
const OptionExponent = "exponent"
const OptionMantissa = "mantissa"
const OptionLazy = "lazy"
const OptionNParams = "nparams"
const OptionNLocals = "nlocals"
const OptionOrigin = "origin"

const ValueParentheses = "parentheses"
const ValueBrackets = "brackets"
const ValueBraces = "braces"
const ValueDef = "def"
const ValueFn = "fn"
const ValueLet = "let"
const ValueIf = "if"
const ValueFor = "for"
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
const ValueInner = "inner"
const ValueOuter = "outer"
const ValueGlobal = "global"
const ValueBlank = ""
const ValueTrue = "true"
const ValueFalse = "false"

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

func (n *Node) UpdateSpan() {
	if len(n.Children) > 0 {
		span := n.Children[0].Span
		for _, child := range n.Children[1:] {
			span = span.MergeSpan(&child.Span)
		}
		n.Span = span
	}
}

func (n *Node) ClearChildren() {
	n.Children = n.Children[:0]
}

func (n *Node) ToInteger() *string {
	if n == nil {
		return nil
	}
	if n.Name != NameNumber {
		return nil
	}
	mantissa, mantissa_ok := n.Options[OptionMantissa]
	if !mantissa_ok {
		return nil
	}
	fraction, fraction_ok := n.Options[OptionFraction]
	if !fraction_ok {
		return nil
	}
	exponent, exponent_ok := n.Options[OptionExponent]
	if !exponent_ok {
		return nil
	}
	base, base_ok := n.Options[OptionBase]
	if !base_ok {
		return nil
	}

	if fraction == "" && exponent == "0" && base == "10" {
		return &mantissa
	}
	return nil
}
