# Smart Aggregation System

## Overview

The capytrace.nvim smart aggregation system transforms raw event streams into meaningful insights through a dual-file architecture and intelligent analytics.

---

## Dual-File Architecture

### 1. The Truth: `{session_id}_raw.json`

**Purpose:** Complete, unmodified event history for machine computation and potential replay features.

**Contains:**
- Every single event with millisecond precision
- Full context: column, line, changed_tick values
- All cursor movements (after smart filtering)
- Complete terminal commands and annotations
- Unprocessed, immutable data

**Use cases:**
- Programmatic analysis and data mining
- Building "Replay" features (reviewing coding sessions like videos)
- Debug troubleshooting of the plugin itself
- Long-term archival of coding sessions

**Example structure:**
```json
{
  "id": "1737000000_myproject",
  "project_path": "/home/user/myproject",
  "start_time": "2026-01-15T10:00:00Z",
  "end_time": "2026-01-15T12:30:00Z",
  "active": false,
  "events": [
    {
      "type": "file_edit",
      "timestamp": "2026-01-15T10:05:23.123Z",
      "data": {
        "filename": "src/auth.lua",
        "line": 45,
        "column": 12,
        "line_count": 120,
        "changed_tick": 1523
      }
    },
    // ... thousands more events
  ]
}
```

---

### 2. The Story: `SESSION_SUMMARY.md`

**Purpose:** Human-readable, aggregated summary with advanced analytics and insights.

**Contains:**
- Activity blocks (merged events within 2 seconds)
- Velocity metrics (flow state detection)
- Focus ratio (main files vs. distractions)
- Error correction patterns
- Idle gap analysis
- Executive summary with key stats

**Use cases:**
- Quick session review
- Identifying productivity patterns
- Understanding workflow bottlenecks
- Sharing session insights with team members
- Self-reflection on coding habits

**Updates:**
- **Automatically** every 5 minutes (configurable)
- **Automatically** when session ends
- **Crash-safe**: Raw JSON is saved immediately on every event

---

## Smart Aggregation Rules

### The Three Golden Rules

#### 1. The 2-Second Rule: Activity Block Merging

**Logic:** If `file_edit` events occur less than 2 seconds apart in the same file, merge them into a single Activity Block.

**Why?** Rapid successive edits represent continuous work, not separate tasks. This reduces noise from hundreds of keystroke events into meaningful "coding bursts."

**Example:**
```
Raw events (50 edits in 8 seconds) → 1 Activity Block
```

**Activity Block includes:**
- Start and end timestamps
- Total duration
- Number of events
- Delta tick (total changes made)
- Velocity (ticks per second)
- How the block was closed

---

#### 2. Context Switch Rule: Immediate Block Closure

**Logic:** If the user switches files or executes a terminal command, immediately close the current Activity Block, even if less than 2 seconds have passed.

**Why?** Context switches indicate a shift in mental focus. Treating them as block boundaries preserves the semantic meaning of each work session.

**Triggers:**
- Opening a different file (`file_open` event)
- Switching buffers
- Running terminal commands (`terminal_command` event)

**Example:**
```
10:05:00 - Edit auth.lua
10:05:01 - Edit auth.lua
10:05:02 - Run 'go test' → Block closed (context_switch)
10:05:05 - Edit handler.go → New block starts
```

---

#### 3. Idle Rule: Gap Detection

**Logic:** If no events occur for more than 5 minutes, record an idle gap. This helps identify:
- Moments when you're stuck debugging
- Breaks for coffee/meetings
- Time spent reading documentation

**Why?** Long pauses reveal workflow inefficiencies or learning opportunities.

**Analysis provided:**
- Number of idle gaps
- Total idle time
- Timestamps of each gap (for correlation with error corrections)

---

## Advanced Metrics

### 1. Velocity: Flow State Detection

**Formula:** `Velocity = Delta Tick / Duration (seconds)`

**What is it?** The rate of code changes per second. High velocity (>10 ticks/sec) indicates you're in a "flow state" - coding fast and efficiently without interruptions.

**Insights:**
- **Average Velocity:** Overall coding speed across the session
- **Peak Velocity:** Your fastest moment
- **Flow Blocks:** Specific periods of peak productivity (highlighted in summary)

**Example:**
```
Flow State Block:
- File: auth.lua
- Velocity: 15.3 ticks/sec 🔥
- Duration: 2m 15s
- Changes: 2,065 ticks across 47 events
```

