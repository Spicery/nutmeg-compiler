package resolver

import (
	"fmt"
	"strconv"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

const (
	VarOption       = "var"
	ConstOption     = "const"
	ScopeOption     = "scope"
	LastOption      = "last"
	ProtectedOption = "protected"
)

// Resolver performs identifier resolution on a Nutmeg AST.
type Resolver struct {
	currentScope *Scope                     // Current scope during traversal.
	nextID       uint64                     // Next available unique ID.
	globalScope  *Scope                     // The global scope.
	idInfo       map[uint64]*IdentifierInfo // Metadata for each identifier name.
	Closures     map[*Scope]bool            // Set of closure scopes encountered.
}

// NewResolver creates a new resolver instance.
func NewResolver() *Resolver {
	globalScope := &Scope{
		Level:        0,
		DynamicLevel: 0,
		Identifiers:  make(map[string]*IdentifierInfo),
		Parent:       nil,
		IsDynamic:    false, // Global scope is lexical.
		Node:         nil,
		Captured:     nil,
	}
	return &Resolver{
		currentScope: globalScope,
		globalScope:  globalScope,
		nextID:       0, // Start IDs at 0.
		idInfo:       make(map[uint64]*IdentifierInfo),
		Closures:     make(map[*Scope]bool),
	}
}

// Resolve performs identifier resolution on the given AST.
// Uses a two-pass approach:
// 1. First pass: Build scope structure, assign IDs, collect identifier metadata
// 2. Second pass: Annotate all nodes with the complete resolution information
func (r *Resolver) Resolve(root *common.Node) error {
	// First pass: collect identifier information
	if err := r.traverse(root); err != nil {
		fmt.Println(">>> Failing 3")
		return err
	}

	// Second pass: annotate all nodes
	if err := r.annotate(root); err != nil {
		return err
	}

	// Third pass: implement closure captures.
	r.closureCaptures()

	return nil
}

func (r *Resolver) closureCaptures() {
	for closureScope := range r.Closures {
		node := closureScope.Node
		// Transform the function node into a partapply node with two arguments:
		// 1. The original function (as a fn node) with additional captured parameters.
		// 2. An arguments node containing the captured identifiers as id nodes.

		// Create the renumbering mapping and populate as we create the args list.
		renumber_str := make(map[string]string)
		// Create the arguments node.
		args := &common.Node{
			Name:     common.NameArguments,
			Children: []*common.Node{},
			Options:  make(map[string]string),
		}
		for _, info := range closureScope.Captured {
			next_id := r.nextID
			next_id_str := fmt.Sprintf("%d", next_id)
			r.nextID++
			renumber_str[fmt.Sprintf("%d", info.UniqueID)] = next_id_str
			r.nextID++
			arg_node := info.toNode(common.ValueInner)
			// arg_node.Options[common.OptionSerialNo] = next_id_str
			args.Children = append(args.Children, arg_node)
		}

		new_fn_node := &common.Node{
			Name:     common.NameFn,
			Children: node.Children,
			Options:  node.Options,
		}
		params := new_fn_node.Children[0]
		for _, info := range closureScope.Captured {
			params.Children = append(params.Children, info.toNode(common.ValueInner))
		}
		node.Name = common.NamePartApply
		node.Children = []*common.Node{new_fn_node, args}
		node.Options = make(map[string]string)

		// Renumber the captured identifiers in the function body.
		renumberIdentifiersInNode(new_fn_node, renumber_str)
	}
}

func renumberIdentifiersInNode(node *common.Node, renumber_str map[string]string) {
	if node != nil && node.Name == common.NameIdentifier {

		no := node.Options[common.OptionSerialNo]
		serial_no, err := strconv.ParseUint(no, 10, 64)
		if err == nil {

			if new_no, found := renumber_str[no]; found {
				node.Options[common.OptionSerialNo] = new_no
			} else {
				fmt.Println("Cannot find serial number in map: %d", serial_no)
			}
		}
	}
	// Renumber the node's identifier if it exists in the mapping.
	// Recursively renumber child nodes.
	for _, child := range node.Children {
		renumberIdentifiersInNode(child, renumber_str)
	}
}

// traverse performs a custom traversal of the AST, handling different node types appropriately.
// First pass only - builds scope structure and assigns IDs, but does not annotate nodes.
func (r *Resolver) traverse(node *common.Node) error {
	if node == nil {
		return fmt.Errorf("invalid node")
	}

	// Handle different node types.
	switch node.Name {
	case common.NameBind:
		return r.handleBind(node)
	case common.NameDef:
		return r.handleDef(node)
	case common.NameFn:
		return r.handleFnScope(node)
	case common.NameLet, common.NameIf, common.NameFor:
		return r.handleLexicalScope(node)
	case common.NameIdentifier:
		err := r.handleIdentifier(node)
		if err != nil {
			fmt.Println(">>> Failing 2")
		}
		return err
	default:
		// For other nodes, just traverse children.
		for _, child := range node.Children {
			if err := r.traverse(child); err != nil {
				fmt.Println(">>> Failing 4")
				return err
			}
		}
	}
	return nil
}

// handleBind processes a bind node: bind(id, expression).
// The first child is the identifier being defined, the second is the value.
func (r *Resolver) handleBind(node *common.Node) error {
	if len(node.Children) > 0 && node.Children[0].Name == "id" {
		// Define the identifier in the current scope.
		r.defineIdentifier(node.Children[0])
	}

	// Traverse remaining children (the value expression).
	for i := 1; i < len(node.Children); i++ {
		err := r.traverse(node.Children[i])
		if err != nil {
			fmt.Println(">>> Failing 5")
			return err
		}
	}

	return nil
}

// handleDef processes a def node: def f(x): body end.
// Structure: def -> [apply(f, x), body]
// The apply node contains the function name and parameters.
func (r *Resolver) handleDef(node *common.Node) error {
	if len(node.Children) != 2 {
		return fmt.Errorf("invalid def node structure")
	}

	// Declare the identifier in the current scope.
	if node.Children[0].Name == "apply" {
		applyNode := node.Children[0]
		// First child of apply is the function name (defining occurrence).
		if len(applyNode.Children) > 0 && applyNode.Children[0].Name == "id" {
			info := r.defineIdentifier(applyNode.Children[0])
			if info.ScopeType == GlobalScope {
				info.IsAssignable = false
				info.IsConst = true
			}
		}
	} else {
		return fmt.Errorf("unimplemented: first child must be apply")
	}

	// Enter a new dynamic scope.
	r.currentScope = r.currentScope.NewChildScope(true, node)

	// Extract function name and parameters from the first child (apply node).
	if node.Children[0].Name == "apply" {
		applyNode := node.Children[0]
		// Second child is an arguments node containing parameters (defining occurrences).
		argsNode := applyNode.Children[1]
		for _, arg := range argsNode.Children {
			if arg.Name == "id" {
				r.defineIdentifier(arg)
			}
		}
	} else {
		return fmt.Errorf("unimplemented: first child must be apply")
	}

	// Traverse remaining children (body of the function), skipping the first (apply).
	if err := r.traverse(node.Children[1]); err != nil {
		return err
	}

	// Restore the previous scope.
	r.currentScope = r.currentScope.Parent

	return nil
}

// handleFnScope processes nodes that introduce a dynamic scope (fn).
// Structure for named fn: fn -> [id(name), params..., body]
// Structure for anonymous fn: fn -> [params..., body]
func (r *Resolver) handleFnScope(node *common.Node) error {
	// Enter a new dynamic scope.
	r.currentScope = r.currentScope.NewChildScope(true, node)
	fmt.Println("Entering fn scope, level:", r.currentScope.Level, "dynamic level:", r.currentScope.DynamicLevel)

	// If the fn has a name (first child is an id), define it in the function's own scope
	// for self-reference (e.g., fn factorial(n) =>> ... factorial(n-1) ...)
	name, params, err := r.extractFnInfo(node)
	if err != nil {
		return err
	}
	if name != nil {
		fmt.Println("Defining fn name in its own scope:", *name)
		r.defineIdentifierByName(*name)
	}
	if len(params) == 0 {
		fmt.Println("No fn params to define")
	}
	for _, param := range params {
		fmt.Println("Defining fn param in scope:", getIdentifierName(param))
		r.defineIdentifier(param)
	}

	if err := r.traverse(node.Children[1]); err != nil {
		return err
	}

	// Restore the previous scope.
	r.currentScope = r.currentScope.Parent
	return nil
}

// This is a helper function that takes a function node and extracts the
// name and parameters.
func (r *Resolver) extractFnInfo(node *common.Node) (*string, []*common.Node, error) {
	if node == nil || len(node.Children) == 0 {
		return nil, nil, fmt.Errorf("invalid function node structure")
	}

	// The first child is either an Apply node or a Seq node or a single id node.
	first := node.Children[0]

	if first.Name == common.NameApply {
		// There must be exactly 2 children, the first of which is an id node.
		if len(first.Children) != 2 {
			return nil, nil, fmt.Errorf("invalid function node structure")
		}
		fn_name := first.Children[0]
		fn_args := first.Children[1]
		if fn_name.Name != common.NameIdentifier {
			return nil, nil, fmt.Errorf("invalid function name node")
		}
		if fn_args.Name != common.NameSeq {
			return nil, nil, fmt.Errorf("invalid function arguments node")
		}
		name := getIdentifierName(fn_name)
		return &name, fn_args.Children, nil
	}

	if first.Name == common.NameIdentifier {
		return nil, []*common.Node{first}, nil
	}

	if first.Name != common.NameSeq {
		return nil, first.Children, nil
	}

	return nil, nil, fmt.Errorf("invalid function node")
}

// handleLexicalScope processes nodes that introduce a lexical scope (let, if, for).
func (r *Resolver) handleLexicalScope(node *common.Node) error {
	// Enter a new lexical scope.
	r.currentScope = r.currentScope.NewChildScope(false, node)

	// TODO: Extract parameters/definitions for let, handle if/for structure.
	// This will depend on the specific structure of these nodes.

	// Traverse all children.
	for _, child := range node.Children {
		if err := r.traverse(child); err != nil {
			return err
		}
	}

	// Restore the previous scope.
	r.currentScope = r.currentScope.Parent

	return nil
}

// handleIdentifier processes an identifier node (a use of an identifier).
// First pass - records the usage for later analysis.
func (r *Resolver) handleIdentifier(node *common.Node) error {
	// During the first pass, we just need to ensure the identifier is known.
	// The actual annotation happens in the second pass.

	// Look up the identifier to ensure it's registered (may be undefined).
	// This has the side effect of registering undefined identifiers as global.
	info, _, err := r.lookupIdentifier(node)
	if err != nil {
		fmt.Println(">>> Failing")
		return err
	}
	// Update the last reference since we're traversing in order.
	info.LastReference = node
	return nil
}

// NewIdentifierInfo creates a new IdentifierInfo with a unique ID.
func (r *Resolver) NewIdentifierInfo(name string) *IdentifierInfo {
	uniqueID := r.nextID
	r.nextID++
	info := r.currentScope.NewIdentifierInfo(name, uniqueID)
	r.idInfo[uniqueID] = info
	return info
}

func (r *Resolver) NewGlobalIdentifierInfo(name string) *IdentifierInfo {
	uniqueID := r.nextID
	r.nextID++
	info := r.globalScope.NewIdentifierInfo(name, uniqueID)
	info.IsAssignable = false
	info.IsConst = false
	r.idInfo[uniqueID] = info
	return info
}

// defineIdentifier defines a new identifier in the current scope.
// First pass only - collects information but does not annotate nodes.
func (r *Resolver) defineIdentifier(node *common.Node) *IdentifierInfo {
	if node.Name != "id" {
		return nil
	}

	name := getIdentifierName(node)
	if name == "" {
		return nil
	}

	// Create and store metadata for this identifier.
	info := r.NewIdentifierInfo(name)
	q, ok := node.Options[VarOption]
	if ok {
		info.IsAssignable = (q == "true")
	}
	q, ok = node.Options[ConstOption]
	if ok {
		info.IsConst = (q == "true")
	}
	q, ok = node.Options[ProtectedOption]
	if ok {
		fmt.Println("Setting protected option for identifier:", name, "to", q)
		info.IsProtected = (q == "true")
	}
	node.Options[common.OptionSerialNo] = fmt.Sprintf("%d", info.UniqueID)
	return info
}

func (r *Resolver) defineIdentifierByName(name string) {
	// Create and store metadata for this identifier.
	r.NewIdentifierInfo(name)
}

// annotate performs the second pass traversal to annotate all nodes with resolution information.
// This re-traverses the tree with scope tracking to properly annotate each identifier.
func (r *Resolver) annotate(node *common.Node) error {
	// Downwards pass
	switch node.Name {
	case common.NameIdentifier:
		info := r.getIdentifierInfo(node)

		node.Options[VarOption] = fmt.Sprintf("%t", info.IsAssignable)
		node.Options[ConstOption] = fmt.Sprintf("%t", info.IsConst)
		node.Options[ScopeOption] = string(info.ScopeType)
		// Check if this is the last reference to this identifier
		if info.LastReference == node {
			node.Options[LastOption] = "true"
		}
	}

	// Recurse into children
	for _, child := range node.Children {
		err := r.annotate(child)
		if err != nil {
			return err
		}
	}

	// Upwards pass
	switch node.Name {
	case common.NameBind:
		// Implement IsShadowable.
		id := node.Children[0]
		if id.Name == common.NameIdentifier {
			info := r.getIdentifierInfo(id)
			if info.ScopeType != GlobalScope {
				// fmt.Printf("Checking scope chain for shadowing of: %s\n", info.Name)
				// Scan the scope chain looking for prior definitions of the same identifier.
				for s := r.currentScope; s != nil; s = s.Parent {
					// fmt.Println("Checking scope level:", s.Level, "dynamic level:", s.DynamicLevel, "isDynamic:", s.IsDynamic)
					// for _, prior := range s.Identifiers {
					// 	fmt.Println("Checking prior identifier:", prior.Name, prior.IsProtected)
					// }
					prior, found := s.Identifiers[info.Name]
					if found && prior != nil {
						// fmt.Println("FOUND PRIOR DEFINITION OF:", info.Name)
						if prior.IsProtected {
							return fmt.Errorf("trying to re-declare protected identifier: %s, at line %d, column %d", info.Name, id.Span.StartLine, id.Span.StartColumn)
						}
					}
				}
			}
		} else {
			return fmt.Errorf("invalid bind structure, at line %d, column %d", node.Span.StartLine, node.Span.StartColumn)
		}
	case common.NameAssign:
		// Implement IsAssignable.
		id := node.Children[0]
		if id.Name == common.NameIdentifier {
			info := r.getIdentifierInfo(id)
			// fmt.Printf("Checking assignability of: %s, %t\n", info.Name, info.IsAssignable)
			if !info.IsAssignable {
				// fmt.Println("THROWING ERROR")
				return fmt.Errorf("assigning to non-assignable identifier: %s, at line %d, column %d", info.Name, id.Span.StartLine, id.Span.StartColumn)
			}
		} else {
			return fmt.Errorf("invalid assign node structure, at line %d, column %d", node.Span.StartLine, node.Span.StartColumn)
		}
	case common.NameUpdate:
		// TODO: Implement IsUpdatable. This requires an analysis of the
		// side-effects on each parameter of the function being invoked. I do
		// not think it is feasible to implement this without a more
		// sophisticated control flow analysis.
		return nil
	}
	return nil
}

func (r *Resolver) getIdentifierInfo(node *common.Node) *IdentifierInfo {
	no := getNumberOption(node, common.OptionSerialNo)
	info := r.idInfo[no]
	if info == nil {
		panic(fmt.Sprintf("no identifier info for node with no=%d", no))
	}
	return info
}

func getNumberOption(node *common.Node, key string) uint64 {
	if node == nil {
		panic("nil node")
	}
	value, ok := node.Options[key]
	if !ok {
		panic(fmt.Sprintf("missing option '%s' in node '%s'", key, node.Name))
	}
	no, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid no option '%s' in node '%s'", key, node.Name))
	}
	return no
}

