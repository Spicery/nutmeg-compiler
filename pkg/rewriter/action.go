package rewriter

import (
	"fmt"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

type Action interface {
	// Returns the node (possibly modified or replaced) and a boolean indicating
	// whether any modification occurred. If modified is false, the returned node
	// should be ignored and the original used.
	Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool)
}

////////////////////////////////////////////////////////////////////////////////
/// Actions
////////////////////////////////////////////////////////////////////////////////

type ClearOptionsAction struct {
}

func (a *ClearOptionsAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	clear(node.Options)
	return node, true
}

type NullAction struct {
}

func (a *NullAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	// Continue action does nothing but reports success, allowing the rule to succeed
	// without modifying the node, so processing can continue to the next rule.
	return node, false
}

type FailAction struct {
	Message string
}

func (a *FailAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	// Output error message with span information and exit immediately.
	if node != nil {
		fmt.Fprintf(os.Stderr, "%s, for node '%s', at line %d, column %d\n", a.Message, node.Name, node.Span.StartLine, node.Span.StartColumn)
	} else {
		fmt.Fprintf(os.Stderr, "Validation error (no node): %s\n", a.Message)
	}
	os.Exit(1)
	return node, false
}

type AssertAction struct {
	AssertPattern *Pattern
}

func (a *AssertAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	// Test if the assertion pattern matches the node.
	matches, _ := a.AssertPattern.Matches(node, path)
	if !matches {
		// Output error message with span information and node name.
		fmt.Fprintf(os.Stderr, "Node '%s' failed to meet pattern conditions, at %s\n", node.Name, node.Span.SpanString())
		os.Exit(1)
		return node, false
	}
	return node, false
}

type ReplaceValueFromAction struct {
	Key    string
	Source string
	From   string
}

func (a *ReplaceValueFromAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	node.Options[a.Key] = fetchFromSource(a.From, a.Source, pattern, childPosition, node, path)
	return node, true
}

type ReplaceValueAction struct {
	Key  string
	With string
}

func (a *ReplaceValueAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	k := a.Key
	node.Options[k] = a.With
	return node, true
}

type ReplaceNameWithAction struct {
	With string
}

func (a *ReplaceNameWithAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	node.Name = a.With
	return node, true
}

type ReplaceNameFromAction struct {
	Source string
	From   string
}

func fetchFrom(from string, key *string, node *common.Node) string {
	switch from {
	case "value":
		if key == nil {
			return ""
		}
		return node.Options[*key]
	case "key":
		if key == nil {
			return ""
		}
		return *key
	case "name":
		return node.Name
	}
	return ""
}

func fetchFromSource(from string, source string, pattern *Pattern, childPosition int, node *common.Node, path *common.Path) string {
	switch source {
	case "self":
		return fetchFrom(from, pattern.Self.Key, node)
	case "parent":
		if path == nil || path.Parent == nil {
			return ""
		}
		return fetchFrom(from, pattern.Parent.Key, path.Parent)
	case "child":
		if pattern.Child == nil || pattern.Child.Key == nil || childPosition < 0 || childPosition >= len(node.Children) {
			return ""
		}
		return fetchFrom(from, pattern.Child.Key, node.Children[childPosition])
	}
	return ""
}

func (a *ReplaceNameFromAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	new_name := fetchFromSource(a.From, a.Source, pattern, childPosition, node, path)
	node.Name = new_name
	return node, true
}

type ReplaceByChildAction struct {
	ChildIndex int
}

func (a *ReplaceByChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	if a.ChildIndex < 0 || a.ChildIndex >= len(node.Children) {
		fmt.Fprintln(os.Stderr, "ReplaceByChild: failed, invalid child index", a.ChildIndex)
		return node, false
	}
	return node.Children[a.ChildIndex], true
}

type InlineChildAction struct {
}

func (a *InlineChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	if childPosition < 0 || childPosition >= len(node.Children) {
		fmt.Fprintln(os.Stderr, "InlineChildAction: invalid child position")
		return node, false
	}

	matched_child := node.Children[childPosition]
	old_children := node.Children
	new_children := make([]*common.Node, 0)
	new_children = append(new_children, old_children[:childPosition]...)
	new_children = append(new_children, matched_child.Children...)
	new_children = append(new_children, old_children[childPosition+1:]...)
	node.Children = new_children

	return node, true
}

type RotateOptionAction struct {
	Key     string
	Values  []string
	Initial string
}

func (a *RotateOptionAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	fmt.Fprintln(os.Stderr, "Pattern:", *pattern)
	k := a.Key
	if node.Options[k] == "" {
		node.Options[k] = a.Initial
	}
	value := node.Options[k]
	for i, v := range a.Values {
		if v == value {
			nextIndex := (i + 1) % len(a.Values)
			node.Options[k] = a.Values[nextIndex]
			return node, true
		}
	}
	return node, false
}