**Use case:** Identify optimal working conditions. Were you in flow during morning hours? After coffee? In quiet environments?

---

### 2. Focus Ratio: Time Management

**Formula:** `Focus Ratio = (Time in main files) / (Time in main files + distraction time)`

**What is it?** The percentage of time spent on actual code files vs. file browsers, chat tools, or navigation buffers.

**Distraction files detected:**
- NvimTree, neo-tree, CHADTree (file explorers)
- copilot-chat (AI chat windows)
- undotree, fern (utility buffers)

**Insights:**
- **Overall Focus:** E.g., "85% of time on main code files"
- **Time by File:** Which files consumed the most attention
- **Distraction Time:** How much time spent in non-code interfaces

**Example:**
```
Focus Ratio: 87.5%

Time Spent by File:
- auth.lua: 45m 30s (35%)
- handler.go: 30m 12s (23%)
- config.yaml: 15m 5s (12%)
...

Distraction Time: 18m 22s in file browsers/tools
```

**Use case:** Optimize your workflow. If distraction time is high, consider using jump-to-definition commands instead of file browsers.

---

### 3. Error Correction Pattern: Debugging Insights

**Logic:** Detects when you add an annotation like "fix error" or "debug issue" and looks back at recent file edits to identify:
- Lines deleted before the fix
- Negative tick deltas (reversing changes)
- Number of affected blocks

**Why?** Understanding your error patterns helps you:
- Identify recurring mistake types
- Improve code review practices
- Learn from debugging sessions

**Example:**
```
Error Correction Pattern:
- Time: 10:45:23
- File: auth.lua
- Annotation: "fix authentication bug - wrong token validation"
- Blocks affected: 3
- Changes reversed: 127 ticks
- Lines deleted: ~8
```

**Use case:** Track which files or functions cause the most errors. Over time, you'll see patterns like "I always mess up error handling in this module."

---

## Configuration Options

### Default Configuration

```lua
require("capytrace").setup({
  -- Existing options...
  output_format = "markdown",
  save_path = "~/capytrace_logs/",
  
  -- Smart Aggregation Settings
  aggregation = {
    -- The 2-Second Rule: merge edits within this window (milliseconds)
    merge_window = 2000,
    
    -- The Idle Rule: detect gaps longer than this (milliseconds)
    idle_threshold = 300000, -- 5 minutes
    
    -- Flow State threshold: velocity > this = flow (ticks/sec)
    flow_velocity_threshold = 10.0,
    
    -- Files/buffers that count as distractions (pattern matching)
    distraction_files = {
      "NvimTree",
      "copilot-chat",
      "neo-tree",
      "CHADTree",
      "fern",
      "undotree",
    },
    
    -- How often to update SESSION_SUMMARY.md (milliseconds)
    periodic_update_interval = 300000, -- 5 minutes
  },
})
```

---

## Data Processing Pipeline

### Pipeline Flow

```
1. Event Occurs in Neovim
   ↓
2. Lua Plugin Captures Event
   ↓
3. Go Backend Receives Event
   ↓
4. [IMMEDIATE] Save to {session_id}_raw.json (crash-safe)
   ↓
5. [BACKGROUND] Aggregator runs every 5 minutes:
   - Build activity blocks (3 golden rules)
   - Calculate velocity metrics
   - Analyze focus ratio
   - Detect error patterns
   - Find idle gaps
   ↓
6. [BACKGROUND] Generate SESSION_SUMMARY.md
   ↓
7. On Session End:
   - Stop periodic updates
   - Generate final SESSION_SUMMARY.md
   - Close all files
```

### Crash Safety

**Q: What happens if Neovim crashes?**
**A:** Raw JSON is saved immediately on every event. You'll have a complete record up to the last keystroke. SESSION_SUMMARY.md can be regenerated from the raw JSON.

**Q: What if the aggregation process fails?**
**A:** The raw JSON is always preserved. Aggregation runs in a separate goroutine and never blocks event recording.

---

## Example SESSION_SUMMARY.md

