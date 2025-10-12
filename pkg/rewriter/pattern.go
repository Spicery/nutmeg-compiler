package rewriter

import (
	"fmt"
	"os"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

type NodePattern struct {
	Name            *string `yaml:"name,omitempty"`
	Key             *string `yaml:"key,omitempty"`
	Value           *string `yaml:"value,omitempty"`
	Cmp             *bool   `yaml:"cmp,omitempty"`
	Count           *int    `yaml:"count,omitempty"`
	SiblingPosition *int    `yaml:"siblingPosition,omitempty"`
}

// GetCmp returns the comparison value, defaulting to true if not set
func (np *NodePattern) GetCmp() bool {
	if np.Cmp == nil {
		return true // Default to true
	}
	return *np.Cmp
}

func (np *NodePattern) IsEmpty() bool {
	return np == nil || (np.Name == nil && np.Key == nil && np.Value == nil && np.Count == nil && np.SiblingPosition == nil)
}

func (np *NodePattern) Matches(node *common.Node, path *Path) bool {
	if node == nil {
		return false
	}
	if np.IsEmpty() {
		return true
	}
	if np.Name != nil && node.Name != *np.Name {
		return false
	}
	if np.Key != nil {
		val, exists := node.Options[*np.Key]
		if !exists {
			return false
		}
		if np.Value != nil && (val == *np.Value) != np.GetCmp() {
			return false
		}
	}
	if np.Count != nil && len(node.Children) != *np.Count {
		return false
	}
	if np.SiblingPosition != nil && path != nil {
		k := mod(*np.SiblingPosition, len(path.Parent.Children))
		if path.SiblingPosition != k {
			return false
		}
	}
	return true
}

type Pattern struct {
	Parent        *NodePattern `yaml:"parent,omitempty"`
	Self          *NodePattern `yaml:"self,omitempty"`
	Child         *NodePattern `yaml:"child,omitempty"`
	PreviousChild *NodePattern `yaml:"previousChild,omitempty"`
	NextChild     *NodePattern `yaml:"nextChild,omitempty"`
}

func (p *Pattern) Matches(node *common.Node, path *Path) (bool, int) {
	childPosition := -1
	if node == nil {
		return false, childPosition
	}
	if p == nil {
		return false, childPosition
	}
	if p.Self != nil && !p.Self.Matches(node, path) {
		return false, childPosition
	}
	if p.Parent != nil {
		if path == nil || path.Parent == nil {
			return false, childPosition
		}
		if !p.Parent.Matches(path.Parent, path.Others) {
			return false, childPosition
		}
	}
	if p.Child != nil {
		matched := false
		for n, child := range node.Children {
			if p.Child.Matches(child, &Path{SiblingPosition: n, Parent: node, Others: path}) {
				matched = true
				childPosition = n
				break
			}
		}
		if !matched {
			return false, -1
		}
	}
	if p.PreviousChild != nil && childPosition >= 1 {
		prevChild := node.Children[childPosition-1]
		if !p.PreviousChild.Matches(prevChild, &Path{SiblingPosition: childPosition - 1, Parent: node, Others: path}) {
			return false, -1
		}
	}
	if p.NextChild != nil && childPosition <= len(node.Children)-2 {
		nextChild := node.Children[childPosition+1]
		fmt.Fprintln(os.Stderr, "NextChild:", nextChild.Name)
		if !p.NextChild.Matches(nextChild, &Path{SiblingPosition: childPosition + 1, Parent: node, Others: path}) {
			return false, -1
		}
	}
	return true, childPosition
}

func (p *Pattern) Validate(name string) error {
	if p == nil {
		return fmt.Errorf("pattern is nil")
	}
	if p.Self == nil && p.Parent == nil && p.Child == nil && p.PreviousChild == nil && p.NextChild == nil {
		return fmt.Errorf("pattern has no conditions: %s", name)
	}
	return nil
}
