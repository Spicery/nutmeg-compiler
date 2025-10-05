package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"slices"
	"strings"

	. "github.com/spicery/nutmeg-compiler/pkg/common"
)

type TokenQueue struct {
	tokens []*Token
}

func NewPeekQueue() TokenQueue {
	return TokenQueue{
		tokens: []*Token{},
	}
}

func (q *TokenQueue) IsEmpty() bool {
	return len(q.tokens) == 0
}

func (q *TokenQueue) Push(token *Token) {
	q.tokens = append(q.tokens, token)
}

func (q *TokenQueue) Peek() *Token {
	if len(q.tokens) == 0 {
		return nil
	}
	return q.tokens[0]
}

func (q *TokenQueue) Pop() *Token {
	if len(q.tokens) == 0 {
		return nil
	}
	token := q.tokens[0]
	q.tokens = q.tokens[1:]
	return token
}

type Parser struct {
	scanner *bufio.Scanner
	peeked  TokenQueue
	fragile bool
}

func StringToParser(input string) *Parser {
	return &Parser{
		scanner: bufio.NewScanner(strings.NewReader(input)),
		peeked:  NewPeekQueue(),
		fragile: true,
	}
}

func NewParser(input *os.File, fragile bool) *Parser {
	return &Parser{
		scanner: bufio.NewScanner(input),
		peeked:  NewPeekQueue(),
		fragile: fragile,
	}
}

func (p *Parser) Clone(fragile bool) *Parser {
	// Create a new Parser instance that shares the underlying store.
	return &Parser{
		scanner: p.scanner,
		peeked:  p.peeked,
		fragile: fragile,
	}
}

// PeekToken returns the next token without consuming it. If there are no more
// tokens, it returns nil.
func (p *Parser) PeekToken() *Token {
	if !p.peeked.IsEmpty() {
		return p.peeked.Peek()
	}
	token, _ := p.GetToken()

	// May push nil if at end of input.
	p.peeked.Push(token)
	return token
}

// DropPeekedToken removes the first token from the peeked list, effectively consuming it.
// If there are no peeked tokens, it does nothing.
func (p *Parser) DropPeekedToken() {
	p.peeked.Pop()
}

func (p *Parser) GetToken() (*Token, error) {
	if !p.peeked.IsEmpty() {
		token := p.peeked.Pop()
		return token, nil
	}
	if !p.scanner.Scan() {
		return nil, p.scanner.Err()
	}
	line := p.scanner.Bytes()
	var token Token
	if err := json.Unmarshal(line, &token); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing token: %v\n", err)
		return nil, err
	}

	return &token, nil
}

func (p *Parser) MustReadToken(expectedType TokenType, text string) (*Token, error) {
	token, err := p.GetToken()
	if err != nil {
		return nil, err
	}
	if token == nil || token.Type != expectedType || token.Text != text {
		if token == nil {
			return nil, fmt.Errorf("found end of input while expecting '%s'", text)
		}
		return nil, fmt.Errorf("found '%s' while expecting '%s' at line %d, column %d", token.Text, text, token.Span.StartLine, token.Span.StartColumn)
	}
	return token, nil
}

func (p *Parser) TryReadToken(expectedType TokenType, text string) *Token {
	token := p.PeekToken()
	if token == nil {
		return nil
	}
	if token.Type == expectedType && token.Text == text {
		p.DropPeekedToken()
		return token
	}
	return nil
}

func (p *Parser) MustReadOneOf(expectedType TokenType, closedBy []string) (*Token, error) {
	token, err := p.GetToken()
	if err != nil {
		return nil, err
	}
	if token == nil || token.Type != expectedType || !slices.Contains(closedBy, token.Text) {
		expecting := ""
		for _, ex := range closedBy {
			if expecting != "" {
				expecting += " or "
			}
			expecting += ex
		}
		if token == nil {
			return nil, fmt.Errorf("found end of input while expecting '%s'", expecting)
		}
		return nil, fmt.Errorf("found '%s' while expecting '%s' at line %d, column %d", token.Text, expecting, token.Span.StartLine, token.Span.StartColumn)
	}
	return token, nil
}

func (p *Parser) TryReadOneOf(expectedType TokenType, closedBy []string) *Token {
	token := p.PeekToken()
	if token == nil {
		return nil
	}
	if token.Type == expectedType && slices.Contains(closedBy, token.Text) {
		p.DropPeekedToken()
		return token
	}
	return nil
}

func (p *Parser) TryReadPrimaryExpr() (*Node, error) {
	return p.DoReadPrimaryExpr(true)
}