type RemoveOptionAction struct {
	Key string
}

func (a *RemoveOptionAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	delete(node.Options, a.Key)
	return node, true
}

type RenameOptionAction struct {
	From string
	To   string
}

func (a *RenameOptionAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	value, exists := node.Options[a.From]
	if !exists {
		return node, false
	}
	node.Options[a.To] = value
	delete(node.Options, a.From)
	return node, true
}

type SequenceAction struct {
	Actions []Action
}

func (a *SequenceAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	anyModified := false
	for _, action := range a.Actions {
		replacement_node, modified := action.Apply(pattern, childPosition, node, path)
		if modified {
			anyModified = true
			node = replacement_node
		}
	}
	return node, anyModified
}

type ChildAction struct {
	Action Action
}

func (a *ChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	if childPosition < 0 || childPosition >= len(node.Children) {
		return node, false
	}
	child := node.Children[childPosition]
	new_child, modified := a.Action.Apply(pattern, -1, child, &common.Path{Parent: node, Others: path})
	if modified {
		node.Children[childPosition] = new_child
		return node, true
	}
	return node, false
}

type MergeChildWithNextAction struct {
	NextTakesPriority bool
}

func (a *MergeChildWithNextAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	if childPosition < 0 || childPosition >= len(node.Children)-1 {
		return node, false
	}
	child := node.Children[childPosition]
	nextChild := node.Children[childPosition+1]
	child.Children = append(child.Children, nextChild.Children...)
	if a.NextTakesPriority {
		child.Options = mergeOptions(child.Options, nextChild.Options)
	} else {
		child.Options = mergeOptions(nextChild.Options, child.Options)
	}
	child.Span = *child.Span.ToSpan(&nextChild.Span)
	// Remove nextChild from node.Children
	node.Children = append(node.Children[:childPosition+1], node.Children[childPosition+2:]...)
	return node, true
}

func mergeOptions(opt1, opt2 map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range opt1 {
		merged[k] = v
	}
	for k, v := range opt2 {
		merged[k] = v
	}
	return merged
}

type NewNodeChildAction struct {
	Name     string
	Key      *string
	Value    *string
	Children *int
	Offset   int
	Length   *int
}

func (a *NewNodeChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	newNode := &common.Node{
		Name:     a.Name,
		Options:  make(map[string]string),
		Children: []*common.Node{},
	}
	if a.Key != nil && a.Value != nil {
		fmt.Fprintln(os.Stderr, "NewNodeChildAction: setting option", *a.Key, "to", *a.Value)
		newNode.Options[*a.Key] = *a.Value
	}

	offset := childPosition + a.Offset
	var length int
	if a.Length == nil {
		length = len(node.Children) - offset
	} else {
		length = *a.Length
	}

	if offset == 0 && length == len(node.Children) {
		newNode.Children = append(newNode.Children, node.Children...)
		node.Children = []*common.Node{newNode}
	} else {
		length := max(0, *a.Length)
		N := min(offset+length, len(node.Children))
		for i := offset; i < N; i++ {
			newNode.Children = append(newNode.Children, node.Children[i])
		}
		node.Children = append(node.Children[:offset], append([]*common.Node{newNode}, node.Children[offset+length:]...)...)
	}
	newNode.UpdateSpan()
	return node, true
}

type PermuteChildrenAction struct {
	NewOrder []int
}

func (a *PermuteChildrenAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil || len(a.NewOrder) < 2 {
		return node, false
	}
	for _, idx := range a.NewOrder {
		if idx < 0 || idx >= len(node.Children) {
			fmt.Fprintln(os.Stderr, "PermuteChildrenAction: invalid index in new order:", idx)
			return node, false
		}
	}
	// a.NewOrder is a permutation "cycle".
	tmp := node.Children[a.NewOrder[0]]
	for i, newIndex := range a.NewOrder {
		if i == 0 {
			continue
		}
		prevIndex := a.NewOrder[i-1]
		node.Children[prevIndex] = node.Children[newIndex]
	}
	// Place the first element in the position of the last element
	node.Children[a.NewOrder[len(a.NewOrder)-1]] = tmp
	return node, true
}

type RemoveChildAction struct {
}

func (a *RemoveChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	if childPosition < 0 || childPosition >= len(node.Children) {
		fmt.Fprintln(os.Stderr, "RemoveChildAction: invalid child position")
		return node, false
	}
	// Remove the child at childPosition
	node.Children = append(node.Children[:childPosition], node.Children[childPosition+1:]...)
	return node, true
}

type RemoveChildrenAction struct {
}

func (a *RemoveChildrenAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *common.Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	if len(node.Children) == 0 {
		return node, false
	}
	// Remove all children
	node.Children = node.Children[:0] // ✓ Reuses capacity, no allocation
	return node, true
}
