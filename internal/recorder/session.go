// Package recorder manages debugging session state, event buffering, and persistence.
// It coordinates with the filter package to ensure only meaningful events are recorded.
package recorder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andev0x/capytrace.nvim/internal/aggregator"
	"github.com/andev0x/capytrace.nvim/internal/exporter"
	"github.com/andev0x/capytrace.nvim/internal/filter"
	"github.com/andev0x/capytrace.nvim/internal/models"
)

var (
	activeSessions   = make(map[string]*Session)
	activeSessionsMu sync.RWMutex
)

// Session wraps a models.Session with additional runtime state and filtering capabilities.
type Session struct {
	*models.Session
	mu               sync.Mutex
	cursorFilter     *filter.CursorFilter
	aggregatorConfig *aggregator.AggregatorConfig
	periodicTicker   *time.Ticker
	stopPeriodicChan chan struct{}
}

// NewSession creates a new debugging session with the specified parameters.
// It initializes the session with a cursor filter to reduce noise from rapid movements
// and starts a background goroutine for periodic SESSION_SUMMARY.md updates.
func NewSession(id, projectPath, savePath, outputFormat string, filterConfig *filter.FilterConfig) *Session {
	session := &Session{
		Session: &models.Session{
			ID:           id,
			ProjectPath:  projectPath,
			SavePath:     savePath,
			OutputFormat: outputFormat,
			StartTime:    time.Now(),
			Events:       []models.Event{},
			Active:       true,
		},
		cursorFilter:     filter.NewCursorFilter(filterConfig),
		aggregatorConfig: aggregator.DefaultConfig(),
		stopPeriodicChan: make(chan struct{}),
	}

	// Start periodic aggregation updates (every 5 minutes)
	session.startPeriodicAggregation(5 * time.Minute)

	return session
}

// startPeriodicAggregation starts a background goroutine that regenerates
// SESSION_SUMMARY.md every interval.
func (s *Session) startPeriodicAggregation(interval time.Duration) {
	s.periodicTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-s.periodicTicker.C:
				// Regenerate SESSION_SUMMARY.md
				s.regenerateSummary()
			case <-s.stopPeriodicChan:
				return
			}
		}
	}()
}

// regenerateSummary updates the SESSION_SUMMARY.md file with current session data.
func (s *Session) regenerateSummary() {
	s.mu.Lock()
	sessionCopy := *s.Session
	s.mu.Unlock()

	// Use SmartMarkdownExporter to generate updated summary
	smartExporter := exporter.NewSmartMarkdownExporter(s.aggregatorConfig)
	if err := smartExporter.Export(&sessionCopy, s.SavePath); err != nil {
		// Log error but don't fail the session
		fmt.Fprintf(os.Stderr, "Failed to regenerate session summary: %v\n", err)
	}
}

// stopPeriodicAggregation stops the background aggregation goroutine.
func (s *Session) stopPeriodicAggregation() {
	if s.periodicTicker != nil {
		s.periodicTicker.Stop()
	}
	close(s.stopPeriodicChan)
}

// Start begins recording a new debugging session and persists it to disk.
func (s *Session) Start() error {
	activeSessionsMu.Lock()
	activeSessions[s.ID] = s
	activeSessionsMu.Unlock()

	// Record initial event
	s.addEvent(models.Event{
		Type:      "session_start",
		Timestamp: s.StartTime,
		Data: models.EventData{
			Note: fmt.Sprintf("Started debugging session in %s", s.ProjectPath),
		},
	})

	return s.save()
}

