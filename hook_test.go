package sloghook

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestHook_Info(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))

	logger.Info("hello world")
	logger.Debug("debug log is not recorded")
	logger.Trace("trace log is not recorded")

	var v struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal(w.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Level != "INFO" {
		t.Errorf("unexpected level: %s", v.Level)
	}
	if v.Msg != "hello world" {
		t.Errorf("unexpected msg: %s", v.Msg)
	}
}

func TestHook_Debug(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))
	logger.SetLevel(logrus.DebugLevel)

	logger.Debug("hello world")
	logger.Trace("trace log is not recorded")

	var v struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal(w.Bytes(), &v); err != nil {
		t.Fatalf("parsing %q: %v", w.String(), err)
	}
	if v.Level != "DEBUG" {
		t.Errorf("unexpected level: %s", v.Level)
	}
	if v.Msg != "hello world" {
		t.Errorf("unexpected msg: %s", v.Msg)
	}
}

func TestHook_Trace(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))
	logger.SetLevel(logrus.TraceLevel)

	logger.Trace("hello world")

	var v struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal(w.Bytes(), &v); err != nil {
		t.Fatalf("parsing %q: %v", w.String(), err)
	}
	if v.Level != "DEBUG-1" {
		t.Errorf("unexpected level: %s", v.Level)
	}
	if v.Msg != "hello world" {
		t.Errorf("unexpected msg: %s", v.Msg)
	}
}

func TestHook_Warn(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))

	logger.Warn("hello world")

	var v struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal(w.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Level != "WARN" {
		t.Errorf("unexpected level: %s", v.Level)
	}
	if v.Msg != "hello world" {
		t.Errorf("unexpected msg: %s", v.Msg)
	}
}

func TestHook_Error(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))

	logger.Error("hello world")

	var v struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal(w.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Level != "ERROR" {
		t.Errorf("unexpected level: %s", v.Level)
	}
	if v.Msg != "hello world" {
		t.Errorf("unexpected msg: %s", v.Msg)
	}
}

func TestHook_Panic(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("panic not recovered")
			}
		}()
		logger.Panic("hello world")
	}()

	var v struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal(w.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Level != "ERROR+1" {
		t.Errorf("unexpected level: %s", v.Level)
	}
	if v.Msg != "hello world" {
		t.Errorf("unexpected msg: %s", v.Msg)
	}
}

