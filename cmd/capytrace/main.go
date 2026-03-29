// capytrace is a high-performance, local-first development tracer for Neovim.
// It records debugging sessions including file edits, terminal commands, and LSP diagnostics.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andev0x/capytrace.nvim/internal/exporter"
	"github.com/andev0x/capytrace.nvim/internal/filter"
	"github.com/andev0x/capytrace.nvim/internal/models"
	"github.com/andev0x/capytrace.nvim/internal/recorder"
)

type daemonRequest struct {
	ID      int      `json:"id"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type daemonResponse struct {
	ID         int    `json:"id"`
	OK         bool   `json:"ok"`
	Result     string `json:"result,omitempty"`
	Error      string `json:"error,omitempty"`
	ReportPath string `json:"report_path,omitempty"`
}

type commandResult struct {
	Message    string
	ReportPath string
}

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
		fmt.Fprintf(os.Stderr, "  daemon             Start long-lived daemon mode\n")
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
	case "daemon":
		runDaemon()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func runDaemon() {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req daemonRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_ = encoder.Encode(daemonResponse{ID: req.ID, OK: false, Error: err.Error()})
			continue
		}

		result, err := executeDaemonCommand(req.Command, req.Args)
		if err != nil {
			_ = encoder.Encode(daemonResponse{ID: req.ID, OK: false, Error: err.Error()})
			continue
		}

		_ = encoder.Encode(daemonResponse{ID: req.ID, OK: true, Result: result.Message, ReportPath: result.ReportPath})
	}
}

func executeDaemonCommand(command string, args []string) (commandResult, error) {
	switch command {
	case "start":
		if len(args) < 4 {
			return commandResult{}, fmt.Errorf("start requires 4 args")
		}
		sessionID, projectPath, savePath, outputFormat := args[0], args[1], args[2], args[3]
		session := recorder.NewSession(sessionID, projectPath, savePath, outputFormat, filter.DefaultFilterConfig())
		if err := session.Start(); err != nil {
			return commandResult{}, err
		}
		return commandResult{Message: "Session started: " + sessionID}, nil
	case "end":
		if len(args) < 2 {
			return commandResult{}, fmt.Errorf("end requires 2 args")
		}
		sessionID, savePath := args[0], args[1]
		session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
		if err != nil {
			return commandResult{}, err
		}
		if err := session.End(); err != nil {
			return commandResult{}, err
		}

		var exp exporter.Exporter
		switch session.OutputFormat {
		case "json":
			exp = &exporter.JSONExporter{}
		case "sqlite":
			home, _ := os.UserHomeDir()
			dataDir := filepath.Join(home, ".local", "share", "capytrace")
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				return commandResult{}, err
			}
			exp = exporter.NewSQLiteExporter(dataDir)
		default:
			exp = &exporter.MarkdownExporter{}
		}

		if err := exp.Export(session.Session, session.SavePath); err != nil {
			return commandResult{}, err
		}

		reportPath := filepath.Join(savePath, sessionID+".md")
		return commandResult{Message: "Session ended and exported: " + sessionID, ReportPath: reportPath}, nil
	case "annotate":
		if len(args) < 3 {
			return commandResult{}, fmt.Errorf("annotate requires 3 args")
		}
		session, err := recorder.LoadSession(args[0], args[1], filter.DefaultFilterConfig())
		if err != nil {
			return commandResult{}, err
		}
		if err := session.AddAnnotation(args[2]); err != nil {
			return commandResult{}, err
		}
		return commandResult{Message: "Annotation added"}, nil
	case "record-edit":
		if len(args) < 8 {
			return commandResult{}, fmt.Errorf("record-edit requires 8 args")
		}
		session, err := recorder.LoadSession(args[0], args[1], filter.DefaultFilterConfig())
		if err != nil {
			return commandResult{}, err
		}
		if err := session.RecordEdit(args[2], args[3], args[4], args[5], args[6], args[7]); err != nil {
			return commandResult{}, err
		}
		return commandResult{}, nil
	case "record-terminal":
		if len(args) < 3 {
			return commandResult{}, fmt.Errorf("record-terminal requires 3 args")
		}
		session, err := recorder.LoadSession(args[0], args[1], filter.DefaultFilterConfig())
		if err != nil {
			return commandResult{}, err
		}
		if err := session.RecordTerminalCommand(args[2]); err != nil {
			return commandResult{}, err
		}
		return commandResult{}, nil
	case "record-cursor":
		if len(args) < 5 {
			return commandResult{}, fmt.Errorf("record-cursor requires 5 args")
		}
		session, err := recorder.LoadSession(args[0], args[1], filter.DefaultFilterConfig())
		if err != nil {
			return commandResult{}, err
		}
		if err := session.RecordCursorMove(args[2], args[3], args[4]); err != nil {
			return commandResult{}, err
		}
		return commandResult{}, nil
	case "record-file-open":
		if len(args) < 4 {
			return commandResult{}, fmt.Errorf("record-file-open requires 4 args")
		}
		session, err := recorder.LoadSession(args[0], args[1], filter.DefaultFilterConfig())
		if err != nil {
			return commandResult{}, err
		}
		if err := session.RecordFileOpen(args[2], args[3]); err != nil {
			return commandResult{}, err
		}
		return commandResult{}, nil
	case "record-lsp-diagnostic":
		if len(args) < 7 {
			return commandResult{}, fmt.Errorf("record-lsp-diagnostic requires 7 args")
		}
		session, err := recorder.LoadSession(args[0], args[1], filter.DefaultFilterConfig())
		if err != nil {
			return commandResult{}, err
		}
		if err := session.RecordLSPDiagnostic(args[2], args[3], args[4], args[5], args[6]); err != nil {
			return commandResult{}, err
		}
		return commandResult{}, nil
	case "resume":
		if len(args) < 2 {
			return commandResult{}, fmt.Errorf("resume requires 2 args")
		}
		session, err := recorder.ResumeSession(args[0], args[1], filter.DefaultFilterConfig())
		if err != nil {
			return commandResult{}, err
		}
		return commandResult{Message: "Session resumed: " + session.ID}, nil
	case "list":
		if len(args) < 1 {
			return commandResult{}, fmt.Errorf("list requires 1 arg")
		}
		sessions, err := recorder.ListSessions(args[0])
		if err != nil {
			return commandResult{}, err
		}
		return commandResult{Message: fmt.Sprintf("%v", sessions)}, nil
	default:
		return commandResult{}, fmt.Errorf("unknown command: %s", command)
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
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create data directory: %v\n", err)
			os.Exit(1)
		}
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
	if len(os.Args) < 10 {
		fmt.Fprintf(os.Stderr, "Usage: record-edit <session_id> <save_path> <filename> <line> <col> <line_count> <changedtick> <line_text>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	filename := os.Args[4]
	line := os.Args[5]
	col := os.Args[6]
	lineCount := os.Args[7]
	changedTick := os.Args[8]
	lineText := os.Args[9]

	session, err := recorder.LoadSession(sessionID, savePath, filter.DefaultFilterConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.RecordEdit(filename, line, col, lineCount, changedTick, lineText); err != nil {
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
