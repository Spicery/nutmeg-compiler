# Task: nutmeg-resolver

This task is concerned with adding a new standalone command that will be
part of the Nutmeg toolchain: namely nutmeg-resolver. The job of this 
tool is to read in a unit-node in JSON format, to scan it for identifers, 
and to annotate them with the following information:

- A unique identifier ID, so that IDs with the same name but in different 
  scopes can be easily distinguished.
- Their scope, which is one of inner, outer or global, where inner and 
  outer are differently scoped locals.
- Whether this id-node is the definition or a use of the definition.

This command will share the same basic options as the other nutmeg-XXX 
commands and how no configuration file.

    -h, --help
    --version
    -i, --input FILE
    -o, --output FILE
    -f, --format FORMAT
    --no-spans
    --trim INT

The basic plan is to implement interface Visitor that supports the method
`OnArrival(node *Node) func(*Node)` that is applied on an in-order traverse of
the tree. A rough sketch looks like this.

```go
type Visitor interface {
    OnArrival(node *Node) func(*Node)
}

func Traverse(a Visitor, node *Node) {
    onDeparture := a.OnArrival(node)
    for _, child := range node.Children {
        Traverse(a, child)
    }
    if onDeparture != nil {
        onDeparture(node)
    }
}
```

This pattern should be implemented in `pkg/common/visitor.go` as it will be
used in other toolchains.

The nutmeg-resolver will use this idiom to manage its traversal of the 
node-tree, adding local variables to a list of scopes that it maintains. As
it returns back up the tree the onDeparture function pops the list of scopes.
