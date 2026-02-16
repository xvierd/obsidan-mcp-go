# MCP-Obsidian-Go: Self-Instructions

**Document for:** AI Assistant (myself)  
**Project:** go-obsidian-mcp  
**Purpose:** Clear instructions and guidelines for development

---

## 🎯 Project Mission

Build a **superior MCP server for Obsidian** in Go that:
1. Replaces py-obsidian-tools with better performance
2. Adds features impossible in Python (graph queries, watch mode)
3. Distributes as single binary (no dependencies)
4. Maintains 100% compatibility with existing MCP clients

---

## 📋 Development Rules

### 1. Always Follow Checklist
- **BEFORE starting work:** Check `CHECKLIST.md`
- **DURING work:** Update task status ([ ] → [~] → [x])
- **AFTER work:** Add notes to daily log section

### 2. Test-Driven Development
- **No code without tests** (except main.go entrypoint)
- Unit tests go in `*_test.go` files
- Mock external dependencies (HTTP, file system)
- Aim for 80%+ coverage

### 3. Code Quality Standards

#### Formatting
```bash
# Always run before commit
make fmt
make lint
make test
```

#### Structure
```go
// File header template
package x

import (
    // stdlib first
    // external second
    // internal third
)

// Types/Interfaces
// Constructors
// Methods
// Private helpers
// Tests
```

#### Error Handling
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to X: %w", err)
}

// Use custom error types for domain errors
var ErrNoteNotFound = errors.New("note not found")
```

#### Logging
```go
// Use structured logging always
slog.Info("operation_completed",
    "tool", toolName,
    "duration_ms", duration.Milliseconds(),
    "success", true,
)

// No fmt.Println in production code
```

### 4. Documentation Requirements
- **Every exported function** must have doc comment
- **Every tool** must have description and example
- **Every package** must have package-level doc
- Update README when adding features

### 5. Git Workflow
```bash
# Commit message format
type: short description

- Detailed point 1
- Detailed point 2

Types:
- feat: new feature
- fix: bug fix
- docs: documentation
- test: tests
- refactor: code refactoring
- perf: performance
- chore: maintenance
```

---

## 🏗️ Architecture Guidelines

### Package Structure
```
cmd/
  server/       # MCP server entrypoint
  indexer/      # CLI indexer entrypoint

internal/
  config/       # Configuration management
  obsidian/     # HTTP client for Obsidian API
  tools/        # MCP tools implementation
    registry.go
    vault.go
    search.go
    commands.go
    active.go
    vector.go
  vector/       # Vector search implementation
    store.go
    indexer.go
    embeddings/
  graph/        # Graph analysis (Phase 3)
  analytics/    # Vault analytics (Phase 3)
  cache/        # LRU cache implementation
  middleware/   # Logging, metrics, rate limiting

pkg/            # Public API (if needed)
```

### Key Interfaces (Must Define First)

```go
// ObsidianClient - HTTP client
 type ObsidianClient interface {
    GetNote(ctx context.Context, path string) (*Note, error)
    ListFiles(ctx context.Context, dir string) ([]string, error)
    // ... etc
}

// Tool - MCP tool
 type Tool interface {
    Name() string
    Description() string
    Schema() json.RawMessage
    Execute(ctx context.Context, params json.RawMessage) (any, error)
}

// VectorStore - Vector search
 type VectorStore interface {
    Index(notes []*Note) error
    Search(query string, limit int) ([]SearchResult, error)
    FindSimilar(notePath string, limit int) ([]SearchResult, error)
}

