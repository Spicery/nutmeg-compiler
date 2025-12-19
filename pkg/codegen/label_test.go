package codegen

import (
	"testing"
)

func TestLabelTypes(t *testing.T) {
	// Test creating different label types
	simpleLabel := NewSimpleLabel("test")
	if simpleLabel.labelType != SimpleLabel {
		t.Errorf("Expected SimpleLabel type, got %v", simpleLabel.labelType)
	}
	if simpleLabel.labelText != "test" {
		t.Errorf("Expected label text 'test', got %s", simpleLabel.labelText)
	}

	continueLabel := NewContinueLabel()
	if continueLabel.labelType != ContinueLabel {
		t.Errorf("Expected ContinueLabel type, got %v", continueLabel.labelType)
	}

	returnLabel := NewReturnLabel()
	if returnLabel.labelType != ReturnLabel {
		t.Errorf("Expected ReturnLabel type, got %v", returnLabel.labelType)
	}
}

func TestLabelAllocation(t *testing.T) {
	cg := NewCodeGenerator()
	fcg := cg.NewFnCodeGenState()

	// Test label allocation with auto-incrementing counter
	label1 := fcg.AllocateLabel()
	if label1.labelText != "L0" {
		t.Errorf("Expected first label to be 'L0', got %s", label1.labelText)
	}
	if label1.labelType != SimpleLabel {
		t.Errorf("Expected SimpleLabel type, got %v", label1.labelType)
	}

	label2 := fcg.AllocateLabel()
	if label2.labelText != "L1" {
		t.Errorf("Expected second label to be 'L1', got %s", label2.labelText)
	}

	label3 := fcg.AllocateLabel()
	if label3.labelText != "L2" {
		t.Errorf("Expected third label to be 'L2', got %s", label3.labelText)
	}

	// Verify counter incremented correctly
	if fcg.labelCounter != 3 {
		t.Errorf("Expected label counter to be 3, got %d", fcg.labelCounter)
	}
}

func TestLabelAllocationIndependence(t *testing.T) {
	cg := NewCodeGenerator()

	// Test that different FnCodeGenStates have independent label counters
	fcg1 := cg.NewFnCodeGenState()
	fcg2 := cg.NewFnCodeGenState()

	label1a := fcg1.AllocateLabel()
	label2a := fcg2.AllocateLabel()
	label1b := fcg1.AllocateLabel()

	if label1a.labelText != "L0" {
		t.Errorf("Expected fcg1 first label to be 'L0', got %s", label1a.labelText)
	}
	if label2a.labelText != "L0" {
		t.Errorf("Expected fcg2 first label to be 'L0', got %s", label2a.labelText)
	}
	if label1b.labelText != "L1" {
		t.Errorf("Expected fcg1 second label to be 'L1', got %s", label1b.labelText)
	}
}
