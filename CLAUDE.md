# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Testing
- `go test ./...` - Run all tests across all packages
- `go test ./package_name` - Run tests for a specific package (e.g., `go test ./report`)
- `go test -v ./...` - Run all tests with verbose output
- `go test ./llm -run Mock` - Run only mock tests for llm package
- `OLLAMA_TEST_URL=http://localhost:11434 OLLAMA_TEST_MODEL=llama3.2:latest go test ./llm -run Integration` - Run integration tests

### Build
- `go build` - Build the main module
- `go mod tidy` - Clean up module dependencies

## Architecture

This is a modular Go utility library (`github.com/denis-kilchichakov/toolbox`) with distinct packages for common application needs:

### Package Structure
- **`llm/`** - LLM client library with Ollama support (Ask/Chat API)
- **`report/`** - Telegram notification system using github.com/nikoksr/notify
- **`secret/`** - AES-256-GCM encryption/decryption for secret management
- **`sqldb/`** - SQLite database wrapper with MD5-based migration tracking
- **`system/`** - Signal handling utilities for graceful shutdown

### Key Design Patterns
- Each package is self-contained with clear responsibilities
- Interface-based design for extensibility (e.g., `LLMClient` and `Model` interfaces)
- Custom error types for better error handling (ValidationError, ModelNotFoundError, APIError)
- Input validation before external API calls
- Uses testify for comprehensive unit testing with mocks
- SQLite migrations support both file-based and embedded approaches
- Type-safe secret handling with `WrappedSecret`/`UnwrappedSecret` types
- Migration state tracking prevents duplicate runs

### Dependencies
- **Database**: `github.com/mattn/go-sqlite3` for SQLite operations
- **Notifications**: `github.com/nikoksr/notify` for Telegram integration  
- **Testing**: `github.com/stretchr/testify` for assertions and mocking

### Testing Strategy
- **Mock tests** (fast, no external dependencies) using `httptest` for HTTP clients
- **Integration tests** (require external services) skip when env vars not set
- Packages use in-memory SQLite for database tests
- Custom mocks for external dependencies (notification services, LLM APIs)
- Tests cover both success paths and error conditions
- Migration tests verify idempotency
- Table-driven tests for comprehensive coverage

### Package-Specific Documentation
- Some packages have their own `CLAUDE.md` for detailed guidance (e.g., `llm/CLAUDE.md`)
- Package-specific documentation includes implementation patterns, testing strategies, and common pitfalls
- Refer to package CLAUDE.md when working within that package

## Versioning and Releases

When creating a new release:

1. **Test everything**: `go test ./...`
2. **Update version**: Increment based on changes (v0.0.x for patches/features)
3. **Commit changes**: Ensure all changes are committed
4. **Create tag**: `git tag v0.0.x` (e.g., v0.0.3)
5. **Push tag**: `git push origin v0.0.x`
6. **Users can import**: `go get github.com/denis-kilchichakov/toolbox@v0.0.x`

### Version Guidelines
- **v0.0.x**: Patch releases (bug fixes, minor improvements)
- **v0.x.0**: Minor releases (new features, non-breaking changes)
- **vx.0.0**: Major releases (breaking changes)

Example workflow:
```bash
# Make changes and test
go test ./...

# Commit changes
git add .
git commit -m "Add new feature"

# Create and push version tag
git tag v0.0.3
git push origin main
git push origin v0.0.3
```