// EmbeddingProvider - Embeddings
 type EmbeddingProvider interface {
    Embed(texts []string) ([][]float32, error)
}
```

### Design Patterns to Use

1. **Registry Pattern** for tools
2. **Interface Segregation** for external deps
3. **Decorator Pattern** for middleware (logging, metrics)
4. **Circuit Breaker** for HTTP client
5. **LRU Cache** for note caching

---

## ⚠️ Technical Constraints

### Must Have
- [ ] stdio transport (Claude Desktop compatibility)
- [ ] Connection pooling for HTTP
- [ ] Structured logging (slog)
- [ ] Context propagation for cancellation
- [ ] Error wrapping with context
- [ ] Configurable timeouts

### Must NOT Have
- [ ] Global state (use dependency injection)
- [ ] fmt.Println (use slog)
- [ ] Panic in production code
- [ ] Hardcoded paths or credentials
- [ ] CGO (for easy cross-compilation)

### Performance Targets
- Startup: <100ms
- Memory: <50MB
- P99 latency: <50ms for cached reads
- Binary size: <25MB

---

## 🔧 Development Commands

### Setup
```bash
cd ~/projects/mcp-obsidian-go
make setup    # Install dependencies
make dev      # Run with hot reload
```

### Build
```bash
make build              # Local build
make build-all          # Cross-compile all platforms
make build-darwin-arm64 # macOS Apple Silicon
make build-linux-amd64  # Linux
```

### Test
```bash
make test           # Run all tests
make test-unit      # Unit tests only
make test-integration # Integration tests
make coverage       # Generate coverage report
```

### Quality
```bash
make fmt      # Format code
make lint     # Run linter
make vet      # Run go vet
make check    # Run all checks
```

### Run
```bash
make run      # Run server locally
make run-indexer full  # Run indexer
```

---

## 🧪 Testing Strategy

### Unit Tests
- Test each function in isolation
- Mock HTTP client with `httptest`
- Mock file system with `afero` or interfaces

### Integration Tests
- Test against real Obsidian instance
- Use test vault with known content
- Run in CI with Obsidian mock if possible

### E2E Tests
- Test with Claude Desktop
- Verify actual MCP protocol communication
- Test error scenarios

---

## 📚 Learning Resources

### MCP Protocol
- https://modelcontextprotocol.io/
- https://github.com/modelcontextprotocol/spec

### Go Best Practices
- https://go.dev/doc/effective_go
- https://github.com/uber-go/guide
- https://golang.org/wiki/CodeReviewComments

### Libraries
- mcp-go SDK: https://github.com/mark3labs/mcp-go
- SQLite: https://modernc.org/sqlite
- Testing: https://github.com/stretchr/testify

---

## 🚨 Common Pitfalls to Avoid

1. **Don't block in tool handlers** - Use goroutines for I/O
2. **Don't ignore context cancellation** - Check `ctx.Done()`
3. **Don't leak goroutines** - Use `errgroup` or `sync.WaitGroup`
4. **Don't use global variables** - Use dependency injection
5. **Don't forget to close resources** - Use `defer`
6. **Don't panic** - Return errors gracefully
7. **Don't hardcode timeouts** - Make them configurable
8. **Don't ignore errors** - Handle or explicitly ignore with comment

---

## ✅ Daily Routine

### Morning
1. Read CHECKLIST.md - what needs to be done today?
2. Update task status to [~] (in progress)
3. Review yesterday's notes

### During Work
1. Write tests first (TDD)
2. Implement feature
3. Run `make check`
4. Commit with descriptive message

### End of Day
1. Update CHECKLIST.md with completed tasks [x]
2. Add notes to daily log
3. Push to GitHub
4. Plan next day's work

---

## 🎯 Success Criteria

### Phase 0 Success
- [ ] Binary compiles and runs
- [ ] `server_status` works end-to-end
- [ ] Tests pass
- [ ] CI/CD working

### Phase 1 Success
- [ ] All 20+ core tools implemented
- [ ] 80%+ test coverage
- [ ] Caching working
- [ ] Parity with py-obsidian-tools

### Phase 2 Success
- [ ] Vector search working
- [ ] Multiple embedding providers
- [ ] Incremental indexing
- [ ] CLI indexer functional

### Phase 3 Success
- [ ] Graph queries working
- [ ] Analytics tools working
- [ ] Watch mode functional
- [ ] Superior to Python version

### Phase 4 Success
- [ ] Release v0.1.0
- [ ] Homebrew installable
- [ ] Docker image available
- [ ] Documentation complete

---

## 📝 Notes to Self

- **Keep it simple** - Don't over-engineer
- **Test everything** - No exceptions
- **Document as you go** - Not at the end
- **Commit often** - Small, focused commits
- **Ask for help** - If stuck for >30min
- **Take breaks** - Don't burn out

---

**Remember:** The goal is working software, not perfect architecture. Ship early, iterate often.

**Last Updated:** 2026-02-16
