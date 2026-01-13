// capytrace is a high-performance, local-first development tracer for Neovim.
// It records debugging sessions including file edits, terminal commands, and LSP diagnostics.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andev0x/capytrace.nvim/internal/exporter"
	"github.com/andev0x/capytrace.nvim/internal/filter"
	"github.com/andev0x/capytrace.nvim/internal/models"
	"github.com/andev0x/capytrace.nvim/internal/recorder"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  start              Start a new session\n")
		fmt.Fprintf(os.Stderr, "  end                End current session\n")
		fmt.Fprintf(os.Stderr, "  annotate           Add annotation to session\n")
		fmt.Fprintf(os.Stderr, "  record-edit        Record file edit event\n")
		fmt.Fprintf(os.Stderr, "  record-terminal    Record terminal command\n")
		fmt.Fprintf(os.Stderr, "  record-cursor      Record cursor movement\n")
		fmt.Fprintf(os.Stderr, "  record-file-open   Record file open event\n")
		fmt.Fprintf(os.Stderr, "  record-lsp-diagnostic  Record LSP diagnostic\n")
		fmt.Fprintf(os.Stderr, "  list               List all sessions\n")
		fmt.Fprintf(os.Stderr, "  resume             Resume a previous session\n")
		fmt.Fprintf(os.Stderr, "  stats              Show session statistics\n")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start":
		handleStart()
	case "end":
		handleEnd()
	case "annotate":
		handleAnnotate()
	case "record-edit":
		handleRecordEdit()
	case "record-terminal":
		handleRecordTerminal()
	case "record-cursor":
		handleRecordCursor()
	case "list":
		handleList()
	case "resume":
		handleResume()
	case "record-file-open":
		handleRecordFileOpen()
	case "record-lsp-diagnostic":
		handleRecordLSPDiagnostic()
	case "stats":
		handleStats()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

// handleStart initializes a new debugging session.
func handleStart() {
	if len(os.Args) < 6 {
		fmt.Fprintf(os.Stderr, "Usage: start <session_id> <project_path> <save_path> <output_format>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	projectPath := os.Args[3]
	savePath := os.Args[4]
	outputFormat := os.Args[5]

	// Use default filter configuration
	session := recorder.NewSession(sessionID, projectPath, savePath, outputFormat, filter.DefaultFilterConfig())
	if err := session.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Session started: %s\n", sessionID)
}

// handleEnd terminates the current session and exports it.
func handleEnd() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: end <session_id> <save_path>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]

	session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.End(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to end session: %v\n", err)
		os.Exit(1)
	}

	// Export session based on format
	var exp exporter.Exporter
	switch session.OutputFormat {
	case "json":
		exp = &exporter.JSONExporter{}
	case "sqlite":
		home, _ := os.UserHomeDir()
		dataDir := filepath.Join(home, ".local", "share", "capytrace")
		os.MkdirAll(dataDir, 0755)
		exp = exporter.NewSQLiteExporter(dataDir)
	default:
		exp = &exporter.MarkdownExporter{}
	}

	if err := exp.Export(session.Session, session.SavePath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Session ended and exported: %s\n", sessionID)
}

// handleAnnotate adds a user note to the current session.
func handleAnnotate() {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Usage: annotate <session_id> <save_path> <note>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	note := os.Args[4]

	session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.AddAnnotation(note); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add annotation: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Annotation added\n")
}

// handleRecordEdit records a file modification event.
func handleRecordEdit() {
	if len(os.Args) < 9 {
		fmt.Fprintf(os.Stderr, "Usage: record-edit <session_id> <save_path> <filename> <line> <col> <line_count> <changedtick>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	filename := os.Args[4]
	line := os.Args[5]
	col := os.Args[6]
	lineCount := os.Args[7]
	changedTick := os.Args[8]

	session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.RecordEdit(filename, line, col, lineCount, changedTick); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to record edit: %v\n", err)
		os.Exit(1)
	}
}

// handleRecordTerminal records a terminal command execution.
func handleRecordTerminal() {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Usage: record-terminal <session_id> <save_path> <command>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	command := os.Args[4]

	session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.RecordTerminalCommand(command); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to record terminal command: %v\n", err)
		os.Exit(1)
	}
}

// handleRecordCursor records cursor position changes with intelligent filtering.
func handleRecordCursor() {
	if len(os.Args) < 7 {
		fmt.Fprintf(os.Stderr, "Usage: record-cursor <session_id> <save_path> <filename> <line> <col>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	filename := os.Args[4]
	line := os.Args[5]
	col := os.Args[6]

	session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.RecordCursorMove(filename, line, col); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to record cursor movement: %v\n", err)
		os.Exit(1)
	}
}

// handleList displays all available sessions.
func handleList() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: list <save_path>\n")
		os.Exit(1)
	}

	savePath := os.Args[2]
	sessions, err := recorder.ListSessions(savePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list sessions: %v\n", err)
		os.Exit(1)
	}

	for _, session := range sessions {
		fmt.Println(session)
	}
}

