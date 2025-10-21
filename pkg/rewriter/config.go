package rewriter

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RewriteConfig represents the top-level configuration structure
type RewriteConfig struct {
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
	Passes      []Pass `yaml:"passes"`
}

// Pass represents a single named pass containing rewrite rules
type Pass struct {
	Name       string        `yaml:"name"`
	SinglePass bool          `yaml:"singlePass,omitempty"`
	Downwards  []RewriteRule `yaml:"downwards,omitempty"`
	Upwards    []RewriteRule `yaml:"upwards,omitempty"`
}

// RewriteRule represents a single rewrite rule with match conditions and actions
type RewriteRule struct {
	Name            string       `yaml:"name,omitempty"`
	Match           Pattern      `yaml:"match"`
	Action          ActionConfig `yaml:"action"`
	OnSuccess       *string      `yaml:"onSuccess,omitempty"`
	OnFailure       *string      `yaml:"onFailure,omitempty"`
	RepeatOnSuccess bool         `yaml:"repeatOnSuccess,omitempty"`
	BreakOnSuccess  bool         `yaml:"breakOnSuccess,omitempty"`
	BreakOnFailure  bool         `yaml:"breakOnFailure,omitempty"`
}

// ActionConfig defines what action to take when a match is found
// This is used for YAML unmarshaling and then converted to concrete Action implementations
type ActionConfig struct {
	ReplaceValue       *ReplaceValueConfig `yaml:"replaceValue,omitempty"`
	ReplaceName        *ReplaceNameConfig  `yaml:"replaceName,omitempty"`
	ReplaceByChild     *int                `yaml:"replaceByChild,omitempty"`
	InlineChild        bool                `yaml:"inlineChild,omitempty"`
	RotateOption       *RotateOptionConfig `yaml:"rotateOption,omitempty"`
	RemoveOption       *RemoveOptionConfig `yaml:"removeOption,omitempty"`
	Sequence           []ActionConfig      `yaml:"sequence,omitempty"`
	ChildAction        *ActionConfig       `yaml:"childAction,omitempty"`
	RemoveChild        bool                `yaml:"removeChild,omitempty"`
	MergeChildWithNext *bool               `yaml:"mergeChildWithNext,omitempty"`
	NewNodeChild       *NewNodeChildConfig `yaml:"newNodeChild,omitempty"`
	PermuteChildren    []int               `yaml:"permuteChildren,omitempty"`
	Continue           bool                `yaml:"continue,omitempty"`
	Fail               *string             `yaml:"fail,omitempty"`
	Assert             *Pattern            `yaml:"assert,omitempty"`
}

type NewNodeChildConfig struct {
	Name   string  `yaml:"name"`
	Key    *string `yaml:"key,omitempty"`
	Value  *string `yaml:"value,omitempty"`
	Offset int     `yaml:"offset,omitempty"`
	Length *int    `yaml:"length,omitempty"`
}

type RemoveOptionConfig struct {
	Key string `yaml:"key"`
}

type RotateOptionConfig struct {
	Key     string   `yaml:"key"`
	Values  []string `yaml:"values"`
	Initial string   `yaml:"initial,omitempty"`
}

type ReplaceNameConfig struct {
	With   *string `yaml:"with"`
	Source string  `yaml:"src,omitempty"`
	From   *string `yaml:"from,omitempty"`
}

// ReplaceValueConfig is the YAML configuration for replace value actions
type ReplaceValueConfig struct {
	Key    *string `yaml:"key,omitempty"`
	With   *string `yaml:"with,omitempty"`
	Source string  `yaml:"src,omitempty"`
	From   *string `yaml:"from,omitempty"`
}

func (ac ActionConfig) Validate() error {
	// Options are mutually exclusive; only one should be set.
	count := 0
	if ac.ReplaceValue != nil {
		if ac.ReplaceValue.Key != nil {
			if ac.ReplaceValue.With != nil || (ac.ReplaceValue.From != nil && ac.ReplaceValue.Source != "") {
				count++
			} else {
				return fmt.Errorf("invalid ReplaceValueConfig: either 'with' or both 'from' and 'src' must be set")
			}
		} else {
			return fmt.Errorf("invalid ReplaceValueConfig: 'key' must be set")
		}
	}
	if ac.ReplaceName != nil {
		count++
	}
	if ac.ReplaceByChild != nil {
		count++
	}
	if ac.InlineChild {
		count++
	}
	// if ac.Repeat != nil {
	// 	count++
	// }
	if ac.RotateOption != nil {
		count++
	}
	if ac.RemoveOption != nil {
		count++
	}
	if len(ac.Sequence) > 0 {
		count++
	}
	if ac.ChildAction != nil {
		count++
	}
	if ac.MergeChildWithNext != nil {
		count++
	}
	if ac.NewNodeChild != nil {
		count++
	}
	if ac.RemoveChild {
		count++
	}
	if len(ac.PermuteChildren) > 0 {
		// fmt.Println("PermuteChildren count:", len(ac.PermuteChildren))
		count++
	}
	if ac.Continue {
		count++
	}
	if ac.Fail != nil {
		count++
	}
	if ac.Assert != nil {
		count++
	}
	if count == 0 {
		return fmt.Errorf("no action specified in ActionConfig: %+v", ac)
	}
	if count > 1 {
		return fmt.Errorf("multiple actions specified in ActionConfig; only one allowed: %+v", ac)
	}
	return nil
}

