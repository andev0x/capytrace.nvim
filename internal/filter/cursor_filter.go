// Package filter provides intelligent event filtering and denoising mechanisms
// to reduce redundant events and improve signal-to-noise ratio in session recordings.
package filter

import (
	"sync"
	"time"

	"github.com/andev0x/capytrace.nvim/internal/models"
)

// CursorFilter implements intelligent filtering for high-frequency cursor movement events.
// It debounces rapid movements and only commits position changes when the cursor
// remains idle or when followed by significant events like text changes.
type CursorFilter struct {
	mu               sync.Mutex
	lastEventTime    time.Time
	pendingEvent     *models.Event
	debounceTimer    *time.Timer
	idleThreshold    time.Duration
	debounceInterval time.Duration
	eventChan        chan *models.Event
	contextTriggers  map[string]bool
	stopChan         chan struct{}
}

// FilterConfig holds configuration parameters for the cursor filter.
type FilterConfig struct {
	// DebounceInterval is the minimum time between cursor movements to be considered separate (default: 200ms)
	DebounceInterval time.Duration
	// IdleThreshold is the time cursor must remain idle before committing position (default: 500ms)
	IdleThreshold time.Duration
	// ContextTriggers are event types that immediately commit pending cursor movements
	ContextTriggers []string
}

// DefaultFilterConfig returns a configuration with recommended defaults.
func DefaultFilterConfig() *FilterConfig {
	return &FilterConfig{
		DebounceInterval: 200 * time.Millisecond,
		IdleThreshold:    500 * time.Millisecond,
		ContextTriggers:  []string{"file_edit", "terminal_command", "session_end"},
	}
}

// NewCursorFilter creates a new cursor filter with the given configuration.
// It starts a background goroutine to handle event filtering without blocking the editor.
func NewCursorFilter(config *FilterConfig) *CursorFilter {
	if config == nil {
		config = DefaultFilterConfig()
	}

	triggers := make(map[string]bool)
	for _, t := range config.ContextTriggers {
		triggers[t] = true
	}

	filter := &CursorFilter{
		idleThreshold:    config.IdleThreshold,
		debounceInterval: config.DebounceInterval,
		eventChan:        make(chan *models.Event, 100),
		contextTriggers:  triggers,
		stopChan:         make(chan struct{}),
	}

	// Start the filtering goroutine
	go filter.run()

	return filter
}

// ProcessEvent filters an incoming event based on the anti-spam rules.
// For cursor_move events, it applies debouncing and idle detection.
// For context trigger events, it immediately commits any pending cursor movement.
// Returns the event to be recorded, or nil if the event should be filtered out.
func (cf *CursorFilter) ProcessEvent(event *models.Event) *models.Event {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	now := time.Now()

	// Handle context trigger events (e.g., text changes, terminal commands)
	if cf.contextTriggers[event.Type] {
		// Commit any pending cursor movement before this significant event
		var result *models.Event
		if cf.pendingEvent != nil {
			result = cf.pendingEvent
			cf.pendingEvent = nil
		}
		cf.lastEventTime = now

		// Cancel any pending timer
		if cf.debounceTimer != nil {
			cf.debounceTimer.Stop()
			cf.debounceTimer = nil
		}

		return result
	}

	// Handle cursor movement events
	if event.Type == "cursor_move" {
		// Debounce: Ignore if movement is too fast
		if !cf.lastEventTime.IsZero() && now.Sub(cf.lastEventTime) < cf.debounceInterval {
			// Update pending event but don't commit yet
			cf.pendingEvent = event
			return nil
		}

		// Update pending event
		cf.pendingEvent = event
		cf.lastEventTime = now

		// Cancel existing timer if any
		if cf.debounceTimer != nil {
			cf.debounceTimer.Stop()
		}

		// Set up idle detection timer
		cf.debounceTimer = time.AfterFunc(cf.idleThreshold, func() {
			cf.commitPendingEvent()
		})

		return nil
	}

	// Pass through all other event types
	cf.lastEventTime = now
	return event
}

// commitPendingEvent commits a pending cursor movement after the idle threshold has passed.
func (cf *CursorFilter) commitPendingEvent() {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	if cf.pendingEvent != nil {
		// Send to event channel for recording
		select {
		case cf.eventChan <- cf.pendingEvent:
		default:
			// Channel full, skip this event
		}
		cf.pendingEvent = nil
	}
}

// run is the main filtering loop that runs in a separate goroutine.
func (cf *CursorFilter) run() {
	for {
		select {
		case <-cf.stopChan:
			return
		case event := <-cf.eventChan:
			// Events on this channel have already been filtered and approved
			// In a real implementation, this would forward to the recorder
			_ = event
		}
	}
}

// FlushPending immediately commits any pending cursor movement event.
// This is useful when ending a session to ensure no events are lost.
func (cf *CursorFilter) FlushPending() *models.Event {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	if cf.debounceTimer != nil {
		cf.debounceTimer.Stop()
		cf.debounceTimer = nil
	}

	event := cf.pendingEvent
	cf.pendingEvent = nil
	return event
}

// Stop gracefully shuts down the cursor filter and its background goroutine.
func (cf *CursorFilter) Stop() {
	close(cf.stopChan)
	if cf.debounceTimer != nil {
		cf.debounceTimer.Stop()
	}
}
