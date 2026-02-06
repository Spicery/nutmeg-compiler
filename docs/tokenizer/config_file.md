# Format of the config file for nutmeg-tokenizer

This document provides a complete reference for the YAML configuration format
used by `nutmeg-tokenizer`.

## Overview

The tokenizer configuration defines how text is classified into tokens. Token
**boundaries** are determined by the tokenizer algorithm (strings, numbers,
alphanumerics, sign sequences, single characters), while token
**classification** is determined by the rules.

## Top-Level Structure

```yaml
bracket:   # Optional: Bracket/delimiter rules
  - # ...
prefix:    # Optional: Prefix form rules
  - # ...
start:     # Optional: Form start tokens
  - # ...
bridge:    # Optional: Bridge tokens (labels)
  - # ...
wildcard:  # Optional: Wildcard tokens 
  - # ...
operator:  # Optional: Operator rules with precedence
  - # ...
mark:      # Optional: Mark tokens (punctuation)
  - # ...
```

**Note:** All categories are optional. If a category is included, it
**replaces** the defaults for that category entirely.

## Bracket Rules

Define opening delimiters and their matching closing delimiters.

```yaml
bracket:
  - text: "("              # Required: Opening bracket text
    closed_by:             # Required: List of valid closing brackets
      - ")"
    infix: 0               # Optional: Infix precedence (0 = not infix)
    prefix: true           # Optional: Can be used as prefix (default: false)
```

**Examples:**
```yaml
bracket:
  - text: "("
    closed_by: [")"]
    infix: 0
    prefix: true
  
  - text: "["
    closed_by: ["]"]
    infix: 2150          # Can be used as infix operator
    prefix: true
  
  - text: "{"
    closed_by: ["}"]
    infix: 0
    prefix: true
```

**Token Type:** `[` (open delimiter)  
**Closes as:** `]` (close delimiter)

## Prefix Rules

Define tokens that start prefix forms (prefix operators or keywords).

```yaml
prefix:
  - text: "return"         # Required: Token text
  - text: "yield"
  - text: "throw"
```

**Token Type:** `P` (prefix)

## Start Rules

Define tokens that start multi-part constructs (like `if...then...end`).

```yaml
start:
  - text: "if"             # Required: Token text
    closed_by:             # Required: List of tokens that close this form
      - "end"
      - "endif"
    expecting:             # Optional: Immediate next expected tokens
      - "then"
    single: false          # Optional: If true, only one instance allowed
```

**Examples:**
```yaml
start:
  - text: "if"
    closed_by: ["end", "endif"]
    expecting: ["then"]
    single: true
  
  - text: "def"
    closed_by: ["end", "enddef"]
    expecting: ["=>>"]
    single: true
  
  - text: "for"
    closed_by: ["end", "endfor"]
    expecting: ["do"]
    single: false
```

**Token Type:** `S` (start)  
**Paired with:** `E` (end) - automatically generated from `closed_by`

## Bridge Rules

Define tokens that appear within multi-part constructs (like `else` in `if...else...end`).

```yaml
bridge:
  - text: "else"           # Required: Token text
    expecting:             # Required: List of tokens that can follow
      - "then"
    in:                    # Required: List of start tokens that can contain this
      - "if"
      - "unless"
```

**Examples:**
```yaml
bridge:
  - text: "else"
    expecting: []          # Nothing expected after
    in: ["if", "unless"]
  
  - text: "=>>"
    expecting: ["do"]
    in: ["def"]
  
  - text: "do"
    expecting: []
    in: ["def", "for", "while"]
  
  - text: "elseif"
    expecting: ["then"]
    in: ["if"]
```

**Token Type:** `B` (bridge)

## Wildcard Rules (`:`)

Define tokens that act as wildcards that, when encountered, are interpreted
as the token that was actually expected. For example a `:` can be used instead
of `then` or `do` and appears in the parse tree as the expected `then` or `do`.

```yaml
wildcard:
  - text: ":"              # Required: Token text
  - text: "..."
```

