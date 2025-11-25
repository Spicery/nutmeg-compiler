# Nutmeg Project Structure Specification

## Overview

This document specifies the file structure, module organization, and dependency management system for Nutmeg projects. Key design principles:

- **No distinction between libraries and applications** - only entry points differ
- **Modules are the unit of encapsulation** - files within modules are transparent
- **Code equals data** - no special treatment for different file types
- **Policy-based visibility** - tags define stability contracts, not access control
- **Convention over configuration** - sensible defaults with explicit overrides

## Project File Structure

### Basic Layout

```
my-nutmeg-project/
├── project.nutconf             # Project configuration (auto-loaded)
├── main.mod                    # Module (auto-discovered)
│   ├── data.csv                # Data files (no distinction from code)
│   ├── imports.nutport         # Import declarations for this module
│   ├── main.nutmeg             # Source files
│   └── utils.nutmeg
└── parser.mod                  # Another module
    ├── imports.nutport
    └── parser.nutmeg
```

- Note that "nutconf" is a contraction of "Nutmeg's Configuration Subset". This is
  a limited functional subset of Nutmeg that is used to write configuration
  constants.
  - If omitted, it is equivalent to having no external dependencies.

- Similarly "nutport" is a contraction of "Nutmeg's Import Format". This is a
  sequence of imports from other modules. This will be one of the least stable
  formats during the early development of Nutmeg. 
  - If omitted, it is equivalent to importing from the public section of the
    standard library (std).

The simplest possible Nutmeg project would therefore have a structure like
this:-

```
simplest-nutmeg-project/
└── main.mod                    # Module (auto-discovered)
    └─ main.nutmeg              # Source files
```

To invoke this simplest project would require a command like this - assuming 
that the `main` module defined a function `main`.

```
nutmeg run --project=simplest-nutmeg-project --entrypoint=main::main
```

### Key Conventions

#### Implicit (Auto-discovered)
- **`project.nutconf`** in project root is automatically loaded
- All **`.mod` folders** are automatically discovered as modules
- **Module name** = folder name (without `.mod` suffix)

#### Explicit
- **Additional configuration files** must be specified: `nutmegc --config dev.nutconf`
- **Dependencies** must be declared in configuration
- **Imports** must be declared in `imports.nutport`

### What's NOT Needed

The following folders are **NOT** part of Nutmeg's project structure:

- **`src/`** - Everything is source
- **`build/`** - Build output is a SQLite bundle
- **`tests/`** - Tests use `[unittest]` decoration, live with code
- **`assets/`** - Data files live alongside code files

## Project Configuration

### File: `project.nutconf`

Configuration uses a consytrained subset of Nutmeg syntax (not TOML/JSON/XML).
The file is evaluated at load-time with access to a filesystem snapshot,
enabling dynamic configuration.

```nutmeg
### Project metadata
project := {
    name: "my-project",
    version: "0.1.0",
    authors: ["Your Name <email@example.com>"],
    license: "Apache-2.0"
}

### Default visibility tag for unmarked definitions
default_tag := "#team"

### Entry points (what makes this runnable)
entry_points := [
    { module: "main", function: "main" },
    { module: "main", function: "start_server", name: "server" }
]

### External dependencies (cached centrally)
dependencies := {
    stdlib: { 
        git: "https://github.com/nutmeg-lang/stdlib", 
        tag: "v1.2.0" 
    },
    json_parser: { 
        git: "https://github.com/user/json-parser", 
        tag: "v0.5.1" 
    }
}

### Internal dependencies (via registry)
internal_dependencies := {
    shared_types: { registry: "shared-types" },
    common: { registry: "common-lib" }
}

### Development-only dependencies
dev_dependencies := {
    test_framework: { 
        git: "https://github.com/nutmeg-lang/test", 
        tag: "v1.0.0" 
    }
}
```

### Dynamic Configuration

Since `project.nutconf` is executable Nutmeg with filesystem access:

```nutmeg
### Auto-discover modules
modules := [ 
    for name in filesystem.list(".")
    where name.endsWith(".mod")
    do name.removeSuffix(".mod") 
    end 
]

### Conditional dependencies based on what exists
dependencies := {
    stdlib: { git: "...", tag: "v1.0" }
} ++ (
    if _ in filesystem.list(".") where _.endsWith("gpu_support.mod") then
        { cuda: { git: "https://github.com/nutmeg/cuda", tag: "v2.0" } }
    else
        {}
    endif
)

### Read version from file if it exists
project := {
    name: "my-project",
    version: if filesystem.exists("VERSION") then
        for content in filesystem.read("VERSION") afterwards "0.1.0"
        do content.trim() end
    else
        "0.1.0"
    endif
}
```

## Modules

### Module Structure

A module is a **folder with `.mod` suffix** containing source files and an optional `imports.nutport`.

