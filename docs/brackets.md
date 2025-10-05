# Bracket Syntax

_Work in progress_

These notes provide a quick summary of the different types of parentheses that are used throughout Nutmeg.

## Construction

| Brackets        |   Data-type              |
|-----------------|--------------------------|
| [ x, y ]        | 1D-array aka list |
| [\| x, y \|]    | tuple |
| { "foo" => 99, "bar" => 88 } | map |
| {\| left => x , right => y \|} |  named-tuple |
| [% x, y %]      | chain |
| [: x, y :]      | stream aka iterator |
| {? x, y ?}      | set |


## Indexing

| Syntax     | Operation                |
|------------|--------------------------|
| x[ n ]     | subscript for array, tuple, masked array, stream, chain, set.  |
| x$[ 0.. ] | slicing for array, tuple, masked array, stream, chain, set. |
