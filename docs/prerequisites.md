# Prerequisites for this project

## Summary

This project requires:

- **Go 1.24+** - [https://go.dev/](https://go.dev/)
- **Just** (command runner) - [https://github.com/casey/just](https://github.com/casey/just)
- **Python 3.x with uv** - [https://docs.astral.sh/uv/](https://docs.astral.sh/uv/)

Optional but recommended:

- **gosec** (security scanner) - [https://github.com/securego/gosec](https://github.com/securego/gosec)
- **golangci-lint** (Go linter) - [https://github.com/golangci/golangci-lint](https://github.com/golangci/golangci-lint)

### Quick Install (Linux)

For Debian/Ubuntu Linux and macOS systems, you can use the jumpstart script to install all prerequisites:

```bash
./.tools/bin/jumpstart.sh
```

This will automatically install all required and optional tools. On macOS, it will use Homebrew (and install it if needed).

---

## Component-by-component installation guidance

### Go 1.24 or later
The Nutmeg compiler toolchain is written in Go and requires Go 1.24 or later.

**Check if installed:**
```bash
go version
```

Expected output: `go version go1.24.x ...` or higher. If the version is lower than 1.24, you need to upgrade.

**Installation:**
- Download from [https://go.dev/dl/](https://go.dev/dl/)
- Or use your system package manager

### Just (command runner)
The project uses [Just](https://github.com/casey/just) as a command runner (similar to Make but simpler).

**Check if installed:**
```bash
just --version
```

Expected output: `just x.y.z` (any version should work)

**Installation:**
```bash
# On macOS
brew install just

# On Linux
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/bin

# Or install via cargo
cargo install just
```

### Python 3.x with uv (for functional tests)
The functional test suite uses Python and the `uv` package manager.

**Check if installed:**
```bash
python3 --version
uv --version
```

Expected: Python 3.6+ and any version of uv. Most modern systems have Python 3 pre-installed.

**Installation:**
```bash
# Install uv (will handle Python if needed)
curl -LsSf https://astral.sh/uv/install.sh | sh

# uv will automatically manage Python environments and dependencies
```

## Optional (but recommended)

### gosec (security scanner)
Gosec is used to scan the codebase for common security issues.

**Check if installed:**
```bash
gosec --version
# Or check in Go bin directory
~/go/bin/gosec --version
```

Expected output: `gosec x.y.z` (any recent version should work)

**Installation:**
```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

The `just test` command will automatically run gosec if available, but will skip it with a warning if not installed.

### golangci-lint (Go linter)
A comprehensive Go linter that runs multiple linters in parallel.

**Check if installed:**
```bash
golangci-lint --version
# Or check in project directory
~/.tools/ext/bin/golangci-lint --version
```

Expected output: `golangci-lint has version x.y.z` (any recent version should work)

**Installation:**
```bash
# Install locally to the project
just install-golangci-lint

# Or install globally
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

The `just test` command will fall back to `go vet` if golangci-lint is not available.

## Quick Start

Once prerequisites are installed:

```bash
# Clone the repository
git clone https://github.com/Spicery/nutmeg-compiler.git
cd nutmeg-compiler

# Run all tests
just test

# Build binaries
just build

# Install binaries to $GOPATH/bin
just install
```

## Available Commands

Run `just` without arguments to see all available commands:

```bash
just
```

Common commands:
- `just test` - Run all tests (functional, unit, linting, security scan)
- `just build` - Build all binaries to `./bin/`
- `just install` - Install binaries to `$GOPATH/bin`
- `just functest` - Run functional tests only
- `just unittest` - Run unit tests only
- `just lint` - Run linter only
- `just gosec` - Run security scanner only
- `just fmt` - Format Go code
