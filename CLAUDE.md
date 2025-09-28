# Nutmeg Compiler Toolchain

## Nutmeg Tokenizer - a standalone tokenizer for the Nutmeg project

### Tokens

We are collaborating on the development of a standalone parser for the Nutmeg
programming language, implemented in the Go programming language. Given a stream
of JSON tokens (one per line), the parser outputs a list of nodes in various
formats.

The format of the incoming tokens is described [in this document](https://github.com/Spicery/nutmeg-tokenizer/blob/main/docs/tokens.md).

### Options

The `nutmeg-parser` command consumes tokens on stdin and generates trees to
stdout. The options it has at this time are:

- `-f,--format` which specifies the output format as JSON, XML or other. By
  default XML is assumed.

- `--src-path` which is used to annotate the `unit` with the origin but
  does not affect the parse. Optional.

- `--trim=N` which specifies if the names should be trimmed for display
  purposes. Optional.

- `--indent=N` which specfies the indentation used, for display purposes.
  Optional.

### Processing

The parser digests tokens and generates AST's internally, using the Node 
struct specified in `./pkg/parset/node.go`, which is a flexible monotype
that resembles a highly stripped down XML. (We use this flexible datatype rather
than a collection of strong types for the freedom of ad hoc annotations.)

It generates a stream of ASTs that are wrapped in a `unit` node, representing
the file that is being built.

The incoming tokens are heavily attributed and these attributes drive the
parser, which has relatively little understanding of the AST format. This is
part of the BYO parser concept (see `docs/byo_parser.md`).

INIIALLY the parser will simply process the options, print them out and
quit.


###  Output Format, Happy Path

Assuming the the whole input parses cleanly then a single `unit` is printed
to stdout. The format that is used is configurable. We support:

- Data-oriented formats
    - XML
    - JSON

- Display-oriented formats
    - Mermaid
    - Asciitree
    - Dot
    - YAML

### Output Format, Unhappy Path

In the event that a parse error occurs, the parse emits a problems node with
a shape like this:

```xml
<problems>
    <problem reason="Explanation of the first problem"/>
    <problem reason="Explanation of the second problem"/>
    ...
</problems>
```

In the future this will be expanded to provide more details.

### Programming Guidelines

- Comments should be proper sentences, with correct grammar and punctuation,
  including the use of capitalization and periods.

- Where defensive checks are added, include a comment explaining why they are
  appropriate (not necessary, since defensive checks are not necessary).

- I prefer text files to use new-line as a terminator rather than a separator
  i.e. newlines at the end of non-empty files.

### Test Guidelines

- When testing the behaviour of the binary, always use `go run ./cmd/nutmeg-tokenizer`
  rather than `./nutmeg-tokenizer` directory. This ensures you are always testing
  the latest code rather than an out-of-date compiled binary. (Unless you are 
  deliberately testing an out-of-date binary).
  
### Collaboration Guidelines

When providing technical assistance:

- **Be objective and critical**: Focus on technical correctness over agreeability
- **Challenge assumptions**: If code has clear technical flaws, point them out directly
- **Prioritize correctness**: Don't compromise on proper implementation to avoid disagreement
- **Think through implications**: Consider how users will actually use features in practice
- **Be direct about problems**: If something is wrong or will cause user confusion, say so clearly

The goal is to build robust, well-designed software, not to avoid technical disagreements.

## Nutmeg Parser - a standalone parser for the Nutmeg project

### Tokens

We are collaborating on the development of a standalone parser for the Nutmeg
programming language, implemented in the Go programming language. Given a stream
of JSON tokens (one per line), the parser outputs a list of nodes in various
formats.

The format of the incoming tokens is described [in this document](https://github.com/Spicery/nutmeg-tokenizer/blob/main/docs/tokens.md).

### Options

The `nutmeg-parser` command consumes tokens on stdin and generates trees to
stdout. The options it has at this time are:

- `-f,--format` which specifies the output format as JSON, XML or other. By
  default XML is assumed.

- `--src-path` which is used to annotate the `unit` with the origin but
  does not affect the parse. Optional.

- `--trim=N` which specifies if the names should be trimmed for display
  purposes. Optional.

- `--indent=N` which specfies the indentation used, for display purposes.
  Optional.

### Processing

The parser digests tokens and generates AST's internally, using the Node 
struct specified in `./pkg/parset/node.go`, which is a flexible monotype
that resembles a highly stripped down XML. (We use this flexible datatype rather
than a collection of strong types for the freedom of ad hoc annotations.)

It generates a stream of ASTs that are wrapped in a `unit` node, representing
the file that is being built.

The incoming tokens are heavily attributed and these attributes drive the
parser, which has relatively little understanding of the AST format. This is
part of the BYO parser concept (see `docs/byo_parser.md`).

INIIALLY the parser will simply process the options, print them out and
quit.


###  Output Format, Happy Path

Assuming the the whole input parses cleanly then a single `unit` is printed
to stdout. The format that is used is configurable. We support:

- Data-oriented formats
    - XML
    - JSON

- Display-oriented formats
    - Mermaid
    - Asciitree
    - Dot
    - YAML

### Output Format, Unhappy Path

In the event that a parse error occurs, the parse emits a problems node with
a shape like this:

```xml
<problems>
    <problem reason="Explanation of the first problem"/>
    <problem reason="Explanation of the second problem"/>
    ...
</problems>
```

In the future this will be expanded to provide more details.

### Programming Guidelines

- Comments should be proper sentences, with correct grammar and punctuation,
  including the use of capitalization and periods.

- Where defensive checks are added, include a comment explaining why they are
  appropriate (not necessary, since defensive checks are not necessary).

- I prefer text files to use new-line as a terminator rather than a separator
  i.e. newlines at the end of non-empty files.

### Test Guidelines

- When testing the behaviour of the binary, always use `go run ./cmd/nutmeg-tokenizer`
  rather than `./nutmeg-tokenizer` directory. This ensures you are always testing
  the latest code rather than an out-of-date compiled binary. (Unless you are 
  deliberately testing an out-of-date binary).
  
### Collaboration Guidelines

When providing technical assistance:

- **Be objective and critical**: Focus on technical correctness over agreeability
- **Challenge assumptions**: If code has clear technical flaws, point them out directly
- **Prioritize correctness**: Don't compromise on proper implementation to avoid disagreement
- **Think through implications**: Consider how users will actually use features in practice
- **Be direct about problems**: If something is wrong or will cause user confusion, say so clearly

The goal is to build robust, well-designed software, not to avoid technical disagreements.