// ToAction converts an ActionConfig to a concrete Action implementation
func (ac ActionConfig) ToAction() (Action, error) {
	// fmt.Println("ac:", ac)
	// Validate the action config first
	if err := ac.Validate(); err != nil {
		return nil, err
	}
	// Determine which action is specified and create the corresponding Action
	if ac.ReplaceValue != nil {
		if ac.ReplaceValue.Key == nil {
			return nil, fmt.Errorf("invalid ReplaceValueConfig: both 'key' must be set")
		}
		if ac.ReplaceValue.With != nil {
			return &ReplaceValueAction{
				With: *ac.ReplaceValue.With,
				Key:  *ac.ReplaceValue.Key,
			}, nil
		} else if ac.ReplaceValue.From != nil && ac.ReplaceValue.Source != "" {
			return &ReplaceValueFromAction{
				Key:    *ac.ReplaceValue.Key,
				From:   *ac.ReplaceValue.From,
				Source: ac.ReplaceValue.Source,
			}, nil
		} else {
			return nil, fmt.Errorf("invalid ReplaceValueConfig: either 'with' or both 'from' and 'src' must be set")
		}
	}
	if ac.ReplaceName != nil {
		if ac.ReplaceName.With != nil {
			return &ReplaceNameWithAction{With: *ac.ReplaceName.With}, nil
		}
		if ac.ReplaceName.From != nil && ac.ReplaceName.Source != "" {
			return &ReplaceNameFromAction{From: *ac.ReplaceName.From, Source: ac.ReplaceName.Source}, nil
		}
		// Return nil if no valid action is found
		with := "nil"
		if ac.ReplaceName.With != nil {
			with = *ac.ReplaceName.With
		}
		from := "nil"
		if ac.ReplaceName.From != nil {
			from = *ac.ReplaceName.From
		}
		src := "nil"
		if ac.ReplaceName.Source != "" {
			src = ac.ReplaceName.Source
		}
		return nil, fmt.Errorf("no valid action found in ReplaceNameConfig: with %s, from %s, src %s", with, from, src)
	}
	if ac.ReplaceByChild != nil {
		return &ReplaceByChildAction{ChildIndex: *ac.ReplaceByChild}, nil
	}
	if ac.InlineChild {
		return &InlineChildAction{}, nil
	}
	if ac.RemoveChild {
		return &RemoveChildAction{}, nil
	}
	if ac.RotateOption != nil {
		if ac.RotateOption.Key != "" && len(ac.RotateOption.Values) >= 2 {
			initial := ac.RotateOption.Values[0]
			return &RotateOptionAction{
				Key:     ac.RotateOption.Key,
				Values:  ac.RotateOption.Values,
				Initial: initial,
			}, nil
		}
		return nil, fmt.Errorf("invalid RotateOptionConfig: key must be set and at least two values are required")
	}
	if ac.RemoveOption != nil {
		if ac.RemoveOption.Key != "" {
			return &RemoveOptionAction{
				Key: ac.RemoveOption.Key,
			}, nil
		}
		return nil, fmt.Errorf("invalid RemoveOptionConfig: key must be set")
	}
	if len(ac.Sequence) > 0 {
		actions := []Action{}
		for i, subAc := range ac.Sequence {
			subAction, err := subAc.ToAction()
			if err != nil {
				return nil, fmt.Errorf("error in nested sequence action, position %d: %w", i, err)
			}
			actions = append(actions, subAction)
		}
		return &SequenceAction{Actions: actions}, nil
	}
	if ac.ChildAction != nil {
		childAction, err := ac.ChildAction.ToAction()
		if err != nil {
			return nil, fmt.Errorf("error in nested child action: %w", err)
		}
		return &ChildAction{Action: childAction}, nil
	}
	if ac.MergeChildWithNext != nil {
		return &MergeChildWithNextAction{NextTakesPriority: *ac.MergeChildWithNext}, nil
	}
	if ac.NewNodeChild != nil {
		return &NewNodeChildAction{Name: ac.NewNodeChild.Name, Key: ac.NewNodeChild.Key, Value: ac.NewNodeChild.Value, Offset: ac.NewNodeChild.Offset, Length: ac.NewNodeChild.Length}, nil
	}
	if len(ac.PermuteChildren) > 0 {
		return &PermuteChildrenAction{NewOrder: ac.PermuteChildren}, nil
	}
	if ac.Continue {
		return &NullAction{}, nil
	}
	if ac.Fail != nil {
		return &FailAction{Message: *ac.Fail}, nil
	}
	if ac.Assert != nil {
		return &AssertAction{AssertPattern: ac.Assert}, nil
	}
	// Future actions can be handled here
	return nil, fmt.Errorf("no valid action found in ActionConfig: %+v", ac)
}

// LoadSubstitutions loads substitutions from a YAML file
func LoadRewriteConfig(filename string) (*RewriteConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var rewriteConfig RewriteConfig
	err = yaml.Unmarshal(data, &rewriteConfig)
	if err != nil {
		return nil, err
	}

	return &rewriteConfig, nil
}

// LoadRewriteConfigFromString loads a RewriteConfig from a YAML string.
func LoadRewriteConfigFromString(yamlContent string) (*RewriteConfig, error) {
	var rewriteConfig RewriteConfig
	err := yaml.Unmarshal([]byte(yamlContent), &rewriteConfig)
	if err != nil {
		return nil, err
	}

	return &rewriteConfig, nil
}