// getIdentifierName extracts the identifier name from a node.
func getIdentifierName(node *common.Node) string {
	// Identifier names are stored in the "name" option.
	if name, ok := node.Options["name"]; ok {
		return name
	}
	return ""
}

// lookupIdentifier searches for an identifier in the scope chain.
// Returns (*IdentifierInfo, definingScope).
func (r *Resolver) lookupIdentifier(node *common.Node) (*IdentifierInfo, *Scope, error) {
	name := getIdentifierName(node)
	if name == "" {
		return nil, nil, fmt.Errorf("invalid identifier node")
	}
	fmt.Printf("Looking up identifier node: %s\n", name)
	info, scope, err := r.currentScope.lookupIdentifier(name, r)
	if err != nil {
		fmt.Println(">>> ERROR")
		return nil, nil, err
	}
	fmt.Println(">>> Found identifier info:", info, "in scope level:", scope.Level)
	if info != nil {
		node.Options[common.OptionSerialNo] = fmt.Sprintf("%d", info.UniqueID)
		return info, scope, nil
	}
	// Not found - treat as global undefined identifier.
	info = r.NewGlobalIdentifierInfo(name)
	return info, r.globalScope, nil
}

func (scope *Scope) lookupIdentifier(name string, r *Resolver) (*IdentifierInfo, *Scope, error) {
	// fmt.Println("Looking up identifier:", name, "in scope level:", scope.Level, "dynamic level:", scope.DynamicLevel)
	s := scope
	for s != nil {
		if info, found := s.Identifiers[name]; found && info != nil {
			if info.ScopeType == InnerScope {
				// fmt.Println(name, "scope.DynamicLevel:", scope.DynamicLevel, "info.DefDynLevel:", info.DefDynLevel, "isDynamic:", scope.IsDynamic)
				if scope.DynamicLevel != info.DefDynLevel && info.DefDynLevel != 0 {
					info.ScopeType = OuterScope
					err := scope.captureOuterIdentifier(info, r)
					if err != nil {
						return nil, nil, err
					}
				}
			}
			return info, s, nil
		}
		s = s.Parent
	}
	return nil, nil, nil
}

func (scope *Scope) captureOuterIdentifier(info *IdentifierInfo, r *Resolver) error {
	// fmt.Println("Capturing outer identifier:", info.Name, "defined at dynamic level:", info.DefDynLevel)
	// fmt.Println("Current scope level:", scope.Level, "dynamic level:", scope.DynamicLevel, "isDynamic:", scope.IsDynamic)
	s := scope
	deflevel := info.DefDynLevel
	// fmt.Println("s != nil", s != nil, "s.DynamicLevel > deflevel", s.DynamicLevel > deflevel)
	for s != nil && s.DynamicLevel > deflevel {
		// fmt.Println("Checking scope level:", s.Level, "dynamic level:", s.DynamicLevel, "isDynamic:", s.IsDynamic)
		if s.IsDynamic {
			err := s.captureIdentifier(info, r)
			if err != nil {
				return err
			}
		}
		s = s.Parent
	}
	return nil
}
