package bundler

// Instruction represents a single instruction in the function body.
// This uses an adjacently tagged union format with a Type field and type-specific fields.
type Instruction struct {
	Type string `json:"type"`

	// Fields for different instruction types.
	// Only the relevant fields for each type will be populated.

	// PushInt, PopLocal, PushLocal.
	Index *int `json:"index,omitempty"`

	// PushString, PushGlobal.
	Value *string `json:"value,omitempty"`

	// SyscallCounted, CallGlobalCounted.
	Name *string `json:"name,omitempty"`
}

// FunctionObject represents a compiled function with its metadata and instructions.
type FunctionObject struct {
	NLocals      int           `json:"nlocals"`
	NParams      int           `json:"nparams"`
	Instructions []Instruction `json:"instructions"`
}

// Constructor functions for creating instructions.

// NewPushInt creates a push.int instruction.
func NewPushInt(value int) Instruction {
	return Instruction{Type: "push.int", Index: &value}
}

// NewPushString creates a push.string instruction.
func NewPushString(value string) Instruction {
	return Instruction{Type: "push.string", Value: &value}
}

// NewStackLength creates a stack.length instruction.
func NewStackLength(offset int) Instruction {
	return Instruction{Type: "stack.length", Index: &offset}
}

// NewPopLocal creates a pop.local instruction.
func NewPopLocal(offset int) Instruction {
	return Instruction{Type: "pop.local", Index: &offset}
}

// NewPushLocal creates a push.local instruction.
func NewPushLocal(offset int) Instruction {
	return Instruction{Type: "push.local", Index: &offset}
}

// NewPushGlobal creates a push.global instruction.
func NewPushGlobal(name string) Instruction {
	return Instruction{Type: "push.global", Value: &name}
}

// NewReturn creates a return instruction.
func NewReturn() Instruction {
	return Instruction{Type: "return"}
}

// NewSyscallCounted creates a syscall.counted instruction.
func NewSyscallCounted(name string, nargs int) Instruction {
	return Instruction{Type: "syscall.counted", Name: &name, Index: &nargs}
}

// NewCallGlobalCounted creates a call.global.counted instruction.
func NewCallGlobalCounted(name string, nargs int) Instruction {
	return Instruction{Type: "call.global.counted", Name: &name, Index: &nargs}
}
