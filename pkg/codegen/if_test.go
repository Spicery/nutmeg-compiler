package codegen

import (
	"testing"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

func TestPlantIfThenElse(t *testing.T) {
	cg := NewCodeGenerator()
	fcg := cg.NewFnCodeGenState()

	// Create an if-then-else node:
	// if true then 1 else 2
	ifNode := &common.Node{
		Name: common.NameIf,
		Children: []*common.Node{
			// Predicate: true
			{
				Name: common.NameBoolean,
				Options: map[string]string{
					common.OptionValue: common.ValueTrue,
				},
			},
			// Then branch: 1
			{
				Name: common.NameNumber,
				Options: map[string]string{
					common.OptionMantissa: "1",
					common.OptionFraction: "",
					common.OptionExponent: "0",
					common.OptionBase:     "10",
				},
			},
			// Else branch: 2
			{
				Name: common.NameNumber,
				Options: map[string]string{
					common.OptionMantissa: "2",
					common.OptionFraction: "",
					common.OptionExponent: "0",
					common.OptionBase:     "10",
				},
			},
		},
	}

	err := fcg.plantInstructions(ifNode)
	if err != nil {
		t.Fatalf("plantInstructions failed: %v", err)
	}

	// Expected instruction sequence:
	// stack.length (for predicate)
	// push.bool (true)
	// check.bool
	// if.not (to else label)
	// push.int (1 - then branch)
	// goto (to end label)
	// label (else label)
	// push.int (2 - else branch)
	// label (end label)

	expectedInsts := []string{
		common.NameStackLength,
		common.NamePushBool,
		common.NameCheckBool,
		common.NameIfNot,
		common.NamePushInt,
		common.NameGoto,
		common.NameLabel,
		common.NamePushInt,
		common.NameLabel,
	}

	insts := fcg.instructions.Items()
	if len(insts) != len(expectedInsts) {
		t.Errorf("Expected %d instructions, got %d", len(expectedInsts), len(insts))
		t.Logf("Generated instructions:")
		for i, inst := range insts {
			t.Logf("  [%d] %s", i, inst.Name)
		}
		return
	}

	for i, expectedName := range expectedInsts {
		if insts[i].Name != expectedName {
			t.Errorf("Instruction %d: expected %s, got %s", i, expectedName, insts[i].Name)
		}
	}

	// Verify label consistency
	ifNotInst := insts[3] // if.not instruction
	elseLabel := ifNotInst.Options[common.OptionValue]

	gotoInst := insts[5] // goto instruction
	endLabel := gotoInst.Options[common.OptionValue]

	elseLabelInst := insts[6] // first label instruction
	elseLabelValue := elseLabelInst.Options[common.OptionValue]

	endLabelInst := insts[8] // second label instruction
	endLabelValue := endLabelInst.Options[common.OptionValue]

	if elseLabel != elseLabelValue {
		t.Errorf("if.not jumps to '%s' but else label is '%s'", elseLabel, elseLabelValue)
	}

	if endLabel != endLabelValue {
		t.Errorf("goto jumps to '%s' but end label is '%s'", endLabel, endLabelValue)
	}

	// Verify labels are different
	if elseLabel == endLabel {
		t.Errorf("else label and end label should be different, both are '%s'", elseLabel)
	}
}

func TestPlantIfThenElseInvalidChildren(t *testing.T) {
	cg := NewCodeGenerator()
	fcg := cg.NewFnCodeGenState()

	// Test with wrong number of children
	ifNode := &common.Node{
		Name: common.NameIf,
		Children: []*common.Node{
			{Name: common.NameBoolean, Options: map[string]string{common.OptionValue: common.ValueTrue}},
			{Name: common.NameNumber, Options: map[string]string{common.OptionMantissa: "1", common.OptionFraction: "", common.OptionExponent: "0", common.OptionBase: "10"}},
		},
	}

	err := fcg.plantInstructions(ifNode)
	if err == nil {
		t.Error("Expected error for if node with 2 children, got nil")
	}
}

func TestPlantLabelAndGoto(t *testing.T) {
	cg := NewCodeGenerator()
	fcg := cg.NewFnCodeGenState()

	label := NewSimpleLabel("TEST_LABEL")

	// Plant goto
	fcg.plantGoto(label)

	// Plant label
	fcg.plantLabel(label)

	insts := fcg.instructions.Items()
	if len(insts) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(insts))
	}

	// Check goto instruction
	if insts[0].Name != common.NameGoto {
		t.Errorf("Expected first instruction to be goto, got %s", insts[0].Name)
	}
	if insts[0].Options[common.OptionValue] != "TEST_LABEL" {
		t.Errorf("Expected goto to 'TEST_LABEL', got '%s'", insts[0].Options[common.OptionValue])
	}

	// Check label instruction
	if insts[1].Name != common.NameLabel {
		t.Errorf("Expected second instruction to be label, got %s", insts[1].Name)
	}
	if insts[1].Options[common.OptionValue] != "TEST_LABEL" {
		t.Errorf("Expected label 'TEST_LABEL', got '%s'", insts[1].Options[common.OptionValue])
	}
}
