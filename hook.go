package sloghook

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"sync"

	"github.com/sirupsen/logrus"
)

// we can't add []string to sync.Pool directly, because converting []string to any
// will allocate additional memory.
// so we use a wrapper to hold []string.
type keySorter struct {
	buf []string
}

func (s *keySorter) keys(f logrus.Fields) []string {
	var keys []string
	if s.buf == nil || cap(s.buf) < len(f) {
		keys = make([]string, 0, len(f))
	} else {
		keys = s.buf[:0]
	}
	for k := range f {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	s.buf = keys
	return keys
}

// NewLogger returns a new logrus.Logger with slog.Handler.
func NewLogger(h slog.Handler) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))
	logger.ReportCaller = true
	logger.SetLevel(logrus.TraceLevel)
	return logger
}

var _ logrus.Hook = (*Hook)(nil)

type Hook struct {
	h      slog.Handler
	sorter sync.Pool
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
	attrs := h.fieldToAttrs(entry.Data)
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
	case logrus.PanicLevel:
		return slog.LevelError + 1
	case logrus.FatalLevel:
		return slog.LevelError + 2
	default:
		return slog.LevelInfo
	}
}

func (h *Hook) newSorter() *keySorter {
	if v := h.sorter.Get(); v != nil {
		return v.(*keySorter)
	}
	return &keySorter{}
}

func (h *Hook) fieldToAttrs(f logrus.Fields) []slog.Attr {
	sorter := h.newSorter()
	keys := sorter.keys(f)

	attrs := make([]slog.Attr, 0, len(f))
	for _, k := range keys {
		attrs = append(attrs, slog.Any(k, f[k]))
	}
	h.sorter.Put(sorter)
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
