package exporter

import "github.com/andev0x/debugstory.nvim/recorder"

type Exporter interface {
	Export(session *recorder.Session, savePath string) error
}
