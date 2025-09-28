
- [x] signs for numbers

- [x] Spans correct everywhere

- [x] Implement infix delimiters

- [x] ensure that the in-constraint is honoured

- [x] Decent handling of `U` tokens

- [x] Decent handling of `X` tokens

- [x] Decent error messages when:
  - [x] Encountering unexpected end tokens
  - [x] Encountering bridge words that are not "in" the surround

- [x] Semi-colons 
  - [x] Line-breaks as expression separators

- [x] let syntax
    - Note that `in` plays the role of a bridging word in `let`
    - But an operator role in queries
    - solution, `in` is not used inside `let`.

- [x] Prefix operators
    - [x] We need to make sure that `-` can be negation in the tokenizer.

- [x] Postfix operators

- [x] Note that the list of possible closers includes end/endif when it shouldn't, really.

    ❯ echo 'if x: 1 endif' | (cd ../nutmeg-tokeniser/; go run ./cmd/nutmeg-tokenizer) | go run ./cmd/nutmeg-parser/
    nutmeg-parser options:
    Format: XML
    Error reading input: unexpected token ':' at line 1, column 5, expecting then/end/endif
    exit status 1

- [x] Incorrect error message

    ❯ echo 'if  endif' | (cd ../nutmeg-tokeniser/; go run ./cmd/nutmeg-tokenizer) | go run ./cmd/nutmeg-parser/
    nutmeg-parser options:
    Format: XML
    Error reading input: unimplemented, got token type: E
    exit status 1