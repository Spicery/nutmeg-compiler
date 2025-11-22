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
		return err
	}

	// Second pass: annotate all nodes
	if err := r.annotate(root); err != nil {
		return err
	}

	// Third pass: implement closure captures.
	r.closureCaptures()

	// Fourth pass: lift lambdas to the top level.
	r.liftLambdas(root)

	return nil
}

func (r *Resolver) liftLambdas(root *common.Node) {
	// Collect the paths of all lambda nodes.
	paths := collectLambdaPaths(root, nil)

	// Exclude top-level definitions: unit[*]/bind[1]/lambda
	paths = filterPaths(paths)

	definitions := make([]*common.Node, 0)
	for _, lambdaPath := range paths {
		definitions = append(definitions, r.convertLambdaToDefinition(lambdaPath))
	}

	// Stitch the definitions at the top level.
	root.Children = append(definitions, root.Children...)
}

func (r *Resolver) NewSerialNo() uint64 {
	no := r.nextID
	r.nextID++
	return no
}

func filterPaths(paths []*common.Path) []*common.Path {
	result := make([]*common.Path, 0)
	for _, path := range paths {
		if !isTopLevelPath(path) {
			result = append(result, path)
		}
	}
	return result
}

// isTopLevelPath determines if a path points to a lambda that is already at the top level.
// Top-level lambdas have the structure: unit[*]/bind[1]/lambda, where the lambda is the
// second child (index 1) of a bind node, which is a direct child of the unit node.
// These lambdas should not be lifted since they're already in the correct position.
func isTopLevelPath(path *common.Path) bool {
	// unit[*]/bind[1]
	if path == nil || path.Parent == nil {
		return false // defensive
	}
	if path.Parent.Name != common.NameBind || path.SiblingPosition != 1 {
		return false
	}
	if path.Others == nil || path.Others.Parent == nil || path.Others.Parent.Name != common.NameUnit {
		return false
	}
	return path.Others.Others == nil
}

func (r *Resolver) convertLambdaToDefinition(path *common.Path) *common.Node {
	// Get a new serial number for the binding-identifier.
	serial_no := r.NewSerialNo()
	serial_no_str := fmt.Sprintf("%d", serial_no)
	idNode := &common.Node{
		Name: common.NameIdentifier,
		Options: map[string]string{
			common.OptionSerialNo: serial_no_str,
			common.OptionName:     fmt.Sprintf("tmp-%s", serial_no_str),
			common.OptionScope:    string(UnitScope),
			common.OptionConst:    "true",
			common.OptionVar:      "false",
		},
	}
	// Create the bind node: bind(id, lambdaNode)
	bindNode := &common.Node{
		Name:     common.NameBind,
		Children: []*common.Node{idNode, path.Node()},
		Options:  make(map[string]string),
	}
	// Replace the lambda node with a reference to the new identifier.
	path.Parent.Children[path.SiblingPosition] = idNode
	return bindNode
}

// collectLambdaPaths traverses the tree and collects common.Path structures to all lambda (fn) nodes.
// Returns a slice of Path pointers, each representing the location of a lambda in the tree.
func collectLambdaPaths(node *common.Node, path *common.Path) []*common.Path {
	list := &common.List[*common.Path]{}
	collectLambdaPathsInto(node, path, list)
	return list.Items()
}

func collectLambdaPathsInto(node *common.Node, path *common.Path, list *common.List[*common.Path]) {
	if node == nil {
		return
	}
	if node.Name == common.NameFn {
		list.Add(path)
	}
	for i, child := range node.Children {
		childPath := &common.Path{
			SiblingPosition: i,
			Parent:          node,
			Others:          path,
		}
		collectLambdaPathsInto(child, childPath, list)
	}
}

