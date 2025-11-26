package bundler

// Instruction represents a single instruction in the function body.
type Instruction struct {
	Name    string            `json:"name"`
	Options map[string]string `json:"options,omitempty"`
}

// FunctionObject represents a compiled function with its metadata and instructions.
type FunctionObject struct {
	NLocals      int           `json:"nlocals"`
	NParams      int           `json:"nparams"`
	Instructions []Instruction `json:"instructions"`
}