```
parser.mod/
├─── imports.nutport          # Import declarations (optional)
├─── lexer.nutmeg            # Source files
├─── parser.nutmeg
├─── ast.nutmeg
└─── grammar.ebnf            # Data files (treated same as code)
```

### Files Within Modules

**Files within a module have no significance** - they're just a convenient division of text. All definitions in all `.nutmeg` files within a module are mutually visible.

```
main.mod/
├─── main.nutmeg       # Can use functions from utils.nutmeg
├─── utils.nutmeg      # Can use functions from helpers.nutmeg
└─── helpers.nutmeg    # All see each other automatically
```

The file split is purely for programmer convenience and organization.

### Nested Modules

Modules can be nested:

```
parser.mod/
├─── imports.nutport
├─── parser.nutmeg
└─── advanced.mod/           # Nested submodule
    ├─── imports.nutport
    └─── optimization.nutmeg
```

Nested modules are separate modules with their own namespace and imports.

## Visibility and Tags

### Tag-Based Visibility

Nutmeg uses **tags** to indicate intended stability and audience, not to enforce access control. Tags define social contracts.

### Standard Tags

| Tag | Audience | Stability Contract |
|-----|----------|-------------------|
| `#int` | This module's developers | Can change/delete freely |
| `#team` | All modules in this program | Can change/delete freely within program |
| `#pub` | Other programs by same team | Type must remain stable, implementation can change |
| `#common` | Partnering development teams | More stable than #pub |
| `#ext` | External users/consumers | Semantic versioning, no breaking changes |

### Tag Application

Tags can be applied at multiple levels:

**1. Project-wide default** (in `project.nutconf`):
```nutmeg
default_tag := "#team"
```

**2. Module-level** (folder naming):
```
main.#pub.mod/          # Everything defaults to #pub
utils.#int.mod/         # Everything defaults to #int
```

**3. File-level** (file naming):
```
parser.mod/
├─── public_api.#pub.nutmeg      # Defaults to #pub
└─── internals.#int.nutmeg       # Defaults to #int
```

**4. Definition-level** (explicit decoration):
```nutmeg
[#pub]
fn public_function() := ...

[#int]
fn internal_helper() := ...
```

**Priority:** Definition > File > Module > Project

### Custom Tags

While the standard tags are conventional, **tags are user-defined**. Projects
can create custom tags for their own purposes:

```nutmeg
[#experimental]
fn new_feature() := ...

[#deprecated]
fn old_api() := ...
```

### Policy Enforcement

