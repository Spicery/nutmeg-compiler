package codegen

import (
	"fmt"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// CodeGenerator transforms a Nutmeg AST by performing code generation.
// In this initial implementation, it walks the tree finding fn nodes
// and applying stub transformations.
type CodeGenerator struct {
	// Future: Add fields for tracking generated code, labels, etc.
	// Maps local variable serial-numbers to their stack offsets.

}

type FnCodeGenState struct {
	CodeGenerator  *CodeGenerator
	instructions   common.List[*common.Node]
	localOffsets   map[string]int
	maxOffsetSoFar int
	freeTmpVars    []*TemporaryVariable
}

type TemporaryVariable struct {
	Offset int
}

func (tv *TemporaryVariable) OffsetString() string {
	return fmt.Sprintf("%d", tv.Offset)
}

func (fcg *FnCodeGenState) offset(serialNo string) int {
	offset, ok := fcg.localOffsets[serialNo]
	if ok {
		return offset
	}
	n := fcg.maxOffsetSoFar
	fcg.localOffsets[serialNo] = n
	fcg.maxOffsetSoFar += 1
	return n
}

// NewCodeGenerator creates a new code generator instance.
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{}
}

func (cg *CodeGenerator) NewFnCodeGenState() *FnCodeGenState {
	return &FnCodeGenState{
		CodeGenerator:  cg,
		instructions:   common.List[*common.Node]{},
		localOffsets:   make(map[string]int),
		maxOffsetSoFar: 0,
	}
}

func (fcg *FnCodeGenState) NewTemporaryVariable() *TemporaryVariable {
	offset := fcg.maxOffsetSoFar
	fcg.maxOffsetSoFar += 1
	return &TemporaryVariable{Offset: offset}
}

func (fcg *FnCodeGenState) AllocateTemporaryVariable() *TemporaryVariable {
	if len(fcg.freeTmpVars) > 0 {
		tv := fcg.freeTmpVars[len(fcg.freeTmpVars)-1]
		fcg.freeTmpVars = fcg.freeTmpVars[:len(fcg.freeTmpVars)-1]
		return tv
	}
	return fcg.NewTemporaryVariable()
}

func (fcg *FnCodeGenState) FreeTemporaryVariable(tv *TemporaryVariable) {
	fcg.freeTmpVars = append(fcg.freeTmpVars, tv)
}

// Generate performs code generation on the given AST.
// It walks the tree, finds fn nodes, and applies transformations.
func (cg *CodeGenerator) Generate(root *common.Node) error {
	if root != nil && root.Name == common.NameUnit {
		for _, child := range root.Children {
			// We expect <unit> to contain <bind> and <annotation> nodes.
			switch child.Name {
			case common.NameBind:
				err := cg.generateBind(child)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unimplemented top-level node: %s", child.Name)
			}
		}
	} else {
		return fmt.Errorf("expected unit node as root")
	}
	return nil
}

func (cg *CodeGenerator) generateBind(bindNode *common.Node) error {
	if len(bindNode.Children) != 2 {
		return fmt.Errorf("bind node must have exactly 2 children")
	}
	valueNode := bindNode.Children[1]
	if valueNode == nil {
		return fmt.Errorf("bind node has nil value child")
	}
	switch valueNode.Name {
	case common.NameFn:
		return cg.transform(valueNode)
	default:
		bindNode.Options[common.OptionLazy] = common.ValueTrue
		fn_node := &common.Node{
			Name:    common.NameFn,
			Options: map[string]string{},
			Children: []*common.Node{
				{
					Name:     common.NameArguments,
					Options:  map[string]string{},
					Children: []*common.Node{},
				},
				valueNode,
			},
		}
		bindNode.Children[1] = fn_node
		return cg.transform(fn_node)
	}
}

// transform recursively walks the tree and transforms nodes.
func (cg *CodeGenerator) transform(node *common.Node) error {
	if node == nil {
		return nil
	}

	// Check if this is a fn node that needs transformation.
	if node.Name == common.NameFn {
		fcg := cg.NewFnCodeGenState()
		return fcg.rewriteFnNode(node)
	}

	// Recursively transform children.
	for _, child := range node.Children {
		err := cg.transform(child)
		if err != nil {
			return err
		}
	}

	return nil
}

// rewriteFnNode transforms a fn node by processing its arguments and body,
// generating instructions, and updating the node to reflect the generated code.
func (fcg *FnCodeGenState) rewriteFnNode(node *common.Node) error {
	argumentsNode := node.Children[0]
	pdnargs := len(argumentsNode.Children)
	node.Options[common.OptionNParams] = fmt.Sprintf("%d", pdnargs)
	bodyNode := node.Children[1]
	fcg.plantPopArguments(argumentsNode)
	err := fcg.plantInstructions(bodyNode)
	// fmt.Println("INSTRUCTIONS", bodyNode.Name, instructions.Items())
	// fmt.Println("ERROR CHECK", bodyNode.Name, err)
	if err != nil {
		// fmt.Println("CLIMBING")
		return err
	}
	fcg.instructions.Add(&common.Node{Name: common.NameReturn})
	node.ClearChildren()
	node.Children = fcg.instructions.Items()
	node.Options[common.OptionNLocals] = fmt.Sprintf("%d", fcg.maxOffsetSoFar)
	return nil
}

