package common

import (
	"encoding/json"
	"io"
)

func PrintASTJSON(root *Node, indentDelta string, output io.Writer, options *PrintOptions) {
	// Ignore indentDelta and options parameters, just output simple JSON
	encoder := json.NewEncoder(output)
	encoder.Encode(root)
}

func ReadASTJSON(input io.Reader) (*Node, error) {
	var root Node
	decoder := json.NewDecoder(input)
	err := decoder.Decode(&root)
	if err != nil {
		return nil, err
	}
	return &root, nil
}
