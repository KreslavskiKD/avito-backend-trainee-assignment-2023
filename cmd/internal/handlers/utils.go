package handlers

import (
	"log/slog"
	"time"
)

type Request struct {
	Name string `json:"name"`
}

type SegmentCreator interface {
	CreateSegment(name string) error
}

type SegmentDeleter interface {
	DeleteSegment(name string) error
}

type SegmentGetter interface {
	GetSegments(userId int64) ([]string, error)
}

type HistoryPreparator interface {
	PrepareUserHistoryFile(userId int64, year int, month time.Month) error
}

func MakeLoggerError(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
