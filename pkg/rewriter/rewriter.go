// This file contains the rewrite rules for the nutmeg-parser.

package rewriter

import (
	"fmt"
	"os"

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
	Name                string `yaml:"name,omitempty"`
	DownwardsRules      []*Rule
	UpwardsRules        []*Rule
	DownwardsStartIndex map[string]int // Maps node name to starting rule index.
	UpwardsStartIndex   map[string]int // Maps node name to starting rule index.
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
			if e := up.Match.Validate(up.Name); e != nil {
				return nil, fmt.Errorf("error in upwards rule %s: %w", passConfig.Name, e)
			}
			upAction, err := up.Action.ToAction()
			if err != nil {
				return nil, fmt.Errorf("error in upwards rule %s: %w", passConfig.Name, err)
			}
			onSuccess := u + 1
			onFailure := u + 1
			fmt.Fprintln(os.Stderr, "Upwards rule", up.Name)
			fmt.Fprintln(os.Stderr, "            ", up.RepeatOnSuccess)
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
		pass := RewriterPass{
			Name:           passConfig.Name,
			DownwardsRules: downwards,
			UpwardsRules:   upwards,
		}

		// Optimization 1: Build name-based start index maps.
		pass.DownwardsStartIndex = buildStartIndexMap(downwards)
		pass.UpwardsStartIndex = buildStartIndexMap(upwards)

		// Optimization 2: Optimize OnSuccess/OnFailure jumps.
		optimizeRuleJumps(downwards)
		optimizeRuleJumps(upwards)

		rewriter.Passes = append(rewriter.Passes, pass)
	}
	return rewriter, nil
}

////////////////////////////////////////////////////////////////////////////////
/// Performance Optimizations
////////////////////////////////////////////////////////////////////////////////
//
// The following functions implement two compile-time optimizations to improve
// runtime performance of rule matching:
//
// 1. Name-based Start Index Maps (buildStartIndexMap):
//    Instead of checking all N rules for every node, we build a map from node
//    name to the first rule that could potentially match. This allows us to
//    skip all rules that require a different name.
//
// 2. Jump Target Optimization (optimizeRuleJumps):
//    After a successful match, we often know what the node's name will be
//    (either unchanged or changed to a predictable value by the action).
//    We optimize OnSuccess/OnFailure jumps to skip rules that cannot possibly
//    match based on name constraints.
//
//    Key insights:
//    - If an action doesn't contain ReplaceNameWithAction or ReplaceNameFromAction,
//      the node name remains unchanged after the action.
//    - ReplaceNameWithAction changes the name to a known constant value.
//    - ReplaceNameFromAction changes the name unpredictably (runtime-dependent).
//    - For SequenceAction, we track name changes through the sequence.
//
// Both optimizations are performed once in NewRewriter at load time, so there
// is no runtime overhead for these optimizations.

// buildStartIndexMap creates a map from node name to the first rule index that could match.
// Returns map[string]int where the value is the starting rule index for that node name.
// If a name is not in the map, it means no rules apply (default to len(rules)).
func buildStartIndexMap(rules []*Rule) map[string]int {
	startIndex := make(map[string]int)
	wildcardIndex := len(rules) // Default: no wildcard found.

	fmt.Fprintln(os.Stderr, "[OPT1] Building start index map for", len(rules), "rules")

	// Scan rules to find first occurrence of each name and any wildcard.
	for i, rule := range rules {
		if rule.Pattern != nil && rule.Pattern.Self != nil {
			if rule.Pattern.Self.Name == nil {
				// This is a wildcard rule.
				if wildcardIndex == len(rules) {
					wildcardIndex = i
					fmt.Fprintf(os.Stderr, "[OPT1]   Rule #%d (%s): wildcard - set as default start index\n", i, rule.Name)
				}
			} else {
				name := *rule.Pattern.Self.Name
				if _, exists := startIndex[name]; !exists {
					// First rule for this name.
					startIndex[name] = i
					fmt.Fprintf(os.Stderr, "[OPT1]   Rule #%d (%s): first rule for name '%s'\n", i, rule.Name, name)
				}
			}
		}
	}

	// For names not in the map, they should start at the wildcard (if any) or skip all rules.
	// We'll handle this by checking in applyRules: if not in map, use wildcardIndex.
	// Store the wildcard index under a special key.
	if wildcardIndex < len(rules) {
		startIndex[""] = wildcardIndex // Empty string = default/wildcard.
		fmt.Fprintf(os.Stderr, "[OPT1] Default start index (wildcard): %d\n", wildcardIndex)
	} else {
		startIndex[""] = len(rules) // No wildcard, skip all.
		fmt.Fprintln(os.Stderr, "[OPT1] No wildcard found - unknown names will skip all rules")
	}

	return startIndex
}

