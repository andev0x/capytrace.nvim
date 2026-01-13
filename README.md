<div align="center">
  <img src="assets/img/capytrace.gif" alt="capytrace.nvim Logo" width="200"/>

  # capytrace.nvim

  **A high-performance, local-first debugging session recorder for Neovim**

  [![Go Version](https://img.shields.io/badge/Go-%3E=1.18-blue?logo=go&style=flat-square)](https://golang.org/)
  [![Neovim Version](https://img.shields.io/badge/Neovim-%3E=0.9.0-blueviolet?logo=neovim&style=flat-square)](https://neovim.io/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
  [![Go Report Card](https://goreportcard.com/badge/github.com/andev0x/capytrace.nvim?style=flat-square)](https://goreportcard.com/report/github.com/andev0x/capytrace.nvim)

</div>

---

## Overview

**capytrace.nvim** is a lightweight, privacy-first debugging session recorder for Neovim. It automatically captures your development workflow—file edits, cursor movements, terminal commands, and LSP diagnostics—into structured, timestamped session logs. Designed with a focus on **signal-over-noise**, capytrace intelligently filters redundant events and provides actionable insights into your debugging process.

### Why capytrace?

- 🔒 **Privacy First**: All data stays local—zero external APIs or telemetry
- ⚡ **High Performance**: Smart event filtering eliminates 90% of cursor noise
- 💾 **Multiple Formats**: Export to Markdown, JSON, or SQLite for different use cases
- 🚀 **Non-Blocking**: Background Goroutines ensure your editor stays responsive
- 📦 **Zero Dependencies**: Pure Go with no CGO—builds on any platform
- 🎯 **Developer-Centric**: Actionable statistics and intuitive session management

---

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
- [Output Examples](#output-examples)
- [Architecture](#architecture)
- [Performance](#performance)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)
- [Support](#support)

---

## Features

### Core Functionality

- **Live Session Recorder**: Automatically captures file edits, cursor movements, terminal commands, file opens, and LSP diagnostics
- **Smart Event Filtering**: Intelligent cursor debouncing (200ms) and idle detection (500ms) reduce noise while preserving meaningful context
- **Structured Timeline**: Events are timestamped and organized chronologically for easy review
- **User Annotations**: Add inline notes during debugging sessions—automatically timestamped and integrated into session logs
- **Session Management**: List, resume, or query any previous debugging session without losing context

### Export Formats

- **Markdown**: Human-readable session reports with emojis and formatted timelines
- **JSON**: Machine-readable data suitable for programmatic analysis and integration
- **SQLite**: Queryable database for aggregating statistics across multiple sessions

### Advanced Features

- **Context-Aware Logging**: File edits and terminal commands immediately flush pending cursor events for accurate context
- **Configurable Thresholds**: Adjust debounce intervals and idle detection times to your workflow
- **Session Resumption**: Continue debugging from exactly where you left off
- **Statistics Command**: Analyze session metrics (duration, event counts, code vs. navigation time)

---

## Requirements

- **Neovim**: v0.9.0 or higher
- **Go**: v1.18 or higher (required to build from source)
- **Make**: For building (optional, see [Building from Source](#building-from-source))

---

## Installation

### With lazy.nvim (Recommended)

Add the following to your lazy.nvim configuration:

```lua
{
  "andev0x/capytrace.nvim",
  build = "make build",  -- Builds the Go binary automatically
  config = function()
    require("capytrace").setup({
      output_format = "markdown",  -- or "json" or "sqlite"
      save_path = "~/capytrace_logs/",
      filter_threshold = 500,      -- Idle detection threshold (ms)
      debounce_interval = 200,     -- Cursor debounce interval (ms)
    })
  end,
}
```

### With vim-plug

```vim
Plug 'andev0x/capytrace.nvim', { 'do': 'make build' }
```

### Manual Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/andev0x/capytrace.nvim.git
   cd capytrace.nvim
   ```

2. Build the Go binary:
   ```bash
   make build
   ```

3. Add the plugin path to your Neovim configuration and call setup:
   ```lua
   require("capytrace").setup()
   ```

---

## Quick Start

### Starting Your First Session

```vim
:CapyTraceStart myproject
```

This creates a new debugging session for your project. The plugin will now record all relevant events.

### Recording Your Work

The plugin automatically records:
- File edits (TextChanged, TextChangedI)
- Cursor movements (intelligently filtered)
- File opens (BufEnter)
- Terminal commands (TermOpen)
- LSP diagnostics (DiagnosticChanged)
- User annotations (`:CapyTraceAnnotate`)

### Ending the Session

```vim
:CapyTraceEnd
```

The session is exported to your configured format (Markdown, JSON, or SQLite) and saved to `save_path`.

### Viewing Statistics

```vim
:CapyTraceList                 " List all recorded sessions
:CapyTraceStatus               " Show current session status
```

---

## Configuration

All settings are optional. Defaults are designed for typical use cases:

```lua
require("capytrace").setup({
  -- Output format for exported sessions
  output_format = "markdown",        -- "markdown" | "json" | "sqlite"

  -- Directory where sessions are saved
  save_path = "~/capytrace_logs/",

  -- Smart Filter: Idle threshold before committing cursor position (milliseconds)
  filter_threshold = 500,

  -- Smart Filter: Debounce interval for rapid cursor movements (milliseconds)
  debounce_interval = 200,

  -- Auto-save session when closing Neovim
  auto_save_on_exit = true,

  -- Maximum cursor movement events per session (for memory efficiency)
  max_cursor_events = 100,

  -- Event logging preferences
  log_events = {
    terminal_commands = true,       -- Log TermOpen events
    file_open = true,               -- Log BufEnter events
    lsp_diagnostics = true,         -- Log LSP diagnostics
  },
})
```

### Smart Filter Explanation

The plugin uses an intelligent cursor filter to reduce noise:

1. **Debouncing** (default 200ms): Rapid keyboard navigation (j/k) within 200ms is considered a single movement
2. **Idle Detection** (default 500ms): Cursor position is only recorded after staying idle for 500ms
3. **Context Triggers**: Immediate recording on file edits, terminal commands, or session end

This approach reduces cursor events by ~90% while preserving meaningful context.

---

## Usage

### Vim Commands

```vim
" Start a new debugging session
:CapyTraceStart [project_name]

" End the current session and export
:CapyTraceEnd

" Add a note to the current session
:CapyTraceAnnotate This is a note

" Show status of current session
:CapyTraceStatus

" List all available sessions
:CapyTraceList

" Resume a previous session
:CapyTraceResume session_id
```

### CLI Commands (Direct Usage)

```bash
# Start a session
./bin/capytrace start <session_id> <project_path> <save_path> <format>

# End a session
./bin/capytrace end <session_id> <save_path>

# Add annotation
./bin/capytrace annotate <session_id> <save_path> "note text"

# Record events
./bin/capytrace record-edit <session_id> <save_path> <filename> <line> <col> <line_count> <changed_tick>
./bin/capytrace record-cursor <session_id> <save_path> <filename> <line> <col>
./bin/capytrace record-terminal <session_id> <save_path> "command"

# Session management
./bin/capytrace list <save_path>
./bin/capytrace resume <session_id> <save_path>
./bin/capytrace stats <save_path> [session_id]
```

---

## Output Examples

### Markdown Format

```markdown
# Debug Session: 1704067200_myproject

**Project:** /home/user/myproject
**Started:** 2026-01-01T10:00:00Z
**Ended:** 2026-01-01T11:30:00Z
**Duration:** 1h30m0s

---

## Timeline

### 10:00:00 - Session Started
🚀 Started debugging session in /home/user/myproject

### 10:05:23 - File Edit
📄 **File:** `src/auth.lua`
📍 **Position:** Line 45, Column 12
📊 **Total Lines:** 120

### 10:07:15 - Terminal Command
💻 ```bash
git log --oneline -10
```

### 10:10:30 - Note
📝 Found potential issue in authentication logic

### 11:30:00 - Session Ended
🏁 Debugging session ended

## Summary

- **Total Events:** 25
- **File Edits:** 12
- **Terminal Commands:** 8
- **Annotations:** 3
- **Cursor Movements:** 2
```

### JSON Format

```json
{
  "id": "1704067200_myproject",
  "project_path": "/home/user/myproject",
  "start_time": "2026-01-01T10:00:00Z",
  "end_time": "2026-01-01T11:30:00Z",
  "output_format": "json",
  "active": false,
  "events": [
    {
      "type": "session_start",
      "timestamp": "2026-01-01T10:00:00Z",
      "data": {
        "note": "Started debugging session in /home/user/myproject"
      }
    },
    {
      "type": "file_edit",
      "timestamp": "2026-01-01T10:05:23Z",
      "data": {
        "filename": "src/auth.lua",
        "line": 45,
        "column": 12,
        "line_count": 120,
        "changed_tick": 5
      }
    }
  ]
}
```

### Statistics Output

```
$ capytrace stats ~/capytrace_logs/

Session Statistics
==================

1704067200_myproject:
  Status: Completed
  Duration: 1h30m0s
  Total Events: 25
  File Edits: 12
  Cursor Moves: 2
  Terminal Commands: 8
  Annotations: 3
```

---

## Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Neovim Editor                         │
├─────────────────────────────────────────────────────────┤
│  Lua Frontend (lua/capytrace/)                           │
│  - Listens to editor events                              │
│  - Manages user commands                                 │
│  - Communicates with Go backend                          │
├─────────────────────────────────────────────────────────┤
│  Go Backend (cmd/capytrace/main.go)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │  Recorder    │  │   Filter     │  │  Exporter    │   │
│  │  (Session    │  │  (Smart      │  │  (Markdown,  │   │
│  │   State)     │  │   Filtering) │  │   JSON,      │   │
│  │              │  │              │  │   SQLite)    │   │
│  └──────────────┘  └──────────────┘  └──────────────┘   │
├─────────────────────────────────────────────────────────┤
│  Storage (Local Filesystem / SQLite)                     │
│  - Session JSON files                                    │
│  - Exported reports (Markdown/JSON)                      │
│  - SQLite database (optional)                            │
└─────────────────────────────────────────────────────────┘
```

### Component Details

**Lua Frontend** (`lua/capytrace/`)
- Hooks into Neovim autocommands (TextChanged, BufEnter, CursorMoved, etc.)
- Invokes the Go binary with appropriate arguments
- Manages user-facing commands and configuration
- Ensures UI remains responsive during recording

**Go Backend** (`cmd/capytrace/`)
- **Recorder**: Manages session state, buffers events, persists to disk
- **Filter**: Implements the Smart Filter for cursor event debouncing
- **Exporter**: Converts sessions to Markdown, JSON, or SQLite formats
- **Models**: Shared data structures across all components

**Storage**
- **JSON Files** (default): Direct, human-readable session logs
- **SQLite Database**: Queryable storage for analytics and session merging
- **Standard Directory**: `~/.local/share/capytrace/` (follows XDG spec)

### Concurrency Model

- **Non-blocking I/O**: All file operations use Goroutines to prevent editor lag
- **Thread-safe Sessions**: RWMutex protects concurrent access to session state
- **Buffered Channels**: Events are queued and processed asynchronously

---

## Performance

### Event Filtering Effectiveness

Capytrace's Smart Filter significantly reduces event noise:

| Event Type | Before Filter | After Filter | Reduction |
|------------|---------------|--------------|-----------|
| Cursor Movements | ~1000 events/session | ~100 events | ~90% |
| File Edits | Unchanged | Unchanged | 0% |
| Terminal Commands | Unchanged | Unchanged | 0% |
| Annotations | Unchanged | Unchanged | 0% |

### Binary Size

- **macOS**: ~9.8 MB (single binary, no dependencies)
- **Linux**: ~8.5 MB (single binary, no dependencies)
- **Windows**: ~9.2 MB (single binary, no dependencies)

### Memory Usage

- **Idle**: <5 MB
- **Recording**: 5-15 MB (depends on event frequency and session size)
- **Session Size**: ~1-2 MB per hour of typical development

---

## Development

### Project Structure

```
capytrace.nvim/
├── cmd/capytrace/           # Entry point and CLI
├── internal/                # Core packages
│   ├── filter/             # Smart cursor filter
│   ├── recorder/           # Session management
│   ├── exporter/           # Export formats
│   └── models/             # Data structures
├── lua/capytrace/          # Lua frontend
├── plugin/                 # Neovim plugin entry
├── assets/                 # Static assets
├── Makefile                # Build system
├── go.mod                  # Go dependencies
└── README.md               # This file
```

### Building from Source

```bash
# Clone repository
git clone https://github.com/andev0x/capytrace.nvim.git
cd capytrace.nvim

# Build binary
make build

# Run tests
make test

# Format code
make fmt

# Clean build artifacts
make clean
```

### Coding Standards

- **Go**: Follow [Effective Go](https://golang.org/doc/effective_go) conventions
- **Documentation**: All exported functions must have doc comments
- **Testing**: Unit tests for filter, recorder, and exporter packages
- **Formatting**: Use `go fmt` before committing

### Dependencies

| Package | Purpose | License |
|---------|---------|---------|
| `modernc.org/sqlite` | Pure-Go SQLite driver | Apache 2.0 |

Minimal dependencies ensure fast builds and maximum portability.

---

## Contributing

We welcome contributions from the community! Whether it's bug fixes, feature requests, or documentation improvements, your help is valuable.

### Getting Started

1. **Fork** the repository on GitHub
2. **Clone** your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/capytrace.nvim.git
   ```
3. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. **Make your changes** and ensure code quality:
   ```bash
   make fmt  # Format code
   make test # Run tests
   ```
5. **Commit** with clear, concise messages:
   ```bash
   git commit -m "feat: add new feature" -m "Detailed description of changes"
   ```
6. **Push** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Open a Pull Request** on GitHub with a clear description

### Reporting Issues

Found a bug? Please open an issue with:
- A clear, descriptive title
- Steps to reproduce the issue
- Expected vs. actual behavior
- Your Neovim version (`nvim --version`)
- Your Go version (`go version`)

### Feature Requests

Have an idea for improvement? Open an issue with:
- A clear use case and motivation
- Description of the proposed feature
- Any relevant examples or mockups

### Development Tips

- Use `nvim -u NORC -c "edit test.nvim"` to test without your config
- Enable verbose logging: `let g:capytrace_debug = 1`
- Check Go code quality: `go vet ./...`
- Profile performance: `go test -bench=. ./internal/...`

---

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for full details.

### Summary

You are free to:
- ✅ Use this software for any purpose
- ✅ Copy, modify, and distribute the software
- ✅ Include the software in proprietary applications

Under the condition that:
- ℹ️ You include a copy of the license and copyright notice

---

## Support

### Getting Help

- 📖 **Documentation**: Check [docs/](docs/) for detailed guides
- 🐛 **Bug Reports**: Open an issue on [GitHub Issues](https://github.com/andev0x/capytrace.nvim/issues)
- 💬 **Discussions**: Join our community discussions
- 📧 **Email**: For security issues, email maintainers directly

### Frequently Asked Questions

**Q: Does capytrace send data to external servers?**
A: No. All data stays on your machine. There's no telemetry or external API calls.

**Q: What's the performance impact?**
A: Minimal. The Smart Filter and non-blocking I/O ensure your editor stays responsive. Typical overhead is <5% CPU.

**Q: Can I export sessions to other formats?**
A: Currently supported: Markdown, JSON, SQLite. More formats can be added via the exporter interface.

**Q: How much disk space do sessions use?**
A: Approximately 1-2 MB per hour of development, depending on event frequency.

**Q: Can I share sessions with teammates?**
A: Yes! Export to JSON or SQLite format and share the files. Markdown exports are also shareable.

---

## Roadmap

### Current Version (v0.2.0+)

- ✅ Smart event filtering
- ✅ Multi-format export (Markdown, JSON, SQLite)
- ✅ Session management and resumption
- ✅ Statistics and analytics
- ✅ Professional architecture (cmd/internal pattern)

### Future Plans

- 🔄 Web-based session viewer
- 🔄 Git integration (correlate with commits)
- 🔄 Multi-session merging and aggregation
- 🔄 Custom event hooks
- 🔄 Session tagging and search
- 🔄 Visual timeline renderer

---

## Acknowledgments

- Built with ❤️ using [Go](https://golang.org/) and [Lua](https://www.lua.org/)
- Thanks to the [Neovim](https://neovim.io/) community
- Inspired by flight data recorders and black-box logging concepts

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and release notes.

---

## Funding & Support

If capytrace.nvim has been helpful in your development workflow, consider supporting the project:

- 🌟 **Star** the repository on GitHub
- 💰 **Sponsor** via [GitHub Sponsors](https://github.com/sponsors/andev0x)
- ☕ **Donate** via [Buy Me a Coffee](https://www.buymeacoffee.com/anvndev)

Your support helps maintain and improve the project for everyone!

---

<div align="center">

Made with ❤️ by the capytrace community

[GitHub](https://github.com/andev0x/capytrace.nvim) · [Issues](https://github.com/andev0x/capytrace.nvim/issues) · [Discussions](https://github.com/andev0x/capytrace.nvim/discussions)

</div>
