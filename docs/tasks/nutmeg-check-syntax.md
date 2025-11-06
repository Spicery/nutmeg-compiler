# Task: nutmeg-check-syntax

## Tool

This task is concerned with adding a new standalone command that will be
part of the Nutmeg toolchain: namely nutmeg-check-syntax. This 
tool will:

- read in the unit-node (in JSON format) that is emitted by nutmeg-parse,
- walk the tree, checking that nodes conform to additional syntactic rules that
  are not enforced by nutmeg-parse,
- if the rules are violated them it will exit with a non-zero status,
- otherwise it simply emits the tree unchanged on the stdout.

This command will share the same basic options as the other nutmeg-XXX 
commands such as nutmeg-parser, nutmeg-rewriter and nutmeg-resolver.

    --help
    --version
    --input FILE
    --output FILE
    --format, -f FORMAT
    --no-spans
    --trim INT

## Validations

The key validations is concerned with the structure of definitions ("def") and
functions ("fn"). In addition we will check the arity of the different types 
of node. Don't try to implement these, simply provide scaffolding code.


## Extending nutmeg-common

We will be extending nutmeg-common with it too - splicing it between the
parse and rewrite stages.
