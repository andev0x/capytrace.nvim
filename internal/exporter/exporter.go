// Package exporter provides interfaces and implementations for exporting
// debugging sessions to various formats (Markdown, JSON, SQLite).
package exporter

import "github.com/andev0x/capytrace.nvim/internal/models"

// Exporter defines the interface for exporting session data to different formats.
type Exporter interface {
	// Export writes a session to the specified save path in the exporter's format.
	Export(session *models.Session, savePath string) error
}
