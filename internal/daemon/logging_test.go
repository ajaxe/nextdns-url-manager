package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/kardianos/service"
)

// mockServiceLogger implements service.Logger for testing
type mockServiceLogger struct {
	infoMessages    []string
	warningMessages []string
	errorMessages   []string
	called          bool
}

func (m *mockServiceLogger) Info(msg string, keyvalues ...interface{}) {
	m.called = true
	m.infoMessages = append(m.infoMessages, msg)
}

func (m *mockServiceLogger) Warning(msg string, keyvalues ...interface{}) {
	m.called = true
	m.warningMessages = append(m.warningMessages, msg)
}

func (m *mockServiceLogger) Error(keyvalues ...interface{}) error {
	m.called = true
	msg := ""
	for _, kv := range keyvalues {
		if s, ok := kv.(string); ok {
			msg = s
			break
		}
	}
	if msg == "" && len(keyvalues) > 0 {
		msg = fmt.Sprintf("%v", keyvalues[0])
	}
	m.errorMessages = append(m.errorMessages, msg)
	return nil
}

func (m *mockServiceLogger) Detail(keyvalues ...interface{}) {}

func TestNewSlogServiceHandler_NilLogger(t *testing.T) {
	_, err := NewSlogServiceHandler(nil)
	if err == nil {
		t.Error("expected error for nil service.Logger")
	}
}

func TestNewSlogServiceHandler_ValidLogger(t *testing.T) {
	mock := &mockServiceLogger{}
	h, err := NewSlogServiceHandler(mock)
	if err != nil {
		t.Fatalf("NewSlogServiceHandler() error = %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if !mock.called {
		t.Error("expected mockServiceLogger to be called during handler creation")
	}
}

func TestSlogServiceHandler_Handle_Info(t *testing.T) {
	mock := &mockServiceLogger{}
	h, _ := NewSlogServiceHandler(mock)

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test info message", 0)
	err := h.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if len(mock.infoMessages) != 1 || mock.infoMessages[0] != "test info message" {
		t.Errorf("expected info message 'test info message', got %v", mock.infoMessages)
	}
}

func TestSlogServiceHandler_Handle_Debug(t *testing.T) {
	mock := &mockServiceLogger{}
	h, _ := NewSlogServiceHandler(mock)

	record := slog.NewRecord(time.Now(), slog.LevelDebug, "test debug message", 0)
	err := h.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if len(mock.infoMessages) != 1 || mock.infoMessages[0] != "test debug message" {
		t.Errorf("expected debug routed to Info, got %v", mock.infoMessages)
	}
}

func TestSlogServiceHandler_Handle_Warning(t *testing.T) {
	mock := &mockServiceLogger{}
	h, _ := NewSlogServiceHandler(mock)

	record := slog.NewRecord(time.Now(), slog.LevelWarn, "test warning message", 0)
	err := h.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if len(mock.warningMessages) != 1 || mock.warningMessages[0] != "test warning message" {
		t.Errorf("expected warning message, got %v", mock.warningMessages)
	}
}

func TestSlogServiceHandler_Handle_Error(t *testing.T) {
	mock := &mockServiceLogger{}
	h, _ := NewSlogServiceHandler(mock)

	record := slog.NewRecord(slog.Now(), slog.LevelError, "test error message", 0)
	err := h.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if len(mock.errorMessages) != 1 || mock.errorMessages[0] != "test error message" {
		t.Errorf("expected error message, got %v", mock.errorMessages)
	}
}

func TestSlogServiceHandler_Handle_DefaultLevel(t *testing.T) {
	mock := &mockServiceLogger{}
	h, _ := NewSlogServiceHandler(mock)

	record := slog.NewRecord(slog.Now(), slog.Level(999), "test default message", 0)
	err := h.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if len(mock.infoMessages) != 1 || mock.infoMessages[0] != "test default message" {
		t.Errorf("expected default level to route to Info, got %v", mock.infoMessages)
	}
}

func TestSlogServiceHandler_Enabled(t *testing.T) {
	mock := &mockServiceLogger{}
	h, _ := NewSlogServiceHandler(mock)

	if !h.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expected Enabled to return true")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Error("expected Enabled to return true for error level")
	}
}

