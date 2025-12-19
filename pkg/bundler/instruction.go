package bundler

// Instruction represents a single instruction in the function body.
// This uses an adjacently tagged union format with a Type field and type-specific fields.
type Instruction struct {
	Type string `json:"type"`

	// Fields for different instruction types.
	// Only the relevant fields for each type will be populated.

	// PushInt, PopLocal, PushLocal.
	Index *int `json:"index,omitempty"`

	IntValue *int `json:"ivalue,omitempty"`

	// PushString, PushGlobal.
	StrValue *string `json:"value,omitempty"`

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
	return Instruction{Type: "push.int", IntValue: &value}
}

func NewPushBool(value string) Instruction {
	return Instruction{Type: "push.bool", StrValue: &value}
}

// NewPushString creates a push.string instruction.
func NewPushString(value string) Instruction {
	return Instruction{Type: "push.string", StrValue: &value}
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
	return Instruction{Type: "push.global", Name: &name}
}

func NewDone(name string, offset int) Instruction {
	return Instruction{Type: "done", Name: &name, Index: &offset}
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

// NewErase creates an erase instruction.
func NewErase() Instruction {
	return Instruction{Type: "erase"}
}

// NewCheckBool creates a check.bool instruction.
func NewCheckBool(offset int) Instruction {
	return Instruction{Type: "check.bool", Index: &offset}
}

// NewLabel creates a label instruction.
func NewLabel(label string) Instruction {
	return Instruction{Type: "label", StrValue: &label}
}

// NewGoto creates a goto instruction.
func NewGoto(label string) Instruction {
	return Instruction{Type: "goto", StrValue: &label}
}

// NewIfNot creates an if.not instruction.
func NewIfNot(label string) Instruction {
	return Instruction{Type: "if.not", StrValue: &label}
}

// NewIfSo creates an if.so instruction.
func NewIfSo(label string) Instruction {
	return Instruction{Type: "if.so", StrValue: &label}
}

// NewIfNotReturn creates an if.not.return instruction.
func NewIfNotReturn() Instruction {
	return Instruction{Type: "if.not.return"}
}

// NewIfSoReturn creates an if.so.return instruction.
func NewIfSoReturn() Instruction {
	return Instruction{Type: "if.so.return"}
}

// NewIfThenElse creates an if.then.else instruction.
func NewIfThenElse(thenLabel string, elseLabel string) Instruction {
	return Instruction{Type: "if.then.else", Name: &thenLabel, StrValue: &elseLabel}
}

// NewInProgress creates an in.progress instruction.
func NewInProgress(name string) Instruction {
	return Instruction{Type: "in.progress", Name: &name}
}