func TestHook_StackTrace(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := NewLogger(h)
	logger.Info("hello world")

	var v struct {
		Level  string `json:"level"`
		Msg    string `json:"msg"`
		Source struct {
			File     string `json:"file"`
			Function string `json:"function"`
		}
	}
	if err := json.Unmarshal(w.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Level != "INFO" {
		t.Errorf("unexpected level: %s", v.Level)
	}
	if v.Msg != "hello world" {
		t.Errorf("unexpected msg: %s", v.Msg)
	}
	base := filepath.Base(v.Source.File)
	if base != "hook_test.go" {
		t.Errorf("unexpected file: %s", base)
	}
	if v.Source.Function != "github.com/shogo82148/logrus-slog-hook.TestHook_StackTrace" {
		t.Errorf("unexpected function: %s", v.Source.Function)
	}
}

func TestHook_Fields(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, nil)

	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))

	logger.WithFields(logrus.Fields{
		"string":   "hello",
		"int":      int(1),
		"uint":     uint(2),
		"int64":    int64(3),
		"uint64":   uint64(4),
		"bool":     true,
		"duration": time.Second,
		"my_time":  time.Unix(1234567890, 0),
		"uint8":    uint8(5),
		"uint16":   uint16(6),
		"uint32":   uint32(7),
		"int8":     int8(8),
		"int16":    int16(9),
		"int32":    int32(10),
		"float32":  float32(1.5),
		"float64":  float64(2.5),
	}).Info("hello world")

	var v struct {
		Level    string    `json:"level"`
		Msg      string    `json:"msg"`
		Int      int       `json:"int"`
		Uint     uint      `json:"uint"`
		Int64    int64     `json:"int64"`
		Uint64   uint64    `json:"uint64"`
		Bool     bool      `json:"bool"`
		Duration int64     `json:"duration"`
		MyTime   time.Time `json:"my_time"`
		Uint8    uint8     `json:"uint8"`
		Uint16   uint16    `json:"uint16"`
		Uint32   uint32    `json:"uint32"`
		Int8     int8      `json:"int8"`
		Int16    int16     `json:"int16"`
		Int32    int32     `json:"int32"`
		Float32  float32   `json:"float32"`
		Float64  float64   `json:"float64"`
	}

	if err := json.Unmarshal(w.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Level != "INFO" {
		t.Errorf("unexpected level: %s", v.Level)
	}
	if v.Msg != "hello world" {
		t.Errorf("unexpected msg: %s", v.Msg)
	}
	if v.Int != 1 {
		t.Errorf("unexpected int: %d", v.Int)
	}
	if v.Uint != 2 {
		t.Errorf("unexpected uint: %d", v.Uint)
	}
	if v.Int64 != 3 {
		t.Errorf("unexpected int64: %d", v.Int64)
	}
	if v.Uint64 != 4 {
		t.Errorf("unexpected uint64: %d", v.Uint64)
	}
	if v.Bool != true {
		t.Errorf("unexpected bool: %t", v.Bool)
	}
	if v.Duration != 1000000000 {
		t.Errorf("unexpected duration: %d", v.Duration)
	}
	if v.MyTime.Unix() != 1234567890 {
		t.Errorf("unexpected my_time: %s", v.MyTime)
	}
	if v.Uint8 != 5 {
		t.Errorf("unexpected uint8: %d", v.Uint8)
	}
	if v.Uint16 != 6 {
		t.Errorf("unexpected uint16: %d", v.Uint16)
	}
	if v.Uint32 != 7 {
		t.Errorf("unexpected uint32: %d", v.Uint32)
	}
	if v.Int8 != 8 {
		t.Errorf("unexpected int8: %d", v.Int8)
	}
	if v.Int16 != 9 {
		t.Errorf("unexpected int16: %d", v.Int16)
	}
	if v.Int32 != 10 {
		t.Errorf("unexpected int32: %d", v.Int32)
	}
	if v.Float32 != 1.5 {
		t.Errorf("unexpected float32: %f", v.Float32)
	}
	if v.Float64 != 2.5 {
		t.Errorf("unexpected float64: %f", v.Float64)
	}
}

func TestHook_SortKeys(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewTextHandler(w, nil)

	logger := NewLogger(h)
	logger.WithFields(logrus.Fields{
		"A": "1",
		"B": "2",
		"C": "3",
		"D": "4",
		"E": "5",
		"F": "6",
		"G": "7",
		"H": "8",
	}).Info("hello world")

	if !strings.HasSuffix(w.String(), " A=1 B=2 C=3 D=4 E=5 F=6 G=7 H=8\n") {
		t.Errorf("field values are not sorted: %s", w.String())
	}
}

func TestHook_Group(t *testing.T) {
	w := &bytes.Buffer{}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := logrus.New()
	logger.SetFormatter(NewFormatter())
	logger.SetOutput(Writer)
	logger.AddHook(New(h))

	logger.WithField("group", map[string]any{
		"foo": "bar",
	}).Info("hello world")

	var v struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	t.Log(w.String())
	if err := json.Unmarshal(w.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Level != "INFO" {
		t.Errorf("unexpected level: %s", v.Level)
	}
	if v.Msg != "hello world" {
		t.Errorf("unexpected msg: %s", v.Msg)
	}
}

func BenchmarkHook(b *testing.B) {
	h := slog.NewTextHandler(io.Discard, nil)
	logger := logrus.New()
	entry := logrus.NewEntry(logger).WithFields(logrus.Fields{
		"A": "1",
		"B": "2",
		"C": "3",
		"D": "4",
		"E": "5",
		"F": "6",
		"G": "7",
		"H": "8",
	})
	hook := New(h)

	for i := 0; i < b.N; i++ {
		hook.Fire(entry)
	}
}
