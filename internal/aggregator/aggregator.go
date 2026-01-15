// Package aggregator implements smart event aggregation logic for transforming
// raw event streams into meaningful activity blocks and analytics.
package aggregator

import (
	"math"
	"strings"
	"time"

	"github.com/andev0x/capytrace.nvim/internal/models"
)

// AggregatorConfig holds configuration for aggregation rules.
type AggregatorConfig struct {
	// MergeWindow is the time window for merging file_edit events (default: 2 seconds)
	MergeWindow time.Duration

	// IdleThreshold is the time after which to record an idle gap (default: 5 minutes)
	IdleThreshold time.Duration

	// FlowVelocityThreshold is the minimum velocity to consider a block as "flow state"
	FlowVelocityThreshold float64

	// DistractionFiles are file patterns that count as distractions
	DistractionFiles []string
}

// DefaultConfig returns the default aggregator configuration.
func DefaultConfig() *AggregatorConfig {
	return &AggregatorConfig{
		MergeWindow:           2 * time.Second,
		IdleThreshold:         5 * time.Minute,
		FlowVelocityThreshold: 10.0, // 10 ticks per second
		DistractionFiles: []string{
			"NvimTree",
			"copilot-chat",
			"neo-tree",
			"CHADTree",
			"fern",
			"undotree",
		},
	}
}

// Aggregator processes raw events into activity blocks and analytics.
type Aggregator struct {
	config *AggregatorConfig
}

// New creates a new Aggregator with the given configuration.
func New(config *AggregatorConfig) *Aggregator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Aggregator{config: config}
}

// AggregateSession processes a session's events and returns activity blocks and analytics.
func (a *Aggregator) AggregateSession(session *models.Session) ([]models.ActivityBlock, *models.SessionAnalytics) {
	blocks := a.buildActivityBlocks(session.Events)
	analytics := a.computeAnalytics(session, blocks)
	return blocks, analytics
}

// buildActivityBlocks implements the three golden rules:
// 1. 2-second rule: Merge file_edit events < 2s apart
// 2. Context Switch rule: Close block on file change or terminal command
// 3. Idle rule: Record gaps > 5 minutes
func (a *Aggregator) buildActivityBlocks(events []models.Event) []models.ActivityBlock {
	var blocks []models.ActivityBlock
	var currentBlock *models.ActivityBlock
	var lastEventTime time.Time

	for i := range events {
		event := &events[i]

		// Skip non-file-edit events for block building, but use them as context triggers
		if event.Type != "file_edit" {
			// Check if this is a context switch trigger
			if currentBlock != nil && (event.Type == "terminal_command" || event.Type == "file_open") {
				currentBlock.ClosedBy = "context_switch"
				blocks = append(blocks, *currentBlock)
				currentBlock = nil
			}
			lastEventTime = event.Timestamp
			continue
		}

		// Rule 1: Check if we should start a new block or merge with current
		if currentBlock == nil {
			// Start new block
			currentBlock = a.startNewBlock(event)
		} else {
			timeSinceLastEvent := event.Timestamp.Sub(lastEventTime)

			// Rule 2: Context switch - different file
			if event.Data.Filename != currentBlock.Filename {
				currentBlock.ClosedBy = "context_switch"
				blocks = append(blocks, *currentBlock)
				currentBlock = a.startNewBlock(event)
			} else if timeSinceLastEvent > a.config.MergeWindow {
				// Rule 1: Timeout - more than 2 seconds
				currentBlock.ClosedBy = "timeout"
				blocks = append(blocks, *currentBlock)
				currentBlock = a.startNewBlock(event)
			} else {
				// Merge into current block
				a.mergeIntoBlock(currentBlock, event)
			}
		}

		lastEventTime = event.Timestamp
	}

	// Close final block
	if currentBlock != nil {
		currentBlock.ClosedBy = "session_end"
		blocks = append(blocks, *currentBlock)
	}

	return blocks
}

// startNewBlock creates a new activity block from the first event.
func (a *Aggregator) startNewBlock(event *models.Event) *models.ActivityBlock {
	return &models.ActivityBlock{
		StartTime:  event.Timestamp,
		EndTime:    event.Timestamp,
		Duration:   0,
		Filename:   event.Data.Filename,
		EventCount: 1,
		StartTick:  event.Data.ChangedTick,
		EndTick:    event.Data.ChangedTick,
		DeltaTick:  0,
		Velocity:   0,
		Events:     []models.Event{*event},
	}
}

// mergeIntoBlock adds an event to an existing activity block.
func (a *Aggregator) mergeIntoBlock(block *models.ActivityBlock, event *models.Event) {
	block.EndTime = event.Timestamp
	block.Duration = block.EndTime.Sub(block.StartTime)
	block.EventCount++
	block.EndTick = event.Data.ChangedTick
	block.DeltaTick = block.EndTick - block.StartTick

	// Calculate velocity: ticks per second
	if block.Duration.Seconds() > 0 {
		block.Velocity = float64(block.DeltaTick) / block.Duration.Seconds()
	}

	block.Events = append(block.Events, *event)
}

