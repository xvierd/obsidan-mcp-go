# MCP-Obsidian-Go

A high-performance MCP (Model Context Protocol) server for Obsidian, written in Go.

## Overview

This project is a Go reimplementation of [py-obsidian-tools](https://github.com/rmc8/py-obsidian-tools) with significant performance improvements.

### Key Features

- 🚀 **Fast**: <100ms startup, parallel batch operations
- 📦 **Single Binary**: No dependencies, easy distribution
- 💾 **Caching**: LRU cache with TTL for improved performance
- ✅ **Complete**: 23 tools covering all Obsidian operations

## Prerequisites

1. **Obsidian** with the [Local REST API](https://github.com/coddingtonbear/obsidian-local-rest-api) plugin installed
2. **Go 1.23+** (for building from source)
3. **API Key** from the Obsidian Local REST API plugin settings

## Installation

### Quick Install (Recommended)

```bash
# Download and run installer (no root required)
curl -fsSL https://raw.githubusercontent.com/xvierd/obsidan-mcp-go/main/install.sh | bash

# Or if you have the repository locally:
./install.sh
```

The installer will:
1. Detect your platform automatically
2. Install the binary to `~/.local/bin/` (no sudo needed)
3. Create config directory at `~/.config/mcp-obsidian/`
4. Show you the exact JSON to add to Claude Desktop

### Uninstall

```bash
./uninstall.sh
```

### From Source

```bash
# Clone the repository
git clone https://github.com/xvierd/obsidan-mcp-go.git
cd obsidan-mcp-go

# Build the binary
make build

# The binary will be at build/obsidan-mcp-go
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OBSIDIAN_API_KEY` | **Required.** Your Obsidian Local REST API key | - |
| `OBSIDIAN_HOST` | Host where Obsidian is running | `127.0.0.1` |
| `OBSIDIAN_PORT` | Port for Obsidian REST API | `27124` |
| `MCP_LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |
| `CACHE_ENABLED` | Enable note caching | `true` |
| `CACHE_TTL` | Cache TTL duration | `30s` |
| `CACHE_SIZE` | Maximum cache size | `1000` |

## Claude Desktop Configuration

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
    "obsidian": {
      "command": "/path/to/obsidan-mcp-go",
      "env": {
        "OBSIDIAN_API_KEY": "your-api-key-here",
        "OBSIDIAN_HOST": "127.0.0.1",
        "OBSIDIAN_PORT": "27124"
      }
    }
  }
}
```

## Available Tools (23 total)

### Vault Operations (8 tools)
- `server_status` - Check if Obsidian server is running
- `list_notes` - List notes in vault or directory
- `read_note` - Read note content
- `create_note` - Create new note
- `update_note` - Update existing note
- `append_note` - Append content to note
- `delete_note` - Delete note
- `patch_note` - Patch specific section

### Search Operations (4 tools)
- `search_notes` - Simple text search
- `complex_search` - JsonLogic-based complex search
- `dataview_query` - Execute Dataview queries
- `batch_read_notes` - Read multiple notes in parallel

### Active Note Operations (6 tools)
- `get_active_note` - Get currently open note
- `update_active_note` - Update active note
- `append_active_note` - Append to active note
- `patch_active_note` - Patch active note
- `delete_active_note` - Delete active note
- `open_note` - Open note in Obsidian

### Command Operations (2 tools)
- `list_commands` - List available Obsidian commands
- `execute_command` - Execute Obsidian command

### Special Operations (2 tools)
- `get_recent_changes` - Get recently modified files
- `get_periodic_note` - Get daily/weekly/monthly notes

## Development

### Build

```bash
make build        # Local build
make test         # Run tests
make coverage     # Coverage report
```

### Testing

```bash
go test ./...     # Run all tests
```

## Project Status

✅ **COMPLETE** - Phase 0 & 1 implemented

- 23 MCP tools working
- Full parity with py-obsidian-tools
- ~7MB binary
- No external dependencies

**Not implemented:** Vector search (requires 5GB model - too large)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by [py-obsidian-tools](https://github.com/rmc8/py-obsidian-tools) by rmc8.
