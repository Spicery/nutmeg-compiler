# Improve the performance of nutmeg-rewriter

This task is concerned with improving the performance of a RewriterPass in
`rewriter.go`. This type has two rule arrays and, at present, when we apply
these rule-sets to a node we essentially step linearly through the rules
attempting to match each one in turn. We will be making two improvements on
this.

The first improvement is that for each rule-set we will compile a `map[string]int`
that takes the _name_ of the node and determines which rule you should start 
from. If there is no rule, this means you either can skip all the rules or there
is a "wildcard" rule; in both cases we just compile a default-start. 

- Note that a wildcard rule simply uses `Self.Name == nil`.
- When there is no match the start index should, of course, the length of the array!

The second improvement utilises the fact that after a rule is run it moves to
the next rule via the OnSuccess/OnFailure fields. As a consequence we can check
to see whether or not the rule being jumped to cannot match - in which case we
can optimise it by bumping the pointer along (and then iterating).

Both of these optimisations are performed when we load rule-set in the 
function NewRewriter.