func (fcg *FnCodeGenState) plantPopArguments(argumentsNode *common.Node) {
	pdnargs := len(argumentsNode.Children)
	for i := pdnargs - 1; i >= 0; i-- {
		child := argumentsNode.Children[i]
		offset := fcg.offset(child.Options[common.OptionSerialNo])
		popArgNode := &common.Node{Name: common.NamePopLocal, Options: map[string]string{common.OptionOffset: fmt.Sprintf("%d", offset)}}
		fcg.instructions.Add(popArgNode)
	}
}

func (fcg *FnCodeGenState) plantInstructions(node *common.Node) error {
	switch node.Name {
	case common.NameSysCall:
		err := fcg.plantChildren(node)
		if err != nil {
			return err
		}
		name := node.Options[common.OptionName]
		fcg.plantSysCall(name)
	case common.NameIdentifier:
		scope := node.Options[common.OptionScope]
		switch scope {
		case common.ValueInner, common.ValueOuter:
			fcg.plantPushLocal(node.Options[common.OptionSerialNo])
		case common.ValueGlobal:
			id_name := node.Options[common.OptionName]
			fcg.plantPushGlobal(id_name)
		default:
			return fmt.Errorf("unknown identifier scope: %s", scope)
		}
	case common.NameNumber:
		mantissa_str := node.ToInteger()
		if mantissa_str == nil {
			return fmt.Errorf("non-integer numbers not implemented")
		}
		fcg.plantPushInt(*mantissa_str)
	case common.NameString:
		str_value, ok := node.Options[common.OptionValue]
		if !ok {
			return fmt.Errorf("string node missing string value option")
		}
		fcg.plantPushString(str_value)
	case common.NameApply:
		// fmt.Println("NameApply", len(bodyNode.Children))
		if len(node.Children) == 2 {
			fn := node.Children[0]
			args := node.Children[1]
			tmpvar := fcg.plantStackLength()
			err := fcg.plantChildren(args)
			if err != nil {
				return err
			}
			err = fcg.plantCall(fn, tmpvar)
			if err != nil {
				return err
			}
			fcg.FreeTemporaryVariable(tmpvar)
		} else {
			return fmt.Errorf("apply with != 2 children not implemented")
		}
	default:
		return fmt.Errorf("unimplemented node type: %s", node.Name)
	}
	return nil
}

func (fcg *FnCodeGenState) plantChildren(node *common.Node) error {
	for _, child := range node.Children {
		err := fcg.plantInstructions(child)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fcg *FnCodeGenState) plantCall(node *common.Node, stackLengthTmpVar *TemporaryVariable) error {
	if node == nil {
		return fmt.Errorf("nil node in plantCall")
	}
	switch node.Name {
	case common.NameIdentifier:
		scope := node.Options[common.OptionScope]
		switch scope {
		case common.ValueInner, common.ValueOuter:
			return fmt.Errorf("cannot call local function: %s", node.Options[common.OptionName])
		case common.ValueGlobal:
			id_name := node.Options[common.OptionName]
			fcg.plantCallGlobal(id_name, stackLengthTmpVar)
			return nil
		default:
			return fmt.Errorf("unknown identifier scope in call: %s", scope)
		}
	default:
		return fmt.Errorf("unimplemented call target node: %s", node.Name)
	}
}

func (fcg *FnCodeGenState) plantCallGlobal(id_name string, stackLengthTmpVar *TemporaryVariable) {
	fcg.instructions.Add(&common.Node{
		Name: common.NameCallGlobal,
		Options: map[string]string{
			common.OptionOffset: stackLengthTmpVar.OffsetString(),
			common.OptionName:   id_name,
		},
		Children: []*common.Node{},
	})
}

func (fcg *FnCodeGenState) plantPushInt(value string) {
	pushNumber := &common.Node{Name: common.NamePushInt, Options: map[string]string{common.OptionDecimal: value}, Children: []*common.Node{}}
	fcg.instructions.Add(pushNumber)
}

func (fcg *FnCodeGenState) plantPushString(value string) {
	pushString := &common.Node{Name: common.NamePushString, Options: map[string]string{common.OptionValue: value}, Children: []*common.Node{}}
	fcg.instructions.Add(pushString)
}

func (fcg *FnCodeGenState) plantSysCall(syscallName string) {
	syscallNode := &common.Node{Name: common.NameSysCall, Options: map[string]string{common.OptionName: syscallName}, Children: []*common.Node{}}
	fcg.instructions.Add(syscallNode)
}

func (fcg *FnCodeGenState) plantPushLocal(serialNo string) {
	offset := fcg.offset(serialNo)
	pushLocalNode := &common.Node{Name: common.NamePushLocal, Options: map[string]string{common.OptionOffset: fmt.Sprintf("%d", offset)}, Children: []*common.Node{}}
	fcg.instructions.Add(pushLocalNode)
}

func (fcg *FnCodeGenState) plantPushGlobal(id_name string) {
	pushGlobalNode := &common.Node{Name: common.NamePushGlobal, Options: map[string]string{common.OptionName: id_name}, Children: []*common.Node{}}
	fcg.instructions.Add(pushGlobalNode)
}

func (fcg *FnCodeGenState) plantStackLength() *TemporaryVariable {
	tmpvar := fcg.AllocateTemporaryVariable()
	stackLengthNode := &common.Node{Name: common.NameStackLength, Options: map[string]string{common.OptionOffset: tmpvar.OffsetString()}, Children: []*common.Node{}}
	fcg.instructions.Add(stackLengthNode)
	return tmpvar
}