func (p *Parser) MustReadPrimaryExpr() (*Node, error) {
	return p.DoReadPrimaryExpr(false)
}

func (p *Parser) DoReadPrimaryExpr(optional bool) (*Node, error) {
	token := p.PeekToken()
	if token == nil {
		if optional {
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected end of input while reading expression")
	}
	switch token.Type {
	case "s", "i", "m", "e":
		p.DropPeekedToken()
		return ConvertStringTokenToNode(token)
	case "n":
		p.DropPeekedToken()
		return p.ReadNumber(token)
	case "V":
		p.DropPeekedToken()
		return p.ReadId(token)
	case "[":
		if token.Prefix != nil && *token.Prefix {
			p.DropPeekedToken()
			return p.ReadDelimited(token)
		} else {
			return nil, fmt.Errorf("unexpected start of expression, token '%s' at line %d, char %d", token.Text, token.Span.StartLine, token.Span.StartColumn)
		}
	case PrefixTokenType:
		p.DropPeekedToken()
		return p.ReadPrefixForm(token)
	case StartTokenType:
		p.DropPeekedToken()
		return p.ReadSurroundForm(token)
	case EndTokenType:
		if optional {
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected end token '%s' at line %d, char %d", token.Text, token.Span.StartLine, token.Span.StartColumn)
	case OperatorTokenType:
		if token.Precedence != nil && len(token.Precedence) > 0 && token.Precedence[0] > 0 {
			p.DropPeekedToken()
			arg, err := p.TryReadExprPrec(token.Precedence[0])
			if err != nil {
				return nil, err
			}
			if arg == nil {
				return nil, fmt.Errorf("unexpected end of input while parsing argument for operator '%s'", token.Text)
			}
			return &Node{
				Name: NameOperator,
				Options: map[string]string{
					OptionName:   token.Text,
					OptionSyntax: ValuePrefix,
				},
				Children: []*Node{arg},
			}, nil
		}
	case MarkTokenType:
		if optional {
			return nil, nil
		}
		return nil, fmt.Errorf("misplaced punctuation mark '%s' at line %d, char %d", token.Text, token.Span.StartLine, token.Span.StartColumn)
	case UnclassifiedTokenType:
		return nil, fmt.Errorf("invalid token '%s' found at line %d, char %d", token.Text, token.Span.StartLine, token.Span.StartColumn)
	case ExceptionTokenType:
		return nil, fmt.Errorf("%s '%s' found at line %d, char %d", *token.Reason, token.Text, token.Span.StartLine, token.Span.StartColumn)
	}
	return nil, fmt.Errorf("unimplemented, got token '%s' of type: %s", token.Text, token.Type)
}

func ConvertStringTokenToNode(token *Token) (*Node, error) {
	switch token.Type {
	case "s":
		return ConvertPlainStringToken(token)
	case "i":
		return ConvertInterpolatedStringToken(token)
	case "m":
		return ConvertMultilineStringToken(token)
	case "e":
		return ConvertExpressionSubtoken(token)
	}
	return nil, fmt.Errorf("internal error, malformed string token for '%s' in string at line %d, char %d", token.Type, token.Span.StartLine, token.Span.StartColumn)
}

func ConvertExpressionSubtoken(token *Token) (*Node, error) {
	valueString := ""
	if token.Value != nil {
		valueString = *token.Value
	}
	// Now we pipe *subtoken.Value through nutmeg-tokenizer to generate the output as a string.
	output, perr := PipeThroughNutmegTokenizer(valueString)
	if perr != nil {
		return nil, fmt.Errorf("error tokenizing interpolated string at line %d, char %d: %s", token.Span.StartLine, token.Span.StartColumn, perr.Error())
	}
	return StringToParser(output).MustReadExpr()
}

func ConvertPlainStringToken(token *Token) (*Node, error) {
	return &Node{
		Name: NameString,
		Options: map[string]string{
			OptionValue: *token.Value,
		},
		Span:     token.Span,
		Children: []*Node{},
	}, nil
}

func ConvertMultilineStringToken(token *Token) (*Node, error) {
	joinLines := &Node{
		Name: NameJoinLines,
		Options: map[string]string{
			OptionQuote: token.Quote,
		},
		Children: []*Node{},
		Span:     token.Span,
	}
	if token.Subtokens != nil {
		for _, subtoken := range token.Subtokens {
			child, err := ConvertStringTokenToNode(subtoken)
			if err != nil {
				return nil, err
			}
			joinLines.Children = append(joinLines.Children, child)
		}
	}
	return joinLines, nil
}

func ConvertInterpolatedStringToken(token *Token) (*Node, error) {
	join := &Node{
		Name: NameJoin,
		Options: map[string]string{
			OptionQuote: token.Quote,
		},
		Children: []*Node{},
		Span:     token.Span,
	}
	if token.Subtokens != nil {
		for _, subtoken := range token.Subtokens {
			child, err := ConvertStringTokenToNode(subtoken)
			if err != nil {
				return nil, err
			}
			join.Children = append(join.Children, child)
		}
	}
	return join, nil
}

func PipeThroughNutmegTokenizer(input string) (string, error) {
	// Call the executable nutmeg-tokenizer with input as stdin and capture stdout.
	cmd := exec.Command("nutmeg-tokenizer")
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (p *Parser) ReadSurroundForm(token *Token) (*Node, error) {
	// Switch to fragile mode.
	p = p.Clone(true)

	// Create the form node
	form := &Node{
		Name: NameForm,
		Options: map[string]string{
			OptionSyntax: ValueSurround,
		},
		Children: []*Node{},
	}

	// Create the first part with the start token
	currentPart := &Node{
		Name: NamePart,
		Options: map[string]string{
			OptionKeyword: token.Text,
		},
		Children: []*Node{},
		Span:     token.Span,
	}

	// Read expressions for the start token
	if err := p.readPartExpressions(currentPart, token.Arity); err != nil {
		return nil, err
	}

	form.Children = append(form.Children, currentPart)

	// Process bridge tokens and end token
	for {
		nextToken := p.PeekToken()
		if nextToken == nil {
			return nil, fmt.Errorf("unexpected end of input inside '%s' on line %d, char %d", token.Text, token.Span.StartLine, token.Span.StartColumn)
		}

		// Check if it's an end token
		if nextToken.Type == EndTokenType && token.ClosedBy != nil && slices.Contains(token.ClosedBy, nextToken.Text) {
			if token.Expecting != nil && slices.Contains(token.Expecting, nextToken.Text) {
				p.DropPeekedToken()
				// Set the span for the entire form
				form.Span = *token.Span.ToSpan(&nextToken.Span)
				return form, nil
			}
			return nil, fmt.Errorf("encountered '%s' unexpectedly early at line %d, column %d, while parsing '%s'", nextToken.Text, nextToken.Span.StartLine, nextToken.Span.StartColumn, token.Text)
		}

		// Check if it's a bridge token
		if nextToken.Type == BridgeTokenType {
			// Verify the bridge token is expected
			var text string
			if nextToken.Alias != nil {
				text = *nextToken.Alias
			} else {
				text = nextToken.Text
			}
			if token.Expecting != nil && !slices.Contains(token.Expecting, text) {
				expecting_text := token.ExpectingMessage(nextToken.Text)
				return nil, fmt.Errorf("unexpected token '%s' at line %d, column %d, but expecting %s", nextToken.Text, nextToken.Span.StartLine, nextToken.Span.StartColumn, expecting_text)
			}

			// Verify the bridge token is valid for this form
			if nextToken.In != nil && !slices.Contains(nextToken.In, token.Text) {
				return nil, fmt.Errorf("misplaced token '%s' is not valid inside '%s' at line %d, column %d", nextToken.Text, token.Text, nextToken.Span.StartLine, nextToken.Span.StartColumn)
			}

			p.DropPeekedToken()

			// Create new part for the bridge token
			bridgePart := &Node{
				Name: NamePart,
				Options: map[string]string{
					OptionKeyword: text,
				},
				Children: []*Node{},
				Span:     nextToken.Span,
			}

			// Read expressions for the bridge token
			if err := p.readPartExpressions(bridgePart, nextToken.Arity); err != nil {
				return nil, err
			}

			form.Children = append(form.Children, bridgePart)

			// Update expectations for next bridge token
			token.Expecting = nextToken.Expecting
		} else {
			expecting := token.ExpectingMessage(nextToken.Text)
			return nil, fmt.Errorf("found '%s' at line %d, column %d but expecting %s", nextToken.Text, nextToken.Span.StartLine, nextToken.Span.StartColumn, expecting)
		}
	}
}

// Helper method to read expressions for a part (either single or multiple)
func (p *Parser) readPartExpressions(part *Node, arity *Arity) error {
	var a Arity
	if arity != nil {
		a = *arity
	}
	switch a {
	case Zero:
		// Do nothing
	case One:
		// Read exactly one expression
		expr, err := p.MustReadExpr()
		if err != nil {
			return err
		}
		if expr == nil {
			// Cannot be nil here because MustReadExpr would error.
			t := p.PeekToken()
			if t == nil {
				return fmt.Errorf("encountered end of input while reading expression for part '%s'", part.Options[OptionKeyword])
			}
			return fmt.Errorf("found %s but expected expression for part '%s'", t.Text, part.Options[OptionKeyword])
		}
		part.Children = append(part.Children, expr)
	case Many:
		// Read multiple expressions separated by semicolons
		for {
			nextToken := p.PeekToken()
			if nextToken == nil {
				break
			}

			// Stop if we hit a bridge or end token
			if nextToken.Type == BridgeTokenType || nextToken.Type == EndTokenType {
				break
			}

			expr, err := p.TryReadExpr()
			if err != nil {
				return err
			}
			if expr == nil {
				break
			}
			part.Children = append(part.Children, expr)

			// Check for semicolon separator
			is_semicolon := p.TryReadToken(MarkTokenType, ";") != nil
			if !is_semicolon {
				nextToken := p.PeekToken()
				if nextToken != nil && nextToken.LnBefore != nil && *nextToken.LnBefore {
					is_semicolon = true
				}
			}
			if !is_semicolon {
				nextToken := p.PeekToken()

				// No semicolon found, check if next token is bridge/end token
				if nextToken != nil && (nextToken.Type == BridgeTokenType || nextToken.Type == EndTokenType) {
					// End of expressions, break out
					break
				}
				// If there's another expression coming but no semicolon, that's an error
				if nextToken != nil {
					return fmt.Errorf("found '%s' but expected semicolon between expressions at line %d, column %d", nextToken.Text, nextToken.Span.StartLine, nextToken.Span.StartColumn)
				}
				break
			}
			// Semicolon found, continue to next expression or check for termination
			nextToken = p.PeekToken()
			if nextToken != nil && (nextToken.Type == BridgeTokenType || nextToken.Type == EndTokenType) {
				// Optional trailing semicolon before bridge/end token
				break
			}
		}
	}
	length := len(part.Children)
	if length > 0 {
		lastSpan := part.Children[length-1].Span
		newSpan := part.Span.ToSpan(&lastSpan)
		part.Span = *newSpan
	}
	return nil
}

func (p *Parser) ReadPrefixForm(token *Token) (*Node, error) {

	form := &Node{
		Name: NameForm,
		Options: map[string]string{
			OptionSyntax: ValuePrefix,
		},
		Span:     token.Span,
		Children: []*Node{},
	}
	if token.Arity == nil || *token.Arity != One {
		part := &Node{
			Name: NamePart,
			Options: map[string]string{
				OptionName: token.Text,
			},
			Span:     token.Span,
			Children: []*Node{},
		}
		form.Children = append(form.Children, part)
	} else if token.Arity != nil && *token.Arity == One {
		// Handle prefix operators like 'return', 'yield', etc.
		operand, err := p.TryReadExprPrec(math.MaxInt)
		if err != nil {
			return nil, err
		}
		if operand == nil {
			return nil, fmt.Errorf("expected expression after prefix operator '%s' at line %d, char %d", token.Text, token.Span.StartLine, token.Span.StartColumn)
		}
		part := &Node{
			Name: NamePart,
			Options: map[string]string{
				OptionName: token.Text,
			},
			Span:     *token.Span.ToSpan(&operand.Span),
			Children: []*Node{operand},
		}
		form.Children = append(form.Children, part)
	}
	return form, nil
}

// TryReadExprPrec tries to read an expression with the given precedence,
// returning nil if no expression is found. If no expression is found then it is
// guaranteed that no tokens were consumed.
func (p *Parser) TryReadExprPrec(outerPrec int) (*Node, error) {
	return p.DoReadExprPrec(outerPrec, true)
}

// TryReadExprPrec reads an expression with the given precedence,
// returning an error if no expression is found.
func (p *Parser) MustReadExprPrec(outerPrec int) (*Node, error) {
	return p.DoReadExprPrec(outerPrec, false)
}

// DoReadExprPrec reads an expression with the given precedence. If optional is
// true then it returns nil if no expression is found, consuming no input;
// otherwise it returns an error if no expression is found.
func (p *Parser) DoReadExprPrec(outerPrec int, optional bool) (*Node, error) {
	lhs, err := p.DoReadPrimaryExpr(optional)
	if err != nil {
		return nil, err
	}
	if lhs == nil {
		return nil, nil
	}
	for {
		op := p.PeekToken()
		if op == nil {
			return lhs, nil
		}
		if p.fragile {
			if op.LnBefore != nil && *op.LnBefore {
				return lhs, nil
			}
		}
		prec := op.InfixPrec()
		if prec > 0 && prec <= outerPrec {
			p.DropPeekedToken()
			switch op.Type {
			case OpenDelimiterTokenType:
				args, err := p.ReadDelimited(op)
				if err != nil {
					return nil, err
				}
				lhs = &Node{
					Name: NameApply,
					Options: map[string]string{
						OptionKind: op.ToKind(),
					},
					Span:     *lhs.Span.ToSpan(&args.Span),
					Children: []*Node{lhs, args},
				}
			case OperatorTokenType:
				rhs, err := p.TryReadExprPrec(prec)
				if err != nil {
					return nil, err
				}
				if rhs == nil {
					var made_progress bool
					lhs, made_progress = p.doReadExprPostfixPrec(op, outerPrec, lhs, true)
					if !made_progress {
						return nil, fmt.Errorf("expected expression after operator '%s' at line %d, char %d", op.Text, op.Span.StartLine, op.Span.StartColumn)
					}
				} else {
					lhs = &Node{
						Name: NameOperator,
						Options: map[string]string{
							OptionName:   op.Text,
							OptionSyntax: ValueInfix,
						},
						Span:     *lhs.Span.ToSpan(&rhs.Span),
						Children: []*Node{lhs, rhs},
					}
				}
			default:
				return nil, fmt.Errorf("unexpected token at start of an expression'%s' at line %d, char %d", op.Text, op.Span.StartLine, op.Span.StartColumn)
			}
		} else {
			var made_progress bool
			lhs, made_progress = p.doReadExprPostfixPrec(op, outerPrec, lhs, false)
			if !made_progress {
				return lhs, nil
			}
		}
	}
}

func (p *Parser) doReadExprPostfixPrec(op *Token, outerPrec int, lhs *Node, dropped bool) (*Node, bool) {
	prec := op.PostfixPrec()
	if prec > 0 && prec <= outerPrec {
		if !dropped {
			p.DropPeekedToken()
		}
		return &Node{
			Name: NameOperator,
			Options: map[string]string{
				OptionName:   op.Text,
				OptionSyntax: ValuePostfix,
			},
			Span:     *lhs.Span.ToSpan(&op.Span),
			Children: []*Node{lhs},
		}, true
	}
	return lhs, false
}

func (p *Parser) TryReadExpr() (*Node, error) {
	return p.TryReadExprPrec(math.MaxInt)
}

func (p *Parser) MustReadExpr() (*Node, error) {
	return p.MustReadExprPrec(math.MaxInt)
}

func (p *Parser) ReadNumber(token *Token) (*Node, error) {
	base := 10
	if token.Base != nil {
		base = *token.Base
	}
	mantissa := "0"
	if token.Mantissa != nil {
		mantissa = *token.Mantissa
	}
	fraction := ""
	if token.Fraction != nil {
		fraction = *token.Fraction
	}
	exponent := 0
	if token.Exponent != nil {
		exponent = *token.Exponent
	}
	return &Node{
		Name: NameNumber,
		Options: map[string]string{
			"base":     fmt.Sprintf("%d", base),
			"mantissa": mantissa,
			"fraction": fraction,
			"exponent": fmt.Sprintf("%d", exponent),
			"sign":     "+",
		},
		Span:     token.Span,
		Children: []*Node{},
	}, nil
}

func (p *Parser) ReadId(token *Token) (*Node, error) {
	return &Node{
		Name: NameIdentifier,
		Options: map[string]string{
			OptionName: token.Text,
		},
		Span:     token.Span,
		Children: []*Node{},
	}, nil
}

func (p *Parser) ReadDelimited(token *Token) (*Node, error) {
	p = p.Clone(false) // Switch back to non-fragile mode.
	result := &Node{
		Name: NameDelimited,
		Options: map[string]string{
			OptionKind: token.ToKind(),
		},
		Children: []*Node{},
	}
	var endToken *Token
	for endToken = p.TryReadOneOf(CloseDelimiterTokenType, token.ClosedBy); endToken == nil; endToken = p.TryReadOneOf(CloseDelimiterTokenType, token.ClosedBy) {
		if len(result.Children) != 0 {
			commaToken, err := p.MustReadToken(MarkTokenType, ",")
			if err != nil {
				return nil, err
			}
			result.Options[OptionSeparator] = commaToken.ToSeparator()
		}
		node, err := p.TryReadExpr()
		if err != nil {
			return nil, err
		}
		if node == nil {
			return nil, fmt.Errorf("expected expression in delimited starting at line %d, char %d", token.Span.StartLine, token.Span.StartColumn)
		}
		result.Children = append(result.Children, node)
	}
	result.Span = *token.Span.ToSpan(&endToken.Span)
	return result, nil
}