// handleResume reactivates a previously saved session.
func handleResume() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: resume <session_name> <save_path>\n")
		os.Exit(1)
	}

	sessionName := os.Args[2]
	savePath := os.Args[3]

	session, err := recorder.ResumeSession(sessionName, savePath, filter.DefaultFilterConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resume session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Session resumed: %s\n", session.ID)
}

// handleRecordFileOpen records when a file is opened in the editor.
func handleRecordFileOpen() {
	if len(os.Args) < 6 {
		fmt.Fprintf(os.Stderr, "Usage: record-file-open <session_id> <save_path> <filename> <filetype>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	filename := os.Args[4]
	filetype := os.Args[5]

	session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.RecordFileOpen(filename, filetype); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to record file open: %v\n", err)
		os.Exit(1)
	}
}

// handleRecordLSPDiagnostic records LSP diagnostic messages.
func handleRecordLSPDiagnostic() {
	if len(os.Args) < 9 {
		fmt.Fprintf(os.Stderr, "Usage: record-lsp-diagnostic <session_id> <save_path> <filename> <line> <col> <message> <level>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	filename := os.Args[4]
	line := os.Args[5]
	col := os.Args[6]
	message := os.Args[7]
	level := os.Args[8]

	session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.RecordLSPDiagnostic(filename, line, col, message, level); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to record lsp diagnostic: %v\n", err)
		os.Exit(1)
	}
}

// handleStats displays statistics for a session or all sessions.
func handleStats() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: stats <save_path> [session_id]\n")
		os.Exit(1)
	}

	savePath := os.Args[2]

	// Check if using SQLite backend
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".local", "share", "capytrace")
	sqliteExporter := exporter.NewSQLiteExporter(dataDir)

	if len(os.Args) >= 4 {
		// Show stats for specific session
		sessionID := os.Args[3]

		// Try SQLite first
		if summary, err := sqliteExporter.GetSessionStats(sessionID); err == nil {
			printSessionSummary(summary)
			return
		}

		// Fall back to JSON file
		session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
			os.Exit(1)
		}
		printSessionStats(session)
	} else {
		// Show stats for all sessions
		sessions, err := recorder.ListSessions(savePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list sessions: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Session Statistics")
		fmt.Println("==================")
		for _, sessionID := range sessions {
			session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
			if err != nil {
				continue
			}
			fmt.Printf("\n%s:\n", sessionID)
			printSessionStats(session)
		}
	}
}

// printSessionStats outputs formatted statistics for a session.
func printSessionStats(session *recorder.Session) {
	duration := session.EndTime.Sub(session.StartTime)
	if session.Active {
		fmt.Printf("  Status: Active\n")
	} else {
		fmt.Printf("  Status: Completed\n")
		fmt.Printf("  Duration: %s\n", duration)
	}

	eventCounts := make(map[string]int)
	for _, event := range session.Events {
		eventCounts[event.Type]++
	}

	fmt.Printf("  Total Events: %d\n", len(session.Events))
	fmt.Printf("  File Edits: %d\n", eventCounts["file_edit"])
	fmt.Printf("  Cursor Moves: %d\n", eventCounts["cursor_move"])
	fmt.Printf("  Terminal Commands: %d\n", eventCounts["terminal_command"])
	fmt.Printf("  Annotations: %d\n", eventCounts["annotation"])
}

// printSessionSummary outputs formatted statistics from a SessionSummary.
func printSessionSummary(summary *models.SessionSummary) {
	fmt.Printf("Session: %s\n", summary.ID)
	fmt.Printf("  Project: %s\n", summary.ProjectPath)
	fmt.Printf("  Duration: %s\n", summary.Duration)
	fmt.Printf("  Total Events: %d\n", summary.TotalEvents)
	fmt.Printf("  File Edits: %d\n", summary.FileEdits)
	fmt.Printf("  Cursor Moves: %d\n", summary.CursorMoves)
	fmt.Printf("  Terminal Commands: %d\n", summary.TerminalCommands)
	fmt.Printf("  Annotations: %d\n", summary.Annotations)
}
