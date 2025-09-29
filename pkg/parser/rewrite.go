// This file contains the rewrite rules for the nutmeg-parser.

package parser

import (
	. "github.com/spicery/nutmeg-parser/pkg/common"
)

type Path struct {
	SiblingPosition int // Position among siblings
	Parent          *Node
	Others          *Path
}

type Rewriter struct {
	Substitutions *Substitutions
}

func NewRewriter(rewriteConfig *RewriteConfig) *Rewriter {
	return &Rewriter{Substitutions: rewriteConfig.Substitutions}
}

func (r *Rewriter) Rewrite(node *Node) *Node {
	return r.DoRewrite(node, nil)
}

func (r *Rewriter) DoRewrite(node *Node, path *Path) *Node {
	if node == nil {
		return nil
	}
	// fmt.Println("Rewriting node:", node.Name)

	node = r.DownwardsRewrites(node, path)
	for i, child := range node.Children {
		node.Children[i] = r.DoRewrite(child, &Path{SiblingPosition: i, Parent: node, Others: path})
	}
	node = r.UpwardsRewrites(node, path)
	return node
}

func (r *Rewriter) DownwardsRewrites(node *Node, path *Path) *Node {
	// fmt.Println("Downwards rewrite for node:", node.Name)
	// fmt.Println("Substitutions?", r.Substitutions != nil)
	if r.Substitutions != nil {
		r.Substitutions.ApplySubstitutionsToNode(node)
	}
	return node
}

func (r *Rewriter) UpwardsRewrites(node *Node, path *Path) *Node {
	return node
}

////////////////////////////////////////////////////////////////////////////////
/// Rewrite Transformations
////////////////////////////////////////////////////////////////////////////////

// ApplySubstitution applies the appropriate substitution based on context
func (s *Substitutions) ApplySubstitutionsToNode(node *Node) {
	if node == nil || s == nil {
		return
	}
	// fmt.Println("Applying substitutions to node:", node.Name)
	switch node.Name {
	case "part":
		keyword, exists := node.Options["keyword"]
		if exists {
			kwmap := s.Part.Keyword
			if kwmap != nil {
				replacement, exists := kwmap[keyword]
				if exists {
					node.Options["keyword"] = replacement
				}
			}
		}
	case "operator":
		name, exists := node.Options["name"]
		if exists {
			nmmap := s.Operator.Name
			if nmmap != nil {
				replacement, exists := nmmap[name]
				if exists {
					node.Options["name"] = replacement
				}
			}
		}
	case "identifier":
		// fmt.Println("Identifier substitution")
		name, exists := node.Options["name"]
		if exists {
			nmmap := s.Identifier.Name
			if nmmap != nil {
				replacement, exists := nmmap[name]
				if exists {
					node.Options["name"] = replacement
				}
			}
		}
	}
}
