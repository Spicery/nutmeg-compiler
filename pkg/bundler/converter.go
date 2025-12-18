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
	case common.NamePushInt:
		value, err := getIntOption(node, common.OptionDecimal)
		if err != nil {
			return nil, fmt.Errorf("push.int missing value: %w", err)
		}
		return []Instruction{NewPushInt(value)}, nil

	case common.NamePushBool:
		valueStr, err := getStringOption(node, common.OptionValue)
		if err != nil {
			return nil, fmt.Errorf("boolean missing value: %w", err)
		}
		return []Instruction{NewPushBool(valueStr)}, nil

	case common.NamePushString:
		value, err := getStringOption(node, common.OptionValue)
		if err != nil {
			return nil, fmt.Errorf("push.string missing value: %w", err)
		}
		return []Instruction{NewPushString(value)}, nil

	case common.NameStackLength:
		offset, err := getIntOption(node, common.OptionOffset)
		if err != nil {
			return nil, fmt.Errorf("stack.length missing offset: %w", err)
		}
		return []Instruction{NewStackLength(offset)}, nil

	case common.NamePopLocal:
		offset, err := getIntOption(node, common.OptionOffset)
		if err != nil {
			return nil, fmt.Errorf("pop.local missing offset: %w", err)
		}
		return []Instruction{NewPopLocal(offset)}, nil

	case common.NamePushLocal:
		offset, err := getIntOption(node, common.OptionOffset)
		if err != nil {
			return nil, fmt.Errorf("push.local missing offset: %w", err)
		}
		return []Instruction{NewPushLocal(offset)}, nil

	case common.NamePushGlobal:
		name, err := getStringOption(node, common.OptionName)
		if err != nil {
			return nil, fmt.Errorf("push.global missing name: %w", err)
		}
		return []Instruction{NewPushGlobal(name)}, nil

	case common.NameDone:
		name, err := getStringOption(node, common.OptionName)
		if err != nil {
			return nil, fmt.Errorf("done missing name")
		}
		offset, err := getIntOption(node, common.OptionOffset)
		if err != nil {
			return nil, fmt.Errorf("done missing offset: %w", err)
		}
		return []Instruction{NewDone(name, offset)}, nil

	case common.NameReturn:
		return []Instruction{NewReturn()}, nil

	case common.NameSysCallCounted:
		name, err := getStringOption(node, common.OptionSysFn)
		if err != nil {
			return nil, fmt.Errorf("syscall.counted missing name: %w", err)
		}
		offset, err := getIntOption(node, common.OptionOffset)
		if err != nil {
			return nil, fmt.Errorf("syscall.counted missing offset: %w", err)
		}
		return []Instruction{NewSyscallCounted(name, offset)}, nil

	case common.NameCallGlobalCounted:
		name, err := getStringOption(node, common.OptionName)
		if err != nil {
			return nil, fmt.Errorf("call.global.counted missing name: %w", err)
		}
		offset, err := getIntOption(node, common.OptionOffset)
		if err != nil {
			return nil, fmt.Errorf("call.global.counted missing offset: %w", err)
		}
		return []Instruction{NewCallGlobalCounted(name, offset)}, nil

	case common.NameSeq, common.NameArguments:
		// For container nodes (like <seq>, <arguments>, etc.), recursively collect instructions from children.
		if len(node.Children) > 0 {
			return convertNodesToInstructions(node.Children)
		}
		return []Instruction{}, nil

	default:
		return nil, fmt.Errorf("unrecognized instruction node: %s", node.Name)
	}
}

// getStringOption extracts a string option from a node.
func getStringOption(node *common.Node, key string) (string, error) {
	value, ok := node.Options[key]
	if !ok {
		return "", fmt.Errorf("missing option: %s", key)
	}
	return value, nil
}

// getIntOption extracts an integer option from a node.
func getIntOption(node *common.Node, key string) (int, error) {
	valueStr, ok := node.Options[key]
	if !ok {
		return 0, fmt.Errorf("missing option: %s", key)
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %w", key, err)
	}
	return value, nil
}
