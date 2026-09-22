# Change Log for Nutmeg Compiler Toolchain

Following the style in https://keepachangelog.com/en/1.0.0/

## [Unreleased]

No tagged release has been made yet; this section tracks the toolchain
as it stands during initial development.

### Added

- `nutmeg-tokenizer`: standalone tokenizer for Nutmeg source, emitting a
  stream of JSON tokens. Configurable via a YAML token-rules file. See
  [docs/tokenizer/](docs/tokenizer).
- `nutmeg-parser`: standalone parser that consumes a token stream and
  produces an AST (a `unit` node), using the "bring your own parser"
  design described in [docs/byo_parser.md](docs/byo_parser.md).
- `nutmeg-check-syntax`: validates an AST against syntactic rules not
  enforced by the parser itself.
- `nutmeg-rewriter`: normalizes an AST via a configurable set of rewrite
  rules (e.g. `ifnot`/`elseifnot` desugaring, `let` handling, `matches`
  patterns). Ships with default rules generated from
  `configs/rewrite.yaml`.
- `nutmeg-resolver`: resolves identifiers in an AST, annotating each with
  a unique ID, scope (global/outer/inner), and definition-vs-use. Lifts
  closures to top-level functions via partial application.
- `nutmeg-codegen`: generates code from a resolved AST.
- `nutmeg-bundler`: packages compiled output into a SQLite bundle
  (`.bundle` file), with an optional `--migrate` mode for schema
  migrations.
- `nutmeg-convert-tree`: standalone converter between AST display formats
  (JSON, XML, YAML, Mermaid, Asciitree, Dot), independent of the
  pipeline.
- `nutmeg-common`: integrates the tokenizer, parser, and rewriter into a
  single tool.
- `nutmeg-compiler`: integrates the full pipeline — tokenization,
  parsing, syntax checking, rewriting, resolution, code generation, and
  bundling — into a single tool.
- Shared `Node` type (`pkg/common`) used as the common AST representation
  across all pipeline stages, in place of a large family of strongly
  typed AST nodes.
- Functional test suite under `functests/`, driven by a `uv`-managed
  Python harness (`functests/functest.py`).
- `Justfile` with `build`, `install`, `test`, `unittest`, `functest`,
  `lint`, `gosec`, `fmt`/`fmt-check`, and `rules` (regenerate default
  rewrite rules) recipes.
