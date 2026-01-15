package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andev0x/capytrace.nvim/internal/aggregator"
	"github.com/andev0x/capytrace.nvim/internal/models"
)

// SmartMarkdownExporter exports sessions as human-readable Markdown files with
// advanced analytics, activity blocks, and smart aggregation.
type SmartMarkdownExporter struct {
	aggregator *aggregator.Aggregator
}

// NewSmartMarkdownExporter creates a new enhanced Markdown exporter with aggregation support.
func NewSmartMarkdownExporter(config *aggregator.AggregatorConfig) *SmartMarkdownExporter {
	return &SmartMarkdownExporter{
		aggregator: aggregator.New(config),
	}
}

// Export writes both a raw JSON file and an aggregated SESSION_SUMMARY.md file.
func (e *SmartMarkdownExporter) Export(session *models.Session, savePath string) error {
	// 1. Save raw JSON (The Truth)
	if err := e.saveRawJSON(session, savePath); err != nil {
		return fmt.Errorf("failed to save raw JSON: %w", err)
	}

	// 2. Generate and save SESSION_SUMMARY.md (The Story)
	if err := e.saveSessionSummary(session, savePath); err != nil {
		return fmt.Errorf("failed to save session summary: %w", err)
	}

	return nil
}

// saveRawJSON saves the complete event history as JSON.
func (e *SmartMarkdownExporter) saveRawJSON(session *models.Session, savePath string) error {
	filename := fmt.Sprintf("%s_raw.json", session.ID)
	fullPath := filepath.Join(savePath, filename)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fullPath, data, 0644)
}

// saveSessionSummary generates and saves the aggregated Markdown summary.
func (e *SmartMarkdownExporter) saveSessionSummary(session *models.Session, savePath string) error {
	// Run aggregation
	blocks, analytics := e.aggregator.AggregateSession(session)

	// Generate Markdown content
	content := e.generateSmartMarkdown(session, blocks, analytics)

	// Save as SESSION_SUMMARY.md
	filename := "SESSION_SUMMARY.md"
	fullPath := filepath.Join(savePath, filename)

	return os.WriteFile(fullPath, []byte(content), 0644)
}

