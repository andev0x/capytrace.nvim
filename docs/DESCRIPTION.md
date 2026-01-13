# capytrace.nvim

**capytrace.nvim** is a modern, lightweight Neovim plugin designed to automatically capture, organize, and export your debugging journey in real time. It seamlessly records terminal commands, file edits, code navigation, and more, providing a structured timeline of your development and debugging sessions. Built with a focus on performance, extensibility, and developer experience, capytrace.nvim empowers you to trace back your steps, annotate your intent, and resume your work contextually—even after switching tasks or machines.

---

## Key Features

- **Context-Aware Debug Logging**: Automatically logs terminal commands, file modifications, and Git diffs during debugging sessions.
- **Live Debug Session Recorder**: Tracks file edits, cursor movements, LSP diagnostics, and breakpoints as you work.
- **Structured Session Timeline**: Exports each session as a detailed Markdown or JSON file, complete with timestamps and event tags.
- **User Annotations**: Add notes, hypotheses, or TODOs inline during debugging—annotations are saved and timestamped.
- **Session Resumption**: Resume any previous session, restoring context and state for seamless continuation.
- **Plugin Friendly & Lazy.nvim Compatible**: Written in Go with a Lua bridge, making it easy to integrate with any `lazy.nvim` setup.

---

## Architecture Overview

capytrace.nvim is a hybrid plugin, combining the power of Go for efficient event recording and exporting, with Lua for deep Neovim integration and user interaction.

- **Lua Frontend**: Handles Neovim events, user commands, and configuration. It communicates with the Go backend by invoking the compiled Go binary with appropriate arguments.
- **Go Backend**: Manages session state, records events, and exports session data to Markdown or JSON. It is responsible for efficient file I/O and data serialization.
- **Session Recording**: The plugin tracks a wide range of events, including file edits, terminal commands, cursor movements, file opens, and LSP diagnostics. Each event is timestamped and stored as part of the session timeline.
- **Exporters**: Sessions can be exported in Markdown (for human-friendly review) or JSON (for programmatic analysis or integration with other tools).

---

## How It Works

1. **Session Lifecycle**: Start a session with `:CapyTraceStart`. The plugin begins recording all relevant events. End the session with `:CapyTraceEnd` to export the timeline.
2. **Event Capture**: Lua autocommands hook into Neovim events (e.g., file edits, cursor moves, terminal opens) and call the Go binary to record each event.
3. **Annotations**: Add notes at any time with `:CapyTraceAnnotate`, which are saved as part of the session.
4. **Session Management**: List, resume, or check the status of sessions with dedicated commands.
5. **Export**: At session end, the Go backend exports the session timeline to Markdown or JSON, including a summary and all recorded events.

---

## Example Use Cases

- **Debugging**: Trace every step, command, and hypothesis during a complex bug hunt.
- **Knowledge Sharing**: Export session logs to share with teammates or for onboarding documentation.
- **Context Switching**: Resume work exactly where you left off, even after a break or on a different machine.

---

## Installation & Requirements

- **Neovim**: v0.9.0 or higher
- **Go**: v1.18 or higher (required to build the backend binary)
- **Plugin Manager**: Easily integrates with `lazy.nvim` or other plugin managers.

---

## Getting Started

1. **Install** via your preferred plugin manager (see README for details).
2. **Build** the Go binary with `make build` (if not using prebuilt binaries).
3. **Configure** output format and save path in your Neovim config.
4. **Use** the provided commands to start, annotate, end, and resume sessions.

---

## Output Example

**Markdown:**
```
# Debug Session: 1704067200_myproject

**Project:** /home/user/myproject
**Started:** 2024-01-01T10:00:00Z
**Ended:** 2024-01-01T11:30:00Z
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
```

```bash
git log --oneline -10
```

```markdown
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

**JSON:**
```json
{
  "type": "file_edit",
  "timestamp": "2025-07-09T07:54:19.922255+07:00",
  "data": {
    "filename": "./go-algorithm/convert/NvimTree_1",
    "line": 3,
    "column": 6,
    "line_count": 4,
    "changed_tick": 12
  }
}
```

---

## Contributing

Contributions are welcome! Please see the README for development setup and guidelines.

---

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## Funding

If you find capytrace.nvim valuable, consider sponsoring or supporting the project (see README for links). 