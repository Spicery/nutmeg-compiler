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
	Name      string   `yaml:"name,omitempty"`
	Pattern   *Pattern `yaml:"match,omitempty"`
	Action    *Action  `yaml:"action,omitempty"`
	OnSuccess int      `yaml:"onSuccess,omitempty"`
	OnFailure int      `yaml:"onFailure,omitempty"`
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
		downToIndex := make(map[string]int)
		for d, down := range passConfig.Downwards {
			downToIndex[down.Name] = d
		}
		upToIndex := make(map[string]int)
		for u, up := range passConfig.Upwards {
			upToIndex[up.Name] = u
		}
		var upwards, downwards []*Rule
		for d, down := range passConfig.Downwards {
			// fmt.Println("Processing downwards rule:", down.Name)
			if e := down.Match.Validate(down.Name); e != nil {
				return nil, fmt.Errorf("error in downwards rule \"%s/%s\": %w", passConfig.Name, down.Name, e)
			}
			downAction, err := down.Action.ToAction()
			if err != nil {
				return nil, fmt.Errorf("error in downwards rule \"%s/%s\": %w", passConfig.Name, down.Name, err)
			}
			onSuccess := d + 1
			onFailure := d + 1
			if down.RepeatOnSuccess {
				onSuccess = d
			} else if down.OnSuccess != nil {
				value, exists := downToIndex[*down.OnSuccess]
				if exists {
					onSuccess = value
				} else {
					return nil, fmt.Errorf("error in downwards rule \"%s/%s\": onSuccess refers to unknown rule \"%s\"", passConfig.Name, down.Name, *down.OnSuccess)
				}
			}
			if down.OnFailure != nil {
				value, exists := downToIndex[*down.OnFailure]
				if exists {
					onFailure = value
				} else {
					return nil, fmt.Errorf("error in downwards rule \"%s/%s\": onFailure refers to unknown rule \"%s\"", passConfig.Name, down.Name, *down.OnFailure)
				}
			}
			downwards = append(downwards, &Rule{
				Name:      down.Name,
				Pattern:   &down.Match,
				Action:    &downAction,
				OnSuccess: onSuccess,
				OnFailure: onFailure,
			})
		}
		for u, up := range passConfig.Upwards {
			// fmt.Println("Processing upwards rule:", up.Name)
			if e := up.Match.Validate(up.Name); e != nil {
				return nil, fmt.Errorf("error in upwards rule %s: %w", passConfig.Name, e)
			}
			upAction, err := up.Action.ToAction()
			if err != nil {
				return nil, fmt.Errorf("error in upwards rule %s: %w", passConfig.Name, err)
			}
			onSuccess := u + 1
			onFailure := u + 1
			fmt.Println("Upwards rule", up.Name)
			fmt.Println("            ", up.RepeatOnSuccess)
			if up.RepeatOnSuccess {
				onSuccess = u
			} else if up.OnSuccess != nil {
				value, exists := upToIndex[*up.OnSuccess]
				if exists {
					onSuccess = value
				} else {
					return nil, fmt.Errorf("error in upwards rule \"%s/%s\": onSuccess refers to unknown rule \"%s\"", passConfig.Name, up.Name, *up.OnSuccess)
				}
			}
			if up.OnFailure != nil {
				value, exists := upToIndex[*up.OnFailure]
				if exists {
					onFailure = value
				} else {
					return nil, fmt.Errorf("error in upwards rule \"%s/%s\": onFailure refers to unknown rule \"%s\"", passConfig.Name, up.Name, *up.OnFailure)
				}
			}
			upwards = append(upwards, &Rule{
				Name:      up.Name,
				Pattern:   &up.Match,
				Action:    &upAction,
				OnSuccess: onSuccess,
				OnFailure: onFailure,
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
		// fmt.Println("Starting pass:", pass.Name)
		node = pass.doRewrite(node, nil)
	}
	return node
}

func (r *RewriterPass) doRewrite(node *common.Node, path *Path) *common.Node {
	if node == nil {
		return nil
	}
	node = r.downwardsRewrites(node, path)
	for i := 0; i < len(node.Children); i++ {
		node.Children[i] = r.doRewrite(node.Children[i], &Path{SiblingPosition: i, Parent: node, Others: path})
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
	currentRule := 0
	for currentRule < len(rules) {
		rule := rules[currentRule]
		if rule != nil && rule.Pattern != nil && rule.Action != nil {
			// fmt.Println("Checking rule:", rule.Name)
			m, n := rule.Pattern.Matches(node, path)
			// fmt.Println("Result:", m, n)

			if m {
				// fmt.Printf("Applying rule: '%s' to node: %v\n", rule.Name, node)
				node = (*rule.Action).Apply(rule.Pattern, n, node, path)
				currentRule = rule.OnSuccess
				fmt.Println("Success with rule", rule.Name, ", moving to rule #", currentRule)
			} else {
				currentRule = rule.OnFailure
				fmt.Println("Failure, moving to rule #", currentRule)
			}
		}
	}
	return node
}