**Token Type:** Varies based on context (can be `S`, `E`, or `B`)  
**Special:** Includes a `value` field with the expected token it represents

## Operator Rules

Define operators with their precedence levels.

```yaml
operator:
  - text: "+"              # Required: Token text
    precedence: [50, 2040, 0]  # Required: [prefix, infix, postfix]
  
  - text: "*"
    precedence: [0, 2010, 0]   # No prefix, infix only
  
  - text: "-"
    precedence: [50, 2050, 0]  # Both prefix and infix
```

**Precedence Format:** `[prefix, infix, postfix]`
- Use `0` to disable that position
- Lower numbers = tighter binding
- Infix typically starts at 2000+
- Postfix typically starts at 1000+

**Default Precedence Calculation:**
- Base precedence from first character (see below)
- Subtract 1 if first character repeats (e.g., `**` tighter than `*`)
- Infix: base + 2000
- Postfix: base + 1000
- Prefix: base value

**Character Precedence:**
```
*  → 10    /  → 20    %  → 30    +  → 40
-  → 50    <  → 60    >  → 70    ~  → 80
!  → 90    &  → 100   ^  → 110   |  → 120
?  → 130   =  → 140   :  → 150
```

**Token Type:** `O` (operator)

## Mark Rules

Define punctuation marks (separators like `,` and `;`).

```yaml
mark:
  - text: ","              # Required: Token text
  - text: ";"
```

**Token Type:** `U` (unclassified) with special handling  
**Default:** `,` and `;`

## Token Boundaries

The tokenizer automatically recognizes these boundaries:
- **Strings:** Quoted text with escape sequences
- **Numbers:** Numeric literals with radix support (decimal, hex, binary, etc.)
- **Alphanumeric:** Letters, digits, underscores bind together
- **Sign characters:** Operator characters bind together
- **Everything else:** Single character tokens

## Token Classification Priority

1. **Strings and numbers** - Recognized by algorithm
2. **Custom rules** - Defined in config file
3. **Default rules** - Built-in classifications
4. **Variables** - Alphanumeric sequences not matching rules

## Complete Example

```yaml
# Complete tokenizer configuration for a small language

bracket:
  - text: "("
    closed_by: [")"]
    infix: 0
    prefix: true
  - text: "["
    closed_by: ["]"]
    infix: 2150
    prefix: true

prefix:
  - text: "return"
  - text: "yield"

start:
  - text: "if"
    closed_by: ["end", "endif"]
    expecting: ["then"]
    single: true
  - text: "def"
    closed_by: ["end", "enddef"]
    expecting: ["=>>"]
    single: true
  - text: "while"
    closed_by: ["end", "endwhile"]
    expecting: ["do"]
    single: false

bridge:
  - text: "else"
    expecting: []
    in: ["if"]
  - text: "elseif"
    expecting: ["then"]
    in: ["if"]
  - text: "=>>
    expecting: ["do"]
    in: ["def"]
  - text: "do"
    expecting: []
    in: ["def", "while", "for"]

wildcard:
  - text: ":"

operator:
  - text: "+"
    precedence: [40, 2040, 0]
  - text: "-"
    precedence: [50, 2050, 0]
  - text: "*"
    precedence: [0, 2010, 0]
  - text: "/"
    precedence: [0, 2020, 0]
  - text: "="
    precedence: [0, 2140, 0]

mark:
  - text: ","
  - text: ";"
```

## Output Format

Tokens are emitted as JSON objects (one per line, JSONL format). See `tokens.md`
for the complete token format reference.

## Tips

- **Category replacement:** Including a category replaces all defaults for that category
- **Wildcard usage:** Wildcards match expected tokens in their position
- **Precedence tuning:** Lower numbers bind tighter (multiply before add)
- **Single flag:** Use `single: true` for constructs that shouldn't nest
- **Testing:** Use `nutmeg-tokenizer` with sample input to verify rules
- **Debugging:** Check token output for type codes and attributes
