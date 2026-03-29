package exporter

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/andev0x/capytrace.nvim/internal/models"
)

//go:embed templates/session.md
var sessionTemplate []byte

// MarkdownExporter exports sessions as human-readable Markdown files.
type MarkdownExporter struct{}

// Export writes a session to disk as a Markdown file with a detailed timeline and summary.
func (e *MarkdownExporter) Export(session *models.Session, savePath string) error {
	filename := fmt.Sprintf("%s.md", session.ID)
	fullPath := filepath.Join(savePath, filename)

	content, err := e.generateMarkdown(session)
	if err != nil {
		return err
	}

	return os.WriteFile(fullPath, []byte(content), 0644)
}

type viewData struct {
	ID               string
	ProjectPath      string
	StartDate        string
	StartTime        string
	Duration         string
	FileEdits        int
	CursorMoves      int
	TerminalCommands int
	Annotations      int
	LSPDiagnostics   int
	EditPercent      int
	NavPercent       int
	TotalEvents      int
	Blocks           int
	GroupedEvents    []timelineEvent
}

type timelineEvent struct {
	Time     string
	Emoji    string
	Title    string
	File     string
	Location string
	Snippet  string
	Details  string
}

// generateMarkdown builds the template data and renders the embedded template.
func (e *MarkdownExporter) generateMarkdown(session *models.Session) (string, error) {
	tmpl, err := template.New("session").Parse(string(sessionTemplate))
	if err != nil {
		return "", err
	}

	counts := countEvents(session.Events)
	editPercent, navPercent := focusRatios(counts["file_edit"], counts["cursor_move"])
	grouped := groupTimeline(session.Events)

	data := viewData{
		ID:               session.ID,
		ProjectPath:      session.ProjectPath,
		StartDate:        session.StartTime.Format("2006-01-02"),
		StartTime:        session.StartTime.Format("15:04:05"),
		Duration:         formatDurationHuman(session.StartTime, session.EndTime),
		FileEdits:        counts["file_edit"],
		CursorMoves:      counts["cursor_move"],
		TerminalCommands: counts["terminal_command"],
		Annotations:      counts["annotation"],
		LSPDiagnostics:   counts["lsp_diagnostic"],
		EditPercent:      editPercent,
		NavPercent:       navPercent,
		TotalEvents:      len(session.Events),
		Blocks:           len(grouped),
		GroupedEvents:    grouped,
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func countEvents(events []models.Event) map[string]int {
	counts := make(map[string]int)
	for _, ev := range events {
		counts[ev.Type]++
	}
	return counts
}

func focusRatios(edits, navs int) (int, int) {
	total := edits + navs
	if total == 0 {
		return 0, 0
	}
	editPct := int((float64(edits) / float64(total)) * 100)
	navPct := 100 - editPct
	return editPct, navPct
}

func formatDurationHuman(start, end time.Time) string {
	if end.IsZero() {
		return "in progress"
	}
	d := end.Sub(start)
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func groupTimeline(events []models.Event) []timelineEvent {
	var grouped []timelineEvent
	var pending *timelineEvent
	var editCount int

	for _, ev := range events {
		if ev.Type == "cursor_move" {
			continue // omit raw cursor spam from timeline
		}

		if ev.Type == "file_edit" {
			if pending != nil && pending.Title == "Edits" && pending.File == ev.Data.Filename {
				editCount++
				pending.Time = ev.Timestamp.Format("15:04:05")
				pending.Location = fmt.Sprintf("L%d:C%d", ev.Data.Line, ev.Data.Column)
				if ev.Data.LineText != "" {
					pending.Snippet = trimSnippet(ev.Data.LineText)
				}
				pending.Details = fmt.Sprintf("%d edits", editCount)
				continue
			}
			editCount = 1
			pending = &timelineEvent{
				Time:     ev.Timestamp.Format("15:04:05"),
				Emoji:    "🛠",
				Title:    "Edits",
				File:     ev.Data.Filename,
				Location: fmt.Sprintf("L%d:C%d", ev.Data.Line, ev.Data.Column),
				Snippet:  trimSnippet(ev.Data.LineText),
				Details:  "1 edit",
			}
			grouped = append(grouped, *pending)
			continue
		}

		pending = nil
		grouped = append(grouped, timelineEvent{
			Time:     ev.Timestamp.Format("15:04:05"),
			Emoji:    emojiFor(ev.Type),
			Title:    titleFor(ev),
			File:     ev.Data.Filename,
			Location: locationFor(ev),
			Snippet:  trimSnippet(ev.Data.LineText),
			Details:  detailFor(ev),
		})
	}

	return grouped
}

func emojiFor(eventType string) string {
	switch eventType {
	case "annotation":
		return "📝"
	case "file_edit":
		return "🛠"
	case "lsp_diagnostic":
		return "⚠️"
	case "terminal_command":
		return "💻"
	case "file_open":
		return "📂"
	case "session_start":
		return "🚀"
	case "session_end":
		return "🏁"
	case "session_resume":
		return "🔄"
	default:
		return "•"
	}
}

func titleFor(ev models.Event) string {
	switch ev.Type {
	case "annotation":
		return "Note"
	case "file_edit":
		return "Edits"
	case "lsp_diagnostic":
		return "LSP"
	case "terminal_command":
		return "Terminal"
	case "file_open":
		return "File Open"
	case "session_start":
		return "Session Started"
	case "session_end":
		return "Session Ended"
	case "session_resume":
		return "Session Resumed"
	default:
		return strings.Title(strings.ReplaceAll(ev.Type, "_", " "))
	}
}

func locationFor(ev models.Event) string {
	if ev.Data.Filename == "" {
		return ""
	}
	if ev.Data.Line > 0 {
		return fmt.Sprintf("%s:%d", ev.Data.Filename, ev.Data.Line)
	}
	return ev.Data.Filename
}

func trimSnippet(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 120 {
		return s[:117] + "..."
	}
	return s
}

func detailFor(ev models.Event) string {
	switch ev.Type {
	case "annotation":
		return ev.Data.Note
	case "lsp_diagnostic":
		return ev.Data.Message
	case "terminal_command":
		return ev.Data.Command
	case "file_open":
		if ev.Data.FileType != "" {
			return ev.Data.FileType
		}
		return "opened"
	default:
		return ""
	}
}
