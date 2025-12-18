# HOW TO: Add boolean literals

## Introduction

This is really a case-study in adding named literals i.e. literals that overlap
with normal identifiers.

The interesting aspect of booleans is that their literals `true` and `false` are
standard identifiers. So the question arises: should the tokeniser categorise
them in the same way as numeric literals and string literals? Or should it rely
on categorising them as `prefix-syntax`, similar to `return` or `yield` but with
arity 0?

Obviously both are viable approaches. But literal-syntax cannot be overridden by
the tokeniser-rules - which means if someone wanted to have (say) `True` and
`False` instead there would be no override. Without a user-base this isn't a
particularly strong point, of course, but the principle indicates that ordinary
identifiers are handled via the prefix mechanism.

## Implementation in the compiler

### Step 1: Tokenizer Rules

Having decided to set up true/false as prefix keywords, I first modified the
`pkg/tokenizer/rules.go` file, which is where the default rules for tokenization
are defined. I added lines to `getDefaultPrefixTokens`, which sets up the
default rules for prefix-tokens.

```go
func getDefaultPrefixTokens() map[string]PrefixTokenData {
	return map[string]PrefixTokenData{
		"return": {Precedence: LOOSE, Arity: common.One},
		"yield":  {Precedence: LOOSE, Arity: common.One},
		"const":  {Precedence: TIGHT, Arity: common.One},
		"var":    {Precedence: TIGHT, Arity: common.One},
		"val":    {Precedence: TIGHT, Arity: common.One},
		"true":   {Precedence: ATOMIC, Arity: common.Zero}, // CHANGED
		"false":  {Precedence: ATOMIC, Arity: common.Zero}, // CHANGED
	}
}
```

**Key points:**
- Precedence is `ATOMIC` (highest) since literals bind tightest
- Arity is `common.Zero` since booleans take no arguments
- This makes `true`/`false` behave like keywords with zero operands

### Step 2: Rewriter Rules

This meant that the parser would generate the following for (say) `true`:

```xml
<form syntax="prefix" span="1 1 1 5">
    <part keyword="true" span="1 1 1 5" />
</form>
```

To get this into the form I wanted, which was `<boolean value="true">`, I
modified the `configs/rewrite.yaml`. This is my working file that contains the
default rule-set for the nutmeg-rewriter. By using the `--rewrite-rules` option,
I was able to quickly experiment with different rewrite rules to get it
into the form I wanted.

The rules I came up with looked like:

```yaml
      - name: booleans
        match:
          self:
            name: form
            count: 1
          child:
            name: part
            key: keyword
            value.regexp: true|false
        action:
          sequence:
            - replaceName:
                with: boolean
            - replaceValue:
                key: value
                src: child
                from: value
            - removeChildren: true
```

**What this rule does:**
1. **Match condition**: Finds `<form>` nodes with one child that is a `<part>` with `keyword` attribute matching `true|false`
2. **Actions**:
   - Renames `form` to `boolean`
   - Copies the keyword value (`true`/`false`) to a new `value` attribute
   - Removes child nodes (no longer needed since value is now an attribute)

And the result for `true` was:

```xml
<boolean value="true" span="1 1 1 5" />
```

**Result**: Clean, self-contained boolean literal node with no child structure.

### Step 3: Syntax Validation

I had to add another clause to the nutmeg-checker to allow this to pass, of
course:

```go
    ...
	case common.ValueTrue, common.ValueFalse: // ADDED THIS CASE.
		c.validateFormBoolean(node)
	default:
		c.addIssue(fmt.Sprintf("unexpected form keyword: %s", keyword), first)
	}
}

// ADDED THIS DEFINITION
func (c *Checker) validateFormBoolean(bool_node *common.Node) {
	if !c.factArity(1, bool_node) {
		return
	}
	if !c.expectArity(0, bool_node.Children[0]) {
		return
	}
}
```

**Validation logic:**
- Ensures the boolean node has exactly 1 child (the `<part>` wrapper from parsing)
- Ensures that child has 0 children (it's just the keyword)
- This validates structure before rewriting transforms it

### Step 4: Code Generation

With this in place, we obviously now needed to generate the right code in
`nutmeg-codegen`. Exactly the same way that numbers become `push.int` and
strings become `push.string`, so booleans become `push.bool`.

**Pattern to follow:**
- Add case in code generator's expression handler
- Map `<boolean value="true">` → `push.bool true`
- Map `<boolean value="false">` → `push.bool false`
- Generate instruction node similar to other push operations

### Step 5: Bundler Support

This isn't quite enough, since the bundler needs to know how to handle the
`push.bool` instruction - but this was quite mechanical.

**Changes needed:**
- Add `push.bool` to instruction type enumeration
- Serialize boolean values into bundle database
- Mirror the patterns used for `push.int` and `push.string`


## Implementation in nutmeg-run

The implementation in nutmeg-run was pretty straightforward and was essentially
a clone of PUSH_INT. It involved adding another instruction opcode in
`instruction.[ch]pp` and adding code planting case - but not adding another
instruction handler, since this was effectively a push-constant.

**Key changes:**
- Add `PUSH_BOOL` opcode to instruction enumeration
- Implement opcode serialization/deserialization
- Runtime execution: push boolean constant onto stack
- Pattern matches other push-constant instructions (PUSH_INT, PUSH_STRING)

## Summary: The Pipeline

When you write `true` in Nutmeg code:

1. **Tokenizer**: Recognizes `true` as prefix keyword with arity 0 → `<part keyword="true">`
2. **Parser**: Wraps in form structure → `<form><part keyword="true"/></form>`
3. **Rewriter**: Transforms to clean literal → `<boolean value="true"/>`
4. **Checker**: Validates structure (before rewrite: has 1 child with 0 children)
5. **Code Generator**: Emits instruction → `push.bool true`
6. **Bundler**: Serializes to bundle database
7. **Runtime**: Executes `PUSH_BOOL` opcode, pushes boolean onto stack

## Alternative Approaches Considered

**Approach 1: Literal tokens** (rejected)
- Tokenizer could categorize `true`/`false` as literal tokens (like numbers/strings)
- **Con**: No override mechanism via tokenizer-rules
- **Con**: Can't support alternative literals like `True`/`False` or `yes`/`no`

**Approach 2: Prefix keywords** (chosen)
- Treat as zero-arity prefix operators
- **Pro**: Users can override via tokenizer configuration
- **Pro**: Consistent with other keywords (`return`, `yield`)
- **Pro**: Extensible pattern for other named literals

## Lessons Learned

1. **Consistency matters**: Following the prefix-keyword pattern for `return`/`yield` made booleans natural
2. **Rewriter is powerful**: Transforming generic form nodes to specific typed nodes keeps parser simple
3. **Clone patterns**: Most literal types follow same pipeline (tokenize → parse → rewrite → validate → codegen → bundle → run)
4. **Test at each stage**: Verify XML output after each transformation to debug the pipeline
