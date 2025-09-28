# The Algorithm Implemented by the Parser

WORK IN PROGRESS

## Token Categories

Tokens are given a 1-letter category code:

- [x] `n` - Numeric literals with radix support
- [x] `s` - String literals with quotes and escapes
- [x] `S` - Start tokens (form start tokens like `def`, `if`, `while`)
- [x] `E` - End tokens (form end tokens like `end`, `endif`, `endwhile`)
- [x] `B` - Bridging tokens
- [x] `P` - Prefix tokens (prefix operators like `return`, `yield`)
- [x] `V` - Variable tokens (variable identifiers)
- [-] `O` - Operator tokens (infix/postfix operators)
  - [ ] Prefix
  - [x] Infix
  - [ ] Postfix
- [x] `[` - Open delimiter tokens (opening brackets/braces/parentheses)
- [x] `]` - Close delimiter tokens (closing brackets/braces/parentheses)
- [ ] `U` - Unclassified tokens
- [ ] `X` - Exception tokens (for invalid constructs)

These categories instruct the parser as to what to do.

## Strategy

- Recursive descent with operator precedence

## Token driven parsing

### Literals: numbers and strings

These literal tokens are immediately translated into nodes. For example:

```
❯ echo '1966' | nutmeg-tokeniser 
{"text":"1966","span":[1,1,1,5],"type":"n","radix":"","base":10,"mantissa":"1966","ln_after":true}
```

And this would be translated by nutmeg-parser into a number-node.
```
❯ echo '1966' | nutmeg-tokeniser | nutmeg-parser
<unit span="1 1 2 1">
  <number span="1 1 1 5" base="10" mantissa="1966" />
</unit>
```

Strings are similar:
```
❯ echo '"quack"' | go run ./cmd/nutmeg-tokenizer/
{"text":"\"quack\"","span":[1,1,1,8],"type":"s","value":"quack","ln_after":true}

❯ echo '"quack"' | nutmeg-tokeniser | nutmeg-parser
<unit span="1 1 2 1">
  <string span="1 1 1 8" value="quack" />
</unit>
```

### Delimiters

Delimiters are the categorisation given to paired brackets, such as parentheses,
braces and brackets. The general form is: 

- `[` EXPRESSION, ... `]`
    - where `[` and `]` are the category codes and not literals.

The output is like this:


