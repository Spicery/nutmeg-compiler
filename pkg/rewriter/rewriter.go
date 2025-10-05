// This file contains the rewrite rules for the nutmeg-parser.

package rewriter

import (
	"fmt"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

type Path struct {
	SiblingPosition int // Position among siblings
	Parent          *common.Node
	Others          *Path
}

type Rule struct {
	Name    string   `yaml:"name,omitempty"`
	Pattern *Pattern `yaml:"match,omitempty"`
	Action  *Action  `yaml:"action,omitempty"`
}

type RewriterPass struct {
	Name           string `yaml:"name,omitempty"`
	DownwardsRules []*Rule
	UpwardsRules   []*Rule
}

type Rewriter struct {
	Name   string         `yaml:"name,omitempty"`
	Passes []RewriterPass `yaml:"passes,omitempty"`
}

// NewRewriter creates a new Rewriter instance from the given RewriteConfig,
// effectively compiling the configuration into executable rules.
func NewRewriter(rewriteConfig *RewriteConfig) (*Rewriter, error) {
	rewriter := &Rewriter{
		Name:   rewriteConfig.Name,
		Passes: []RewriterPass{},
	}
	for _, passConfig := range rewriteConfig.Passes {
		var upwards, downwards []*Rule
		for _, down := range passConfig.Downwards {
			fmt.Println("Processing downwards rule:", down.Name)
			if e := down.Match.Validate(down.Name); e != nil {
				return nil, fmt.Errorf("error in downwards rule %s: %w", passConfig.Name, e)
			}
			downAction, err := down.Action.ToAction()
			if err != nil {
				return nil, fmt.Errorf("error in downwards rule %s: %w", passConfig.Name, err)
			}
			downwards = append(downwards, &Rule{
				Name:    down.Name,
				Pattern: &down.Match,
				Action:  &downAction,
			})
		}
		for _, up := range passConfig.Upwards {
			fmt.Println("Processing upwards rule:", up.Name)
			if e := up.Match.Validate(up.Name); e != nil {
				return nil, fmt.Errorf("error in upwards rule %s: %w", passConfig.Name, e)
			}
			upAction, err := up.Action.ToAction()
			if err != nil {
				return nil, fmt.Errorf("error in upwards rule %s: %w", passConfig.Name, err)
			}
			upwards = append(upwards, &Rule{
				Name:    up.Name,
				Pattern: &up.Match,
				Action:  &upAction,
			})
		}
		rewriter.Passes = append(rewriter.Passes, RewriterPass{
			Name:           passConfig.Name,
			DownwardsRules: downwards,
			UpwardsRules:   upwards,
		})
	}
	return rewriter, nil
}

func (r *Rewriter) Rewrite(node *common.Node) *common.Node {
	for _, pass := range r.Passes {
		node = pass.doRewrite(node, nil)
	}
	return node
}

func (r *RewriterPass) doRewrite(node *common.Node, path *Path) *common.Node {
	if node == nil {
		return nil
	}
	node = r.downwardsRewrites(node, path)
	for i, child := range node.Children {
		node.Children[i] = r.doRewrite(child, &Path{SiblingPosition: i, Parent: node, Others: path})
	}
	node = r.upwardsRewrites(node, path)
	return node
}

func (r *RewriterPass) downwardsRewrites(node *common.Node, path *Path) *common.Node {
	return applyRules(node, path, r.DownwardsRules)
}

func (r *RewriterPass) upwardsRewrites(node *common.Node, path *Path) *common.Node {
	return applyRules(node, path, r.UpwardsRules)
}

func applyRules(node *common.Node, path *Path, rules []*Rule) *common.Node {
	for _, rule := range rules {
		if rule != nil && rule.Pattern != nil && rule.Action != nil {
			m, n := rule.Pattern.Matches(node, path)
			if m {
				fmt.Printf("Applying rule: '%s' to node: %v\n", rule.Name, node)
				node = (*rule.Action).Apply(rule.Pattern, n, node, path)
			}
		}
	}
	return node
}
