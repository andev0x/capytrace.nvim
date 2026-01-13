package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/andev0x/capytrace.nvim/internal/models"
)

// JSONExporter exports sessions as formatted JSON files.
type JSONExporter struct{}

// Export writes a session to disk as a JSON file with the naming pattern: {session_id}_export.json
func (e *JSONExporter) Export(session *models.Session, savePath string) error {
	filename := session.ID + "_export.json"
	fullPath := filepath.Join(savePath, filename)

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fullPath, data, 0644)
}
