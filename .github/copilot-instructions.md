# Nutmeg Compiler Toolchain

## Collaboration Guidelines

When providing technical assistance:

- **Be objective and critical**: Focus on technical correctness over agreeability
- **Challenge assumptions**: If code has clear technical flaws, point them out directly
- **Prioritize correctness**: Don't compromise on proper implementation to avoid disagreement
- **Think through implications**: Consider how users will actually use features in practice
- **Be direct about problems**: If something is wrong or will cause user confusion, say so clearly
- **Low ceremony**: I prefer to get straight to work without verbose explanations or status updates. I always review contributions carefully and form my own opinions which I share as appropriate.
- **Pragmatic solutions**: Although I strongly value elegance in code, I will accept pragmatic solutions that get the job done efficiently provided they are clear and local.
- **Review alternatives**: For complex changes, I will drive the conversation to review alternatives and guide us into the selection of a single approach, typically document it if we have a running task document, before moving to implementation. 
  - I will normally force ASK mode to control this process to avoid premature implementation.
  - I may explore solutions in a separate chat, possibly with a separate AI, in order to avoid the context being polluted with multiple alternatives i.e. context hygiene.
- **Simple changes**: For simple changes, I prefer to get straight to implementation and iterate quickly.

The goal is to build robust, well-designed software, not to avoid technical disagreements.

## Programming Guidelines

- Comments should be proper sentences, with correct grammar and punctuation,
  including the use of capitalization and periods.
- Where defensive checks are added, include a comment explaining why they are
  appropriate (not necessary, since defensive checks are not necessary).
- I prefer text files to use new-line as a terminator rather than a separator
  i.e. newlines at the end of non-empty files.
- And lines should not have trailing whitespace.
- All command-line tools use pflag (github.com/spf13/pflag) for POSIX-compliant
  flag parsing, supporting both short forms (e.g., `-i`, `-f`) and long forms
  (e.g., `--input`, `--format`), and enabling flag bundling (e.g., `-if`).

## Test Guidelines

- When testing the behaviour of a binary, such as nutmeg-tokenizer, always use 
  `go run ./cmd/nutmeg-tokenizer`
  rather than `./nutmeg-tokenizer` directory. This ensures you are always testing
  the latest code rather than an out-of-date compiled binary. (Unless you are 
  deliberately testing an out-of-date binary).

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

  

## Nutmeg Rewriter - a standalone rewriter for the Nutmeg project

We are collaborating on the development of a standalone rewrite engine for the Nutmeg
programming language, implemented in the Go programming language. Given a stream
of node in JSON format, the rewriter transforms each node in turn and then outputs 
each node in various formats.

## Nutmeg Convert Tree

We are collaborating on the development of a standalone tree converter for the Nutmeg
programming language, implemented in the Go programming language. Given a stream
of node in JSON format, the converter prints each node in the specified target format.

## Nutmeg Common

This repository creates a single Go application that chains together the tokenizer,
parser and rewriter to form a complete parser for the common syntax.

## Nutmeg Resolver - a standalone name resolver for the Nutmeg project

We are collaborating on the development of a standalone name resolver for the Nutmeg
programming language, implemented in the Go programming language. Given a stream
of nodes in JSON format, the resolver identifies the lexical scope of each identifier and
annotates the nodes accordingly.

At the time of writing there is no package system. However we have a plan for
one [here](../docs/nutmeg_package_system.md). When this is implemented the resolver
will also handle cross-package name resolution.

## Nutmeg Codegen - a standalone code generator for the Nutmeg project

We are collaborating on the development of a standalone code generator for the Nutmeg
programming language, implemented in the Go programming language. Given a stream
of nodes in JSON format, the code generator transforms each top-level function into
a sequence of low-level instructions. We use the same flexible, unstructured node
format as input to this pipeline but transforms into a final structure format
in preparation for bundling.

## Nutmeg Bundler - a standalone bundler for the Nutmeg project

Bundling is the process of taking multiple Nutmeg source files, compiling them
into low-level code and then upserting them into a single database file that can
be efficiently loaded at runtime. This sqlite3 database file plays the role of
an executable in Nutmeg. However, unlike a normal executable, it includes 
assets such as images, data files and so forth.

## Nutmeg Compiler - a synthesis of the toolchain

Our goal is to build a new tool `nutmeg-compiler` that is a synthesis of
the individual tools. It strings together the tokenizer, parser,
rewriter, resolver, code generator and bundler into a single command line 
tool.

