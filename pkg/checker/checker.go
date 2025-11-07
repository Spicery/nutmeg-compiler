package checker

import (
	"fmt"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

type Bug struct {
	Message string
	Node    *common.Node
}

type Issue struct {
	Message string
	Node    *common.Node
}

// Checker performs syntax validation on a Nutmeg AST.
type Checker struct {
	Bugs   []Bug   // Accumulated internal errors (bugs).
	Issues []Issue // Accumulated validation errors.
}

func (c *Checker) ReportErrors() {
	// First report any bugs and then move onto issues.
	if len(c.Bugs) > 0 {
		fmt.Fprintln(os.Stderr, "Bug in parser detected; the output of the parser is faulty:")
		count := 0
		for _, bug := range c.Bugs {
			count++
			fmt.Fprintf(os.Stderr, "  [%d]. %s, at line %d, column %d\n", count, bug.Message, bug.Node.Span.StartLine, bug.Node.Span.StartColumn)
		}
	}
	if len(c.Issues) > 0 {
		fmt.Fprintln(os.Stderr, "Errors found in the source code:")
		count := 0
		for _, issue := range c.Issues {
			count++
			fmt.Fprintf(os.Stderr, "  [%d]. %s, at line %d, column %d\n", count, issue.Message, issue.Node.Span.StartLine, issue.Node.Span.StartColumn)
		}
	}
}

// NewChecker creates a new checker instance.
func NewChecker() *Checker {
	return &Checker{
		Bugs:   []Bug{},
		Issues: []Issue{},
	}
}

// Check performs syntax validation on the given AST.
// Returns an error if validation fails.
func (c *Checker) Check(node *common.Node) bool {
	if node == nil {
		c.addBug("invalid node: nil", node)
		return false
	}

	if node.Name != common.NameUnit {
		c.addIssue("expected unit node as root", node)
		return false
	}

	c.validateChildren(node)

	return len(c.Issues) == 0 && len(c.Bugs) == 0
}

func (c *Checker) validateChildren(node *common.Node) {
	for _, child := range node.Children {
		c.validate(child)
	}
}

// validate performs recursive validation on the AST.
func (c *Checker) validate(node *common.Node) {
	if node == nil {
		c.addBug("invalid node: nil", node)
		return
	}

	// Check node-specific rules.
	switch node.Name {
	case common.NameApply:
		c.validateApply(node)
	case common.NameDelimited:
		c.validateDelimited(node)
	case common.NameForm:
		c.validateForm(node)
	case common.NameIdentifier:
		c.validateIdentifier(node)
	case common.NameNumber:
		c.validateNumber(node)
	case common.NameOperator:
		c.validateOperator(node)
	case common.NamePart:
		c.addBug("misplaced Part", node)
	case common.NameString:
		c.validateString(node)
	default:
		c.addBug(fmt.Sprintf("unexpected node type: %s", node.Name), node)
	}
}

func (c *Checker) validateString(node *common.Node) {
	c.factArity(0, node)
}

func (c *Checker) validateNumber(node *common.Node) {
	c.factArity(0, node)
}

func (c *Checker) validateOperator(node *common.Node) {
	if _, ok := node.Options[common.OptionName]; !ok {
		c.addBug("operator node missing name option", node)
	}
	c.validateChildren(node)
}

func (c *Checker) validateApply(node *common.Node) {
	if len(node.Children) != 2 {
		c.addBug("apply node must have exactly two children", node)
	}
	lhs := node.Children[0]
	rhs := node.Children[1]
	c.validate(lhs)
	if rhs.Name != common.NameArguments {
		c.addBug("invalid apply arguments node", rhs)
		return
	}
	c.validateChildren(rhs)
}

func (c *Checker) validateDelimited(node *common.Node) {
	val, ok := node.Options[common.OptionKind]
	if !ok {
		c.addBug("delimited node missing kind option", node)
		return
	}
	switch val {
	case common.ValueParentheses, common.ValueBraces, common.ValueBrackets:
	default:
		c.addBug(fmt.Sprintf("unexpected delimited kind: %s", val), node)
	}
	c.validateChildren(node)
}

func (c *Checker) validateIdentifier(node *common.Node) {
	if _, ok := node.Options[common.OptionName]; !ok {
		c.addBug("identifier node missing name option", node)
	}
	if len(node.Children) != 0 {
		c.addBug("identifier node should not have children", node)
	}
}

func (c *Checker) validateForm(node *common.Node) {
	if len(node.Children) == 0 {
		c.addBug("form node must have at least one child", node)
		return
	}
	for _, part := range node.Children {
		if part.Name != common.NamePart {
			c.addBug("form node children must be part nodes", part)
			return
		}
		_, ok := part.Options[common.OptionKeyword]
		if !ok {
			c.addBug("part node missing keyword option", part)
		}
	}
	first := node.Children[0]
	keyword := first.Options[common.OptionKeyword]
	switch keyword {
	case common.ValueDef:
		c.validateFormDef(node)
	case common.ValueFn:
		c.validateFormFn(node)
	case common.ValueIf:
		c.validateFormIf(node)
	case common.ValueFor:
		c.validateFormFor(node)
	case common.ValueLet:
		c.validateFormLet(node)
	default:
		c.addIssue(fmt.Sprintf("unexpected form keyword: %s", keyword), first)
	}
}

// checkDef validates the structure of a "def" node.
func (c *Checker) validateFormDef(form_node *common.Node) {
	if !c.factArity(2, form_node) {
		return
	}

	first_part := form_node.Children[0]
	if !c.factArity(1, first_part) {
		return
	}
	c.validateDefPattern(first_part.Children[0])

	body_part := form_node.Children[1]
	c.validateChildren(body_part)
}

func (c *Checker) validateDefPattern(node *common.Node) {
	// This verifies that the definition pattern consists of
	// the application of an identifier to zero or more
	// identifiers, e.g., "x" or "f a b c". Allowed nodes are
	// "apply", "operator .", "delimited parentheses", and "id".
	// fmt.Println("Checking def pattern:", node.Name)
	switch node.Name {
	case common.NameOperator:
		c.validateDefDot(node)
	case common.NameApply:
		c.validateDefApply(node)
	case common.NameDelimited:
		if node.Options[common.OptionKind] != common.ValueParentheses {
			c.addIssue("invalid delimited kind in def pattern", node)
			return
		}
		if !c.expectArity(1, node) {
			return
		}
		c.validateDefPattern(node.Children[0])
	default:
		c.addIssue("invalid node in def pattern", node)
	}
}

func (c *Checker) validateDefApply(node *common.Node) {
	// fmt.Println("Checking def apply:", node.Name)
	if !c.factArity(2, node) {
		return
	}
	lhs := node.Children[0]
	rhs := node.Children[1]
	switch lhs.Name {
	case common.NameIdentifier:
		c.validateIdentifier(lhs)
	case common.NameOperator:
		c.validateDefDot(lhs)
		c.validateDefArgs(rhs)
	default:
		c.addIssue("invalid lhs in def pattern apply", node)
	}
	c.validateDefArgs(rhs)
}

func (c *Checker) validateDefDot(node *common.Node) {
	// fmt.Println("Checking dot pattern")
	if node.Options[common.OptionName] != "." {
		c.addIssue("invalid operator in def pattern", node)
		return
	}
	if !c.factArity(2, node) {
		return
	}
	lhs := node.Children[0]
	rhs := node.Children[1]
	c.validateDefArg(lhs)
	c.validateDefFn(rhs)
}

func (c *Checker) validateDefArgs(node *common.Node) {
	// fmt.Println("Checking args:", node.Name)
	switch node.Name {
	case common.NameArguments:
		if node.Options[common.OptionKind] != common.ValueParentheses {
			c.addIssue("invalid brackets for function parameters", node)
			return
		}
		for _, child := range node.Children {
			c.validateDefArg(child)
		}
	default:
		c.addIssue("args must be a delimited node", node)
		return
	}
}

func (c *Checker) validateDefArg(node *common.Node) {
	// fmt.Println("Checking arg:", node.Name)
	switch node.Name {
	case common.NameIdentifier:
		c.validateIdentifier(node)
	case common.NameDelimited:
		if !c.expectArity(1, node) {
			return
		}
		c.validateDefArg(node.Children[0])
	default:
		c.addIssue("invalid parameter", node)
	}
}

func (c *Checker) validateDefFn(node *common.Node) {
	// fmt.Println("Checking fn:", node.Name)
	switch node.Name {
	case common.NameIdentifier:
		c.validateIdentifier(node)
	case common.NameDelimited:
		if !c.expectArity(1, node) {
			return
		}
		c.validateDefFn(node.Children[0])
	default:
		c.addIssue("invalid fn in def pattern", node)
	}
}

// validateFormFn validates the structure of a "fn" node.
// TODO: Implement fn-specific validation rules.
func (c *Checker) validateFormFn(node *common.Node) {
	if !c.factArity(2, node) {
		return
	}
	params_part := node.Children[0]
	if !c.factArity(1, params_part) {
		return
	}
	p := params_part.Children[0]
	switch p.Name {
	case common.NameDelimited:
		c.validateDefArgs(p)
	case common.NameIdentifier:
		c.validateIdentifier(p)
	default:
		c.addIssue("fn parameters must be delimited or identifier", node)
		return
	}
	body_part := node.Children[1]
	c.validateChildren(body_part)
}

// checkIf validates the structure of an "if" node.
func (c *Checker) validateFormIf(if_node *common.Node) {
	c.validateGrandChildren(if_node)
}

// validateFormFor validates the structure of a "for" node.
func (c *Checker) validateFormFor(for_node *common.Node) {
	if !c.factArity(2, for_node) {
		return
	}
	c.validateQuery(for_node.Children[0])
	c.validate(for_node.Children[1])
}

func (c *Checker) validateQuery(query *common.Node) {
	switch query.Name {
	case common.NameOperator:
		c.validateQueryOperator(query)
	default:
		c.addIssue("invalid query node", query)
	}
}

func (c *Checker) validateQueryOperator(node *common.Node) {
	switch node.Options[common.OptionName] {
	case "in":
		if !c.factArity(2, node) {
			return
		}
		lhs := node.Children[0]
		rhs := node.Children[1]
		c.validateIdentifier(lhs)
		c.validate(rhs)
	default:
		c.addIssue(fmt.Sprintf("invalid query operator: %s", node.Options[common.OptionName]), node)
	}
}

// checkLet validates the structure of a "let" node.
func (c *Checker) validateFormLet(let_node *common.Node) {
	c.validateGrandChildren(let_node)
}

func (c *Checker) validateGrandChildren(form_node *common.Node) {
	for _, part := range form_node.Children {
		c.validateChildren(part)
	}
}

func (c *Checker) factArity(arity int, node *common.Node) bool {
	if len(node.Children) != arity {
		c.addBug(fmt.Sprintf("expected %d children, got %d", arity, len(node.Children)), node)
		return false
	}
	return true
}

func (c *Checker) expectArity(arity int, node *common.Node) bool {
	if len(node.Children) != arity {
		c.addIssue(fmt.Sprintf("expected %d children, got %d", arity, len(node.Children)), node)
		return false
	}
	return true
}

// We add a bug if nutmeg-parser is supposed to guarantee the condition
// but it is violated.
func (c *Checker) addBug(message string, node *common.Node) {
	c.Bugs = append(c.Bugs, Bug{Message: message, Node: node})
}

// We add an issue if the user can write code that nutmeg-parser accepts
// but that code is invalid according to our rules.
func (c *Checker) addIssue(message string, node *common.Node) {
	c.Issues = append(c.Issues, Issue{Message: message, Node: node})
}
