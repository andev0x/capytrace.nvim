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
