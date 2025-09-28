package parser

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
const NameIdentifier = "identifier"
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

// type FormBuilder struct {
// 	node         *Node
// 	startForm    LineCol // The span of the form
// 	startPart    LineCol // The span of the current part
// 	includeSpans bool    // Whether to include spans in the form
// }

// func NewFormBuilder(partName string, lc LineCol, includeSpans bool, usePrefixSyntax bool) *FormBuilder {
// 	syntax := ValueSurround
// 	if usePrefixSyntax {
// 		syntax = ValuePrefix
// 	}
// 	return &FormBuilder{
// 		node: &Node{
// 			Name: NameForm,
// 			Options: map[string]string{
// 				OptionSyntax: syntax,
// 			},
// 			Children: []*Node{
// 				{
// 					Name: NamePart,
// 					Options: map[string]string{
// 						OptionKeyword: partName,
// 					},
// 					Children: []*Node{},
// 				},
// 			},
// 		},
// 		startForm:    lc,
// 		startPart:    lc,
// 		includeSpans: includeSpans,
// 	}
// }

// func (b *FormBuilder) AddChild(child *Node) {
// 	parts := b.node.Children
// 	lastpart := parts[len(parts)-1]
// 	lastpart.Children = append(lastpart.Children, child)
// }

// func (b *FormBuilder) _endPartSpan(endPart LineCol) {
// 	if b.includeSpans {
// 		parts := b.node.Children
// 		lastpart := parts[len(parts)-1]
// 		lastpart.Options[OptionSpan] = b.startPart.SpanString(endPart)
// 	}
// }

// func (b *FormBuilder) BeginNextPart(partName string, endOldPart LineCol, startNewPart LineCol) {

// 	// Set the span of the last part
// 	b._endPartSpan(endOldPart)

// 	// Create a new part
// 	b.node.Children = append(b.node.Children, &Node{
// 		Name: NamePart,
// 		Options: map[string]string{
// 			OptionKeyword: partName,
// 		},
// 		Children: []*Node{},
// 	})

// 	// Set the start of the new part
// 	b.startPart = startNewPart

// }

// func (b *FormBuilder) Build(endForm LineCol, separator string) *Node {
// 	b.node.Options[OptionSeparator] = separator
// 	if b.includeSpans {
// 		b._endPartSpan(endForm)
// 		b.node.Options[OptionSpan] = b.startForm.SpanString(endForm)
// 	}
// 	return b.node
// }

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