func TestSlogServiceHandler_WithAttrs(t *testing.T) {
	mock := &mockServiceLogger{}
	h, _ := NewSlogServiceHandler(mock)

	newHandler := h.WithAttrs([]slog.Attr{slog.String("key", "value")})
	if newHandler == nil {
		t.Fatal("expected non-nil handler from WithAttrs")
	}
	// Verify the returned handler is also a SlogServiceHandler
	sh, ok := newHandler.(*SlogServiceHandler)
	if !ok {
		t.Errorf("expected *SlogServiceHandler, got %T", newHandler)
	}
	if sh == nil {
		t.Error("expected non-nil SlogServiceHandler from WithAttrs")
	}
}

func TestSlogServiceHandler_WithGroup(t *testing.T) {
	mock := &mockServiceLogger{}
	h, _ := NewSlogServiceHandler(mock)

	newHandler := h.WithGroup("test-group")
	if newHandler == nil {
		t.Fatal("expected non-nil handler from WithGroup")
	}
	sh, ok := newHandler.(*SlogServiceHandler)
	if !ok {
		t.Errorf("expected *SlogServiceHandler, got %T", newHandler)
	}
	if sh == nil {
		t.Error("expected non-nil SlogServiceHandler from WithGroup")
	}
}

func TestInitServiceLogger(t *testing.T) {
	mock := &mockServiceLogger{}
	err := InitServiceLogger(mock)
	if err != nil {
		t.Fatalf("InitServiceLogger() error = %v", err)
	}

	// Verify the default logger is set up by logging a message
	slog.Info("init test message")
	if len(mock.infoMessages) != 1 {
		t.Errorf("expected 1 info message from InitServiceLogger, got %d", len(mock.infoMessages))
	}
	if mock.infoMessages[0] != "init test message" {
		t.Errorf("expected 'init test message', got '%s'", mock.infoMessages[0])
	}
}

func TestSlogHandler_RoutesMultipleLevels(t *testing.T) {
	mock := &mockServiceLogger{}
	h, _ := NewSlogServiceHandler(mock)

	levels := []slog.Level{
		{slog.LevelDebug, 0},
		{slog.LevelInfo, 0},
		{slog.LevelWarn, 0},
		{slog.LevelError, 0},
	}
	// Create proper slog levels
	debugLevel := slog.LevelDebug
	infoLevel := slog.LevelInfo
	warnLevel := slog.LevelWarn
	errorLevel := slog.LevelError

	tests := []struct {
		level    slog.Level
		msg      string
		target   *[]string
		expected int
	}{
		{debugLevel, "debug msg", &mock.infoMessages, 1},
		{infoLevel, "info msg", &mock.infoMessages, 2},
		{warnLevel, "warn msg", &mock.warningMessages, 1},
		{errorLevel, "error msg", &mock.errorMessages, 1},
	}

	for _, tt := range tests {
		record := slog.NewRecord(slog.Now(), tt.level, tt.msg, 0)
		if err := h.Handle(context.Background(), record); err != nil {
			t.Fatalf("Handle() level %v error = %v", tt.level, err)
		}
		if len(*tt.target) != tt.expected {
			t.Errorf("after level %v: expected %d messages in %s, got %d: %v",
				tt.level, tt.expected, formatType(tt.target), len(*tt.target))
		}
	}
}

func formatType(s *[]string) string {
	switch s {
	case nil:
		return "nil"
	default:
		return "slice"
	}
}

func TestNewSlogServiceHandler_MultipleCalls(t *testing.T) {
	mock := &mockServiceLogger{}

	h1, err := NewSlogServiceHandler(mock)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if h1 == nil {
		t.Fatal("first handler is nil")
	}

	h2, err := NewSlogServiceHandler(mock)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if h2 == nil {
		t.Fatal("second handler is nil")
	}
	if h1 == h2 {
		t.Error("expected different handler instances")
	}
}

// Ensure mockServiceLogger implements service.Logger
var _ service.Logger = (*mockServiceLogger)(nil)

// Verify slog.Handler implementation
var _ slog.Handler = (*SlogServiceHandler)(nil)