// generateSmartMarkdown creates an enhanced Markdown document with analytics.
func (e *SmartMarkdownExporter) generateSmartMarkdown(
	session *models.Session,
	blocks []models.ActivityBlock,
	analytics *models.SessionAnalytics,
) string {
	var sb strings.Builder

	// ===== HEADER SECTION =====
	sb.WriteString("# Session Summary Report\n\n")
	sb.WriteString(fmt.Sprintf("**Session ID:** `%s`\n", session.ID))
	sb.WriteString(fmt.Sprintf("**Project:** `%s`\n", session.ProjectPath))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n", session.StartTime.Format("2006-01-02 15:04:05")))

	if !session.EndTime.IsZero() {
		sb.WriteString(fmt.Sprintf("**Ended:** %s\n", session.EndTime.Format("2006-01-02 15:04:05")))
		duration := session.EndTime.Sub(session.StartTime)
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(duration)))
	} else {
		sb.WriteString("**Status:** Active\n")
	}

	sb.WriteString("\n---\n\n")

	// ===== EXECUTIVE SUMMARY =====
	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Events:** %d\n", len(session.Events)))
	sb.WriteString(fmt.Sprintf("- **Activity Blocks:** %d\n", len(blocks)))
	sb.WriteString(fmt.Sprintf("- **Average Velocity:** %.2f ticks/sec\n", analytics.AverageVelocity))
	sb.WriteString(fmt.Sprintf("- **Peak Velocity:** %.2f ticks/sec\n", analytics.PeakVelocity))
	sb.WriteString(fmt.Sprintf("- **Focus Ratio:** %.1f%%\n", analytics.FocusRatio*100))
	sb.WriteString(fmt.Sprintf("- **Flow State Blocks:** %d\n", len(analytics.FlowBlocks)))
	sb.WriteString(fmt.Sprintf("- **Idle Gaps:** %d (Total: %s)\n", len(analytics.IdleGaps), formatDuration(analytics.TotalIdleTime)))
	sb.WriteString(fmt.Sprintf("- **Error Corrections:** %d\n", len(analytics.ErrorCorrections)))
	sb.WriteString("\n")

	// ===== VELOCITY ANALYSIS =====
	sb.WriteString("## Velocity Analysis\n\n")
	sb.WriteString("**What is Velocity?** Delta Tick / Duration. High velocity (>10 ticks/sec) indicates \"Flow State\" - you're coding fast and efficiently.\n\n")

	if len(analytics.FlowBlocks) > 0 {
		sb.WriteString(fmt.Sprintf("### Flow State Blocks (%d)\n\n", len(analytics.FlowBlocks)))
		sb.WriteString("These are your most productive moments:\n\n")

		for i, block := range analytics.FlowBlocks {
			sb.WriteString(fmt.Sprintf("%d. **%s** - `%s`\n", i+1, block.StartTime.Format("15:04:05"), filepath.Base(block.Filename)))
			sb.WriteString(fmt.Sprintf("   - Velocity: **%.2f ticks/sec** 🔥\n", block.Velocity))
			sb.WriteString(fmt.Sprintf("   - Duration: %s\n", formatDuration(block.Duration)))
			sb.WriteString(fmt.Sprintf("   - Changes: %d ticks across %d events\n", block.DeltaTick, block.EventCount))
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("*No flow state blocks detected. Try to minimize interruptions for deeper focus.*\n\n")
	}

	// ===== FOCUS RATIO =====
	sb.WriteString("## Focus Ratio\n\n")
	sb.WriteString(fmt.Sprintf("**Overall Focus:** %.1f%% of time spent on main code files\n\n", analytics.FocusRatio*100))

	if len(analytics.MainFiles) > 0 {
		sb.WriteString("### Time Spent by File\n\n")

		// Sort files by time spent (descending)
		type fileTime struct {
			name string
			time int
		}
		var files []fileTime
		for name, timeSpent := range analytics.MainFiles {
			files = append(files, fileTime{name, timeSpent})
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].time > files[j].time
		})

		for i, ft := range files {
			if i >= 10 {
				break // Show top 10
			}
			percentage := float64(ft.time) / float64(analytics.DistractionTime+ft.time) * 100
			sb.WriteString(fmt.Sprintf("- `%s`: %s (%.1f%%)\n",
				filepath.Base(ft.name),
				formatDuration(time.Duration(ft.time)*time.Second),
				percentage))
		}
		sb.WriteString("\n")
	}

	if analytics.DistractionTime > 0 {
		sb.WriteString(fmt.Sprintf("**Distraction Time:** %s in file browsers/tools\n\n",
			formatDuration(time.Duration(analytics.DistractionTime)*time.Second)))
	}

	// ===== ERROR CORRECTION PATTERNS =====
	if len(analytics.ErrorCorrections) > 0 {
		sb.WriteString("## Error Correction Patterns\n\n")
		sb.WriteString("Detected moments where you fixed errors after making mistakes:\n\n")

		for i, pattern := range analytics.ErrorCorrections {
			sb.WriteString(fmt.Sprintf("### %d. %s - `%s`\n\n", i+1,
				pattern.Timestamp.Format("15:04:05"),
				filepath.Base(pattern.Filename)))
			sb.WriteString(fmt.Sprintf("**Annotation:** \"%s\"\n\n", pattern.Annotation))
			sb.WriteString(fmt.Sprintf("- Blocks affected: %d\n", pattern.BlocksAffected))
			if pattern.TicksReversed > 0 {
				sb.WriteString(fmt.Sprintf("- Changes reversed: %d ticks\n", pattern.TicksReversed))
			}
			if pattern.LinesDeleted > 0 {
				sb.WriteString(fmt.Sprintf("- Lines deleted: ~%d\n", pattern.LinesDeleted))
			}
			sb.WriteString("\n")
		}
	}

	// ===== IDLE GAPS =====
	if len(analytics.IdleGaps) > 0 {
		sb.WriteString("## Idle Periods\n\n")
		sb.WriteString("Gaps > 5 minutes where you might have been stuck or took a break:\n\n")

		for i, gap := range analytics.IdleGaps {
			sb.WriteString(fmt.Sprintf("%d. **%s** - %s idle\n", i+1,
				gap.StartTime.Format("15:04:05"),
				formatDuration(gap.Duration)))
		}
		sb.WriteString("\n")
	}

	// ===== ACTIVITY TIMELINE =====
	sb.WriteString("## Activity Timeline\n\n")
	sb.WriteString("Aggregated blocks of continuous work (events < 2 seconds apart):\n\n")

	for i, block := range blocks {
		if i >= 50 {
			sb.WriteString(fmt.Sprintf("\n*...and %d more blocks (see raw JSON for full details)*\n", len(blocks)-i))
			break
		}

		sb.WriteString(fmt.Sprintf("### %s - `%s`\n\n",
			block.StartTime.Format("15:04:05"),
			filepath.Base(block.Filename)))

		sb.WriteString(fmt.Sprintf("- **Duration:** %s\n", formatDuration(block.Duration)))
		sb.WriteString(fmt.Sprintf("- **Events:** %d edits\n", block.EventCount))
		sb.WriteString(fmt.Sprintf("- **Changes:** %d → %d ticks (Δ%d)\n",
			block.StartTick, block.EndTick, block.DeltaTick))

		if block.Velocity > 0 {
			velocityEmoji := ""
			if block.Velocity >= 10 {
				velocityEmoji = " 🔥"
			}
			sb.WriteString(fmt.Sprintf("- **Velocity:** %.2f ticks/sec%s\n", block.Velocity, velocityEmoji))
		}

		sb.WriteString(fmt.Sprintf("- **Closed by:** %s\n", block.ClosedBy))
		sb.WriteString("\n")
	}

	// ===== RAW EVENT SUMMARY =====
	sb.WriteString("---\n\n")
	sb.WriteString("## Raw Event Statistics\n\n")

	eventCounts := make(map[string]int)
	for _, event := range session.Events {
		eventCounts[event.Type]++
	}

	for eventType, count := range eventCounts {
		sb.WriteString(fmt.Sprintf("- **%s:** %d\n", formatEventType(eventType), count))
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString("*Generated by capytrace.nvim with smart aggregation*\n")
	sb.WriteString(fmt.Sprintf("*Raw event data available in `%s_raw.json`*\n", session.ID))

	return sb.String()
}

// formatDuration converts a duration to a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
