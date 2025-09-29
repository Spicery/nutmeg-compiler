package rewriter

import (
	"os"

	"gopkg.in/yaml.v3"
)

type RewriteConfig struct {
	Substitutions *Substitutions `yaml:"substitutions,omitempty"`
}

// Substitutions represents the YAML structure for configurable renamings
type Substitutions struct {
	Part       PartSubstitutions       `yaml:"part"`
	Operator   OperatorSubstitutions   `yaml:"operator"`
	Identifier IdentifierSubstitutions `yaml:"identifier"`
}

type PartSubstitutions struct {
	Keyword map[string]string `yaml:"keyword"`
}

type OperatorSubstitutions struct {
	Name map[string]string `yaml:"name"`
}

type IdentifierSubstitutions struct {
	Name map[string]string `yaml:"name"`
}

// LoadSubstitutions loads substitutions from a YAML file
func LoadRewriteConfig(filename string) (*RewriteConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var rewriteConfig RewriteConfig
	err = yaml.Unmarshal(data, &rewriteConfig)
	if err != nil {
		return nil, err
	}

	return &rewriteConfig, nil
}
