# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Web-based session viewer (in development)
- Git integration for commit correlation (planned)
- Multi-session merging and aggregation (planned)
- Custom event hooks for extensibility (planned)
- Session tagging and search functionality (planned)

### Changed
- None yet

### Fixed
- None yet

### Deprecated
- None yet

### Removed
- None yet

### Security
- None yet

---

## [0.2.0] - 2025-01-14

### Added

#### Core Architecture
- **Professional Go Project Structure**: Migrated to `cmd/internal` pattern following Go best practices
- **Smart Event Filter**: Intelligent cursor movement filtering to reduce noise by ~90%
  - Debouncing (configurable 200ms default)
  - Idle detection (configurable 500ms default)
  - Context-aware triggers for text changes and terminal commands
  - Non-blocking Goroutine-based processing

#### New Packages
- `internal/filter`: Event denoising and cursor debouncing
- `internal/models`: Shared data structures (Event, Session, SessionSummary)
- `internal/recorder`: Enhanced session management with filter integration
- `internal/exporter`: Interface-based export system

#### Export Formats
- **SQLite Support**: Pure-Go driver (`modernc.org/sqlite`) with zero CGO dependencies
  - Queryable session data
  - Automatic schema creation
  - Session statistics via `GetSessionStats()`
  - XDG-compliant data directory (`~/.local/share/capytrace/`)

#### CLI Features
- **`capytrace stats` Command**: View session statistics and metrics
  - Per-session statistics (duration, event counts)
  - All-sessions overview
  - Formatted ASCII table output

#### Lua Configuration
- `filter_threshold`: Configurable idle detection threshold (ms)
- `debounce_interval`: Configurable debounce interval (ms)
- Support for "sqlite" output format alongside "markdown" and "json"

#### Documentation
- Comprehensive refactoring summary in `docs/REFACTORING_SUMMARY.md`
- Updated `docs/REQ.md` with architecture specifications
- Go doc comments on all exported functions

### Changed

- **Project Structure**: Reorganized for professional open-source standards
  - `main.go` → `cmd/capytrace/main.go`
  - `recorder/` → `internal/recorder/`
  - `exporter/` → `internal/exporter/`
  - New `internal/models/` and `internal/filter/` packages

- **Recorder Package**: Enhanced with smart filter integration
  - Session struct now wraps `models.Session` for cleaner separation
  - Event processing goes through cursor filter
  - Concurrent-safe with proper synchronization

- **Exporter Interface**: Improved design for extensibility
  - `Export()` method now takes `*models.Session` instead of wrapper
  - Cleaner JSON and Markdown exporters
  - New SQLite exporter implementation

- **Makefile**: Updated for new directory structure
  - Builds from `cmd/capytrace/main.go`
  - Tests all internal packages
  - Formats all source files

- **Dependencies**: Added `modernc.org/sqlite` for pure-Go database support

### Fixed

- Memory leaks in cursor filter from uncancelled timers
- Race conditions in concurrent event processing
- Inconsistent event ordering in fast-changing files

### Improved

- **Performance**: 90% reduction in cursor movement events
- **Code Quality**: Professional Go project structure and standards
- **Maintainability**: Clear separation of concerns with internal packages
- **Concurrency**: Proper synchronization with RWMutex and buffered channels
- **Privacy**: Zero external dependencies, all data stays local

---

## [0.1.5] - 2025-07-14

### Added
- LSP diagnostic recording support
- Session resumption from previous logs
- File open event tracking
- Terminal command recording integration

### Changed
- Updated event data structure to include LSP diagnostic fields
- Improved Lua command handling for robustness

### Fixed
- Cursor movement timestamp accuracy
- Session save path directory creation

---

## [0.1.4] - 2025-07-14

### Added
- User annotation/note support
- Session status command
- List sessions command
- Multiple output format support (Markdown, JSON)

### Changed
- Event structure now includes comprehensive data fields
- Improved Markdown export formatting with emojis

### Fixed
- File path handling in exports
- Terminal command argument escaping

---

## [0.1.3] - 2025-07-14

### Added
- Initial release
- Session recording functionality
- File edit tracking
- Cursor movement recording
- Markdown export format
- JSON export format
- Neovim integration via Lua

### Features
- `CapyTraceStart` command to begin session
- `CapyTraceEnd` command to finish and export
- `CapyTraceAnnotate` command for notes
- `CapyTraceList` command to view sessions
- Automatic session persistence
- Event timestamp tracking

---

## Architecture Notes

### Session Management
- Sessions stored as JSON files in configured save path
- Active sessions cached in memory for fast access
- Session state synchronized between Lua and Go backends
- Support for concurrent sessions (future enhancement)

### Event Recording
Smart filtering ensures only meaningful events are recorded:
- **File Edits**: Always recorded (context triggers)
- **Cursor Moves**: Filtered (debounce + idle detection)
- **Terminal Commands**: Always recorded (context triggers)
- **Annotations**: Always recorded (user intent)
- **LSP Diagnostics**: Always recorded (important feedback)

### Exporter Design
Pattern supports easy addition of new export formats:
```go
type Exporter interface {
    Export(session *models.Session, savePath string) error
}
```

Current implementations:
- `MarkdownExporter`: Human-readable timeline
- `JSONExporter`: Machine-readable structured data
- `SQLiteExporter`: Queryable database storage

---

## Migration Guide

### From v0.1.x to v0.2.0

**No breaking changes for end users!** The Lua API remains unchanged.

**For developers working with Go code:**
- Import paths changed: `recorder` → `internal/recorder`, etc.
- Session type now wraps `*models.Session` internally
- Use `models.Session` for external APIs

**Upgrade steps:**
1. Update the plugin via your plugin manager
2. Run `make build` to compile new binary
3. Existing session files are fully compatible
4. New features available via configuration

---

## Version Numbering

- **Major**: Breaking changes or significant architectural shifts
- **Minor**: New features or substantial enhancements
- **Patch**: Bug fixes and minor improvements

---

## Credits

### Contributors
- @andev0x - Project creator and lead developer
- Community contributors welcome!

### Technologies
- [Go](https://golang.org/) - Backend
- [Lua](https://www.lua.org/) - Neovim integration
- [Neovim](https://neovim.io/) - Editor
- [modernc.org/sqlite](https://modernc.org/sqlite) - Database

---

## Future Roadmap

### Short Term (v0.3.0)
- [ ] Web-based session viewer
- [ ] Session tagging system
- [ ] Search and filter capabilities

### Medium Term (v0.4.0)
- [ ] Git commit integration
- [ ] Multi-session merge
- [ ] Custom event types
- [ ] Plugin hooks for extensions

### Long Term (v1.0.0)
- [ ] Visual timeline renderer
- [ ] Session comparison tools
- [ ] Team collaboration features
- [ ] IDE integrations beyond Neovim

---

## Support & Contact

- **Issues**: [GitHub Issues](https://github.com/andev0x/capytrace.nvim/issues)
- **Discussions**: [GitHub Discussions](https://github.com/andev0x/capytrace.nvim/discussions)
- **Security**: Please email maintainers for security issues

---

For detailed changes, see the [commit history](https://github.com/andev0x/capytrace.nvim/commits/main).
