# Quick Start: Smart Aggregation

## Installation

Follow the standard installation instructions in the main README.md.

## Basic Usage

### 1. Start a Session

In Neovim:
```vim
:CapytraceStart
```

This will:
- Create a new session with ID `{timestamp}_{projectname}`
- Start recording all events
- Save raw events to `{session_id}_raw.json` immediately
- Begin periodic SESSION_SUMMARY.md updates every 5 minutes

---

### 2. Work Normally

The plugin automatically tracks:
- ✅ File edits (every keystroke)
- ✅ Cursor movements (smart-filtered)
- ✅ Terminal commands
- ✅ File opens/switches
- ✅ LSP diagnostics

No action required - just code!

---

### 3. Add Annotations (Optional)

When you fix a bug or reach a milestone:
```vim
:CapytraceAnnotate Fixed authentication bug - wrong token validation
```

This helps the error correction pattern detector correlate fixes with preceding changes.

---

### 4. Check Your Progress

While the session is active, you can view:
- **Real-time summary**: `~/capytrace_logs/SESSION_SUMMARY.md`
- **Raw events**: `~/capytrace_logs/{session_id}_raw.json`

SESSION_SUMMARY.md updates automatically every 5 minutes.

---

### 5. End the Session

```vim
:CapytraceEnd
```

This will:
- Stop recording events
- Generate final SESSION_SUMMARY.md with complete analytics
- Save all data to disk

---

## What You'll See

### SESSION_SUMMARY.md Structure

```markdown
# Session Summary Report

## Executive Summary
- Total events, activity blocks, velocity stats
- Focus ratio, flow blocks, idle gaps
- Error corrections

## Velocity Analysis
- Flow State Blocks (high productivity moments)
- Average and peak velocity

## Focus Ratio
- Time spent per file
- Distraction time breakdown

## Error Correction Patterns
- Detected fixes with context

## Idle Periods
- Gaps where you might have been stuck

## Activity Timeline
- Aggregated blocks of continuous work
```

---

## Example Workflow

### Morning Coding Session

```vim
" Start session
:CapytraceStart

" Work on authentication module
" ... make edits to auth.lua ...

" Add annotation when you hit a problem
:CapytraceAnnotate Stuck on OAuth flow - need to read docs

" ... continue coding ...

" Fixed the issue
:CapytraceAnnotate Fixed OAuth - needed to pass state parameter

" Run tests
:terminal go test ./...

" Switch to handler
" ... work on handler.go ...

" End session
:CapytraceEnd
```

### Review Your SESSION_SUMMARY.md

```markdown
## Executive Summary
- Focus Ratio: 88.5%
- Flow State Blocks: 3
- Idle Gaps: 1 (15m 30s) ← You were reading docs!

## Error Correction Patterns

### 1. 10:45:23 - `auth.lua`
**Annotation:** "Fixed OAuth - needed to pass state parameter"
- Blocks affected: 5
- Changes reversed: 203 ticks
```

**Insight:** The 15-minute idle gap correlates with your OAuth research. The summary shows exactly which blocks you had to rewrite afterward.

---

## Tips for Better Analytics

### 1. Use Meaningful Annotations

**Good:**
```vim
:CapytraceAnnotate Fixed race condition in request handler
:CapytraceAnnotate Refactored auth to use middleware pattern
```

**Less useful:**
```vim
:CapytraceAnnotate Fixed bug
:CapytraceAnnotate Update
```

---

### 2. Break Sessions into Logical Units

Instead of one 8-hour session:
```vim
" Morning: Feature work
:CapytraceStart
... work ...
:CapytraceEnd

" Afternoon: Bug fixes
:CapytraceStart
... work ...
:CapytraceEnd
```

This makes summaries more focused and actionable.

---

### 3. Review Summaries Weekly

Look for patterns:
- Are you more productive in the morning or afternoon?
- Which files/modules cause the most errors?
- Is your focus ratio dropping over time?
- Are idle gaps increasing (sign of complexity/blockers)?

---

## Configuration

### Basic Setup (Defaults)

```lua
require("capytrace").setup({
  output_format = "markdown",
  save_path = "~/capytrace_logs/",
})
```

### Advanced Setup

```lua
require("capytrace").setup({
  save_path = "~/coding_sessions/",
  
  aggregation = {
    -- Merge edits within 3 seconds (default: 2s)
    merge_window = 3000,
    
    -- Detect idle after 10 minutes (default: 5min)
    idle_threshold = 600000,
    
    -- Flow state threshold (default: 10 ticks/sec)
    flow_velocity_threshold = 15.0,
    
    -- Add custom distraction patterns
    distraction_files = {
      "NvimTree",
      "Telescope",
      "lazy.nvim",
    },
    
    -- Update summary every 10 minutes (default: 5min)
    periodic_update_interval = 600000,
  },
})
```

---

## Troubleshooting

### SESSION_SUMMARY.md is empty

**Check:**
1. Is the session active? (`:CapytraceStats`)
2. Have you made any edits yet?
3. Wait for the first 5-minute update

**Solution:** End and restart the session to force generation.

---

### Raw JSON is huge

**Normal!** Long sessions generate lots of events.

**Solutions:**
- Use shorter sessions
- The summary is always small and readable
- Raw JSON is only for machine analysis/replay

---

### Velocity seems inconsistent

**Explanation:** Velocity depends on:
- LSP auto-formatting (inflates ticks)
- Copy-pasting (sudden spikes)
- Your editing style

**Focus on:** Relative velocity (your fast vs. slow moments), not absolute numbers.

---

## Next Steps

- Read `docs/SMART_AGGREGATION.md` for deep dive into algorithms
- Check `docs/REFACTORING_SUMMARY.md` for technical details
- Explore the raw JSON for custom analysis

---

*Happy coding! 🚀*