// End terminates the current session and performs final cleanup.
func (s *Session) End() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop periodic aggregation
	s.stopPeriodicAggregation()

	// Flush any pending cursor events
	if pendingEvent := s.cursorFilter.FlushPending(); pendingEvent != nil {
		s.Session.Events = append(s.Session.Events, *pendingEvent)
	}

	s.cursorFilter.Stop()
	s.EndTime = time.Now()
	s.Active = false

	// Record end event
	s.Session.Events = append(s.Session.Events, models.Event{
		Type:      "session_end",
		Timestamp: s.EndTime,
		Data: models.EventData{
			Note: "Debugging session ended",
		},
	})

	activeSessionsMu.Lock()
	delete(activeSessions, s.ID)
	activeSessionsMu.Unlock()

	// Save raw JSON
	if err := s.save(); err != nil {
		return err
	}

	// Generate final SESSION_SUMMARY.md
	s.regenerateSummary()

	return nil
}

// AddAnnotation adds a user-provided note to the session timeline.
func (s *Session) AddAnnotation(note string) error {
	event := models.Event{
		Type:      "annotation",
		Timestamp: time.Now(),
		Data: models.EventData{
			Note: note,
		},
	}

	return s.addEvent(event)
}

// RecordEdit records a file modification event with position and metadata.
func (s *Session) RecordEdit(filename, line, col, lineCount, changedTick string) error {
	lineNum, _ := strconv.Atoi(line)
	colNum, _ := strconv.Atoi(col)
	lineCountNum, _ := strconv.Atoi(lineCount)
	changedTickNum, _ := strconv.Atoi(changedTick)

	event := models.Event{
		Type:      "file_edit",
		Timestamp: time.Now(),
		Data: models.EventData{
			Filename:    filename,
			Line:        lineNum,
			Column:      colNum,
			LineCount:   lineCountNum,
			ChangedTick: changedTickNum,
		},
	}

	// File edits are context triggers - process through filter first
	if filteredEvent := s.cursorFilter.ProcessEvent(&event); filteredEvent != nil {
		s.addEvent(*filteredEvent)
	}

	return s.addEvent(event)
}

// RecordTerminalCommand records a terminal command execution.
func (s *Session) RecordTerminalCommand(command string) error {
	event := models.Event{
		Type:      "terminal_command",
		Timestamp: time.Now(),
		Data: models.EventData{
			Command: command,
		},
	}

	// Terminal commands are context triggers
	if filteredEvent := s.cursorFilter.ProcessEvent(&event); filteredEvent != nil {
		s.addEvent(*filteredEvent)
	}

	return s.addEvent(event)
}

// RecordCursorMove records cursor position changes with intelligent filtering.
// Rapid movements are debounced and only committed when cursor remains idle.
func (s *Session) RecordCursorMove(filename, line, col string) error {
	lineNum, _ := strconv.Atoi(line)
	colNum, _ := strconv.Atoi(col)

	event := models.Event{
		Type:      "cursor_move",
		Timestamp: time.Now(),
		Data: models.EventData{
			Filename: filename,
			Line:     lineNum,
			Column:   colNum,
		},
	}

	// Process through cursor filter - may return nil if debounced
	if filteredEvent := s.cursorFilter.ProcessEvent(&event); filteredEvent != nil {
		return s.addEvent(*filteredEvent)
	}

	return nil
}

// RecordFileOpen records when a file is opened in the editor.
func (s *Session) RecordFileOpen(filename, filetype string) error {
	event := models.Event{
		Type:      "file_open",
		Timestamp: time.Now(),
		Data: models.EventData{
			Filename: filename,
			FileType: filetype,
		},
	}

	return s.addEvent(event)
}

// RecordLSPDiagnostic records LSP diagnostic messages (errors, warnings, etc.).
func (s *Session) RecordLSPDiagnostic(filename, line, col, message, level string) error {
	lineNum, _ := strconv.Atoi(line)
	colNum, _ := strconv.Atoi(col)

	event := models.Event{
		Type:      "lsp_diagnostic",
		Timestamp: time.Now(),
		Data: models.EventData{
			Filename: filename,
			Line:     lineNum,
			Column:   colNum,
			Message:  message,
			Level:    level,
		},
	}

	return s.addEvent(event)
}

// addEvent appends an event to the session and persists it.
func (s *Session) addEvent(event models.Event) error {
	s.mu.Lock()
	s.Session.Events = append(s.Session.Events, event)
	s.mu.Unlock()

	return s.save()
}

