package logtest

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"

	"bwawan.com/openuss/internal/logging"
)

// Recorder is an io.Writer collecting log records.
type Recorder struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (r *Recorder) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.Write(p)
}

// String returns everything written so far.
func (r *Recorder) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.String()
}

// Entries decodes every record written so far.
func (r *Recorder) Entries() []map[string]any {
	var entries []map[string]any

	dec := json.NewDecoder(strings.NewReader(r.String()))
	for {
		var entry map[string]any
		if err := dec.Decode(&entry); err != nil {
			return entries
		}
		entries = append(entries, entry)
	}
}

// Find returns the first record with the given msg, or nil if there is none.
func (r *Recorder) Find(msg string) map[string]any {
	for _, entry := range r.Entries() {
		if entry["msg"] == msg {
			return entry
		}
	}
	return nil
}

// New returns a debug-level logger writing into the returned Recorder.
func New() (*slog.Logger, *Recorder) {
	rec := &Recorder{}
	return logging.New(rec, slog.LevelDebug), rec
}
