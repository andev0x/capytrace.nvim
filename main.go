package main

import (
	"fmt"
	"os"

	"github.com/andev0x/capytrace.nvim/exporter"
	"github.com/andev0x/capytrace.nvim/recorder"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
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
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func handleStart() {
	if len(os.Args) < 6 {
		fmt.Fprintf(os.Stderr, "Usage: start <session_id> <project_path> <save_path> <output_format>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	projectPath := os.Args[3]
	savePath := os.Args[4]
	outputFormat := os.Args[5]

	session := recorder.NewSession(sessionID, projectPath, savePath, outputFormat)
	if err := session.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Session started: %s\n", sessionID)
}

func handleEnd() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: end <session_id> <save_path>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	session, err := recorder.LoadSession(sessionID, savePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.End(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to end session: %v\n", err)
		os.Exit(1)
	}

	// Export session
	var exp exporter.Exporter
	if session.OutputFormat == "json" {
		exp = &exporter.JSONExporter{}
	} else {
		exp = &exporter.MarkdownExporter{}
	}

	if err := exp.Export(session, session.SavePath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Session ended and exported: %s\n", sessionID)
}

func handleAnnotate() {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Usage: annotate <session_id> <save_path> <note>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	note := os.Args[4]

	session, err := recorder.LoadSession(sessionID, savePath)
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

	session, err := recorder.LoadSession(sessionID, savePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.RecordEdit(filename, line, col, lineCount, changedTick); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to record edit: %v\n", err)
		os.Exit(1)
	}
}

func handleRecordTerminal() {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Usage: record-terminal <session_id> <save_path> <command>\n")
		os.Exit(1)
	}

	sessionID := os.Args[2]
	savePath := os.Args[3]
	command := os.Args[4]

	session, err := recorder.LoadSession(sessionID, savePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.RecordTerminalCommand(command); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to record terminal command: %v\n", err)
		os.Exit(1)
	}
}

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

	session, err := recorder.LoadSession(sessionID, savePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load session: %v\n", err)
		os.Exit(1)
	}

	if err := session.RecordCursorMove(filename, line, col); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to record cursor movement: %v\n", err)
		os.Exit(1)
	}
}

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

func handleResume() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: resume <session_name> <save_path>\n")
		os.Exit(1)
	}

	sessionName := os.Args[2]
	savePath := os.Args[3]

	session, err := recorder.ResumeSession(sessionName, savePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resume session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Session resumed: %s\n", session.ID)
}
