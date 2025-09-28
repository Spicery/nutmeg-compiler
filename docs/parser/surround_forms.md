# Parsing Surround Forms

Surround forms are the core of the system. The basic idea is that they start
and finish with distinctive tokens.

    START .... END

In between these, we have a series of expressions which can be "punctuated"
by bridging-tokens:

    START .... BRIDGE1 .... BRIDGE2 ... END

As an example that fits this, here is a "let" expression with "in" acting as a
bridge.

    let ...... in .... endlet

There are two types of bridge though - those followed by a single expression and
those followed by a series of expressions (statements) - which I can illustrate
with an if-form:

    if EXPR then
        EXPR ...
    elseif EXPR then
        EXPR ...
    else
        EXPR ...
    endif

Multi-expression labels are 'then' and 'else'. But 'elseif' is a single-expression
label. Note that the `if` keyword itself only allows one expression before the
next bridge. Whether a single expression is required or a series is allowed is
indicated by the "single" attribute.

There are consistency checks that are automatically fulfilled during the parse.
Bridging words are associated with allowed start-tokens, so may only be used
within the right form, determined by the in-attribute. And the sequence of
bridge-words is constrained by the expecting-attribute - each bridge word must
be expected by the previous.

For example, let's examine a simple example: `if x then y else z endif`. This
is tokenized as follows:

    {"text":"if","span":[1,1,1,3],"type":"S","expecting":["then"],"closed_by":["end","endif"],"single":true}
    {"text":"x","span":[1,4,1,5],"type":"V"}
    {"text":"then","span":[1,6,1,10],"type":"B","expecting":["elseif","else"],"in":["if","ifnot"],"single":false}
    {"text":"y","span":[1,11,1,12],"type":"V"}
    {"text":"else","span":[1,13,1,17],"type":"B","in":["if","ifnot"],"single":false}
    {"text":"z","span":[1,18,1,19],"type":"V"}
    {"text":"endif","span":[1,20,1,25],"type":"E","ln_after":true}

1. The parser detects that `if` is a start-token (type S), that should be
   followed by a single expression, can be continued with a bridging word `then`
   and closed by either `end` or `endif`. 

2. The parser classifies `x` as a variable. This will become the following
   expression.

3. The parser classifies `then` as a bridging word (type B). It was expecting
   that word, so that's OK. In addition `then` should appear inside an `if` or
   `ifnot`. In fact we are within an `if`, so again that is OK. Then consumes
   multiple expressions. It is expecting the next bridging word to be `else`
   or `elseif`.

4. It then classifies `y` as a variable and the next expression, first of 
   a series of length 1.

5. `else` is classified as a bridging-word (type B). It was expected and it
   is inside an `if`, so constraints are passed. It can consume a series of
   expressions but expects no further bridging tokens.

6. `z` is a variable and the last expression.

7. `endif` is classified as an end-token (type E). It is in the closed_by
   list of the original start-token, so again everything is OK.
   