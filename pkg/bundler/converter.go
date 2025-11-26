package bundler

import (
	"fmt"
	"strconv"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// ConvertFnToFunctionObject converts a <fn> node to a FunctionObject.
func ConvertFnToFunctionObject(fnNode *common.Node) (*FunctionObject, error) {
	if fnNode.Name != common.NameFn {
		return nil, fmt.Errorf("expected fn node, got %s", fnNode.Name)
	}

	// Extract nparams and nlocals from the fn node options.
	nparams := 0
	nlocals := 0

	if nparamsStr, ok := fnNode.Options[common.OptionNParams]; ok {
		var err error
		nparams, err = strconv.Atoi(nparamsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid nparams value: %w", err)
		}
	}

	if nlocalsStr, ok := fnNode.Options[common.OptionNLocals]; ok {
		var err error
		nlocals, err = strconv.Atoi(nlocalsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid nlocals value: %w", err)
		}
	}

	// Convert the function body (children) to instructions.
	instructions, err := convertNodesToInstructions(fnNode.Children)
	if err != nil {
		return nil, fmt.Errorf("failed to convert function body: %w", err)
	}

	return &FunctionObject{
		NLocals:      nlocals,
		NParams:      nparams,
		Instructions: instructions,
	}, nil
}

// convertNodesToInstructions converts a list of nodes to instructions.
// This function flattens container nodes (like <seq>) and extracts instruction nodes.
func convertNodesToInstructions(nodes []*common.Node) ([]Instruction, error) {
	instructions := make([]Instruction, 0)

	for _, node := range nodes {
		nodeInstructions, err := collectInstructions(node)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, nodeInstructions...)
	}

	return instructions, nil
}

// collectInstructions recursively collects instructions from a node and its children.
func collectInstructions(node *common.Node) ([]Instruction, error) {
	// Check if this is an instruction node.
	switch node.Name {
	case common.NamePushInt,
		common.NamePushString,
		common.NameStackLength,
		common.NamePopLocal,
		common.NamePushLocal,
		common.NamePushGlobal,
		common.NameReturn,
		common.NameSysCallCounted,
		common.NameCallGlobalCounted:
		// This is a recognized instruction node.
		return []Instruction{{
			Name:    node.Name,
			Options: copyOptions(node.Options),
		}}, nil

	default:
		// For container nodes (like <seq>, <arguments>, etc.), recursively collect instructions from children.
		if len(node.Children) > 0 {
			return convertNodesToInstructions(node.Children)
		}

		// Empty nodes or unrecognized leaf nodes are skipped (e.g., <arguments> with no children).
		return []Instruction{}, nil
	}
}

// copyOptions creates a copy of the options map.
func copyOptions(options map[string]string) map[string]string {
	if len(options) == 0 {
		return nil
	}
	copied := make(map[string]string, len(options))
	for k, v := range options {
		copied[k] = v
	}
	return copied
}
