package codegen

import (
	"testing"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

func TestPlantQueryInstructions(t *testing.T) {
	tests := []struct {
		name          string
		successLabel  Label
		failureLabel  Label
		expectedInsts []string // Expected instruction names in order
	}{
		{
			name:          "Continue-Continue: erase result",
			successLabel:  NewContinueLabel(),
			failureLabel:  NewContinueLabel(),
			expectedInsts: []string{common.NameStackLength, common.NamePushBool, common.NameCheckBool, common.NameErase},
		},
		{
			name:          "Continue-Simple: if.not",
			successLabel:  NewContinueLabel(),
			failureLabel:  NewSimpleLabel("L0"),
			expectedInsts: []string{common.NameStackLength, common.NamePushBool, common.NameCheckBool, common.NameIfNot},
		},
		{
			name:          "Continue-Return: if.not.return",
			successLabel:  NewContinueLabel(),
			failureLabel:  NewReturnLabel(),
			expectedInsts: []string{common.NameStackLength, common.NamePushBool, common.NameCheckBool, common.NameIfNotReturn},
		},
		{
			name:          "Simple-Continue: if.so",
			successLabel:  NewSimpleLabel("L1"),
			failureLabel:  NewContinueLabel(),
			expectedInsts: []string{common.NameStackLength, common.NamePushBool, common.NameCheckBool, common.NameIfSo},
		},
		{
			name:          "Simple-Simple: if.then.else",
			successLabel:  NewSimpleLabel("L2"),
			failureLabel:  NewSimpleLabel("L3"),
			expectedInsts: []string{common.NameStackLength, common.NamePushBool, common.NameCheckBool, common.NameIfThenElse},
		},
		{
			name:          "Simple-Return: if.so + return",
			successLabel:  NewSimpleLabel("L4"),
			failureLabel:  NewReturnLabel(),
			expectedInsts: []string{common.NameStackLength, common.NamePushBool, common.NameCheckBool, common.NameIfSo, common.NameReturn},
		},
		{
			name:          "Return-Continue: if.so.return",
			successLabel:  NewReturnLabel(),
			failureLabel:  NewContinueLabel(),
			expectedInsts: []string{common.NameStackLength, common.NamePushBool, common.NameCheckBool, common.NameIfSoReturn},
		},
		{
			name:          "Return-Simple: if.not + return",
			successLabel:  NewReturnLabel(),
			failureLabel:  NewSimpleLabel("L5"),
			expectedInsts: []string{common.NameStackLength, common.NamePushBool, common.NameCheckBool, common.NameIfNot, common.NameReturn},
		},
		{
			name:          "Return-Return: unconditional return",
			successLabel:  NewReturnLabel(),
			failureLabel:  NewReturnLabel(),
			expectedInsts: []string{common.NameStackLength, common.NamePushBool, common.NameCheckBool, common.NameReturn},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := NewCodeGenerator()
			fcg := cg.NewFnCodeGenState()

			// Create a simple boolean node to compile
			boolNode := &common.Node{
				Name: common.NameBoolean,
				Options: map[string]string{
					common.OptionValue: common.ValueTrue,
				},
			}

			err := fcg.plantQueryInstructions(boolNode, tt.successLabel, tt.failureLabel)
			if err != nil {
				t.Fatalf("plantQueryInstructions failed: %v", err)
			}

			// Verify the generated instructions
			if len(fcg.instructions.Items()) != len(tt.expectedInsts) {
				t.Errorf("Expected %d instructions, got %d", len(tt.expectedInsts), len(fcg.instructions.Items()))
				t.Logf("Generated instructions:")
				for i, inst := range fcg.instructions.Items() {
					t.Logf("  [%d] %s", i, inst.Name)
				}
			}

			for i, expectedName := range tt.expectedInsts {
				if i >= len(fcg.instructions.Items()) {
					t.Errorf("Missing instruction at index %d: expected %s", i, expectedName)
					continue
				}
				actualName := fcg.instructions.Items()[i].Name
				if actualName != expectedName {
					t.Errorf("Instruction %d: expected %s, got %s", i, expectedName, actualName)
				}
			}
		})
	}
}