// optimizeRuleJumps optimizes OnSuccess and OnFailure jumps by skipping rules that cannot match.
// For each rule, if its OnSuccess/OnFailure points to a rule that cannot possibly match
// the same node (based on name constraint), advance the pointer to skip impossible rules.
func optimizeRuleJumps(rules []*Rule) {
	fmt.Fprintln(os.Stderr, "[OPT2] Optimizing jump targets for", len(rules), "rules")
	optimizationCount := 0

	for i, rule := range rules {
		oldSuccess := rule.OnSuccess
		oldFailure := rule.OnFailure

		rule.OnSuccess = optimizeJump(rule.OnSuccess, rules, rule, true)
		rule.OnFailure = optimizeJump(rule.OnFailure, rules, rule, false)

		if oldSuccess != rule.OnSuccess || oldFailure != rule.OnFailure {
			optimizationCount++
			fmt.Fprintf(os.Stderr, "[OPT2]   Rule #%d (%s):", i, rule.Name)
			if oldSuccess != rule.OnSuccess {
				fmt.Fprintf(os.Stderr, " OnSuccess: %d→%d", oldSuccess, rule.OnSuccess)
			}
			if oldFailure != rule.OnFailure {
				fmt.Fprintf(os.Stderr, " OnFailure: %d→%d", oldFailure, rule.OnFailure)
			}
			fmt.Fprintln(os.Stderr)
		}

		_ = i // Suppress unused variable warning.
	}

	fmt.Fprintf(os.Stderr, "[OPT2] Optimized %d jump targets\n", optimizationCount)
}

// actionChangesName determines if an action might change the Self.Name of a node.
// Returns: (changesName bool, newName *string, isDefinite bool)
// - changesName: true if the action might change the name
// - newName: the specific new name if predictable, nil otherwise
// - isDefinite: true if we're certain about the new name
func actionChangesName(action Action) (bool, *string, bool) {
	if action == nil {
		return false, nil, true
	}

	switch a := action.(type) {
	case *ReplaceNameWithAction:
		// Definite name change to a known value.
		return true, &a.With, true

	case *ReplaceNameFromAction:
		// Name changes but we don't know to what (depends on runtime values).
		return true, nil, false

	case *SequenceAction:
		// Check all actions in sequence.
		// If any changes the name, we need to track it.
		changesName := false
		var finalName *string
		isDefinite := true
		for _, subAction := range a.Actions {
			changes, name, definite := actionChangesName(subAction)
			if changes {
				changesName = true
				finalName = name
				if !definite {
					isDefinite = false
				}
			}
		}
		return changesName, finalName, isDefinite

	case *ChildAction:
		// ChildAction modifies a child, not Self.Name.
		return false, nil, true

	case *ReplaceByChildAction:
		// Replaces the entire node with a child - name will change unpredictably.
		return true, nil, false

	default:
		// Other actions (ReplaceValueAction, RemoveOptionAction, etc.) don't change Self.Name.
		return false, nil, true
	}
}

