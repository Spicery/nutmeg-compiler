package rewriter

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spicery/nutmeg-compiler/pkg/common"
	"gopkg.in/yaml.v3"
)

type NodePattern struct {
	Name              *string        `yaml:"name,omitempty"`
	NameRegexpString  *string        `yaml:"name.regexp,omitempty"`
	NameRegexp        *regexp.Regexp `yaml:"-"` // Compiled regexp, not marshaled.
	Key               *string        `yaml:"key,omitempty"`
	Value             *string        `yaml:"value,omitempty"`
	ValueRegexpString *string        `yaml:"value.regexp,omitempty"`
	ValueRegexp       *regexp.Regexp `yaml:"-"` // Compiled regexp, not marshaled.
	Cmp               *bool          `yaml:"cmp,omitempty"`
	Count             *int           `yaml:"count,omitempty"`
	SiblingPosition   *int           `yaml:"siblingPosition,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshaling with validation.
func (np *NodePattern) UnmarshalYAML(node *yaml.Node) error {
	// Create an alias type to avoid infinite recursion.
	type nodePatternAlias NodePattern
	aux := (*nodePatternAlias)(np)

	if err := node.Decode(aux); err != nil {
		return err
	}

	// Compile the regexp if pattern string is present.
	if np.ValueRegexpString != nil {
		// Anchor the pattern at both start and end to ensure full string match.
		// Use a non-capturing group to treat the user's pattern as a single unit.
		anchoredPattern := "^(?:" + *np.ValueRegexpString + ")$"
		compiled, err := regexp.Compile(anchoredPattern)
		if err != nil {
			return fmt.Errorf("invalid regexp in 'matches': %w", err)
		}
		np.ValueRegexp = compiled
	}

	if np.NameRegexpString != nil {
		anchoredPattern := "^(?:" + *np.NameRegexpString + ")$"
		compiled, err := regexp.Compile(anchoredPattern)
		if err != nil {
			return fmt.Errorf("invalid regexp in 'name.regexp': %w", err)
		}
		np.NameRegexp = compiled
	}

	return nil
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

func (np *NodePattern) Matches(node *common.Node, path *common.Path) bool {
	if node == nil {
		return false
	}
	if np.IsEmpty() {
		return true
	}
	if np.Name != nil && node.Name != *np.Name {
		return false
	}
	if np.NameRegexp != nil {
		if !np.NameRegexp.MatchString(node.Name) {
			return false
		}
	}
	if np.Key != nil {
		val, exists := node.Options[*np.Key]
		if !exists {
			return false
		}
		if np.Value != nil && (val == *np.Value) != np.GetCmp() {
			return false
		}
		if np.ValueRegexp != nil {
			matched := np.ValueRegexp.MatchString(val)
			if matched != np.GetCmp() {
				return false
			}
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

func (p *Pattern) Matches(node *common.Node, path *common.Path) (bool, int) {
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
			if p.Child.Matches(child, &common.Path{SiblingPosition: n, Parent: node, Others: path}) {
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
		if !p.PreviousChild.Matches(prevChild, &common.Path{SiblingPosition: childPosition - 1, Parent: node, Others: path}) {
			return false, -1
		}
	}
	if p.NextChild != nil && childPosition <= len(node.Children)-2 {
		nextChild := node.Children[childPosition+1]
		fmt.Fprintln(os.Stderr, "NextChild:", nextChild.Name)
		if !p.NextChild.Matches(nextChild, &common.Path{SiblingPosition: childPosition + 1, Parent: node, Others: path}) {
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