// save persists the session state to disk as JSON (The Truth).
func (s *Session) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Save as {session_id}_raw.json to distinguish from exported versions
	sessionPath := filepath.Join(s.SavePath, s.ID+"_raw.json")
	data, err := json.MarshalIndent(s.Session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sessionPath, data, 0644)
}

// LoadSession retrieves a session from memory or disk.
func LoadSession(sessionID string, savePath string, filterConfig *filter.FilterConfig) (*Session, error) {
	activeSessionsMu.RLock()
	if session, exists := activeSessions[sessionID]; exists {
		activeSessionsMu.RUnlock()
		return session, nil
	}
	activeSessionsMu.RUnlock()

	// Try to load from file (try both _raw.json and .json for backwards compatibility)
	sessionPath := filepath.Join(savePath, sessionID+"_raw.json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		// Try old naming scheme
		sessionPath = filepath.Join(savePath, sessionID+".json")
		data, err = os.ReadFile(sessionPath)
		if err != nil {
			return nil, err
		}
	}

	var modelSession models.Session
	if err := json.Unmarshal(data, &modelSession); err != nil {
		return nil, err
	}

	session := &Session{
		Session:          &modelSession,
		cursorFilter:     filter.NewCursorFilter(filterConfig),
		aggregatorConfig: aggregator.DefaultConfig(),
		stopPeriodicChan: make(chan struct{}),
	}

	if session.Active {
		activeSessionsMu.Lock()
		activeSessions[sessionID] = session
		activeSessionsMu.Unlock()

		// Restart periodic aggregation for active sessions
		session.startPeriodicAggregation(5 * time.Minute)
	}

	return session, nil
}

// ListSessions returns a list of all saved session IDs in the given directory.
func ListSessions(savePath string) ([]string, error) {
	files, err := os.ReadDir(savePath)
	if err != nil {
		return nil, err
	}

	var sessions []string
	seen := make(map[string]bool)

	for _, file := range files {
		name := file.Name()

		// Handle both _raw.json and .json extensions
		if strings.HasSuffix(name, "_raw.json") {
			sessionID := strings.TrimSuffix(name, "_raw.json")
			if !seen[sessionID] {
				sessions = append(sessions, sessionID)
				seen[sessionID] = true
			}
		} else if strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, "_export.json") {
			sessionID := strings.TrimSuffix(name, ".json")
			if !seen[sessionID] {
				sessions = append(sessions, sessionID)
				seen[sessionID] = true
			}
		}
	}

	return sessions, nil
}

// ResumeSession loads a previously saved session and marks it as active again.
func ResumeSession(sessionName, savePath string, filterConfig *filter.FilterConfig) (*Session, error) {
	// Try to load from file (try both _raw.json and .json for backwards compatibility)
	sessionPath := filepath.Join(savePath, sessionName+"_raw.json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		// Try old naming scheme
		sessionPath = filepath.Join(savePath, sessionName+".json")
		data, err = os.ReadFile(sessionPath)
		if err != nil {
			return nil, err
		}
	}

	var modelSession models.Session
	if err := json.Unmarshal(data, &modelSession); err != nil {
		return nil, err
	}

	session := &Session{
		Session:          &modelSession,
		cursorFilter:     filter.NewCursorFilter(filterConfig),
		aggregatorConfig: aggregator.DefaultConfig(),
		stopPeriodicChan: make(chan struct{}),
	}

	session.Active = true

	activeSessionsMu.Lock()
	activeSessions[session.ID] = session
	activeSessionsMu.Unlock()

	// Start periodic aggregation
	session.startPeriodicAggregation(5 * time.Minute)

	// Record resume event
	session.addEvent(models.Event{
		Type:      "session_resume",
		Timestamp: time.Now(),
		Data: models.EventData{
			Note: "Session resumed",
		},
	})

	return session, session.save()
}
