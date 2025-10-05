package rewriter

import (
	"fmt"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

type Action interface {
	Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) *common.Node
}

////////////////////////////////////////////////////////////////////////////////
/// Actions
////////////////////////////////////////////////////////////////////////////////

type ReplaceValueAction struct {
	With string `yaml:"new_value,omitempty"`
}

func (a *ReplaceValueAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) *common.Node {
	if node == nil {
		return node
	}
	k := pattern.Self.Key
	if k == nil {
		return node
	}
	node.Options[*k] = a.With
	return node
}

type ReplaceNameWithAction struct {
	With string
}

func (a *ReplaceNameWithAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) *common.Node {
	if node == nil {
		return node
	}
	node.Name = a.With
	return node
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
		fmt.Println("fetchFromSource: fetching from child", pattern.Child)
		if pattern.Child == nil || pattern.Child.Key == nil || childPosition < 0 || childPosition >= len(node.Children) {
			return ""
		}
		fmt.Println("fetchFromSource: fetching from child key", *pattern.Child.Key)
		return fetchFrom(from, pattern.Child.Key, node.Children[childPosition])
	}
	return ""
}

func (a *ReplaceNameFromAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) *common.Node {
	fmt.Println("ReplaceNameFromAction: replacing name from", a.From, "of", a.Source)
	if node == nil {
		return node
	}
	new_name := fetchFromSource(a.From, a.Source, pattern, childPosition, node, path)
	fmt.Println("ReplaceNameFromAction: new name is", new_name)
	node.Name = new_name
	return node
}

type ReplaceByChildAction struct {
	ChildIndex int
}

func (a *ReplaceByChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) *common.Node {
	if node == nil {
		return node
	}
	if a.ChildIndex < 0 || a.ChildIndex >= len(node.Children) {
		return node
	}
	return node.Children[a.ChildIndex]
}

type RepeatAction struct {
	Action Action
}

func (a *RepeatAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) *common.Node {
	fmt.Println("RepeatAction: repeating action")
	if node == nil {
		return node
	}

	for {
		fmt.Println("RepeatAction: applying action, inner loop")
		fmt.Println("Children = ", len(node.Children), ", position = ", childPosition)
		node := a.Action.Apply(pattern, childPosition, node, path)
		var m bool
		m, childPosition = pattern.Matches(node, path)
		if !m {
			break
		}
	}
	return node
}

type InlineChildAction struct {
}

func (a *InlineChildAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) *common.Node {
	if node == nil {
		return node
	}
	if childPosition < 0 || childPosition >= len(node.Children) {
		panic("InlineChildAction: invalid child position")
	}

	matched_child := node.Children[childPosition]
	old_children := node.Children
	new_children := make([]*common.Node, 0)
	new_children = append(new_children, old_children[:childPosition]...)
	new_children = append(new_children, matched_child.Children...)
	new_children = append(new_children, old_children[childPosition+1:]...)
	node.Children = new_children

	return node
}

type RotateOptionAction struct {
	Key     string
	Values  []string
	Initial string
}

func (a *RotateOptionAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) *common.Node {
	if node == nil {
		return node
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
			return node
		}
	}
	return node
}

type RemoveOptionAction struct {
	Key string
}

func (a *RemoveOptionAction) Apply(pattern *Pattern, childPosition int, node *common.Node, path *Path) *common.Node {
	if node == nil {
		return node
	}
	delete(node.Options, a.Key)
	return node
}
