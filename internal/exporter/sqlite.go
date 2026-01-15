package exporter

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/andev0x/capytrace.nvim/internal/models"
	_ "modernc.org/sqlite" // Pure-Go SQLite driver
)

// SQLiteExporter exports sessions to a SQLite database for queryable persistence.
// Uses a pure-Go driver (modernc.org/sqlite) to maintain zero CGO dependencies.
type SQLiteExporter struct {
	dbPath string
}

// NewSQLiteExporter creates a new SQLite exporter that stores data in the user's data directory.
func NewSQLiteExporter(dataDir string) *SQLiteExporter {
	return &SQLiteExporter{
		dbPath: filepath.Join(dataDir, "capytrace.db"),
	}
}

// Export writes a session and its events to the SQLite database.
// Creates tables if they don't exist and handles session updates idempotently.
func (e *SQLiteExporter) Export(session *models.Session, savePath string) error {
	db, err := sql.Open("sqlite", e.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to close database: %v\n", closeErr)
		}
	}()

	// Create tables if they don't exist
	if err := e.createTables(db); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			fmt.Fprintf(os.Stderr, "Failed to rollback transaction: %v\n", rollbackErr)
		}
	}()

	// Insert or update session
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO sessions (id, project_path, start_time, end_time, active, output_format)
		VALUES (?, ?, ?, ?, ?, ?)
	`, session.ID, session.ProjectPath, session.StartTime, session.EndTime, session.Active, session.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	// Delete existing events for this session (for updates)
	_, err = tx.Exec(`DELETE FROM events WHERE session_id = ?`, session.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	// Insert all events
	for _, event := range session.Events {
		err = e.insertEvent(tx, session.ID, &event)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// createTables creates the necessary database schema if it doesn't exist.
func (e *SQLiteExporter) createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		project_path TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		active BOOLEAN NOT NULL,
		output_format TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		type TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		filename TEXT,
		line INTEGER,
		column INTEGER,
		line_count INTEGER,
		changed_tick INTEGER,
		file_type TEXT,
		message TEXT,
		level TEXT,
		command TEXT,
		note TEXT,
		prev_line INTEGER,
		prev_column INTEGER,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_events_session_id ON events(session_id);
	CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	`

	_, err := db.Exec(schema)
	return err
}

// insertEvent inserts a single event into the database.
func (e *SQLiteExporter) insertEvent(tx *sql.Tx, sessionID string, event *models.Event) error {
	_, err := tx.Exec(`
		INSERT INTO events (
			session_id, type, timestamp, filename, line, column, 
			line_count, changed_tick, file_type, message, level, 
			command, note, prev_line, prev_column
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		sessionID,
		event.Type,
		event.Timestamp,
		nullString(event.Data.Filename),
		nullInt(event.Data.Line),
		nullInt(event.Data.Column),
		nullInt(event.Data.LineCount),
		nullInt(event.Data.ChangedTick),
		nullString(event.Data.FileType),
		nullString(event.Data.Message),
		nullString(event.Data.Level),
		nullString(event.Data.Command),
		nullString(event.Data.Note),
		nullInt(event.Data.PrevLine),
		nullInt(event.Data.PrevColumn),
	)
	return err
}

// GetSessionStats retrieves statistics for a session from the database.
func (e *SQLiteExporter) GetSessionStats(sessionID string) (*models.SessionSummary, error) {
	db, err := sql.Open("sqlite", e.dbPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to close database: %v\n", closeErr)
		}
	}()

	var summary models.SessionSummary
	var startTime, endTime time.Time

	// Get session metadata
	err = db.QueryRow(`
		SELECT id, project_path, start_time, end_time
		FROM sessions WHERE id = ?
	`, sessionID).Scan(&summary.ID, &summary.ProjectPath, &startTime, &endTime)
	if err != nil {
		return nil, err
	}

	summary.Duration = endTime.Sub(startTime)

	// Get event counts
	rows, err := db.Query(`
		SELECT type, COUNT(*) as count
		FROM events
		WHERE session_id = ?
		GROUP BY type
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to close rows: %v\n", closeErr)
		}
	}()

	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err != nil {
			return nil, err
		}

		summary.TotalEvents += count
		switch eventType {
		case "file_edit":
			summary.FileEdits = count
		case "terminal_command":
			summary.TerminalCommands = count
		case "annotation":
			summary.Annotations = count
		case "cursor_move":
			summary.CursorMoves = count
		}
	}

	return &summary, nil
}

// Helper functions for NULL handling
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(i int) interface{} {
	if i == 0 {
		return nil
	}
	return i
}
