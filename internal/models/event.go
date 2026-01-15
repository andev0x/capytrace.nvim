// Package models contains shared data structures used throughout the capytrace application.
package models

import "time"

// Event represents a single recorded event in a debugging session.
// Each event has a type, timestamp, and associated data.
type Event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      EventData `json:"data"`
}

// EventData contains the payload information for different event types.
// Fields are populated based on the event type being recorded.
type EventData struct {
	// File edit events
	Filename    string `json:"filename,omitempty"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	LineCount   int    `json:"line_count,omitempty"`
	ChangedTick int    `json:"changed_tick,omitempty"`

	// File open events
	FileType string `json:"file_type,omitempty"`

	// LSP diagnostics events
	Message string `json:"message,omitempty"`
	Level   string `json:"level,omitempty"`

	// Terminal events
	Command string `json:"command,omitempty"`

	// Annotation events
	Note string `json:"note,omitempty"`

	// Cursor events
	PrevLine   int `json:"prev_line,omitempty"`
	PrevColumn int `json:"prev_column,omitempty"`
}

// Session represents a complete debugging session with all recorded events.
type Session struct {
	ID           string    `json:"id"`
	ProjectPath  string    `json:"project_path"`
	SavePath     string    `json:"save_path"`
	OutputFormat string    `json:"output_format"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time,omitempty"`
	Events       []Event   `json:"events"`
	Active       bool      `json:"active"`
}

// SessionSummary provides statistics about a session for display purposes.
type SessionSummary struct {
	ID               string
	ProjectPath      string
	Duration         time.Duration
	TotalEvents      int
	FileEdits        int
	TerminalCommands int
	Annotations      int
	CursorMoves      int
}

// ActivityBlock represents a merged group of related events within a short time window.
// Used for aggregating file_edit events that occur within 2 seconds.
type ActivityBlock struct {
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	Filename   string        `json:"filename"`
	EventCount int           `json:"event_count"`
	StartTick  int           `json:"start_tick"`
	EndTick    int           `json:"end_tick"`
	DeltaTick  int           `json:"delta_tick"`
	Velocity   float64       `json:"velocity"` // Delta Tick / Duration in seconds
	Events     []Event       `json:"events"`
	ClosedBy   string        `json:"closed_by"` // "context_switch", "idle", "timeout"
}

// IdleGap represents a period of inactivity during the session.
// Used to identify moments when the developer might be stuck or taking a break.
type IdleGap struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

// SessionAnalytics provides advanced metrics about a coding session.
type SessionAnalytics struct {
	// Velocity metrics
	AverageVelocity float64         `json:"average_velocity"`
	PeakVelocity    float64         `json:"peak_velocity"`
	FlowBlocks      []ActivityBlock `json:"flow_blocks"` // High velocity blocks

	// Focus metrics
	FocusRatio      float64        `json:"focus_ratio"`      // Main file time / Total time
	MainFiles       map[string]int `json:"main_files"`       // File -> time spent (seconds)
	DistractionTime int            `json:"distraction_time"` // Time in NvimTree, copilot-chat, etc.

	// Error correction patterns
	ErrorCorrections []ErrorPattern `json:"error_corrections"`

	// Idle analysis
	IdleGaps      []IdleGap     `json:"idle_gaps"`
	TotalIdleTime time.Duration `json:"total_idle_time"`
}

// ErrorPattern represents a detected error correction event.
type ErrorPattern struct {
	Timestamp      time.Time `json:"timestamp"`
	Filename       string    `json:"filename"`
	Annotation     string    `json:"annotation"`
	LinesDeleted   int       `json:"lines_deleted"`
	TicksReversed  int       `json:"ticks_reversed"` // Negative delta
	BlocksAffected int       `json:"blocks_affected"`
}
