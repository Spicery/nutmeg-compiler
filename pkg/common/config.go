package parser

type ConfigurableOptions struct {
	Format string `yaml:"option-format,omitempty"`
	Indent int    `yaml:"option-indent,omitempty"`
	// DefaultLabel      string `yaml:"option-default-label,omitempty"`
	IncludeSpans  bool `yaml:"option-include-spans,omitempty"`
	Decimal       bool `yaml:"option-decimal,omitempty"`
	CheckLiterals bool `yaml:"option-check-literals,omitempty"`
	// UseClassifier     string `yaml:"option-use-classifier,omitempty"`
	TrimTokenOnOutput int `yaml:"option-trim-token-on-output,omitempty"`
}
