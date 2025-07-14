package recorder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      EventData `json:"data"`
}

type EventData struct {
	// File edit events
	Filename    string `json:"filename,omitempty"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	LineCount   int    `json:"line_count,omitempty"`
	ChangedTick int    `json:"changed_tick,omitempty"`

	// Terminal events
	Command string `json:"command,omitempty"`

	// Annotation events
	Note string `json:"note,omitempty"`

	// Cursor events
	PrevLine   int `json:"prev_line,omitempty"`
	PrevColumn int `json:"prev_column,omitempty"`
}

// Add event types for channel-based debounced logging

type editEvent struct {
	filename    string
	line        string
	col         string
	lineCount   string
	changedTick string
}

type cursorEvent struct {
	filename string
	line     string
	col      string
}

// Add channels to Session

type Session struct {
	ID           string    `json:"id"`
	ProjectPath  string    `json:"project_path"`
	SavePath     string    `json:"save_path"`
	OutputFormat string    `json:"output_format"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time,omitempty"`
	Events       []Event   `json:"events"`
	Active       bool      `json:"active"`

	editChan   chan editEvent
	cursorChan chan cursorEvent
}

var activeSessions = make(map[string]*Session)

func NewSession(id, projectPath, savePath, outputFormat string) *Session {
	s := &Session{
		ID:           id,
		ProjectPath:  projectPath,
		SavePath:     savePath,
		OutputFormat: outputFormat,
		StartTime:    time.Now(),
		Events:       []Event{},
		Active:       true,
		editChan:     make(chan editEvent, 100),
		cursorChan:   make(chan cursorEvent, 100),
	}
	go s.debouncedEditLogger()
	go s.debouncedCursorLogger()
	return s
}

func (s *Session) Start() error {
	activeSessions[s.ID] = s

	// Record initial event
	s.Events = append(s.Events, Event{
		Type:      "session_start",
		Timestamp: s.StartTime,
		Data: EventData{
			Note: fmt.Sprintf("Started debugging session in %s", s.ProjectPath),
		},
	})

	return s.save()
}

func (s *Session) End() error {
	s.EndTime = time.Now()
	s.Active = false

	// Record end event
	s.Events = append(s.Events, Event{
		Type:      "session_end",
		Timestamp: s.EndTime,
		Data: EventData{
			Note: "Debugging session ended",
		},
	})

	delete(activeSessions, s.ID)
	return s.save()
}

func (s *Session) AddAnnotation(note string) error {
	event := Event{
		Type:      "annotation",
		Timestamp: time.Now(),
		Data: EventData{
			Note: note,
		},
	}

	s.Events = append(s.Events, event)
	return s.save()
}

// Debounced logger for file edits
func (s *Session) debouncedEditLogger() {
	var (
		lastEvent editEvent
		timer     *time.Timer
	)
	for evt := range s.editChan {
		lastEvent = evt
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(500*time.Millisecond, func() {
			s.logEditEvent(lastEvent)
		})
	}
}

// Debounced logger for cursor moves
func (s *Session) debouncedCursorLogger() {
	var (
		lastEvent cursorEvent
		timer     *time.Timer
	)
	for evt := range s.cursorChan {
		lastEvent = evt
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(500*time.Millisecond, func() {
			s.logCursorEvent(lastEvent)
		})
	}
}

// Push file edit event to channel
func (s *Session) PushEditEvent(filename, line, col, lineCount, changedTick string) {
	s.editChan <- editEvent{filename, line, col, lineCount, changedTick}
}

// Push cursor event to channel
func (s *Session) PushCursorEvent(filename, line, col string) {
	s.cursorChan <- cursorEvent{filename, line, col}
}

// Actual logging logic for file edit (called by debounced goroutine)
func (s *Session) logEditEvent(evt editEvent) {
	lineNum, _ := strconv.Atoi(evt.line)
	colNum, _ := strconv.Atoi(evt.col)
	lineCountNum, _ := strconv.Atoi(evt.lineCount)
	changedTickNum, _ := strconv.Atoi(evt.changedTick)

	event := Event{
		Type:      "file_edit",
		Timestamp: time.Now(),
		Data: EventData{
			Filename:    evt.filename,
			Line:        lineNum,
			Column:      colNum,
			LineCount:   lineCountNum,
			ChangedTick: changedTickNum,
		},
	}

	s.Events = append(s.Events, event)
	_ = s.save()
}

// Actual logging logic for cursor move (called by debounced goroutine)
func (s *Session) logCursorEvent(evt cursorEvent) {
	lineNum, _ := strconv.Atoi(evt.line)
	colNum, _ := strconv.Atoi(evt.col)

	event := Event{
		Type:      "cursor_move",
		Timestamp: time.Now(),
		Data: EventData{
			Filename: evt.filename,
			Line:     lineNum,
			Column:   colNum,
		},
	}

	s.Events = append(s.Events, event)
	_ = s.save()
}

// Update RecordEdit and RecordCursorMove to use channels
func (s *Session) RecordEdit(filename, line, col, lineCount, changedTick string) error {
	s.PushEditEvent(filename, line, col, lineCount, changedTick)
	return nil
}

func (s *Session) RecordCursorMove(filename, line, col string) error {
	s.PushCursorEvent(filename, line, col)
	return nil
}

func (s *Session) RecordTerminalCommand(command string) error {
	event := Event{
		Type:      "terminal_command",
		Timestamp: time.Now(),
		Data: EventData{
			Command: command,
		},
	}

	s.Events = append(s.Events, event)
	return s.save()
}

func (s *Session) save() error {
	sessionPath := filepath.Join(s.SavePath, s.ID+".json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sessionPath, data, 0644)
}

func LoadSession(sessionID string, savePath string) (*Session, error) {
	if session, exists := activeSessions[sessionID]; exists {
		return session, nil
	}

	// Try to load from file
	sessionPath := filepath.Join(savePath, sessionID+".json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	if session.Active {
		activeSessions[sessionID] = &session
	}

	return &session, nil
}

func ListSessions(savePath string) ([]string, error) {
	files, err := os.ReadDir(savePath)
	if err != nil {
		return nil, err
	}

	var sessions []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			sessions = append(sessions, file.Name()[:len(file.Name())-5]) // Remove .json extension
		}
	}

	return sessions, nil
}

func ResumeSession(sessionName, savePath string) (*Session, error) {
	sessionPath := filepath.Join(savePath, sessionName+".json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	session.Active = true
	activeSessions[session.ID] = &session

	// Record resume event
	session.Events = append(session.Events, Event{
		Type:      "session_resume",
		Timestamp: time.Now(),
		Data: EventData{
			Note: "Session resumed",
		},
	})

	return &session, session.save()
}

func getDefaultSavePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "capytrace_logs")
}
