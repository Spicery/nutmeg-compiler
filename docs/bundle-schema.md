# Bundle Database Schema

The Nutmeg bundle file is a SQLite database that serves as the "executable" format for compiled Nutmeg programs. It contains all the compiled code, dependencies, metadata, and source files.

**Note:** GORM automatically converts Go struct field names to snake_case for column names and pluralizes table names.

## SQL Schema

```sql
CREATE TABLE `migrations` (
    `id` text,
    PRIMARY KEY (`id`)
);

CREATE TABLE `entry_points` (
    `id_name` text,
    PRIMARY KEY (`id_name`)
);

CREATE TABLE `depends_ons` (
    `id_name` text,
    `needs` text,
    PRIMARY KEY (`id_name`,`needs`)
);
CREATE INDEX `idx_depends_ons_needs` ON `depends_ons`(`needs`);
CREATE INDEX `idx_depends_ons_id_name` ON `depends_ons`(`id_name`);

CREATE TABLE `bindings` (
    `id_name` text,
    `lazy` numeric,
    `value` text,
    `file_name` text,
    PRIMARY KEY (`id_name`)
);

CREATE TABLE `source_files` (
    `file_name` text,
    `contents` text,
    PRIMARY KEY (`file_name`)
);

CREATE TABLE `annotations` (
    `id_name` text,
    `annotation_key` text,
    `annotation_value` text,
    PRIMARY KEY (`id_name`,`annotation_key`)
);
CREATE INDEX `idx_annotations_id_name` ON `annotations`(`id_name`);
```

## Tables

### migrations

Tracks which schema migrations have been applied (managed by gormigrate).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | text | PRIMARY KEY | Migration identifier |

### entry_points

Specifies which bindings are entry points that can be invoked from outside the bundle.

| Column  | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id_name | text | PRIMARY KEY | The name of the entry point binding |

### depends_ons

Tracks dependencies between bindings for proper initialization order and lazy evaluation.

| Column  | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id_name | text | PRIMARY KEY, INDEX | The binding that has a dependency |
| needs   | text | PRIMARY KEY, INDEX | The binding that is depended upon |

### bindings

Stores compiled function objects and their metadata.

| Column    | Type    | Constraints | Description |
|-----------|---------|-------------|-------------|
| id_name   | text    | PRIMARY KEY | The name of the binding |
| lazy      | numeric |             | Whether the binding is lazy (deferred evaluation) |
| value     | text    |             | JSON-serialized FunctionObject or Node |
| file_name | text    |             | Source file path (may be empty) |

### source_files

Preserves the original source code for debugging and error reporting.

| Column    | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| file_name | text | PRIMARY KEY | Path to the source file |
| contents  | text |             | Full source code content |

### annotations

Stores metadata annotations associated with bindings.

| Column           | Type | Constraints         | Description |
|------------------|------|---------------------|-------------|
| id_name          | text | PRIMARY KEY, INDEX  | The binding being annotated |
| annotation_key   | text | PRIMARY KEY         | The annotation name/key |
| annotation_value | text |                     | The annotation value (currently unused) |

## Function Object Format

The `Value` column in the `Bindings` table contains JSON-serialized function objects. A function object has the following structure:

```json
{
  "nlocals": <integer>,
  "nparams": <integer>,
  "instructions": [
    { "type": "<instruction-type>", ... },
    ...
  ]
}
```

### Fields

- **nlocals**: Number of local variable slots needed in the call frame
- **nparams**: Number of parameters the function accepts
- **instructions**: Array of instruction objects (see Instruction Set below)

## Instruction Set

Each instruction is a JSON object with a `type` field and additional type-specific fields.

### Value Stack Operations

#### push.int
Pushes an integer constant onto the value stack.
```json
{ "type": "push.int", "ivalue": <integer> }
```

#### push.bool
Pushes a boolean constant onto the value stack.
```json
{ "type": "push.bool", "value": "true" | "false" }
```

#### push.string
Pushes a string constant onto the value stack.
```json
{ "type": "push.string", "value": "<string>" }
```

#### erase
Discards the top value from the value stack.
```json
{ "type": "erase" }
```

### Local Variable Operations

#### push.local
Pushes a local variable's value onto the value stack.
```json
{ "type": "push.local", "index": <offset> }
```

#### pop.local
Pops the top of the value stack into a local variable slot.
```json
{ "type": "pop.local", "index": <offset> }
```

#### stack.length
Stores the current value stack length into a local variable slot.
```json
{ "type": "stack.length", "index": <offset> }
```

### Global Variable Operations

#### push.global
Pushes a global variable's value onto the value stack.
```json
{ "type": "push.global", "name": "<identifier>" }
```

### Call Operations

#### syscall.counted
Invokes a system function with dynamic argument count checking.
```json
{ "type": "syscall.counted", "name": "<sysfn-name>", "index": <stack-length-offset> }
```

#### call.global.counted
Invokes a global function with dynamic argument count checking.
```json
{ "type": "call.global.counted", "name": "<function-name>", "index": <stack-length-offset> }
```

### Control Flow Operations

#### return
Returns from the current function.
```json
{ "type": "return" }
```

#### label
Defines a jump target.
```json
{ "type": "label", "value": "<label-name>" }
```

#### goto
Unconditional jump to a label.
```json
{ "type": "goto", "value": "<label-name>" }
```

### Conditional Operations

#### check.bool
Verifies that the top of the stack is a boolean value.
```json
{ "type": "check.bool", "index": <stack-length-offset> }
```

#### if.not
Jumps to a label if the top of the stack is false.
```json
{ "type": "if.not", "value": "<label-name>" }
```

#### if.so
Jumps to a label if the top of the stack is true.
```json
{ "type": "if.so", "value": "<label-name>" }
```

#### if.not.return
Returns from the function if the top of the stack is false.
```json
{ "type": "if.not.return" }
```

#### if.so.return
Returns from the function if the top of the stack is true.
```json
{ "type": "if.so.return" }
```

#### if.then.else
Branches to one of two labels based on the boolean value on the stack.
```json
{ "type": "if.then.else", "name": "<then-label>", "value": "<else-label>" }
```

### Lazy Evaluation Operations

#### in.progress
Marks a global binding as being initialized (for cycle detection).
```json
{ "type": "in.progress", "name": "<binding-name>" }
```

#### done
Marks a global binding as fully initialized and stores the result.
```json
{ "type": "done", "name": "<binding-name>", "index": <stack-length-offset> }
```

## Migration Version

Current schema version: `202511250001`

The schema is managed using GORM migrations. Use the `--migrate` flag with nutmeg-bundler to update the schema when needed.
