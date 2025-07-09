package exporter

import "github.com/andev0x/capytrace.nvim/recorder"

type Exporter interface {
	Export(session *recorder.Session, savePath string) error
}
