package eventhandler

import (
	"context"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// NOTE: this is to come in the future
// PLAN:
// to make a file handler so that on change of a specific filetype it runs a command

func NewFSEventHandler(log *slog.Logger) *FSEventHandler {
	fseh := &FSEventHandler{
		Log: log,

		// NOTE: temp setup this till go in config later
		fileHandler: map[string]string{
			"templ": "templ generate",
		},
	}
	return fseh
}

type FSEventHandler struct {
	Log *slog.Logger

	fileHandler map[string]string
}

func (h *FSEventHandler) HandleEvent(ctx context.Context, event fsnotify.Event) (bool, error) {
	// Handle _templ.go files.
	if !event.Has(fsnotify.Remove) && strings.HasSuffix(event.Name, "_templ.go") {
		return false, nil
	}

	// Handle _templ.txt files.
	if !event.Has(fsnotify.Remove) && strings.HasSuffix(event.Name, "_templ.txt") {
		return false, nil
	}

	splitName := strings.Split(event.Name, ".")
	nameExt := splitName[len(splitName)]

	cmd, ok := h.fileHandler[nameExt]
	if !ok {
		return false, nil
	}

	if cmd == "IGNORE" {
		return false, nil
	}

	err := h.handleFileChange(ctx, event.Name, cmd)
	if err != nil {
		h.Log.Warn("Failed to handle file command", slog.Any("error", err))
		return false, err
	}

	return true, nil
}

func (h *FSEventHandler) handleFileChange(ctx context.Context, fileName string, cmd string) error {
	if strings.Contains(cmd, "$filename") {
		cmd = strings.Replace(cmd, "$filename", fileName, -1)
	}

	command := exec.Command(cmd)
	if err := command.Run(); err != nil {
		slog.Error("error running command: %w", err)
		return err
	}

	return nil
}
