package reload

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors profile directories for changes and triggers reloads.
type Watcher struct {
	dirs     []string
	onChange func()
	fsw      *fsnotify.Watcher
	stop     chan struct{}
	once     sync.Once
}

// New creates a file watcher. The onChange callback is called after a debounce
// period when files in any watched directory change.
func New(dirs []string, onChange func()) *Watcher {
	return &Watcher{
		dirs:     dirs,
		onChange: onChange,
		stop:     make(chan struct{}),
	}
}

// Start begins watching directories for file changes.
func (w *Watcher) Start(ctx context.Context) error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.fsw = fsw

	for _, dir := range w.dirs {
		if err := fsw.Add(dir); err != nil {
			slog.Warn("watch dir", "dir", dir, "error", err)
		}
	}

	go w.loop(ctx)
	slog.Debug("file watcher started", "dirs", w.dirs)
	return nil
}

func (w *Watcher) loop(ctx context.Context) {
	var debounce *time.Timer

	for {
		select {
		case <-ctx.Done():
			if debounce != nil {
				debounce.Stop()
			}
			return
		case <-w.stop:
			if debounce != nil {
				debounce.Stop()
			}
			return
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			// Only react to writes and creates of .toml files.
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			// Debounce: reset timer on each event, fire after 100ms of quiet.
			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(100*time.Millisecond, func() {
				slog.Debug("file change detected", "file", event.Name)
				w.onChange()
			})
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			slog.Warn("fsnotify error", "error", err)
		}
	}
}

// Stop halts the watcher.
func (w *Watcher) Stop() {
	w.once.Do(func() {
		close(w.stop)
		if w.fsw != nil {
			w.fsw.Close()
		}
	})
}
