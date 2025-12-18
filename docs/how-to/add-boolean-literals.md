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

Having decided to set up true/false as prefix keywords, I first modified the
`pkg/tokenizer/rules.go` file, which is where the defaut rules for tokenization
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

And the result for `true` was:

```xml
<boolean value="true" span="1 1 1 5" />
```

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

With this in place, we obviously now needed to generate the right code in
`nutmeg-codegen`. Exactly the same way that numbers become `push.int` and
strings become `push.string`, so booleans become `push.bool`. 

This isn't quite enough, since the bundler needs to know how to handle the
`push.bool` instruction - but this was quite mechanical.


## Implementation in nutmeg-run

The implementation in nutmeg-run was pretty straightforward and was essentially
a clone of PUSH_INT. It involved adding another instruction opcode `instruction.[ch]pp`
and adding code planting case - but not adding another instruction, since this
was effectively a push-constant.
