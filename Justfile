default:
    @just --list

test: functest unittest lint fmt-check gosec tidy build

functest:
    (cd functests && uv run python3 functest.py --quiet --tests */*.yaml)

unittest:
    go test ./...

lint:
    echo "Running linter..."
    @if command -v ~/.tools/ext/bin/golangci-lint >/dev/null 2>&1; then \
        ~/.tools/ext/bin/golangci-lint run; \
    elif command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run; \
    else \
        echo "golangci-lint not found, falling back to go vet"; \
        echo "To install golangci-lint locally, run: just install-golangci-lint"; \
        go vet ./...; \
    fi

gosec:
    @echo "Running security scanner..."
    @if command -v ~/go/bin/gosec >/dev/null 2>&1; then \
        ~/go/bin/gosec -quiet -fmt=text ./...; \
    elif command -v gosec >/dev/null 2>&1; then \
        gosec -quiet -fmt=text ./...; \
    else \
        echo "gosec not found, skipping security scan"; \
        echo "To install gosec, run: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
    fi

# Install golangci-lint to ~/.tools/ext/bin  
install-golangci-lint:
    @mkdir -p ~/.tools/ext/bin
    GOBIN=~/.tools/ext/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @echo "golangci-lint installed locally to this project in ~/.tools/ext/bin/"
    @echo "Note that ~/.tools/ext/bin is not assumed to be in your PATH"

fmt:
    go fmt ./...

# Check formatting without modifying files
fmt-check:
    ./.tools/bin/go-fmt-check

tidy:
    go mod tidy

build:
    mkdir -p bin
    go build -o bin/nutmeg-tokenizer ./cmd/nutmeg-tokenizer
    go build -o bin/nutmeg-parser ./cmd/nutmeg-parser
    go build -o bin/nutmeg-rewriter ./cmd/nutmeg-rewriter
    go build -o bin/nutmeg-resolver ./cmd/nutmeg-resolver
    go build -o bin/nutmeg-common ./cmd/nutmeg-common
    go build -o bin/nutmeg-convert-tree ./cmd/nutmeg-convert-tree

install:
    go install ./cmd/nutmeg-tokenizer
    go install ./cmd/nutmeg-parser
    go install ./cmd/nutmeg-rewriter
    go install ./cmd/nutmeg-resolver
    go install ./cmd/nutmeg-common
    go install ./cmd/nutmeg-convert-tree

# Copy the rewrite rules over.
rules:
    @echo "Generating default-rewrite-rules.go from configs/rewrite.yaml..."
    @echo 'package rewriter' > pkg/rewriter/default-rewrite-rules.go
    @echo '' >> pkg/rewriter/default-rewrite-rules.go
    @echo 'const DefaultRewriteRules = `' >> pkg/rewriter/default-rewrite-rules.go
    @cat configs/rewrite.yaml >> pkg/rewriter/default-rewrite-rules.go
    @echo '`' >> pkg/rewriter/default-rewrite-rules.go
    @echo "Done! Generated pkg/rewriter/default-rewrite-rules.go"
