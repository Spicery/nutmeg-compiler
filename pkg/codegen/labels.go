package codegen

// LabelType represents the different kinds of labels used in code generation.
type LabelType int

const (
	SimpleLabel LabelType = iota
	ContinueLabel
	ReturnLabel
)

// Label represents a jump target in generated code.
// SimpleLabels correspond to ordinary jump targets.
// ContinueLabels represent drop-through (no jump needed).
// ReturnLabels will be translated into immediate returns.
type Label struct {
	labelType LabelType
	labelText string
}

// NewSimpleLabel creates a new simple label with the given text.
func NewSimpleLabel(text string) Label {
	return Label{labelType: SimpleLabel, labelText: text}
}

// NewContinueLabel creates a new continue label (drop-through).
func NewContinueLabel() Label {
	return Label{labelType: ContinueLabel, labelText: ""}
}

// NewReturnLabel creates a new return label.
func NewReturnLabel() Label {
	return Label{labelType: ReturnLabel, labelText: ""}
}
