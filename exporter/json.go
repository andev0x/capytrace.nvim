package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/andev0x/capytrace.nvim/recorder"
)

type JSONExporter struct{}

func (e *JSONExporter) Export(session *recorder.Session, savePath string) error {
	filename := session.ID + "_export.json"
	filepath := filepath.Join(savePath, filename)

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}
