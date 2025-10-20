# Format of the configuration file for nutmeg-rewriter

This document provides a complete reference for the YAML configuration format used by `nutmeg-rewriter`.

## Top-Level Structure

```yaml
name: "MyRewriter"              # Optional: Name of the rewrite configuration
description: "Description"      # Optional: Human-readable description
passes:                          # Required: List of rewrite passes
  - name: "pass1"
    # ... pass configuration
```

## Pass Configuration

```yaml
- name: "PassName"               # Required: Name of this pass
  singlePass: false              # Optional: If true, run only once then skip (default: false)
  downwards:                     # Optional: Rules applied top-down
    - # ... rule configuration
  upwards:                       # Optional: Rules applied bottom-up
    - # ... rule configuration
```

## Rule Configuration

```yaml
- name: "RuleName"               # Optional: Name for debugging
  match:                         # Required: Pattern to match
    # ... pattern configuration
  action:                        # Required: Action to perform
    # ... action configuration
  onSuccess: "NextRuleName"      # Optional: Jump to named rule on success
  onFailure: "OtherRuleName"     # Optional: Jump to named rule on failure
  repeatOnSuccess: false         # Optional: If true, repeat this rule on success
```

Note: `repeatOnSuccess` is just a convenient shortcut for `onSuccess: {{THIS RULENAME}}`
and both compile to the same code.

**Default behavior:** On success/failure, continue to next rule in sequence.

## Pattern Configuration

Patterns match against nodes in the AST tree. All fields are optional; an empty pattern matches anything.

```yaml
match:
  self:                          # Match properties of the current node
    name: "NodeName"             # Match node name exactly
    key: "optionKey"             # Match presence of this option key
    value: "expectedValue"       # Match option value (requires key)
    matches: "regex"             # Match option value against regex (requires key)
    cmp: true                    # If false, inverts value/matches comparison
    count: 3                     # Match number of children
    siblingPosition: 0           # Match position among siblings (modulo)
  
  parent:                        # Match parent node (same fields as self)
    name: "ParentName"
    # ... any self fields
  
  child:                         # Match any child (returns first match)
    name: "ChildName"
    # ... any self fields
  
  previousChild:                 # Match child before matched child
    name: "PrevName"
    # ... any self fields
  
  nextChild:                     # Match child after matched child
    name: "NextName"
    # ... any self fields
```

**Note:** `previousChild` and `nextChild` require `child` to be specified.

## Action Configuration

**Only one action per rule.** For multiple actions, use `sequence`.

### Replace Node Name

```yaml
action:
  replaceName:
    with: "NewName"              # Set name to constant value
    # OR
    src: "self|parent|child"     # Source node
    from: "name|value|key"       # What to copy (requires src)
```

### Replace Option Value

```yaml
action:
  replaceValue:
    key: "optionKey"             # Required: Which option to modify
    with: "NewValue"             # Set to constant value
    # OR
    src: "self|parent|child"     # Source node
    from: "name|value|key"       # What to copy (requires src)
```

### Replace Node with Child

```yaml
action:
  replaceByChild: 0              # Replace node with child at index
```

### Inline Child

```yaml
action:
  inlineChild: true              # Replace matched child with its children
```

### Remove Child

```yaml
action:
  removeChild: true              # Remove matched child from parent
```

### Remove Option

```yaml
action:
  removeOption:
    key: "optionKey"             # Remove this option from node
```

### Rotate Option Value

```yaml
action:
  rotateOption:
    key: "optionKey"             # Required: Which option to rotate
    values: ["A", "B", "C"]      # Required: Cycle through these values
    initial: "A"                 # Optional: Starting value (default: first)
```

### Merge Child with Next

```yaml
action:
  mergeChildWithNext: true       # true = next wins, false = child wins
```

### Create New Child Node

```yaml
action:
  newNodeChild:
    name: "NewNode"              # Required: Name of new child
    key: "optionKey"             # Optional: Option key for new node
    value: "optionValue"         # Optional: Option value for new node
    offset: 0                    # Optional: Insert position (default: 0)
    length: 1                    # Optional: Number of children to replace
```

### Permute Children

```yaml
action:
  permuteChildren: [2, 0, 1]     # Reorder children by indices
```

### Action Sequence

```yaml
action:
  sequence:                      # Execute actions in order
    - replaceName:
        with: "First"
    - replaceValue:
        key: "attr"
        with: "Second"
    # ... more actions
```

### Child Action

```yaml
action:
  childAction:                   # Apply action to matched child
    replaceName:
      with: "ModifiedChild"
```

## Complete Example

```yaml
name: "Expression Optimizer"
description: "Optimizes arithmetic expressions"

passes:
  - name: "Initialization"
    singlePass: true
    downwards:
      - name: "Setup constants"
        match:
          self: { name: "const" }
        action:
          replaceValue:
            key: "evaluated"
            with: "true"
  
  - name: "Constant Folding"
    downwards:
      - name: "Fold addition"
        match:
          self: { name: "add" }
          child: { name: "const", key: "value" }
        action:
          sequence:
            - replaceByChild: 0
            - replaceName: { with: "folded" }
        repeatOnSuccess: true
      
      - name: "Remove identity"
        match:
          self: { name: "add" }
          child: { name: "const", key: "value", value: "0" }
        action:
          removeChild: true
```

## Tips

- **Patterns**: More specific patterns = fewer false matches
- **Actions**: Use `sequence` for multi-step transformations
- **Control Flow**: Use `onSuccess`/`onFailure` for complex rule chains
- **Passes**: Use `singlePass: true` for setup/initialization
- **Iteration**: Rewriter loops until no changes (or `--max-rewrites` limit)
- **Regex**: `matches` field automatically anchors patterns (`^...$`)
- **Debugging**: Check stderr for optimization and execution logs
