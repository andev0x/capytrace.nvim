# Research Analysis: capytrace.nvim

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture Analysis](#architecture-analysis)
3. [Core Algorithms and Data Structures](#core-algorithms-and-data-structures)
4. [Event Processing Pipeline](#event-processing-pipeline)
5. [Smart Filtering Mechanisms](#smart-filtering-mechanisms)
6. [Aggregation and Analytics Algorithms](#aggregation-and-analytics-algorithms)
7. [Concurrency and Performance](#concurrency-and-performance)
8. [Storage and Export Strategies](#storage-and-export-strategies)
9. [Key Technical Innovations](#key-technical-innovations)
10. [Performance Characteristics](#performance-characteristics)

---

## Project Overview

**capytrace.nvim** is a high-performance, local-first debugging session recorder for Neovim that combines Lua frontend components with a Go backend to capture and analyze developer workflow data. The project implements sophisticated event filtering, aggregation, and analytics algorithms to transform raw development events into actionable insights.

### Key Technologies

- **Frontend**: Lua (Neovim plugin)
- **Backend**: Go 1.24.0
- **Storage**: JSON files, SQLite (via modernc.org/sqlite - pure Go implementation)
- **Architecture Pattern**: Client-Server with dual-mode operation (CLI + Daemon)

### Core Value Proposition

The project addresses the problem of "development workflow black boxes" by implementing a flight data recorder for coding sessions, featuring:
- 90% reduction in cursor movement noise through smart filtering
- Real-time activity aggregation with flow state detection
- Privacy-first design (all data stays local)
- Crash-safe dual-file architecture

---

## Architecture Analysis

### Multi-Layer Architecture

```
┌─────────────────────────────────────────────────┐
│         Neovim Editor (User Interface)          │
├─────────────────────────────────────────────────┤
│  Lua Frontend Layer (lua/capytrace/)            │
│  - Event capture via autocommands               │
│  - Binary lifecycle management                  │
│  - User command interface                       │
├─────────────────────────────────────────────────┤
│  Go Backend Layer (cmd/capytrace/)              │
│  ┌──────────────┐  ┌──────────────┐            │
│  │  Recorder    │  │   Filter     │            │
│  │  (Session    │  │  (Debounce   │            │
│  │   State)     │  │   + Idle)    │            │
│  └──────────────┘  └──────────────┘            │
│  ┌──────────────┐  ┌──────────────┐            │
│  │  Aggregator  │  │  Exporter    │            │
│  │  (Analytics) │  │  (MD/JSON/   │            │
│  │              │  │   SQLite)    │            │
│  └──────────────┘  └──────────────┘            │
├─────────────────────────────────────────────────┤
│  Storage Layer                                  │
│  - {session_id}_raw.json (immutable truth)     │
│  - SESSION_SUMMARY.md (aggregated insights)    │
│  - SQLite database (optional queryable store)  │
└─────────────────────────────────────────────────┘
```

### Communication Patterns

**1. CLI Mode (One-shot Operations)**
```
Lua → exec_go_command() → Go Binary (new process) → Exit
```
- Used for: start, end, list sessions
- Synchronous execution via `vim.fn.system()`

**2. Daemon Mode (Long-lived Process)**
```
Lua → send_daemon_request() → JSON over stdin/stdout → Go Daemon
```
- Used for: high-frequency events (cursor moves, edits)
- Asynchronous via `vim.fn.jobstart()` and channel communication
- Reduces process spawn overhead by ~95%

---

## Core Algorithms and Data Structures

### 1. Event Model (internal/models/event.go)

**Event Structure:**
```go
type Event struct {
    Type      string    `json:"type"`      // Event classification
    Timestamp time.Time `json:"timestamp"` // Millisecond precision
    Data      EventData `json:"data"`      // Polymorphic payload
}
```

**Event Types:**
- `session_start`, `session_end`, `session_resume`
- `file_edit` (line, column, changed_tick, line_text)
- `cursor_move` (line, column)
- `file_open` (filename, filetype)
- `terminal_command` (command string)
- `lsp_diagnostic` (message, level, position)
- `annotation` (user notes)

**Session Model:**
```go
type Session struct {
    ID           string
    ProjectPath  string
    StartTime    time.Time
    EndTime      time.Time
    Events       []Event  // Append-only event log
    Active       bool     // State flag
}
```

### 2. Activity Block Model (Advanced Aggregation)

```go
type ActivityBlock struct {
    StartTime  time.Time
    EndTime    time.Time
    Duration   time.Duration
    Filename   string
    EventCount int
    StartTick  int    // Buffer change counter (start)
    EndTick    int    // Buffer change counter (end)
    DeltaTick  int    // Total changes (EndTick - StartTick)
    Velocity   float64 // DeltaTick / Duration.Seconds()
    Events     []Event // Original events in block
    ClosedBy   string  // "context_switch" | "idle" | "timeout"
}
```

**Key Insight:** Activity blocks represent semantic units of work, not just temporal groupings.

---

## Event Processing Pipeline

### Full Pipeline Flow

```
1. User Action in Neovim
   ↓
2. Autocommand Triggers (TextChanged, CursorMoved, etc.)
   ↓
3. Lua Callback (lua/capytrace/init.lua:460-540)
   ↓
4. Daemon Request or CLI Execution
   ↓
5. Go Backend Event Reception
   ↓
6. [CRITICAL PATH] CursorFilter.ProcessEvent()
   ├─ Debounce check (200ms window)
   ├─ Idle detection timer (500ms)
   └─ Context trigger handling
   ↓
7. Session.addEvent() - Append to Events[]
   ↓
8. Session.save() - Immediate JSON persistence (crash-safe)
   ↓
9. [BACKGROUND] Periodic Aggregation (every 5 minutes)
   ├─ Build activity blocks
   ├─ Calculate analytics
   └─ Generate SESSION_SUMMARY.md
   ↓
10. [ON END] Final export to chosen format
```

---

## Smart Filtering Mechanisms

### Algorithm 1: Cursor Movement Debouncing

**Location:** `internal/filter/cursor_filter.go`

**Problem:** Rapid keyboard navigation (jjjjkkkk) generates hundreds of cursor events per minute, creating 90% noise.

**Solution: Two-Stage Filtering**

#### Stage 1: Debounce Filter (200ms Window)
```
Algorithm: Temporal Clustering
- Maintain lastEventTime
- If now - lastEventTime < debounceInterval:
    - Update pendingEvent (keep latest position)
    - Return nil (suppress event)
- Else:
    - Store as pendingEvent
    - Start idle detection timer
```

**Code Implementation (cursor_filter.go:102-109):**
```go
if !cf.lastEventTime.IsZero() && now.Sub(cf.lastEventTime) < cf.debounceInterval {
    cf.pendingEvent = event  // Replace previous pending
    return nil               // Suppress
}
```

#### Stage 2: Idle Detection (500ms Timer)
```
Algorithm: Deferred Commit with Timer
- After debounce passes, set timer for idleThreshold
- If timer expires:
    - Commit pendingEvent to session
- If new event arrives before expiry:
    - Cancel timer, restart process
```

**Code Implementation (cursor_filter.go:120-122):**
```go
cf.debounceTimer = time.AfterFunc(cf.idleThreshold, func() {
    cf.commitPendingEvent()
})
```

#### Stage 3: Context Triggers (Immediate Flush)
```
Algorithm: Semantic Event Boundary Detection
- On file_edit, terminal_command, or session_end:
    1. Immediately commit pending cursor event
    2. Cancel any active timers
    3. Process context event
- Result: Cursor position always captured before significant actions
```

**Code Implementation (cursor_filter.go:83-98):**
```go
if cf.contextTriggers[event.Type] {
    var result *models.Event
    if cf.pendingEvent != nil {
        result = cf.pendingEvent  // Flush cursor
        cf.pendingEvent = nil
    }
    return result  // Return flushed event
}
```

**Performance Impact:**
- Raw cursor events: ~1000/session
- After filtering: ~100/session
- Reduction: **90%**

---

## Aggregation and Analytics Algorithms

### Algorithm 2: The Three Golden Rules (Activity Block Construction)

**Location:** `internal/aggregator/aggregator.go:69-122`

**Objective:** Merge raw events into semantic work units.

#### Rule 1: The 2-Second Merge Window
```
Algorithm: Temporal Window Aggregation
For each file_edit event:
    If currentBlock exists AND same file AND timeSinceLastEvent ≤ 2s:
        - Merge into currentBlock
        - Update duration, tick counts, velocity
    Else:
        - Close currentBlock
        - Start new block
```

**Code Implementation (aggregator.go:101-109):**
```go
if timeSinceLastEvent > a.config.MergeWindow {
    currentBlock.ClosedBy = "timeout"
    blocks = append(blocks, *currentBlock)
    currentBlock = a.startNewBlock(event)
}
```

**Rationale:** Rapid successive edits (e.g., refactoring a function) represent continuous work, not separate tasks.

#### Rule 2: Context Switch Detection
```
Algorithm: Semantic Boundary Detection
On file_open OR terminal_command OR filename change:
    - Immediately close current block
    - Mark ClosedBy = "context_switch"
    - Start new block
```

**Code Implementation (aggregator.go:97-100):**
```go
if event.Data.Filename != currentBlock.Filename {
    currentBlock.ClosedBy = "context_switch"
    blocks = append(blocks, *currentBlock)
    currentBlock = a.startNewBlock(event)
}
```

**Insight:** File switches indicate mental context changes, even if only 500ms apart.

#### Rule 3: Idle Gap Detection
```
Algorithm: Threshold-based Inactivity Tracking
For consecutive events:
    gap = events[i].Timestamp - events[i-1].Timestamp
    If gap > 5 minutes:
        - Record as IdleGap
        - Track start, end, duration
```

**Code Implementation (aggregator.go:205-220):**
```go
for i := 1; i < len(events); i++ {
    timeBetween := events[i].Timestamp.Sub(events[i-1].Timestamp)
    if timeBetween > a.config.IdleThreshold {
        gaps = append(gaps, models.IdleGap{...})
    }
}
```

### Algorithm 3: Velocity-Based Flow State Detection

**Formula:**
```
Velocity = (EndTick - StartTick) / Duration.Seconds()

Flow State = Velocity ≥ 10.0 ticks/second
```

**Code Implementation (aggregator.go:169-182):**
```go
for _, block := range blocks {
    if block.Velocity > 0 {
        totalVelocity += block.Velocity
        if block.Velocity >= a.config.FlowVelocityThreshold {
            analytics.FlowBlocks = append(analytics.FlowBlocks, block)
        }
    }
}
```

**Interpretation:**
- Low velocity (< 5 ticks/sec): Thoughtful coding, debugging
- Medium velocity (5-10 ticks/sec): Normal coding
- High velocity (> 10 ticks/sec): Flow state - rapid, uninterrupted work

### Algorithm 4: Focus Ratio Calculation

**Formula:**
```
Focus Ratio = TotalMainFileTime / (TotalMainFileTime + DistractionTime)
```

**Algorithm: Time Attribution via Event Spacing**
```
For consecutive events:
    duration = events[i].Timestamp - events[i-1].Timestamp
    currentFile = events[i-1].Data.Filename

    If isDistractionFile(currentFile):
        DistractionTime += duration
    Else:
        MainFiles[currentFile] += duration
```

**Distraction Detection (aggregator.go:260-267):**
```go
func (a *Aggregator) isDistractionFile(filename string) bool {
    for _, pattern := range a.config.DistractionFiles {
        if strings.Contains(filename, pattern) {
            return true
        }
    }
    return false
}
```

**Default Distraction Patterns:**
- NvimTree, neo-tree, CHADTree (file browsers)
- copilot-chat (AI assistants)
- undotree, fern (utility buffers)

### Algorithm 5: Error Correction Pattern Detection

**Algorithm: Backward Event Analysis with Annotation Triggers**
```
For each annotation event:
    If containsErrorKeywords(annotation):
        Look back up to 10 events OR 5 minutes:
            - Count file_edit events in same file
            - Detect negative tick deltas (deletions)
            - Approximate lines deleted
        If corrections found:
            - Create ErrorPattern record
```

**Code Implementation (aggregator.go:304-357):**
```go
func (a *Aggregator) analyzeErrorContext(events []models.Event, annotationIndex int) {
    for i := annotationIndex - 1; i >= 0 && i >= annotationIndex-10; i-- {
        if event.Type == "file_edit" {
            if lastTick > 0 && event.Data.ChangedTick < lastTick {
                ticksReversed += lastTick - event.Data.ChangedTick
            }
        }
    }
}
```

**Error Keywords:** fix, error, bug, issue, mistake, correct, debug

---

## Concurrency and Performance

### Concurrency Model

**1. Goroutines in Filter (cursor_filter.go:66-67)**
```go
go filter.run()  // Background event processing
```
- Non-blocking cursor event handling
- Channel-based communication (buffer size: 100)

**2. Periodic Aggregation (recorder/session.go:63-77)**
```go
go func() {
    for {
        select {
        case <-s.periodicTicker.C:
            s.regenerateSummary()  // Every 5 minutes
        case <-s.stopPeriodicChan:
            return
        }
    }
}()
```

**3. Daemon Mode RPC Loop (cmd/capytrace/main.go:89-113)**
```go
for scanner.Scan() {
    go handleRequest(req)  // Concurrent request processing
}
```

### Thread Safety

**Mutex Protection (recorder/session.go:29-34):**
```go
type Session struct {
    *models.Session
    mu sync.Mutex  // Protects Events[] and state
}
```

**Critical Sections:**
- `addEvent()`: Locks before appending to Events[]
- `save()`: Locks during JSON serialization
- `regenerateSummary()`: Creates session copy under lock

### Performance Optimizations

**1. Buffered Channels**
```go
eventChan: make(chan *models.Event, 100)
```
- Prevents blocking on high-frequency events

**2. Lazy Aggregation**
- Aggregation runs every 5 minutes, not per-event
- Trades real-time accuracy for performance

**3. Binary Auto-Download**
- Checks cache before network requests
- Tries multiple artifact naming conventions
- Reduces manual installation overhead

---

## Storage and Export Strategies

### Dual-File Architecture

**Philosophy: Separation of Truth and Story**

#### File 1: `{session_id}_raw.json` (The Truth)
- **Purpose:** Complete, immutable event history
- **Update frequency:** Every event (crash-safe)
- **Size:** ~1-2 MB per hour
- **Use cases:**
  - Machine analysis
  - Replay features
  - Debugging the plugin itself

#### File 2: `SESSION_SUMMARY.md` (The Story)
- **Purpose:** Human-readable insights
- **Update frequency:** Every 5 minutes + on session end
- **Size:** ~50-200 KB regardless of session length
- **Use cases:**
  - Quick review
  - Pattern identification
  - Team sharing

### Export Formats

**1. Markdown Exporter (exporter/markdown.go)**
- Simple timeline with event types
- Emoji-enhanced for readability
- Chronological event list

**2. Smart Markdown Exporter (exporter/smart_markdown.go)**
- Full analytics dashboard
- Activity blocks with velocity
- Focus ratio and idle gaps
- Error correction patterns
- Top 10 files by time spent

**3. JSON Exporter (exporter/json.go)**
- Machine-readable format
- Identical structure to raw JSON
- For programmatic analysis

**4. SQLite Exporter (exporter/sqlite.go)**
- Schema:
  ```sql
  sessions (id, project_path, start_time, end_time, duration)
  events (id, session_id, type, timestamp, data_json)
  ```
- Enables cross-session queries
- Aggregation and reporting

---

## Key Technical Innovations

### 1. Context-Aware Filtering

**Innovation:** Instead of dumb time-based debouncing, the filter understands semantic event boundaries.

**Example:**
```
10:00:00 - Cursor move to line 45
10:00:01 - Cursor move to line 46  [SUPPRESSED - within 200ms]
10:00:02 - File edit at line 46    [TRIGGERS: Flush cursor + record edit]
```

**Result:** Cursor position is always accurate relative to edits, even with aggressive filtering.

### 2. Velocity-Based Productivity Metrics

**Innovation:** Using `changed_tick` (Vim's buffer modification counter) as a proxy for coding speed.

**Why it works:**
- Each keystroke increments changed_tick
- Undo/redo affects tick count
- Tick delta ÷ time = quantifiable productivity

**Limitations:**
- Auto-formatters inflate velocity
- Copy-paste shows high velocity but low cognitive load

### 3. Crash-Safe Dual Architecture

**Innovation:** Separate "source of truth" (raw JSON) from "derived insights" (summary).

**Benefits:**
- Raw JSON persists on every event → survive Neovim crashes
- Summary can be regenerated from raw JSON → safe to fail
- Periodic updates keep summary fresh without blocking

### 4. Pure Go SQLite

**Innovation:** Using `modernc.org/sqlite` instead of CGO-based drivers.

**Benefits:**
- No C compiler required for builds
- Cross-platform compatibility (Windows, Linux, macOS)
- Smaller binary size
- Easier distribution

---

## Performance Characteristics

### Time Complexity

**Event Recording:** O(1)
- Append to Events[] array
- Single file write

**Cursor Filtering:** O(1)
- Hash map lookup for context triggers
- Timer management

**Aggregation:** O(n)
- Single pass through events for block building
- Second pass for idle gaps
- Third pass for focus metrics

**Session List:** O(n·log n)
- Directory scan: O(n files)
- Deduplication: O(n) with hash map

### Space Complexity

**Memory Usage:**
- Session struct: ~100 bytes + Events array
- Events array: ~200 bytes per event
- 1-hour session (~500 events): ~100 KB in memory
- CursorFilter: ~500 bytes (timers + pending event)

**Disk Usage:**
- Raw JSON: ~2 KB per 10 events → ~1 MB per hour
- SESSION_SUMMARY.md: ~100 KB (constant size)
- SQLite: ~1.5x raw JSON size (with indexes)

### Benchmark Results (from README)

**Event Filtering:**
- Before: ~1000 cursor events/session
- After: ~100 cursor events/session
- Reduction: **90%**

**Binary Size:**
- macOS: 9.8 MB
- Linux: 8.5 MB
- Windows: 9.2 MB

**Runtime Overhead:**
- Idle: < 5 MB RAM
- Recording: 5-15 MB RAM
- CPU: < 5% (event spikes to ~10% during aggregation)

---

## Conclusion

**capytrace.nvim** demonstrates sophisticated software engineering through:

1. **Multi-language architecture:** Lua for Neovim integration, Go for performance-critical backend
2. **Advanced filtering algorithms:** Debouncing + idle detection + context awareness reduces noise by 90%
3. **Intelligent aggregation:** Three Golden Rules transform raw events into semantic work units
4. **Analytics innovation:** Velocity-based flow state detection and focus ratio metrics
5. **Crash-safe design:** Dual-file architecture with immediate persistence and periodic aggregation
6. **Concurrency patterns:** Background goroutines, mutexes, channels for non-blocking operation

The project successfully bridges the gap between high-frequency event capture (cursor movements, edits) and meaningful insights (productivity patterns, error correction), making it a valuable tool for developer workflow analysis and self-improvement.

### Future Research Directions

1. **Machine Learning Integration:** Predict bugs based on error correction patterns
2. **Collaborative Analytics:** Team-wide velocity and focus benchmarking
3. **LSP Integration:** Correlate diagnostics with productivity drops
4. **Replay Mode:** Visualize coding sessions as playback timeline
5. **Git Correlation:** Link activity blocks to commit messages for better commit granularity

---

**Document Version:** 1.0
**Analysis Date:** April 7, 2026
**Total Lines of Code Analyzed:** ~2,500 (Go) + ~700 (Lua)
**Key Files Reviewed:** 10 core implementation files + 3 documentation files
