# Nutmeg Compiler Toolchain

A collection of standalone Go command-line tools that together form the
compiler pipeline for the [Nutmeg](docs/nutmeg-project-structure.md)
programming language. Each stage of the pipeline is a separate binary
that reads a stream of JSON on stdin and writes a transformed stream of
JSON (or another display format) to stdout, so the tools can be composed
with pipes or driven independently for testing and debugging.

This project is under active development; the language, the AST format,
and the tool interfaces should all be considered unstable.

## Pipeline

```
source (.nutmeg)
  │
  ▼
nutmeg-tokenizer   tokenizes source text into a stream of JSON tokens
  │
  ▼
nutmeg-parser      parses tokens into an AST (a `unit` node)
  │
  ▼
nutmeg-check-syntax  validates syntactic rules not enforced by the parser
  │
  ▼
nutmeg-rewriter    normalizes the AST via configurable rewrite rules
  │
  ▼
nutmeg-resolver    resolves identifiers (scope, definition vs. use)
  │
  ▼
nutmeg-codegen     generates bytecode-level instructions from the AST
  │
  ▼
nutmeg-bundler     packages the result into a SQLite bundle
```

`nutmeg-common` integrates the tokenizer, parser, and rewriter into a
single tool, and `nutmeg-compiler` integrates the entire pipeline end to
end. `nutmeg-convert-tree` is a standalone utility for converting a JSON
AST between the supported display formats (XML, JSON, YAML, Mermaid,
Asciitree, Dot), independent of the pipeline itself.

## Tools

| Command | Purpose |
| --- | --- |
| `nutmeg-tokenizer` | Tokenizes Nutmeg source into a stream of JSON tokens. |
| `nutmeg-parser` | Parses a token stream into an AST. |
| `nutmeg-check-syntax` | Validates an AST against syntactic rules not enforced by the parser. |
| `nutmeg-rewriter` | Rewrites/normalizes an AST using a configurable rule set. |
| `nutmeg-resolver` | Resolves identifiers, annotating scope and definition/use. |
| `nutmeg-codegen` | Generates code from a resolved AST. |
| `nutmeg-bundler` | Packages compiled output into a SQLite bundle. |
| `nutmeg-convert-tree` | Converts an AST between output formats (XML, JSON, YAML, Mermaid, Asciitree, Dot). |
| `nutmeg-common` | Integrated tokenizer + parser + rewriter. |
| `nutmeg-compiler` | The full pipeline, end to end. |

Every tool supports `-h`/`--help` and `--version`. Most accept `-f`/`--format`
to select the output format, `--src-path` to annotate the AST with its
origin, and `--trim`/`--indent` to control display formatting.

## AST format

Internally, all stages exchange a single flexible `Node` type (see
[`pkg/common`](pkg/common)), a stripped-down, XML-like tree with ad hoc
attributes rather than a large set of strongly-typed AST nodes. This
keeps the tools loosely coupled and lets new syntax be prototyped without
threading a new type through every stage. See
[docs/byo_parser.md](docs/byo_parser.md) for the rationale.

## Building and testing

The project uses [`just`](https://github.com/casey/just) as its task
runner.

```bash
just build     # build all binaries into ./bin
just install   # go install all binaries
just test      # functests, unit tests, lint, fmt-check, gosec, tidy, build
just unittest  # go test ./...
just functest  # functional tests (functests/, driven by a uv-managed venv)
```

While developing, test a tool with `go run` rather than a stale compiled
binary, e.g.:

```bash
go run ./cmd/nutmeg-tokenizer < snippets/helloworld.nutmeg
```

`just hello` demonstrates the full pipeline by tokenizing, parsing,
rewriting, resolving, generating code for, and bundling
`snippets/helloworld.nutmeg`.

## Documentation

- [docs/nutmeg-project-structure.md](docs/nutmeg-project-structure.md) — project/module layout, imports, and dependency management (design spec).
- [docs/byo_parser.md](docs/byo_parser.md) — the "bring your own parser" design concept.
- [docs/tokenizer/](docs/tokenizer) — tokenizer rules and configuration.
- [docs/parser/](docs/parser) — parser algorithm and examples.
- [docs/rewriter/](docs/rewriter) — rewrite rule configuration.
- [docs/how-to/](docs/how-to) — task-oriented guides.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
