# Bundler

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

That's a good start. 
