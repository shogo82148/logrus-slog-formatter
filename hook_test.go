package sloghook

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestHook(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))

	logger.Info("hello world")
	t.Log(w.String())
}