// optimizeJump advances a jump target past rules that cannot possibly match.
// Given a jump destination and the source rule, skip ahead if the destination
// rule has a name constraint that conflicts with what we know about the node after the action.
func optimizeJump(jumpTarget int, rules []*Rule, sourceRule *Rule, isSuccess bool) int {
	if sourceRule == nil || sourceRule.Pattern == nil || sourceRule.Pattern.Self == nil {
		return jumpTarget
	}

	originalTarget := jumpTarget

	// Determine what we know about the node name after the action.
	var expectedName *string
	if isSuccess && sourceRule.Action != nil {
		// After a successful match, we know the matched name (if not a wildcard).
		matchedName := sourceRule.Pattern.Self.Name

		// Check if the action changes the name.
		changesName, newName, isDefinite := actionChangesName(*sourceRule.Action)

		if changesName {
			if isDefinite && newName != nil {
				// We know the exact new name.
				expectedName = newName
				fmt.Fprintf(os.Stderr, "[OPT2]     Action changes name to '%s'\n", *expectedName)
			} else {
				// Name changes unpredictably, can't optimize.
				fmt.Fprintln(os.Stderr, "[OPT2]     Action changes name unpredictably - can't optimize")
				return jumpTarget
			}
		} else {
			// Name doesn't change, use the matched name.
			expectedName = matchedName
			if expectedName != nil {
				fmt.Fprintf(os.Stderr, "[OPT2]     Name unchanged: '%s'\n", *expectedName)
			} else {
				fmt.Fprintln(os.Stderr, "[OPT2]     Name unchanged: wildcard")
			}
		}
	} else {
		// On failure or no action, we can't predict the node.
		return jumpTarget
	}

	// Now skip rules that can't match the expected name.
	skipped := 0
	for jumpTarget < len(rules) {
		targetRule := rules[jumpTarget]
		if targetRule == nil || targetRule.Pattern == nil {
			// Invalid rule, skip it.
			jumpTarget++
			skipped++
			continue
		}

		if targetRule.Pattern.Self == nil {
			// No constraints on Self - matches anything (wildcard) - stop here.
			if skipped > 0 {
				fmt.Fprintf(os.Stderr, "[OPT2]     Stopped at rule #%d (no Self constraint) after skipping %d rules\n", jumpTarget, skipped)
			}
			break
		}

		targetName := targetRule.Pattern.Self.Name

		if targetName == nil {
			// Target is a wildcard, can match anything - stop here.
			if skipped > 0 {
				fmt.Fprintf(os.Stderr, "[OPT2]     Stopped at rule #%d (%s - wildcard) after skipping %d rules\n", jumpTarget, targetRule.Name, skipped)
			}
			break
		}

		if expectedName == nil {
			// We don't know the name (wildcard source), target requires specific name - skip it.
			jumpTarget++
			skipped++
			continue
		}

		if *targetName == *expectedName {
			// Names match, this rule could apply - stop here.
			if skipped > 0 {
				fmt.Fprintf(os.Stderr, "[OPT2]     Stopped at rule #%d (%s - matches '%s') after skipping %d rules\n", jumpTarget, targetRule.Name, *expectedName, skipped)
			}
			break
		}

		// Names don't match, skip this rule.
		jumpTarget++
		skipped++
	}

	if jumpTarget != originalTarget && skipped == 0 {
		fmt.Fprintf(os.Stderr, "[OPT2]     No optimization possible (target=%d)\n", jumpTarget)
	}

	return jumpTarget
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
	for i := 0; i < len(node.Children); i++ {
		node.Children[i] = r.doRewrite(node.Children[i], &Path{SiblingPosition: i, Parent: node, Others: path})
	}
	node = r.upwardsRewrites(node, path)
	return node
}

func (r *RewriterPass) downwardsRewrites(node *common.Node, path *Path) *common.Node {
	return applyRules(node, path, r.DownwardsRules, r.DownwardsStartIndex)
}

func (r *RewriterPass) upwardsRewrites(node *common.Node, path *Path) *common.Node {
	return applyRules(node, path, r.UpwardsRules, r.UpwardsStartIndex)
}

func applyRules(node *common.Node, path *Path, rules []*Rule, startIndexMap map[string]int) *common.Node {
	// Optimization 1: Start at the first rule that could match this node's name.
	currentRule := getStartIndex(node.Name, startIndexMap, len(rules))

	if currentRule > 0 {
		fmt.Fprintf(os.Stderr, "[OPT1] Node '%s': starting at rule #%d (skipped %d rules)\n", node.Name, currentRule, currentRule)
	}

	for currentRule < len(rules) {
		rule := rules[currentRule]
		if rule != nil && rule.Pattern != nil && rule.Action != nil {
			m, n := rule.Pattern.Matches(node, path)

			if m {
				replacement_node, changed := (*rule.Action).Apply(rule.Pattern, n, node, path)
				if changed {
					node = replacement_node
				}
				currentRule = rule.OnSuccess
				fmt.Fprintln(os.Stderr, "Success with rule", rule.Name, ", moving to rule #", currentRule)
			} else {
				currentRule = rule.OnFailure
				fmt.Fprintln(os.Stderr, "Failure, moving to rule #", currentRule)
			}
		}
	}
	return node
}

// getStartIndex returns the starting rule index for a given node name.
// If the name is in the map, return that index.
// Otherwise, return the wildcard index (stored under "").
// If no wildcard, return the default (skip all rules).
func getStartIndex(nodeName string, startIndexMap map[string]int, defaultIndex int) int {
	if idx, exists := startIndexMap[nodeName]; exists {
		return idx
	}
	// Try wildcard (empty string key).
	if idx, exists := startIndexMap[""]; exists {
		return idx
	}
	// No matching rules at all.
	return defaultIndex
}
