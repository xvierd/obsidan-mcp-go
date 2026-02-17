# Hexagonal Architecture (Ports and Adapters)

This document describes the hexagonal architecture used in the MCP-Obsidian-Go project.

## Overview

The project follows the Hexagonal Architecture pattern (also known as Ports and Adapters), which separates the core business logic from external concerns like HTTP clients, databases, and user interfaces.

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                      │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Adapters                                              │  │
│  │  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐  │  │
│  │  │  Obsidian   │  │    Memory    │  │   SQLite    │  │  │
│  │  │   HTTP      │  │    Cache     │  │  (future)   │  │  │
│  │  │  Client     │  │              │  │             │  │  │
│  │  └─────────────┘  └──────────────┘  └─────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Services (Use Cases)                                  │  │
│  │  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐  │  │
│  │  │   Note      │  │   Search     │  │   Command   │  │  │
│  │  │  Service    │  │   Service    │  │   Service   │  │  │
│  │  └─────────────┘  └──────────────┘  └─────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Ports (Interfaces)                                    │  │
│  │  • NoteRepository         • CacheRepository          │  │
│  │  • ActiveNoteRepository   • SearchRepository         │  │
│  │  • CommandRepository      • ServerStatusRepository   │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              ▲
                              │
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                            │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Core Entities                                         │  │
│  │  • Note              • Command         • Tag         │  │
│  │  • SearchResult      • Link            • Task        │  │
│  │  • PatchRequest      • RecentChange    • Dataview    │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Domain Errors                                         │  │
│  │  • ErrNoteNotFound     • ErrUnauthorized             │  │
│  │  • ErrConnectionFailed • ErrTimeout                  │  │
│  │  • DomainError struct with IsNotFound(), etc.        │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Dependency Direction

Dependencies point **INWARD**:

```
Infrastructure ──► Application ──► Domain
     Adapters          Services        Entities
```

- **Domain**: No external dependencies. Pure business logic.
- **Application**: Depends only on Domain and Ports (interfaces).
- **Infrastructure**: Depends on Application (implements ports).

## Directory Structure

```
internal/
├── domain/                 # Core domain (NO external dependencies)
│   ├── note.go            # Entities: Note, Command, SearchResult, etc.
│   └── errors.go          # Domain errors
│
├── application/           # Use cases
│   ├── ports/             # Interfaces (driving and driven ports)
│   │   ├── note_repository.go
│   │   ├── cache_repository.go
│   │   ├── command_repository.go
│   │   ├── search_repository.go
│   │   └── server_repository.go
│   └── services/          # Business logic
│       ├── note_service.go
│       ├── search_service.go
│       └── command_service.go
│
└── infrastructure/        # External adapters
    └── adapters/
        ├── obsidian/      # HTTP adapter for Obsidian API
        │   ├── client.go
        │   └── client_test.go
        └── memory/        # In-memory cache adapter
            ├── cache.go
            └── cache_test.go
```

## Ports (Interfaces)

### NoteRepository
```go
type NoteRepository interface {
    GetNote(ctx context.Context, path string) (*domain.Note, error)
    ListNotes(ctx context.Context, directory string) ([]string, error)
    CreateNote(ctx context.Context, path string, content string) error
    UpdateNote(ctx context.Context, path string, content string) error
    DeleteNote(ctx context.Context, path string) error
    AppendNote(ctx context.Context, path string, content string) error
    PatchNote(ctx context.Context, path string, patch domain.PatchRequest) error
    OpenNote(ctx context.Context, path string) error
    GetRecentChanges(ctx context.Context, limit int) ([]domain.RecentChange, error)
    GetPeriodicNote(ctx context.Context, period string, offset int) (*domain.Note, error)
}
```

### ActiveNoteRepository
```go
type ActiveNoteRepository interface {
    GetActiveNote(ctx context.Context) (*domain.Note, error)
    UpdateActiveNote(ctx context.Context, content string) error
    AppendActiveNote(ctx context.Context, content string) error
    DeleteActiveNote(ctx context.Context) error
    PatchActiveNote(ctx context.Context, patch domain.PatchRequest) error
}
```

### SearchRepository
```go
type SearchRepository interface {
    Search(ctx context.Context, query string) ([]domain.SearchResult, error)
    ComplexSearch(ctx context.Context, query map[string]interface{}) ([]domain.SearchResult, error)
    DataviewQuery(ctx context.Context, query string) (*domain.DataviewResult, error)
}
```

### CommandRepository
```go
type CommandRepository interface {
    ListCommands(ctx context.Context) ([]domain.Command, error)
    ExecuteCommand(ctx context.Context, commandID string) error
}
```

### CacheRepository
```go
type CacheRepository interface {
    Get(key string) (*domain.Note, bool)
    Set(key string, value *domain.Note)
    Delete(key string)
    Clear()
    Stats() CacheStats
}
```

## Adapters

### Obsidian HTTP Adapter
Located in `internal/infrastructure/adapters/obsidian/client.go`

Implements:
- `NoteRepository`
- `ActiveNoteRepository`
- `SearchRepository`
- `CommandRepository`
- `ServerStatusRepository`

Translates domain operations to HTTP requests against the Obsidian Local REST API.

### Memory Cache Adapter
Located in `internal/infrastructure/adapters/memory/cache.go`

Implements:
- `CacheRepository`

Provides LRU caching with TTL support for notes.

## Services

Services orchestrate the use cases and depend only on ports (interfaces), not concrete implementations:

### NoteService
- Handles note CRUD operations
- Manages caching (invalidation on updates)
- Delegates to repositories

### SearchService
- Handles search operations
- Input validation

### CommandService
- Handles Obsidian command execution
- Input validation

## Dependency Injection

The `cmd/server/main.go` wires all dependencies:

```go
// Create infrastructure adapters
obsidianClient := obsidian.NewClient(apiKey, host, port)
cacheAdapter := memory.NewCache(size, ttl)

// Create application services
noteService := services.NewNoteService(
    obsidianClient,    // NoteRepository
    cacheAdapter,      // CacheRepository
    obsidianClient,    // ActiveNoteRepository
    logger,
)
searchService := services.NewSearchService(obsidianClient, logger)
commandService := services.NewCommandService(obsidianClient, logger)

// Create registry with services
registry := tools.NewRegistry(logger, noteService, searchService, commandService)
```

## Benefits of This Architecture

1. **Testability**: Services can be tested with mock implementations of ports
2. **Flexibility**: Easy to swap implementations (e.g., change cache from memory to Redis)
3. **Maintainability**: Clear separation of concerns
4. **Domain-Driven**: Domain logic is isolated and protected from external concerns
5. **Framework Independence**: Domain and Application don't depend on HTTP, JSON, etc.

## Adding New Functionality

1. **New Entity**: Add to `internal/domain/`
2. **New Operation**: 
   - Add method to appropriate port in `internal/application/ports/`
   - Implement in adapter(s) in `internal/infrastructure/adapters/`
   - Use in service in `internal/application/services/`
3. **New Adapter**: Implement existing ports with new technology
