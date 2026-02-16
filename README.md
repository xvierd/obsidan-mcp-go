# MCP-Obsidian-Go

A high-performance MCP (Model Context Protocol) server for Obsidian, written in Go.

## Overview

This project is a Go reimplementation of [py-obsidian-tools](https://github.com/rmc8/py-obsidian-tools) with significant performance improvements and additional features.

### Key Features

- 🚀 **Fast**: <100ms startup, parallel batch operations
- 📦 **Single Binary**: No dependencies, easy distribution
- 🔍 **Vector Search**: Semantic search with multiple embedding providers
- 🕸️ **Graph Queries**: Analyze note relationships
- 👁️ **Watch Mode**: Auto-indexing on file changes
- 📊 **Analytics**: Vault insights and metrics

## Quick Start

### Installation

```bash
# Homebrew (macOS/Linux)
brew tap xvierd/mcp-obsidian-go
brew install mcp-obsidian-go

# Or download binary from releases
curl -L https://github.com/xvierd/mcp-obsidian-go/releases/latest/download/mcp-obsidian-go-$(uname -s)-$(uname -m) -o mcp-obsidian-go
chmod +x mcp-obsidian-go
```

### Configuration

1. Install [Obsidian Local REST API](https://github.com/coddingtonbear/obsidian-local-rest-api) plugin
2. Copy your API key from Obsidian settings
3. Configure Claude Desktop:

```json
{
  "mcpServers": {
    "obsidian": {
      "command": "mcp-obsidian-go",
      "env": {
        "OBSIDIAN_API_KEY": "your-api-key",
        "OBSIDIAN_HOST": "127.0.0.1",
        "OBSIDIAN_PORT": "27124"
      }
    }
  }
}
```

## Development

### Prerequisites

- Go 1.23+
- Make
- Obsidian with Local REST API plugin

### Build

```bash
make build
```

### Test

```bash
make test
```

### Run

```bash
make run
```

## Architecture

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed design documentation.

## Roadmap

- [x] Phase 0: Foundation
- [ ] Phase 1: Core Features
- [ ] Phase 2: Vector Search
- [ ] Phase 3: Advanced Features
- [ ] Phase 4: Distribution

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by [py-obsidian-tools](https://github.com/rmc8/py-obsidian-tools) by rmc8.
