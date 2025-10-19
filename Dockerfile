# Multi-stage build for Nutmeg compiler toolchain
FROM alpine:latest

# Install ca-certificates for HTTPS support
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -g 1000 nutmeg && \
    adduser -D -u 1000 -G nutmeg nutmeg

# Create directory for binaries
WORKDIR /usr/local/bin

# Copy all four binaries from the artifact download
COPY nutmeg-tokenizer /usr/local/bin/nutmeg-tokenizer
COPY nutmeg-parser /usr/local/bin/nutmeg-parser
COPY nutmeg-rewriter /usr/local/bin/nutmeg-rewriter
COPY nutmeg-convert-tree /usr/local/bin/nutmeg-convert-tree

# Make all binaries executable
RUN chmod +x /usr/local/bin/nutmeg-* && \
    chown nutmeg:nutmeg /usr/local/bin/nutmeg-*

# Switch to non-root user
USER nutmeg

# Set working directory for user
WORKDIR /workspace

# Default command shows help for tokenizer
CMD ["nutmeg-tokenizer", "--help"]