**Freedom with accountability:**
- You can import ANY tag from ANY module (language doesn't prevent it)
- Your program is marked as "policy-violating" if you import inappropriate tags
- Consumers can reject dependencies that violate policy
- Policy violations propagate transitively up the dependency chain

```bash
$ nutmegc link
Error: Dependency 'sketchy-lib' violates policy:
  - Imports #int variables from 'parser' module
  
Run with --ignore-policy to override (not recommended)
```

## Import System

### File: `imports.nutport`

Each module can have an `imports.nutport` file declaring what it imports from other modules.

```nutmeg
### Basic import with shortcode
import #pub from parser as p
### Usage: p::readExpr, p::parseStmt

### Auto-naming (uses first component)
import #pub from stdlib
### Usage: stdlib::print, stdlib::map

### Optional qualification
import #pub from common optionally as c
### Usage: either print(...) or c::print(...)

### Multiple tag sets
import #pub, #exp from advanced_parser as adv
### Gets both #pub and #exp tagged variables

### Include (re-export)
include #pub from json_parser
### Variables become part of THIS module's exports
```

### Import Modes

| Mode | Syntax | Behavior |
|------|--------|----------|
| **Qualified** | `import #pub from parser as p` | Use as `p::readExpr` |
| **Auto-named** | `import #pub from parser` | Use as `parser::readExpr` |
| **Optional** | `import #pub from parser optionally as p` | Use as `readExpr` or `p::readExpr` |
| **Include** | `include #pub from parser` | Re-export as if defined here |

### Use Cases

**Creating a facade:**
```nutmeg
### facade.mod/imports.nutport
include #pub from json_parser
include #pub from xml_parser
include #pub from yaml_parser
### Consumers import facade, get all parsers
```

**Heavy usage patterns:**
```nutmeg
### Common library with optional qualification
import #pub from stdlib optionally as std
### Can write: print(...) or std::print(...) for clarity
```

**Avoiding conflicts:**
```nutmeg
import #pub from parser as p
import #pub from lexer as l
### Clear: p::parse() vs l::lex()
```

## Dependency Management

### External Dependencies

External dependencies are fetched via Git and **cached centrally**, not stored in the project.

**Cache location:** `~/.local/nutmeg/cache/`

```
~/.local/nutmeg/
└─── cache/
    └─── github.com/
        └─── nutmeg-lang/
            └─── stdlib/
                └─── v1.2.0/
```

**Declaration in `project.nutconf`:**
```nutmeg
dependencies := {
    stdlib: { 
        git: "https://github.com/nutmeg-lang/stdlib", 
        tag: "v1.2.0" 
    },
    ### Can also use branch or commit
    experimental: { 
        git: "https://github.com/user/experimental", 
        branch: "main" 
    },
    pinned: { 
        git: "https://github.com/user/utils", 
        commit: "abc123def456" 
    }
}
```

**Benefits:**
- Projects stay clean (no dependency folders to gitignore)
- Cache shared across all projects
- Easy to clean/flush cache without affecting projects
- Grep works on your code only

### Internal Dependencies

Internal dependencies reference other Nutmeg projects on the local filesystem, resolved via a **registry**.

**Registry files:**
- **System-wide:** `/etc/nutmeg/registry.nutconf`
- **User-local:** `~/.local/nutmeg/registry.nutconf` (overrides system)
- **Project-local:** `./nutmeg-local.nutconf` (overrides user)

**Registry format:**
```nutmeg
registry := {
    "shared-types": "/workspace/shared-types",
    "common-lib": "/workspace/common-lib"
}
```

**Declaration in `project.nutconf`:**
```nutmeg
internal_dependencies := {
    shared_types: { registry: "shared-types" },
    common: { registry: "common-lib" }
}
```

### Interactive Registration

When an unknown dependency is encountered, the tooling prompts interactively:

```bash
$ nutmegc build
Error: Unknown internal dependency "shared-types"
? Where is "shared-types" located? 
  [Browse...] or enter path: /home/user/workspace/shared-types

✓ Found valid Nutmeg project

? Register "shared-types" â†’ /home/user/workspace/shared-types
  Where should I save this?
   ◌ Just for this project (./nutmeg-local.nutconf)
   ● For me, all projects (~/.local/nutmeg/registry.nutconf)
   ◌ For my team (/etc/nutmeg/registry.nutconf)
  
✓ Registered! Continuing build...

💡 To learn more: nutmeg register --help
```

**Manual registration:**
```bash
nutmeg register shared-types /workspace/shared-types --user
nutmeg register common-lib https://github.com/team/common --team
nutmeg register --list
```

### Dependency Resolution Tradeoffs

| Approach | Breaks When | Fix Requires |
|----------|-------------|--------------|
| Absolute paths | Workspace moves | Update N projects Ã— M refs |
| Relative paths | Project moves | Update 1 file Ã— M refs |
| Registry | Workspace moves | Update 1 entry in 1 file ✓ |

Registry provides **single point of change** for maximum maintainability.

## Build Output

Nutmeg produces a **SQLite bundle** as build output, not a folder hierarchy.

```bash
$ nutmegc build
✓ Built: my-project.bundle
```

The bundle contains:
- Compiled modules
- Metadata
- Resources

This keeps the project directory clean and makes distribution simple.

## Complete Example

### Project Structure
```
calculator-app/
├─── project.nutconf
├─── app.#pub.mod/
│   ├─── imports.nutport
│   ├─── main.nutmeg
│   └─── ui.nutmeg
├─── math.#pub.mod/
│   ├─── imports.nutport
│   ├─── basic.nutmeg
│   └─── advanced.#int.nutmeg
└─── utils.#team.mod/
    ├─── imports.nutport
    └─── helpers.nutmeg
```

### project.nutconf
```nutmeg
project := {
    name: "calculator-app",
    version: "1.0.0",
    authors: ["Calculator Team"],
    license: "MIT"
}

default_tag := "#team"

entry_points := [
    { module: "app", function: "main" }
]

dependencies := {
    stdlib: { 
        git: "https://github.com/nutmeg-lang/stdlib", 
        tag: "v2.0.0" 
    }
}
```

### app.#pub.mod/imports.nutport
```nutmeg
import #pub from stdlib optionally as std
import #pub from math as m
import #team from utils
```

### math.#pub.mod/imports.nutport
```nutmeg
import #pub from stdlib as std
```

### utils.#team.mod/imports.nutport
```nutmeg
import #pub from stdlib
```

## Migration and Compatibility

### Adding to Existing Code

To add project structure to existing Nutmeg code:

1. Create `project.nutconf` with basic metadata
2. Organize files into `.mod` folders
3. Add `imports.nutport` to modules that need cross-module access
4. Tag definitions as needed (defaults apply if omitted)

### Gradual Adoption

- Start with one module
- Use default tags initially
- Add explicit tags as APIs stabilize
- Refine imports as structure becomes clear

## Tooling Commands

```bash
# Build project
nutmegc build

# Build with override config
nutmegc build --config dev.nutconf

# Register internal dependency
nutmeg register <name> <path> [--user|--team|--project]

# List registered dependencies
nutmeg register --list

# Check policy compliance
nutmegc check-policy

# Clean cache
nutmeg cache clean
```
