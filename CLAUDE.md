# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Testing
- `go test ./...` - Run all tests across all packages
- `go test ./package_name` - Run tests for a specific package (e.g., `go test ./report`)
- `go test -v ./...` - Run all tests with verbose output

### Build
- `go build` - Build the main module
- `go mod tidy` - Clean up module dependencies

## Architecture

This is a modular Go utility library (`github.com/denis-kilchichakov/toolbox`) with distinct packages for common application needs:

### Package Structure
- **`report/`** - Telegram notification system using github.com/nikoksr/notify
- **`secret/`** - AES-256-GCM encryption/decryption for secret management  
- **`sqldb/`** - SQLite database wrapper with MD5-based migration tracking
- **`system/`** - Signal handling utilities for graceful shutdown

### Key Design Patterns
- Each package is self-contained with clear responsibilities
- Uses testify for comprehensive unit testing with mocks
- SQLite migrations support both file-based and embedded approaches
- Type-safe secret handling with `WrappedSecret`/`UnwrappedSecret` types
- Migration state tracking prevents duplicate runs

### Dependencies
- **Database**: `github.com/mattn/go-sqlite3` for SQLite operations
- **Notifications**: `github.com/nikoksr/notify` for Telegram integration  
- **Testing**: `github.com/stretchr/testify` for assertions and mocking

### Testing Strategy
- Packages use in-memory SQLite for database tests
- Custom mocks for external dependencies (notification services)
- Tests cover both success paths and error conditions
- Migration tests verify idempotency