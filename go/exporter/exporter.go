package exporter

import "github.com/debugstory/recorder"

type Exporter interface {
	Export(session *recorder.Session, savePath string) error
}