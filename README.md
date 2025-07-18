<div align="center">
  <img src="assets/img/capytrace.gif" alt="Capytrace Logo" width="200"/>

# capytrace.nvim

[![Go](https://img.shields.io/badge/Go-%3E=1.18-blue?logo=go)](https://golang.org/) [![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE) [![Neovim](https://img.shields.io/badge/Neovim-%3E=0.9.0-blueviolet?logo=neovim)](https://neovim.io/)
</div>

---

**[capytrace.nvim](https://github.com/andev0x/capytrace.nvim.git)** is a modern, lightweight Neovim plugin that automatically captures, organizes, and exports your debugging journey in real time. It records terminal commands, file edits, code navigation, and more, providing a structured timeline of your development and debugging sessions. Built for performance and extensibility, capytrace.nvim empowers you to trace back your steps, annotate your intent, and resume your work contextually—even after switching tasks or machines.

---

## 🚀 Features

- **Context-Aware Debug Logging**: Automatically logs terminal commands, file modifications, and Git diffs during debugging sessions.
- **Live Debug Session Recorder**: Tracks file edits, cursor movements, LSP diagnostics, and breakpoints as you work.
- **Structured Session Timeline**: Exports each session as a detailed Markdown or JSON file, complete with timestamps and event tags.
- **User Annotations**: Add notes, hypotheses, or TODOs inline during debugging—annotations are saved and timestamped.
- **Session Resumption**: Resume any previous session, restoring context and state for seamless continuation.
- **Plugin Friendly & Lazy.nvim Compatible**: Written in Go with a Lua bridge, making it easy to integrate with any `lazy.nvim` setup.

---

## 🏗️ Architecture Overview

capytrace.nvim is a hybrid plugin:
- **Lua Frontend**: Handles Neovim events, user commands, and configuration. It communicates with the Go backend by invoking the compiled Go binary with appropriate arguments.
- **Go Backend**: Manages session state, records events, and exports session data to Markdown or JSON. Responsible for efficient file I/O and data serialization.
- **Session Recording**: Tracks file edits, terminal commands, cursor movements, file opens, and LSP diagnostics. Each event is timestamped and stored as part of the session timeline.
- **Exporters**: Sessions can be exported in Markdown (for human-friendly review) or JSON (for programmatic analysis or integration with other tools).

---

## 📦 Installation

**With [lazy.nvim](https://github.com/folke/lazy.nvim):**

```lua
{
  "andev0x/capytrace.nvim",
  build = "make", -- Optional: if building Go binary
  config = function()
    require("capytrace").setup({
      output_format = "markdown", -- or "json"
      save_path = "~/capytrace_logs/",
    })
  end,
}
```

> ⚠️ Requires Go installed (`go version`) if you plan to build from source.

---

## 🛠️ Building from Source

```bash
git clone https://github.com/andev0x/capytrace.nvim.git
cd capytrace.nvim
make go-mod-init
make build
```

---

## ✨ Usage

### Vim Commands

```vim
:CapyTraceStart [project_name]    " Start a new debug session
:CapyTraceEnd                     " End current session
:CapyTraceAnnotate [note]         " Add annotation to current session
:CapyTraceStatus                  " Show current session status
:CapyTraceList                    " List all available sessions
:CapyTraceResume <session_name>   " Resume a previous session
```

---

## 📁 Output Example

```markdown
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

### Example:
```json
{
  "type": "cursor_move",
  "timestamp": "2025-07-09T07:54:19.904485+07:00",
  "data": {
    "filename": "./go-algorithm/convert/NvimTree_1",
    "line": 3,
    "column": 6
  }
},
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

## 🤝 Contributing

Pull requests, bug reports, and feature ideas are all welcome! Please open an issue to discuss changes or share feedback.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `make test` to ensure tests pass
5. Submit a pull request

---

## 📜 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

---

## 💖 Funding

If you find capytrace.nvim valuable, please consider supporting its development! Your sponsorship helps maintain and improve the project for everyone.

- [Sponsor on GitHub](https://github.com/sponsors/andev0x)
- [Buy Me a Coffee](https://www.buymeacoffee.com/anvndev)

Thank you for your support! 