// computeAnalytics calculates advanced metrics for the session.
func (a *Aggregator) computeAnalytics(session *models.Session, blocks []models.ActivityBlock) *models.SessionAnalytics {
	analytics := &models.SessionAnalytics{
		MainFiles:        make(map[string]int),
		ErrorCorrections: []models.ErrorPattern{},
		IdleGaps:         []models.IdleGap{},
		FlowBlocks:       []models.ActivityBlock{},
	}

	// Calculate velocity metrics
	var totalVelocity float64
	var velocityCount int

	for _, block := range blocks {
		if block.Velocity > 0 {
			totalVelocity += block.Velocity
			velocityCount++

			if block.Velocity > analytics.PeakVelocity {
				analytics.PeakVelocity = block.Velocity
			}

			// Identify flow state blocks
			if block.Velocity >= a.config.FlowVelocityThreshold {
				analytics.FlowBlocks = append(analytics.FlowBlocks, block)
			}
		}
	}

	if velocityCount > 0 {
		analytics.AverageVelocity = totalVelocity / float64(velocityCount)
	}

	// Calculate focus ratio and idle gaps
	analytics.IdleGaps = a.findIdleGaps(session.Events)
	for _, gap := range analytics.IdleGaps {
		analytics.TotalIdleTime += gap.Duration
	}

	// Track file focus time
	a.calculateFocusMetrics(session.Events, analytics)

	// Detect error correction patterns
	analytics.ErrorCorrections = a.detectErrorPatterns(session.Events)

	return analytics
}

// findIdleGaps identifies periods of inactivity > 5 minutes.
func (a *Aggregator) findIdleGaps(events []models.Event) []models.IdleGap {
	var gaps []models.IdleGap

	for i := 1; i < len(events); i++ {
		timeBetween := events[i].Timestamp.Sub(events[i-1].Timestamp)

		if timeBetween > a.config.IdleThreshold {
			gaps = append(gaps, models.IdleGap{
				StartTime: events[i-1].Timestamp,
				EndTime:   events[i].Timestamp,
				Duration:  timeBetween,
			})
		}
	}

	return gaps
}

// calculateFocusMetrics computes focus ratio and distraction time.
func (a *Aggregator) calculateFocusMetrics(events []models.Event, analytics *models.SessionAnalytics) {
	var lastEventTime time.Time
	var currentFile string

	for _, event := range events {
		// Calculate time spent on previous file
		if !lastEventTime.IsZero() && currentFile != "" {
			duration := event.Timestamp.Sub(lastEventTime)

			if a.isDistractionFile(currentFile) {
				analytics.DistractionTime += int(duration.Seconds())
			} else {
				analytics.MainFiles[currentFile] += int(duration.Seconds())
			}
		}

		// Update current file
		if event.Data.Filename != "" {
			currentFile = event.Data.Filename
		}
		lastEventTime = event.Timestamp
	}

	// Calculate focus ratio
	totalMainTime := 0
	for _, timeSpent := range analytics.MainFiles {
		totalMainTime += timeSpent
	}

	totalTime := totalMainTime + analytics.DistractionTime
	if totalTime > 0 {
		analytics.FocusRatio = float64(totalMainTime) / float64(totalTime)
	}
}

// isDistractionFile checks if a filename matches distraction patterns.
func (a *Aggregator) isDistractionFile(filename string) bool {
	for _, pattern := range a.config.DistractionFiles {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	return false
}

// detectErrorPatterns identifies error correction events with preceding deletions.
func (a *Aggregator) detectErrorPatterns(events []models.Event) []models.ErrorPattern {
	var patterns []models.ErrorPattern

	for i := range events {
		event := &events[i]

		// Look for "fix error" annotations
		if event.Type == "annotation" && a.containsErrorKeywords(event.Data.Note) {
			// Look back at recent file_edit events for deletions
			pattern := a.analyzeErrorContext(events, i)
			if pattern != nil {
				patterns = append(patterns, *pattern)
			}
		}
	}

	return patterns
}

// containsErrorKeywords checks if an annotation mentions error fixing.
func (a *Aggregator) containsErrorKeywords(note string) bool {
	keywords := []string{"fix", "error", "bug", "issue", "mistake", "correct", "debug"}
	noteLower := strings.ToLower(note)

	for _, keyword := range keywords {
		if strings.Contains(noteLower, keyword) {
			return true
		}
	}

	return false
}

// analyzeErrorContext examines events before an error annotation.
func (a *Aggregator) analyzeErrorContext(events []models.Event, annotationIndex int) *models.ErrorPattern {
	if annotationIndex == 0 {
		return nil
	}

	annotation := &events[annotationIndex]

	// Look back up to 10 events or 5 minutes
	lookbackWindow := 5 * time.Minute
	var linesDeleted int
	var ticksReversed int
	var blocksAffected int
	var lastTick int

	for i := annotationIndex - 1; i >= 0 && i >= annotationIndex-10; i-- {
		event := &events[i]

		// Stop if outside time window
		if annotation.Timestamp.Sub(event.Timestamp) > lookbackWindow {
			break
		}

		// Count file edits in the same file
		if event.Type == "file_edit" && event.Data.Filename == annotation.Data.Filename {
			blocksAffected++

			// Detect deletions (negative tick delta)
			if lastTick > 0 && event.Data.ChangedTick < lastTick {
				ticksReversed += lastTick - event.Data.ChangedTick
			}

			lastTick = event.Data.ChangedTick

			// Approximate line deletions (rough heuristic)
			if event.Data.ChangedTick < lastTick {
				linesDeleted += int(math.Abs(float64(event.Data.LineCount - lastTick)))
			}
		}
	}

	// Only create pattern if we detected actual corrections
	if blocksAffected > 0 {
		return &models.ErrorPattern{
			Timestamp:      annotation.Timestamp,
			Filename:       annotation.Data.Filename,
			Annotation:     annotation.Data.Note,
			LinesDeleted:   linesDeleted,
			TicksReversed:  ticksReversed,
			BlocksAffected: blocksAffected,
		}
	}

	return nil
}