```markdown
# Session Summary Report

**Session ID:** `1737000000_myproject`
**Project:** `/home/user/myproject`
**Started:** 2026-01-15 10:00:00
**Ended:** 2026-01-15 12:30:00
**Duration:** 2h 30m

---

## Executive Summary

- **Total Events:** 1,523
- **Activity Blocks:** 87
- **Average Velocity:** 8.5 ticks/sec
- **Peak Velocity:** 18.2 ticks/sec
- **Focus Ratio:** 82.3%
- **Flow State Blocks:** 5
- **Idle Gaps:** 3 (Total: 22m 15s)
- **Error Corrections:** 2

## Velocity Analysis

**What is Velocity?** Delta Tick / Duration. High velocity (>10 ticks/sec) indicates "Flow State" - you're coding fast and efficiently.

### Flow State Blocks (5)

These are your most productive moments:

1. **10:15:30** - `auth.lua`
   - Velocity: **18.2 ticks/sec** 🔥
   - Duration: 3m 45s
   - Changes: 4,095 ticks across 52 events

2. **11:05:12** - `handler.go`
   - Velocity: **15.7 ticks/sec** 🔥
   - Duration: 2m 20s
   - Changes: 2,198 ticks across 38 events

...

## Focus Ratio

**Overall Focus:** 82.3% of time spent on main code files

### Time Spent by File

- `auth.lua`: 45m 30s (30.3%)
- `handler.go`: 38m 22s (25.5%)
- `config.yaml`: 18m 5s (12.0%)
...

**Distraction Time:** 26m 40s in file browsers/tools

## Error Correction Patterns

Detected moments where you fixed errors after making mistakes:

### 1. 10:45:23 - `auth.lua`

**Annotation:** "fix authentication bug - wrong token validation"

- Blocks affected: 3
- Changes reversed: 127 ticks
- Lines deleted: ~8

...

## Idle Periods

Gaps > 5 minutes where you might have been stuck or took a break:

1. **10:35:00** - 8m 22s idle
2. **11:30:00** - 10m 15s idle
3. **12:00:00** - 3m 38s idle

## Activity Timeline

Aggregated blocks of continuous work (events < 2 seconds apart):

### 10:15:30 - `auth.lua`

- **Duration:** 3m 45s
- **Events:** 52 edits
- **Changes:** 3,200 → 7,295 ticks (Δ4,095)
- **Velocity:** 18.2 ticks/sec 🔥
- **Closed by:** context_switch

...

---

*Generated by capytrace.nvim with smart aggregation*
*Raw event data available in `1737000000_myproject_raw.json`*
```

---

## Customization Tips

### Adjust Merge Window for Different Workflows

**Fast typists:**
```lua
aggregation = {
  merge_window = 1000, -- 1 second (more granular blocks)
}
```

**Thoughtful coders:**
```lua
aggregation = {
  merge_window = 5000, -- 5 seconds (larger blocks)
}
```

---

### Custom Distraction Files

**Add your own patterns:**
```lua
aggregation = {
  distraction_files = {
    "NvimTree",
    "Telescope", -- If you spend too much time searching
    "lazy.nvim", -- Plugin manager
    "Mason",     -- LSP installer
  },
}
```

---

### Change Flow State Threshold

**For high-velocity coders:**
```lua
aggregation = {
  flow_velocity_threshold = 20.0, -- Higher bar for flow
}
```

**For careful coders:**
```lua
aggregation = {
  flow_velocity_threshold = 5.0, -- Lower threshold
}
```

---

## Future Enhancements

Potential features for smart aggregation:

1. **Replay Mode:** Play back coding sessions as a video using raw JSON timestamps
2. **Team Analytics:** Compare velocity/focus across team members
3. **Learning Patterns:** Identify correlation between documentation time and error rates
4. **Time-of-Day Insights:** When are you most productive?
5. **LSP Integration:** Correlate diagnostics with velocity drops
6. **Git Integration:** Link activity blocks to commit messages

---

## Troubleshooting

**Q: SESSION_SUMMARY.md isn't updating**
- Check that `periodic_update_interval` is set
- Ensure the session is active (not just loaded)
- Look for errors in Neovim's `:messages`

**Q: Raw JSON file is huge**
- This is normal for long sessions with many edits
- Consider splitting sessions or using shorter intervals
- The summary file is always small and readable

**Q: Velocity seems off**
- Velocity depends on your editor setup (LSP auto-formatting can inflate it)
- Compare relative velocities (your fast vs. slow moments) rather than absolute numbers

**Q: Focus ratio is lower than expected**
- Check `distraction_files` patterns - you may need to add more
- Some time in file browsers is normal for exploratory work
- Use the "Time by File" breakdown to see where time went

---

## Related Documentation

- **REFACTORING_SUMMARY.md** - Technical implementation details
- **README.md** - User guide and quick start
- **internal/aggregator/aggregator.go** - Source code for aggregation logic

---

*Generated by capytrace.nvim - Smart session tracking for Neovim*
