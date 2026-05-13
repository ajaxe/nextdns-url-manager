package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/kardianos/service"
)

// SlogServiceHandler bridges slog to service.Logger.
type SlogServiceHandler struct {
	logger service.Logger
}

// NewSlogServiceHandler creates a new SlogServiceHandler wrapping the given service.Logger.
func NewSlogServiceHandler(logger service.Logger) (*SlogServiceHandler, error) {
	if logger == nil {
		return nil, fmt.Errorf("service.Logger is nil")
	}
	h := &SlogServiceHandler{logger: logger}
	return h, nil
}

// Handle implements slog.Handler.
func (h *SlogServiceHandler) Handle(_ context.Context, r slog.Record) error {
	switch r.Level {
	case slog.LevelDebug, slog.LevelInfo:
		h.logger.Info(r.Message)
	case slog.LevelWarn:
		h.logger.Warning(r.Message)
	case slog.LevelError:
		h.logger.Error(r.Message)
	default:
		h.logger.Info(r.Message)
	}
	return nil
}

// Enabled implements slog.Handler.
func (h *SlogServiceHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

// WithAttrs implements slog.Handler.
func (h *SlogServiceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SlogServiceHandler{logger: h.logger}
}

// WithGroup implements slog.Handler.
func (h *SlogServiceHandler) WithGroup(name string) slog.Handler {
	return &SlogServiceHandler{logger: h.logger}
}

// InitServiceLogger sets up the slog default logger to use the service logger.
// Returns a cleanup function to restore the original logger.
func InitServiceLogger(logger service.Logger) error {
	handler, err := NewSlogServiceHandler(logger)
	if err != nil {
		return err
	}
	slog.SetDefault(slog.New(handler))
	return nil
}

// InitCLILogger sets up the slog default logger to use a file handler (default for CLI mode).
func InitCLILogger(path string) error {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	handler := slog.NewJSONHandler(f, opts)
	slog.SetDefault(slog.New(handler))
	return nil
}
