# A synthesis of the toolchain: nutmeg-compiler

Our goal is to build a new tool `nutmeg-compiler` that is a synthesis of
the individual tools. It should be based on the structure of the `nutmeg-common`
tool, which is itself a synthesis of `nutmeg-tokenizer`, `nutmeg-parser`, 
`nutmeg-rewriter` and `nutmeg-check-syntax`.

This tool is effectively the same as a pipeline of:

- `nutmeg-tokenizer`
- `nutmeg-parser`
- `nutmeg-check-syntax`
- `nutmeg-rewriter`
- `nutmeg-resolver`
- `nutmeg-codegen` and
- `nutmeg-bundler`

## Options

Command line options should include (in no particular order)

  -bundle string
    	Bundle file path (required)
  -bundle string
    	Bundle file path (required)
  -debug
    	Enable debug output to stderr
  -f string
    	Output format (JSON, XML, etc.) (default "JSON")
  -format string
    	Output format (JSON, XML, etc.) (default "JSON")
  -h	Show help
  -help
    	Show help
  -input string
    	Input file (does NOT default to stdin, used for srcPath)
  -rewrite-rules string
    	YAML file containing rewrite rules (optional)
  -skip-optional
    	Skip optional rewrite passes
  -token-rules string
    	YAML file containing tokenizer rules (optional)
  -version
    	Show version
