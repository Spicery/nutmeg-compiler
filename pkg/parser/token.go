package parser

import (
	"encoding/json"
	"strings"
)

// TokenType represents the different types of tokens.
type TokenType string

const (
	// Literal constants
	NumericLiteralTokenType     TokenType = "n" // Numeric literals with radix support
	StringLiteralTokenType      TokenType = "s" // String literals with quotes and escapes
	InterpolatedStringTokenType TokenType = "i" // Interpolated string literals e.g. `Hello, \(name)!`
	ExpressionTokenType         TokenType = "e" // Expression tokens (e.g., (1 + 2))

	// Identifier tokens
	StartTokenType    TokenType = "S" // Form start tokens (def, if, while)
	EndTokenType      TokenType = "E" // Form end tokens (end, endif, endwhile)
	BridgeTokenType   TokenType = "B" // Bridge tokens (., ::)
	PrefixTokenType   TokenType = "P" // Prefix operators (return, yield)
	VariableTokenType TokenType = "V" // Variable identifiers

	// Other tokens
	OperatorTokenType       TokenType = "O" // Infix/postfix operators
	OpenDelimiterTokenType  TokenType = "[" // Opening brackets/braces/parentheses
	CloseDelimiterTokenType TokenType = "]" // Closing brackets/braces/parentheses
	UnclassifiedTokenType   TokenType = "U" // Unclassified tokens
	ExceptionTokenType      TokenType = "X" // Exception tokens for invalid constructs
	MarkTokenType           TokenType = "M" // Mark tokens (commas and semicolons)
)

// MarshalJSON implements custom JSON marshaling for Span.
func (s Span) MarshalJSON() ([]byte, error) {
	arr := [4]int{s.StartLine, s.StartColumn, s.EndLine, s.EndColumn}
	return json.Marshal(arr)
}

// UnmarshalJSON implements custom JSON unmarshaling for Span.
func (s *Span) UnmarshalJSON(data []byte) error {
	var arr [4]int
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	s.StartLine = arr[0]
	s.StartColumn = arr[1]
	s.EndLine = arr[2]
	s.EndColumn = arr[3]
	return nil
}

type Arity int

const (
	Zero Arity = iota
	One
	Many
)

// Token represents a single token from the Nutmeg source code.
type Token struct {
	// Common fields for all tokens
	Text  string    `json:"text"`
	Span  Span      `json:"span"`
	Type  TokenType `json:"type"`
	Alias *string   `json:"alias,omitempty"` // The node alias, if any

	// String token fields
	Quote     string   `json:"quote,omitempty"`
	Value     *string  `json:"value,omitempty"`
	Specifier *string  `json:"specifier,omitempty"`
	Subtokens []*Token `json:"subtokens,omitempty"`

	// Numeric token fields
	Radix    *string `json:"radix,omitempty"` // Textual radix prefix (e.g., "0x", "2r", "0t", "" for decimal)
	Base     *int    `json:"base,omitempty"`  // Numeric base (e.g., 16, 2, 3, 10)
	Mantissa *string `json:"mantissa,omitempty"`
	Fraction *string `json:"fraction,omitempty"`
	Exponent *int    `json:"exponent,omitempty"`
	Balanced *bool   `json:"balanced,omitempty"` // For balanced ternary numbers

	// Start token, Label token, and Compound token fields
	Expecting []string `json:"expecting,omitempty"` // For start tokens (immediate next tokens) and label/compound tokens (what can follow them)
	In        []string `json:"in,omitempty"`        // For label and compound tokens - what can contain them
	ClosedBy  []string `json:"closed_by,omitempty"` // For start tokens and delimiter tokens - what can close them
	Arity     *Arity   `json:"arity,omitempty"`     // For start tokens - whether they introduce a single statement block

	// Operator token fields
	Precedence *[3]int `json:"precedence,omitempty"` // [prefix, infix, postfix] precedence values

	// Delimiter fields (for '[' tokens)
	InfixPrecedence *int  `json:"infix,omitempty"`  // For delimiter infix usage
	Prefix          *bool `json:"prefix,omitempty"` // For delimiter prefix usage

	// Exception token fields
	Reason *string `json:"reason,omitempty"` // For exception tokens - explanation of the error

	// Newline tracking fields
	LnBefore *bool `json:"ln_before,omitempty"` // True if token was preceded by a newline
	LnAfter  *bool `json:"ln_after,omitempty"`  // True if token was followed by a newline
}

// NewToken creates a new token with the basic required fields.
func NewToken(text string, tokenType TokenType, span Span) *Token {
	return &Token{
		Text: text,
		Type: tokenType,
		Span: span,
	}
}

func (t *Token) ToKind() string {
	switch t.Text {
	case "[", "]":
		return "brackets"
	case "{", "}":
		return "braces"
	case "(", ")":
		return "parentheses"
	default:
		return string(t.Text)
	}
}

func (t *Token) ToSeparator() string {
	switch t.Text {
	case ",":
		return "comma"
	case ";":
		return "semicolon"
	default:
		return "unknown"
	}
}

func (t *Token) InfixPrec() int {
	if t.InfixPrecedence != nil {
		return *t.InfixPrecedence
	}
	if t.Precedence != nil {
		return t.Precedence[1]
	}
	return 0
}

func (t *Token) PrefixPrec() int {
	if t.Precedence != nil {
		return t.Precedence[0]
	}
	return 0
}

func (t *Token) PostfixPrec() int {
	if t.Precedence != nil {
		return t.Precedence[2]
	}
	return 0
}

func (t *Token) ExpectingMessage(excluding string) string {
	message := strings.Builder{}
	if len(t.Expecting) > 0 {
		for i, x := range t.Expecting {
			if x == excluding {
				continue
			}
			if i > 0 {
				message.WriteString("/")
			}
			message.WriteString(x)
		}
	}
	return message.String()
}
