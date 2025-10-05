# Standard Structure for Parser Output

This notes describes the parser-output format for the nutmeg-compiler. This
is the format that the first-stage of the compiler toolchain must emit.

The first stage includes:

- tokenization
- direct parse
- rewrite to standard format

Later stages include:

- Name resolution
- Package binding
- Type analysis
- Optimisation
- Code Generation

## Format

- `<unit src=FILENAME>`
- `<number ...>`
- `<string ...>`
- `<id name=NAME>` = ID
- `<bind>ID EXPR</bind>`
- `<bind><seq>ID*</seq> EXPR</bind>`
- `<assign>ID EXPR</assign>`
- `<assign><seq>ID*</seq>EXPR</assign>`
- `<update>EXPR EXPR</update>`
- `<seq>EXPR*</seq>`
- `<apply>FUN ARG</apply>`
- `<if>EXPR EXPR EXPR</if>`
- `<for>EXPR EXPR</for>`
- `<def>EXPR EXPR</def>`
- `<annotate>ANNOTATION EXPR</annotate>`


