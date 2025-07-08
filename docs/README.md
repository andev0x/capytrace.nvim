# debugstory.nvim — Your Personal Debugging Story Recorder for Neovim

**debugstory.nvim** is a lightweight Neovim plugin that automatically captures and organizes your debugging journey in real time — across terminal commands, file edits, and code hops. Built with performance and extensibility in mind, it empowers developers to **trace back their debugging steps**, **annotate intent**, and **resume from context** even after switching tasks or machines.

---

## 🔧 Features

- 🧠 **Context-Aware Debug Logging**  
  Automatically captures terminal commands, modified files, and Git diffs during debugging sessions.

- ✍️ **Live Debug Session Recorder**  
  Tracks file edits, cursor jumps, LSP diagnostics, and breakpoints.

- 🗃️ **Structured Session Timeline**  
  Outputs each session into structured Markdown or JSON formats with timestamps and tags.

- 🏷️ **User Annotations**  
  Add notes, hypotheses, or TODOs inline during debugging — all saved automatically.

- 🔁 **Session Resumption**  
  Rehydrate session state into a new Neovim instance for context-aware continuation.

- 🔌 **Plugin Friendly & Lazy.nvim Compatible**  
  Written in Go with a Lua bridge — easily pluggable with any `lazy.nvim` setup.

---

## 📦 Installation (with [lazy.nvim](https://github.com/folke/lazy.nvim))

```lua
{
  "andev0x/debugstory.nvim",
  build = "make", -- Optional: if building Go binary
  config = function()
    require("debugstory").setup({
      output_format = "markdown", -- or "json"
      save_path = "~/debugstories/",
    })
  end,
}
```

> ⚠️ Requires Go installed (`go version`) if you plan to build from source.

---

## 🚀 Usage

### Basic Commands

```vim
:DebugStoryStart [project_name]    " Start a new debug session
:DebugStoryEnd                     " End current session
:DebugStoryAnnotate [note]         " Add annotation to current session
:DebugStoryStatus                  " Show current session status
:DebugStoryList                    " List all available sessions
:DebugStoryResume <session_name>   " Resume a previous session
```

### Lua API

```lua
local debugstory = require("debugstory")

-- Start a session
debugstory.start_session("my_project")

-- Add annotation
debugstory.add_annotation("Found the bug in auth.lua")

-- Get session status
local status = debugstory.get_status()

-- End session
debugstory.end_session()
```

---

## ⚙️ Configuration Options

```lua
require("debugstory").setup({
  output_format = "markdown", -- or "json"
  save_path = "~/debugstories/",
  record_terminal = true,
  record_git_diff = true,
  auto_save_on_exit = true,
  max_cursor_events = 100, -- Limit cursor movement recordings
  debounce_ms = 500, -- Debounce time for events
})
```

---

## 📁 Output Format

### Markdown Example

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

---

## 🛠️ Building from Source

```bash
# Clone the repository
git clone https://github.com/andev0x/debugstory.nvim.git
cd debugstory.nvim

# Initialize Go module
make go-mod-init

# Build the binary
make build

# Clean build artifacts
make clean
```

---

## 🧪 Roadmap

- [ ] Telescope-powered session browser
- [ ] GitHub Gist export
- [ ] Session timeline viewer
- [ ] AI-assisted summary for each session
- [ ] Git diff integration
- [ ] LSP diagnostics tracking
- [ ] Breakpoint recording
- [ ] Performance metrics

---

## 💡 Use Cases

- Keep a timeline of what you've tried while fixing complex bugs
- Create reproducible bug reports with context
- Document exploratory work on unknown codebases
- Reflect on debugging techniques and patterns
- Share debugging approaches with team members
- Resume complex debugging sessions after interruptions

---

## 🤝 Contributing

Pull requests, bug reports, and feature ideas are all welcome!  
Please open an issue to discuss changes or share feedback.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `make test` to ensure tests pass
5. Submit a pull request

---

## 📜 License

MIT © 2025 andev0x

---

## 🛠️ Built With

- [Go](https://golang.org/) - Backend performance and reliability
- [Lua](https://www.lua.org/) - Neovim integration
- [Neovim API](https://neovim.io/) - Editor integration
- [lazy.nvim](https://github.com/folke/lazy.nvim) - Plugin management

---

## 🙏 Acknowledgments

- Inspired by the need for better debugging workflow documentation
- Thanks to the Neovim community for excellent plugin architecture
- Built with performance and simplicity in mind