package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andev0x/capytrace.nvim/recorder"
)

type MarkdownExporter struct{}

func (e *MarkdownExporter) Export(session *recorder.Session, savePath string) error {
	filename := fmt.Sprintf("%s.md", session.ID)
	filepath := filepath.Join(savePath, filename)

	content := e.generateMarkdown(session)
	return os.WriteFile(filepath, []byte(content), 0644)
}

func (e *MarkdownExporter) generateMarkdown(session *recorder.Session) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# Debug Session: %s\n\n", session.ID))
	sb.WriteString(fmt.Sprintf("**Project:** %s\n", session.ProjectPath))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n", session.StartTime.Format(time.RFC3339)))
	if !session.EndTime.IsZero() {
		sb.WriteString(fmt.Sprintf("**Ended:** %s\n", session.EndTime.Format(time.RFC3339)))
		duration := session.EndTime.Sub(session.StartTime)
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", duration.String()))
	}
	sb.WriteString("\n---\n\n")

	// Timeline
	sb.WriteString("## Timeline\n\n")

	for _, event := range session.Events {
		timestamp := event.Timestamp.Format("15:04:05")

		switch event.Type {
		case "session_start":
			sb.WriteString(fmt.Sprintf("### %s - Session Started\n", timestamp))
			sb.WriteString(fmt.Sprintf("🚀 %s\n\n", event.Data.Note))

		case "session_end":
			sb.WriteString(fmt.Sprintf("### %s - Session Ended\n", timestamp))
			sb.WriteString(fmt.Sprintf("🏁 %s\n\n", event.Data.Note))

		case "annotation":
			sb.WriteString(fmt.Sprintf("### %s - Note\n", timestamp))
			sb.WriteString(fmt.Sprintf("📝 %s\n\n", event.Data.Note))

		case "file_edit":
			sb.WriteString(fmt.Sprintf("### %s - File Edit\n", timestamp))
			sb.WriteString(fmt.Sprintf("📄 **File:** `%s`\n", event.Data.Filename))
			sb.WriteString(fmt.Sprintf("📍 **Position:** Line %d, Column %d\n", event.Data.Line, event.Data.Column))
			sb.WriteString(fmt.Sprintf("📊 **Total Lines:** %d\n\n", event.Data.LineCount))

		case "terminal_command":
			sb.WriteString(fmt.Sprintf("### %s - Terminal Command\n", timestamp))
			sb.WriteString(fmt.Sprintf("💻 ```bash\n%s\n```\n\n", event.Data.Command))

		case "cursor_move":
			sb.WriteString(fmt.Sprintf("### %s - Cursor Movement\n", timestamp))
			sb.WriteString(fmt.Sprintf("👆 **File:** `%s`\n", event.Data.Filename))
			sb.WriteString(fmt.Sprintf("📍 **Position:** Line %d, Column %d\n\n", event.Data.Line, event.Data.Column))

		case "session_resume":
			sb.WriteString(fmt.Sprintf("### %s - Session Resumed\n", timestamp))
			sb.WriteString(fmt.Sprintf("🔄 %s\n\n", event.Data.Note))
		}
	}

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Events:** %d\n", len(session.Events)))

	eventCounts := make(map[string]int)
	for _, event := range session.Events {
		eventCounts[event.Type]++
	}

	for eventType, count := range eventCounts {
		sb.WriteString(fmt.Sprintf("- **%s:** %d\n", strings.Title(strings.ReplaceAll(eventType, "_", " ")), count))
	}

	return sb.String()
}
