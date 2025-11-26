# Bundler

## Part 1

In this task we are going to add another tool: `nutmeg-bundler`. This will
use the same input flags as other commands, such as `nutmeg-parser` but it
will require a mandatory `--bundle=FILE` option. This is a SQLITE file that
acts as a unified binary "executable" by the nutmeg runtime.

The bundle file will have the following schema:

```sql
CREATE TABLE IF NOT EXISTS "EntryPoints" (
	"IdName"	TEXT,
	PRIMARY KEY("IdName")
);
CREATE TABLE IF NOT EXISTS "DependsOn" (
	"IdName"	TEXT NOT NULL,
	"Needs"	TEXT NOT NULL,
	PRIMARY KEY("IdName","Needs")
);
CREATE TABLE IF NOT EXISTS "Bindings" (
	"IdName"	TEXT,
    "Lazy"  BOOLEAN,
	"Value"	TEXT,
	"FileName"	TEXT,
	PRIMARY KEY("IdName")
);
CREATE TABLE IF NOT EXISTS "SourceFiles" ( "FileName" TEXT NOT NULL, "Contents" TEXT NOT NULL, PRIMARY KEY("FileName") );
CREATE TABLE IF NOT EXISTS "Annotations" ( "IdName" TEXT, "AnnotationKey" TEXT NOT NULL, "AnnotationValue" TEXT NOT NULL, PRIMARY KEY("IdName", "AnnotationKey") );
```

`nutmeg-bundler` should use a database migration tool. I suggest GORM and
gormigrate are a good combination. It should check if the bundle file is
up to date - if not it should refuse to proceed but tell the user about
the `--migrate` option. And we should implement that option of course.

The bundler itself will iterate across the children `<unit>` node that is
supplied as input and add entries to the tables as follows:

- If the child is an <annotations> node then the contents are added to
  an accumulating list of annotations.

- If the child is a <bind> node then it will have 2 children: an <id> node
  and a <fn> node. 
  - A Bindings row is upserted with the `name` of the <id>
  - And the value is the function
  - And Lazy is the `lazy` value of the <id>
  - And Filename is the `src` of the unit or NULL if not known
  - If there were any annotations then each annotation is upserted as a 
    row with name and annotation key set (the value is not used at present)
  - The annotations list is cleared.

## Part 2

The second part of the bundler is to convert the <fn> nodes into our target
instruction set. IMPORTANT: Do NOT implement the instruction set as part of this
task.

The instruction set (so far) is:

- `<push.int decimal="INTEGER_BASE_10" />`, pushes the referenced integer onto
  the value-stack.
- `<push.string value="TEXT" />`, pushes the referenced string onto the
  value-stack.
- `<stack.length offset="NUMBER" />`, this finds the current stack length and
  puts it in the appropriate slot in the call-stack. 
- `<pop.local offset="NUMBER" />`, pops the top of the value-stack into the
  numbered slot of the call-stack.
- `<push.local offset="NUMBER" />`, pushes the top of the value-stack into the
  numbered slot of the call-stack.
- `<pop.global name="name" />`, pops the top of the value-stack into the
  named global variable.
- `<push.global name="NAME" />`, pushes the top of the value-stack into the
  named global variable.
- `<syscall.counted name="NAME" offset="NUMBER" />`, invokes a named
  system-function with a dynamic check of the number of values being passed.
  The old stack-length is passed in the numbered call-frame slot.
- `<call.global.counted name="NAME" offset="NUMBER" />`, invokes a named
  global variable with a dynamic check of the number of values being passed.
  The old stack-length is passed in the numbered call-frame slot.
- `<return />`, simply exits from a function normally.

A function-object is:

- An object with two integer fields: nlocals and nparams. 
- And a list of instructions.

The task is:

- [ ] To create a suitable DTO for rendering to JSON
- [ ] A conversion function for mapping from <fn> nodes to function-objects
- [ ] And then use that conversion function to serialise the function-object
      into the Binding.Value slot, instead of a Node.

