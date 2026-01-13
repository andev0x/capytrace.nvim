# Contributing to capytrace.nvim

Thank you for your interest in contributing to capytrace.nvim! This document provides guidelines and instructions for getting involved.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)
- [Feature Requests](#feature-requests)

---

## Code of Conduct

We are committed to providing a welcoming and inspiring community for all. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

In summary:
- Be respectful and inclusive
- Welcome diverse perspectives
- Focus on constructive criticism
- Report unacceptable behavior to maintainers

---

## Getting Started

### Prerequisites

- **Neovim**: v0.9.0 or higher
- **Go**: v1.18 or higher
- **Make**: For building and testing
- **Git**: For version control
- **GitHub Account**: To fork and submit PRs

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/capytrace.nvim.git
   cd capytrace.nvim
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/andev0x/capytrace.nvim.git
   ```

---

## Development Setup

### Building from Source

```bash
# Install dependencies
go mod download

# Build the binary
make build

# Run tests
make test

# Format code
make fmt

# Verify code quality
go vet ./...
```

### Setting Up Your Editor

**VS Code / VSCodium:**
```json
{
  "[go]": {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "golang.go"
  }
}
```

**Vim / Neovim:**
```vim
autocmd BufWritePre *.go !gofmt -w %
```

### Running Tests

```bash
# Run all tests with verbose output
make test

# Run specific test package
go test ./internal/filter -v

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Making Changes

### Create a Feature Branch

Always create a new branch for your work:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test additions/improvements
- `perf/` - Performance improvements

### Project Structure

```
capytrace.nvim/
├── cmd/capytrace/           # CLI entry point
├── internal/
│   ├── filter/             # Event filtering logic
│   ├── recorder/           # Session management
│   ├── exporter/           # Export formats
│   └── models/             # Data structures
├── lua/capytrace/          # Lua frontend
├── plugin/                 # Neovim plugin entry
├── tests/                  # Test files
├── docs/                   # Documentation
└── README.md              # Main documentation
```

### Key Components

**Internal Packages:**
- `internal/filter`: Cursor event debouncing and filtering
- `internal/recorder`: Session state and event persistence
- `internal/exporter`: Export to Markdown, JSON, SQLite
- `internal/models`: Shared data types

**Lua Frontend:**
- `lua/capytrace/init.lua`: Main plugin logic
- `lua/capytrace/config.lua`: Configuration management

**Entry Points:**
- `cmd/capytrace/main.go`: CLI implementation
- `plugin/capytrace.lua`: Neovim plugin setup

---

## Coding Standards

### Go Code Style

Follow [Effective Go](https://golang.org/doc/effective_go):

```go
// Good: Clear, concise variable names
sessionID := generateSessionID()
filters := make(map[string]bool)

// Good: Descriptive function names
func (s *Session) RecordEdit(filename string, line int) error {
    // Implementation
}

// Good: Proper error handling
if err != nil {
    return fmt.Errorf("failed to save session: %w", err)
}
```

### Documentation

All exported functions and types must have doc comments:

```go
// Session represents a complete debugging session with recorded events.
type Session struct {
    ID       string
    Events   []Event
}

// NewSession creates a new session with the given parameters.
// It initializes empty event slice and sets current time as start time.
func NewSession(id, path string) *Session {
    return &Session{
        ID: id,
        Events: make([]Event, 0),
    }
}
```

### Naming Conventions

- **Packages**: Lower case, single word preferred
- **Functions**: CamelCase, exported functions capitalized
- **Variables**: camelCase for local variables, CamelCase for exported
- **Constants**: ALL_CAPS with underscores
- **Interfaces**: End with `er` or `or` (Reader, Writer, Filter)

### Error Handling

Always wrap errors with context:

```go
// Good
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("failed to read session file %q: %w", path, err)
}

// Avoid generic errors
if err != nil {
    return err  // Loss of context
}
```

### Concurrency

Use proper synchronization primitives:

```go
// Good: Protected with mutex
func (s *Session) addEvent(e Event) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.events = append(s.events, e)
}

// Use channels for communication between goroutines
eventChan := make(chan Event, 100)
go func() {
    for event := range eventChan {
        s.addEvent(event)
    }
}()
```

---

## Testing

### Writing Tests

Test files should be in the same package:

```go
// math.go
package math

func Add(a, b int) int {
    return a + b
}

// math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    tests := []struct {
        a, b, want int
    }{
        {1, 2, 3},
        {0, 0, 0},
        {-1, 1, 0},
    }
    
    for _, tt := range tests {
        got := Add(tt.a, tt.b)
        if got != tt.want {
            t.Errorf("Add(%d, %d) = %d, want %d", 
                tt.a, tt.b, got, tt.want)
        }
    }
}
```

### Test Coverage Goals

- **Critical paths**: >90% coverage
- **Helper functions**: >70% coverage
- **Overall target**: >80% coverage

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestName ./path/to/package
```

---

## Documentation

### README Updates

If your change affects user-facing functionality, update the README.md:

```markdown
### New Feature Name

Brief description of the feature.

```lua
-- Example configuration
require("capytrace").setup({
    new_option = true,
})
```

Usage example...
```

### Code Comments

- **Package-level**: Explain the package's purpose
- **Function-level**: Explain what the function does, not how
- **Complex logic**: Explain why, not just what
- **TODO comments**: Format as `// TODO(issue#): description`

### Example Documentation

```go
// Package filter provides event filtering and denoising for reducing noise
// from high-frequency cursor movement events.
package filter

// CursorFilter implements intelligent filtering for cursor movement events.
// It debounces rapid movements and only commits positions after idle detection.
type CursorFilter struct {
    // ...
}

// ProcessEvent filters an incoming event based on debounce and idle rules.
// Returns the event to record, or nil if the event should be filtered.
func (cf *CursorFilter) ProcessEvent(event *models.Event) *models.Event {
    // Implementation...
}
```

---

## Commit Message Guidelines

Follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `test`: Test additions or changes
- `refactor`: Code refactoring without feature changes
- `perf`: Performance improvements
- `chore`: Maintenance tasks, dependency updates
- `ci`: CI/CD changes

### Scope (Optional)

Specific component: `filter`, `recorder`, `exporter`, `lua`, etc.

### Subject

- Imperative mood ("add feature" not "adds feature")
- No capitalization
- No period at end
- Limit to 50 characters

### Body

- Explain what and why, not how
- Wrap at 72 characters
- Separate from subject with blank line

### Footer

Reference related issues:
```
Fixes #123
Related-to #456
```

### Examples

```
feat(filter): implement cursor debouncing

Add intelligent debouncing of cursor movement events to reduce
noise in session recordings by approximately 90%. Cursor positions
are now only recorded after idle detection or when followed by
significant events like text changes.

Fixes #45
```

```
fix(exporter): handle empty session correctly

Ensure SQLite exporter doesn't crash when session has no events.
Add validation check before schema creation.

Fixes #89
```

---

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Format code**:
   ```bash
   make fmt
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Check code quality**:
   ```bash
   go vet ./...
   ```

5. **Update documentation**:
   - README.md (if user-facing changes)
   - CHANGELOG.md (in Unreleased section)
   - Code comments (if complex changes)

### Creating a Pull Request

1. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open PR on GitHub with:
   - Clear, descriptive title
   - Reference related issues
   - Summary of changes
   - Any breaking changes
   - How to test the changes

### PR Title Format

```
[Type] Brief description of changes

Examples:
[Feature] Add session tagging support
[Fix] Correct cursor filter timing issue
[Docs] Update configuration examples
[Refactor] Improve exporter interface
```

### PR Description Template

```markdown
## Description
Brief summary of the changes made.

## Related Issues
Fixes #123, Related to #456

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Documentation update
- [ ] Code refactoring
- [ ] Performance improvement

## Changes Made
- Specific change 1
- Specific change 2
- Specific change 3

## Testing
How to test these changes:
```bash
# Commands to verify
```

## Breaking Changes
None / Describe any breaking changes

## Checklist
- [ ] Code follows project style guidelines
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests pass (`make test`)
- [ ] No new warnings from `go vet`
```

### During Review

- Be responsive to feedback
- Make requested changes in new commits
- Don't force-push after review begins (unless asked)
- Keep discussion focused and respectful

---

## Reporting Issues

### Bug Reports

Include:
1. **Environment**: Neovim version, Go version, OS
2. **Reproduction Steps**: Clear, numbered steps
3. **Expected vs Actual**: What should happen vs what happens
4. **Logs/Error Messages**: Full error output if applicable
5. **Configuration**: Your `setup()` options

**Example:**
```markdown
## Issue: Cursor filter not debouncing properly

### Environment
- Neovim: v0.9.2
- Go: 1.21
- OS: macOS 13.5

### Steps to Reproduce
1. Start capytrace session
2. Navigate file with 'jjjjjjjj' (hold j key)
3. Stop after 1 second
4. Check session log

### Expected
- Cursor movement recorded once

### Actual
- Cursor movement recorded 5+ times

### Configuration
```lua
require("capytrace").setup({
    filter_threshold = 200,
    debounce_interval = 100,
})
```
```

### Feature Requests

Include:
1. **Use Case**: Why this feature is needed
2. **Proposed Solution**: How you envision it working
3. **Alternatives**: Other approaches considered
4. **Context**: Related features or issues

**Example:**
```markdown
## Feature Request: Session tagging

### Use Case
Users working on multiple projects need to organize sessions.
Currently, sessions are only identified by ID, making it hard
to find sessions related to specific projects or bugs.

### Proposed Solution
Add a tagging system allowing users to tag sessions:
```lua
capytrace.tag_session("session_id", {"bug", "performance"})
```

### Benefits
- Better organization of large session libraries
- Easier filtering and searching
- Support for workflow analysis
```

---

## Troubleshooting

### Build Issues

```bash
# Clean rebuild
make clean
make build

# Check Go version
go version  # Should be 1.18+

# Update modules
go mod tidy
```

### Test Failures

```bash
# Verbose test output
go test -v ./...

# Run single test
go test -run TestName ./path/to/package -v

# Check race conditions
go test -race ./...
```

### Import Path Issues

If you see import errors:
```bash
# Update module paths
go mod tidy

# Verify module
cat go.mod
```

---

## Getting Help

- 📖 Check [README.md](README.md) and [docs/](docs/)
- 💬 Open a discussion on GitHub
- 🐛 Search existing issues
- 📧 Contact maintainers for guidance

---

## Acknowledgments

Thank you for contributing to capytrace.nvim! Your efforts help make debugging easier for everyone.

Happy contributing! 🎉