func (r *Resolver) closureCaptures() {
	for closureScope := range r.Closures {
		partApplyNode := closureScope.Node
		// Transform the function node into a partapply node with two arguments:
		// 1. The original function (as a fn node) with additional captured parameters.
		// 2. An arguments node containing the captured identifiers as id nodes.

		// Create the renumbering mapping and populate as we create the args list.
		renumber_str := make(map[string]string)

		// Create the arguments node for the partApply.
		args := &common.Node{
			Name:     common.NameArguments,
			Children: []*common.Node{},
			Options:  make(map[string]string),
		}
		// For each captured identifier, create a new id node and add to args.
		captured := make([]*IdentifierInfo, 0, len(closureScope.Captured))
		for _, info := range closureScope.Captured {
			captured = append(captured, info)
		}
		for _, info := range captured {
			next_id := r.nextID
			r.nextID++
			next_id_str := fmt.Sprintf("%d", next_id)
			renumber_str[fmt.Sprintf("%d", info.UniqueID)] = next_id_str
			arg_node := info.toNode(common.ValueInner)
			args.Children = append(args.Children, arg_node)
		}

		// The original function node, with added parameters.
		new_fn_node := &common.Node{
			Name:     common.NameFn,
			Children: partApplyNode.Children,
			Options:  partApplyNode.Options,
		}
		params := new_fn_node.Children[0]
		for _, info := range captured {
			params.Children = append(params.Children, info.toNode(common.ValueInner))
		}

		partApplyNode.Name = common.NamePartApply
		partApplyNode.Children = []*common.Node{new_fn_node, args}
		partApplyNode.Options = make(map[string]string)

		// Renumber the captured identifiers in the function body.
		renumberIdentifiersInNode(new_fn_node, renumber_str)
	}
}

func renumberIdentifiersInNode(node *common.Node, renumber_str map[string]string) {
	if node != nil && node.Name == common.NameIdentifier {
		no := node.Options[common.OptionSerialNo]
		if new_no, found := renumber_str[no]; found {
			node.Options[common.OptionSerialNo] = new_no
		}
	}
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
	case common.NameFn:
		return r.handleFnScope(node)
	case common.NameLet, common.NameIf, common.NameFor:
		return r.handleLexicalScope(node)
	case common.NameIdentifier:
		err := r.handleIdentifier(node)
		return err
	default:
		// For other nodes, just traverse children.
		for _, child := range node.Children {
			if err := r.traverse(child); err != nil {
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
			return err
		}
	}

	return nil
}

// handleFnScope processes nodes that introduce a dynamic scope (fn).
// Structure for named fn: fn -> [id(name), params..., body]
// Structure for anonymous fn: fn -> [params..., body]
func (r *Resolver) handleFnScope(node *common.Node) error {
	// Enter a new dynamic scope.
	r.currentScope = r.currentScope.NewChildScope(true, node)

	// If the fn has a name (first child is an id), define it in the function's own scope
	// for self-reference (e.g., fn factorial(n) =>> ... factorial(n-1) ...)
	params, err := r.extractFnInfo(node)
	if err != nil {
		return err
	}

	for _, param := range params {
		r.defineIdentifier(param)
	}

	err = r.traverse(node.Children[1])
	if err != nil {
		return err
	}

	// Restore the previous scope.
	r.currentScope = r.currentScope.Parent
	return nil
}

// This is a helper function that takes a function node and extracts the
// arameters.
func (r *Resolver) extractFnInfo(node *common.Node) ([]*common.Node, error) {
	if node == nil || len(node.Children) == 0 {
		return nil, fmt.Errorf("invalid function node structure")
	}

	// The first child must be an Arguments node.
	first := node.Children[0]

	if first.Name != common.NameArguments {
		return nil, fmt.Errorf("invalid function node")
	}

	return first.Children, nil
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
		info.IsProtected = (q == "true")
	}
	node.Options[common.OptionSerialNo] = fmt.Sprintf("%d", info.UniqueID)
	return info
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
				// Scan the scope chain looking for prior definitions of the same identifier.
				for s := info.DefiningScope.Parent; s != nil; s = s.Parent {
					prior, found := s.Identifiers[info.Name]
					if found && prior != nil {
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
			if !info.IsAssignable {
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
	info, scope, err := r.currentScope.lookupIdentifier(name, r)
	if err != nil {
		return nil, nil, err
	}
	if info != nil {
		node.Options[common.OptionSerialNo] = fmt.Sprintf("%d", info.UniqueID)
		return info, scope, nil
	}
	// Not found - treat as global undefined identifier.
	info = r.NewGlobalIdentifierInfo(name)
	node.Options[common.OptionSerialNo] = fmt.Sprintf("%d", info.UniqueID)
	return info, r.globalScope, nil
}

func (scope *Scope) lookupIdentifier(name string, r *Resolver) (*IdentifierInfo, *Scope, error) {
	s := scope
	for s != nil {
		if info, found := s.Identifiers[name]; found && info != nil {
			if info.ScopeType == InnerScope {
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
	s := scope
	deflevel := info.DefDynLevel
	for s != nil && s.DynamicLevel > deflevel {
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
