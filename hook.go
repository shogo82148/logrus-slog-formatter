package sloghook

import (
	"context"
	"io"
	"log/slog"

	"github.com/sirupsen/logrus"
)

var _ logrus.Hook = (*Hook)(nil)

type Hook struct {
	h slog.Handler
}

func New(h slog.Handler) *Hook {
	return &Hook{h: h}
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	ctx := entry.Context
	if ctx == nil {
		ctx = context.Background()
	}

	t := entry.Time
	lv := slogLevel(entry.Level)
	var pc uintptr
	if entry.Caller != nil {
		pc = entry.Caller.PC
	}
	record := slog.NewRecord(t, lv, entry.Message, pc)
	attrs := fieldToAttrs(entry.Data)
	record.AddAttrs(attrs...)
	h.h.Handle(ctx, record)
	return nil
}

// Levels implements logrus.Hook interface.
func (h *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func slogLevel(l logrus.Level) slog.Level {
	switch l {
	case logrus.TraceLevel:
		return slog.LevelDebug - 1
	case logrus.DebugLevel:
		return slog.LevelDebug
	case logrus.InfoLevel:
		return slog.LevelInfo
	case logrus.WarnLevel:
		return slog.LevelWarn
	case logrus.ErrorLevel:
		return slog.LevelError
	case logrus.FatalLevel:
		return slog.LevelError + 1
	case logrus.PanicLevel:
		return slog.LevelError + 2
	default:
		return slog.LevelInfo
	}
}

func fieldToAttrs(f logrus.Fields) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(f))
	for k, v := range f {
		attrs = append(attrs, slog.Any(k, v))
	}
	return attrs
}

type Formatter struct{}

var _ logrus.Formatter = (*Formatter)(nil)

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte{}, nil
}

var Writer = io.Discard
