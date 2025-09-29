package common

type PrintOptions struct {
	Format            string `yaml:"option-format,omitempty"`
	Indent            int    `yaml:"option-indent,omitempty"`
	IncludeSpans      bool   `yaml:"option-include-spans,omitempty"`
	Decimal           bool   `yaml:"option-decimal,omitempty"`
	TrimTokenOnOutput int    `yaml:"option-trim-token-on-output,omitempty"`
}
