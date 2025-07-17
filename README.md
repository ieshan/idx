# IDX - ULID-based ID Management Library

A Go library providing ULID-based unique identifier management with comprehensive database support.

## Features

- **ULID-based IDs**: Universally Unique Lexicographically Sortable Identifiers
- **JSON Support**: Built-in JSON marshaling/unmarshaling
- **Database Compatibility**: Works with MongoDB, MySQL, PostgreSQL, and SQLite
- **Type Safety**: Strong typing with Go's type system
- **Zero Dependencies**: Core library only depends on `github.com/oklog/ulid/v2`
- **Clean Architecture**: Uses Go workspaces to completely separate test dependencies

## Installation

```bash
go get github.com/ieshan/idx
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/ieshan/idx"
)

func main() {
    // Create a new ID
    id := idx.NewID()
    fmt.Println("New ID:", id.String())
    
    // Parse from string
    parsed, err := idx.FromString(id.String())
    if err != nil {
        panic(err)
    }
    
    // Compare IDs
    if id.Compare(parsed) == 0 {
        fmt.Println("IDs are equal")
    }
    
    // Check if zero
    if id.IsZero() {
        fmt.Println("ID is zero")
    }
}
```

## Testing

This library uses **Go workspaces** to completely separate database dependencies from the core library, ensuring zero pollution for downstream projects.

### Unit Tests (Fast - Zero Database Dependencies)

```bash
# Run unit tests only - ultra fast (1-2ms)
go test -v

# Or using Docker
docker-compose --profile unit up unit-tests --abort-on-container-exit
```

### Integration Tests (Separate Module with Database Dependencies)

```bash
# Run integration tests from workspace
cd integration-tests && go test -v

# Or using Docker
docker-compose --profile integration up integration-tests --abort-on-container-exit
```

### All Tests

```bash
# Run both unit and integration tests
docker-compose --profile all up all-tests --abort-on-container-exit

# Or the default service (backward compatibility)
docker-compose up exec-test --abort-on-container-exit
```

## Go Workspace Architecture

This project uses Go workspaces to achieve true dependency separation:

```
idx/
├── go.work                   # Workspace configuration
├── go.mod                    # Core library (minimal dependencies)
├── idx.go                    # Core library implementation
├── idx_test.go              # Unit tests (no databases)
├── integration-tests/        # Separate module for database tests
│   ├── go.mod               # Database dependencies isolated here
│   └── idx_integration_test.go # Database integration tests
└── docker-compose.yml       # Multi-profile testing setup
```

### Why Go Workspaces?

**The Problem with Build Tags**: Even with `//go:build integration` tags, Go still parses import statements and forces database dependencies into the main `go.mod`, polluting downstream projects.

**The Workspace Solution**: 
- **Main module** (`github.com/ieshan/idx`): Only `github.com/oklog/ulid/v2`
- **Test module** (`idx-tests`): All database drivers isolated here
- **True Separation**: Import statements in integration tests don't affect main module

### For Library Consumers

When someone runs `go get github.com/ieshan/idx`, they get:
- ✅ **Only** `github.com/oklog/ulid/v2` (core dependency)
- ❌ **No** MongoDB drivers, GORM, or any database dependencies
- ⚡ **Zero** dependency pollution

### For Developers/Maintainers

```bash
# Fast development cycle
go test                              # 1-2ms unit tests

# Full testing when needed  
cd integration-tests && go test     # 100ms database tests

# Docker-based testing
docker-compose --profile all up all-tests --abort-on-container-exit
```

## Docker Compose Services

The enhanced `docker-compose.yml` provides multiple testing scenarios:

| Service | Command | Purpose | Speed |
|---------|---------|---------|-------|
| `unit-tests` | `docker-compose --profile unit up unit-tests --abort-on-container-exit` | Fast unit tests only | ⚡ ~2ms |
| `integration-tests` | `docker-compose --profile integration up integration-tests --abort-on-container-exit` | Database integration tests | 🔧 ~100ms |
| `all-tests` | `docker-compose --profile all up all-tests --abort-on-container-exit` | Sequential unit + integration | 📊 Complete |
| `exec-test` | `docker-compose up exec-test --abort-on-container-exit` | Default (backward compatibility) | 🔄 Legacy |

### Key Features

1. **Health Checks**: Database services include health checks to ensure they're ready before tests run
2. **Workspace Support**: Commands handle both modules correctly
3. **Profiles**: Different Docker Compose profiles for different testing scenarios
4. **Clear Output**: Tests include descriptive banners showing what's running
5. **True Isolation**: Unit tests run without any database dependencies

## Database Support

The library is tested against:
- **MongoDB** 8.0.11 (CRUD operations with BSON)
- **MySQL/MariaDB** 11.8.2 (Binary ID storage)
- **PostgreSQL** 17.5 (BYTEA column support)
- **SQLite** (In-memory, BLOB storage)

Each database test performs comprehensive CRUD operations to ensure compatibility.

## Performance

- **Unit Tests**: ~1-2ms (zero database dependencies)
- **Integration Tests**: ~100ms (includes database setup/teardown)
- **Memory Usage**: Minimal - IDs are 16-byte arrays

## Dependency Management Comparison

### ❌ Before (Build Tags - Still Polluted)
```go
// Even with build tags, this pollutes go.mod:
//go:build integration

import (
    "go.mongodb.org/mongo-driver/v2/mongo"  // ← Forces into main go.mod
    "gorm.io/gorm"                          // ← Forces into main go.mod
)
```

### ✅ After (Go Workspaces - True Isolation)
```
Main module go.mod:
require github.com/oklog/ulid/v2 v2.1.1  # ← Only this!

integration-tests/go.mod:
require (
    github.com/ieshan/idx v0.0.0-...        # ← Local reference
    go.mongodb.org/mongo-driver/v2 v2.2.2   # ← Isolated here
    gorm.io/driver/mysql v1.6.0             # ← Isolated here
    gorm.io/driver/postgres v1.6.0          # ← Isolated here
    gorm.io/driver/sqlite v1.6.0            # ← Isolated here
    gorm.io/gorm v1.30.0                    # ← Isolated here
)
```

## Contributing

1. Make sure unit tests pass: `go test -v`
2. Make sure integration tests pass: `cd integration-tests && go test -v`
3. Or use Docker: `docker-compose --profile all up all-tests --abort-on-container-exit`

## License

Private