func TestPlantQueryInstructionsLabelValues(t *testing.T) {
	cg := NewCodeGenerator()
	fcg := cg.NewFnCodeGenState()

	successLabel := NewSimpleLabel("SUCCESS")
	failureLabel := NewSimpleLabel("FAILURE")

	boolNode := &common.Node{
		Name: common.NameBoolean,
		Options: map[string]string{
			common.OptionValue: common.ValueTrue,
		},
	}

	err := fcg.plantQueryInstructions(boolNode, successLabel, failureLabel)
	if err != nil {
		t.Fatalf("plantQueryInstructions failed: %v", err)
	}

	// Find the if.then.else instruction
	var ifThenElseNode *common.Node
	for _, inst := range fcg.instructions.Items() {
		if inst.Name == common.NameIfThenElse {
			ifThenElseNode = inst
			break
		}
	}

	if ifThenElseNode == nil {
		t.Fatal("Expected if.then.else instruction not found")
	}

	// Verify the label values
	thenValue := ifThenElseNode.Options[common.OptionName]
	elseValue := ifThenElseNode.Options[common.OptionValue]

	if thenValue != "SUCCESS" {
		t.Errorf("Expected then label 'SUCCESS', got '%s'", thenValue)
	}
	if elseValue != "FAILURE" {
		t.Errorf("Expected else label 'FAILURE', got '%s'", elseValue)
	}
}

func TestPlantConditionalJumpHelpers(t *testing.T) {
	cg := NewCodeGenerator()
	fcg := cg.NewFnCodeGenState()

	// Test plantIfNot
	fcg.plantIfNot(NewSimpleLabel("FAIL"))
	if len(fcg.instructions.Items()) != 1 {
		t.Fatalf("Expected 1 instruction after plantIfNot, got %d", len(fcg.instructions.Items()))
	}
	if fcg.instructions.Items()[0].Name != common.NameIfNot {
		t.Errorf("Expected if.not instruction, got %s", fcg.instructions.Items()[0].Name)
	}
	if fcg.instructions.Items()[0].Options[common.OptionValue] != "FAIL" {
		t.Errorf("Expected label 'FAIL', got '%s'", fcg.instructions.Items()[0].Options[common.OptionValue])
	}

	// Reset and test plantIfSo
	fcg = cg.NewFnCodeGenState()
	fcg.plantIfSo(NewSimpleLabel("SUCCESS"))
	if len(fcg.instructions.Items()) != 1 {
		t.Fatalf("Expected 1 instruction after plantIfSo, got %d", len(fcg.instructions.Items()))
	}
	if fcg.instructions.Items()[0].Name != common.NameIfSo {
		t.Errorf("Expected if.so instruction, got %s", fcg.instructions.Items()[0].Name)
	}
	if fcg.instructions.Items()[0].Options[common.OptionValue] != "SUCCESS" {
		t.Errorf("Expected label 'SUCCESS', got '%s'", fcg.instructions.Items()[0].Options[common.OptionValue])
	}

	// Reset and test plantIfSoReturn
	fcg = cg.NewFnCodeGenState()
	fcg.plantIfSoReturn()
	if len(fcg.instructions.Items()) != 1 {
		t.Fatalf("Expected 1 instruction after plantIfSoReturn, got %d", len(fcg.instructions.Items()))
	}
	if fcg.instructions.Items()[0].Name != common.NameIfSoReturn {
		t.Errorf("Expected if.so.return instruction, got %s", fcg.instructions.Items()[0].Name)
	}

	// Reset and test plantIfNotReturn
	fcg = cg.NewFnCodeGenState()
	fcg.plantIfNotReturn()
	if len(fcg.instructions.Items()) != 1 {
		t.Fatalf("Expected 1 instruction after plantIfNotReturn, got %d", len(fcg.instructions.Items()))
	}
	if fcg.instructions.Items()[0].Name != common.NameIfNotReturn {
		t.Errorf("Expected if.not.return instruction, got %s", fcg.instructions.Items()[0].Name)
	}

	// Reset and test plantIfThenElse
	fcg = cg.NewFnCodeGenState()
	fcg.plantIfThenElse(NewSimpleLabel("THEN"), NewSimpleLabel("ELSE"))
	if len(fcg.instructions.Items()) != 1 {
		t.Fatalf("Expected 1 instruction after plantIfThenElse, got %d", len(fcg.instructions.Items()))
	}
	inst := fcg.instructions.Items()[0]
	if inst.Name != common.NameIfThenElse {
		t.Errorf("Expected if.then.else instruction, got %s", inst.Name)
	}
	if inst.Options[common.OptionName] != "THEN" {
		t.Errorf("Expected then label 'THEN', got '%s'", inst.Options[common.OptionName])
	}
	if inst.Options[common.OptionValue] != "ELSE" {
		t.Errorf("Expected else label 'ELSE', got '%s'", inst.Options[common.OptionValue])
	}
}
