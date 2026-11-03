// Package platform contains cross-cutting infrastructure helpers.
package platform

import (
	"io"
	"log/slog"
)

// NewLogger returns a JSON structured logger.
func NewLogger(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}
