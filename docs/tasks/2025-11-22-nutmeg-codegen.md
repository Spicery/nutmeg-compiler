# Code Generation

In this task we are going to add another tool: `nutmeg-codegen`. This will be
written in the style of the other tools such as `nutmeg-parser`,
`nutmeg-resolver` and so on. Specifically it will use the same command-line
structure and the pkg/common library.

In this first task we will write this as a skeleton application that walks
the node-tree, finding and rewriting any `fn` nodes (`node.Name == NameFn`).
We will leave the rewriting as a stub.

We will need to ensure that the Justfile is updated to reflect this new tool.