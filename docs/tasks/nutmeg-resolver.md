This task is concerned with adding a new standalone command that will be
part of the Nutmeg toolchain: namely nutmeg-resolver. The job of this 
tool is to read in a unit-node in JSON format, to scan it for identifers, 
and to annotate them with the following information:

- A unique identifier ID, so that IDs with the same name but in different 
  scopes can be easily distinguished.
- Their scope, which is one of inner, outer or global, where inner and 
  outer are differently scoped locals.
- Whether this id-node is the definition or a use of the definition.

This command will share the same basic options as the other nutmeg-XXX 
commands and how no configuration file.

    --help
    --version
    --input FILE
    --output FILE
    --format, -f FORMAT
    --no-spans
    --trim INT


