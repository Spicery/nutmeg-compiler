package rewriter

import (
	"fmt"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

type Action interface {
	// Returns the node (possibly modified or replaced) and a boolean indicating
	// whether any modification occurred. If modified is false, the returned node
	// should be ignored and the original used.
	Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool)
}

////////////////////////////////////////////////////////////////////////////////
/// Actions
////////////////////////////////////////////////////////////////////////////////

type ReplaceValueFromAction struct {
	Key    string
	Source string
	From   string
}

func (a *ReplaceValueFromAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
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

func (a *ReplaceValueAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	// fmt.Println("ReplaceValueAction: replacing value with", a.With, "for key", a.Key, "pattern.Self.Key =", pattern.Self.Key)
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

func (a *ReplaceNameWithAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
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

func fetchFromSource(from string, source string, pattern *Pattern, childPosition int, node *common.Node, path *Path) string {
	switch source {
	case "self":
		return fetchFrom(from, pattern.Self.Key, node)
	case "parent":
		if path == nil || path.Parent == nil {
			return ""
		}
		return fetchFrom(from, pattern.Parent.Key, path.Parent)
	case "child":
		// fmt.Println("fetchFromSource: fetching from child", pattern.Child)
		if pattern.Child == nil || pattern.Child.Key == nil || childPosition < 0 || childPosition >= len(node.Children) {
			return ""
		}
		// fmt.Println("fetchFromSource: fetching from child key", *pattern.Child.Key)
		return fetchFrom(from, pattern.Child.Key, node.Children[childPosition])
	}
	return ""
}

func (a *ReplaceNameFromAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	// fmt.Println("ReplaceNameFromAction: replacing name from", a.From, "of", a.Source)
	if node == nil {
		return node, false
	}
	new_name := fetchFromSource(a.From, a.Source, pattern, childPosition, node, path)
	// fmt.Println("ReplaceNameFromAction: new name is", new_name)
	node.Name = new_name
	return node, true
}

type ReplaceByChildAction struct {
	ChildIndex int
}

func (a *ReplaceByChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	// fmt.Println("ReplaceByChild: repeating action")
	if node == nil {
		return node, false
	}
	if a.ChildIndex < 0 || a.ChildIndex >= len(node.Children) {
		fmt.Println("ReplaceByChild: failed, invalid child index", a.ChildIndex)
		return node, false
	}
	// fmt.Println("ReplaceByChild: OK!, replaced by child index", a.ChildIndex)
	return node.Children[a.ChildIndex], true
}

type InlineChildAction struct {
}

func (a *InlineChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	if childPosition < 0 || childPosition >= len(node.Children) {
		fmt.Println("InlineChildAction: invalid child position")
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

func (a *RotateOptionAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	fmt.Println("Pattern:", *pattern)
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

func (a *RemoveOptionAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	delete(node.Options, a.Key)
	return node, true
}

type SequenceAction struct {
	Actions []Action
}

func (a *SequenceAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
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

func (a *ChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	if childPosition < 0 || childPosition >= len(node.Children) {
		return node, false
	}
	child := node.Children[childPosition]
	new_child, modified := a.Action.Apply(pattern, -1, child, &Path{Parent: node, Others: path})
	if modified {
		node.Children[childPosition] = new_child
		return node, true
	}
	return node, false
}

type MergeChildWithNextAction struct {
}

func (a *MergeChildWithNextAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	if node == nil {
		return node, false
	}
	if childPosition < 0 || childPosition >= len(node.Children)-1 {
		return node, false
	}
	child := node.Children[childPosition]
	nextChild := node.Children[childPosition+1]
	child.Children = append(child.Children, nextChild.Children...)
	child.Options = mergeOptions(child.Options, nextChild.Options)
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

func (a *NewNodeChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	// fmt.Println("NewNodeChildAction: inserting new node", a.Name, "at position", childPosition)
	if node == nil {
		return node, false
	}
	newNode := &common.Node{
		Name:     a.Name,
		Options:  make(map[string]string),
		Children: []*common.Node{},
	}
	if a.Key != nil && a.Value != nil {
		fmt.Println("NewNodeChildAction: setting option", *a.Key, "to", *a.Value)
		newNode.Options[*a.Key] = *a.Value
	}
	if a.Length == nil {
		newNode.Children = append(newNode.Children, node.Children...)
		node.Children = []*common.Node{newNode}
	} else {
		// fmt.Println("NewNodeChildAction: inserting with offset", a.Offset, "and length", *a.Length)
		offset := childPosition + a.Offset
		// fmt.Println("NewNodeChildAction: calculated offset is", offset)
		length := max(0, *a.Length)
		N := min(offset+length, len(node.Children))
		for i := offset; i < N; i++ {
			newNode.Children = append(newNode.Children, node.Children[i])
		}
		// before := append(make([]*common.Node, 0), node.Children[:offset]...)
		// after := append(make([]*common.Node, 0), node.Children[offset+length:]...)
		// node.Children = append(before, newNode)
		// node.Children = append(node.Children, after...)
		// fmt.Println("Length of before:", len(before), "after:", len(after), "children now:", len(node.Children))
		node.Children = append(node.Children[:offset], append([]*common.Node{newNode}, node.Children[offset+length:]...)...)
	}
	newNode.UpdateSpan()
	return node, true
}

type PermuteChildrenAction struct {
	NewOrder []int
}

func (a *PermuteChildrenAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) (*common.Node, bool) {
	// fmt.Println("PermuteChildrenAction: permuting children to new order", a.NewOrder)
	if node == nil || len(a.NewOrder) < 2 {
		return node, false
	}
	for _, idx := range a.NewOrder {
		if idx < 0 || idx >= len(node.Children) {
			fmt.Println("PermuteChildrenAction: invalid index in new order:", idx)
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
