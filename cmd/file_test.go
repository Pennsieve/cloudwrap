package cmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pennsieve/cloudwrap/internal/params"
)

func TestWriteExportsSuccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.env")
	pairs := []params.Pair{
		{Key: "ONE_KEY", Value: "valueone"},
		{Key: "TWO", Value: "valuetwo"},
	}

	if err := writeExports(path, pairs); err != nil {
		t.Fatalf("writeExports: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "export ONE_KEY=valueone\nexport TWO=valuetwo\n"
	if string(got) != want {
		t.Errorf("file contents = %q, want %q", got, want)
	}
}

func TestWriteExportsRemovesFileOnWriteFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.env")

	// Seed a stale file at the path: os.Create truncates it, so a mid-write
	// failure must still leave nothing behind.
	if err := os.WriteFile(path, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	writeErr := errors.New("boom")
	err := writeExportsWith(path, []params.Pair{{Key: "K", Value: "v"}},
		func(io.Writer, []params.Pair) error { return writeErr })

	if !errors.Is(err, writeErr) {
		t.Fatalf("expected write error, got %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("expected incomplete file to be removed, stat err = %v", statErr)
	}
}

func TestWriteExportLinesSurfacesWriterError(t *testing.T) {
	err := writeExportLines(failingWriter{}, []params.Pair{{Key: "K", Value: "v"}})
	if err == nil {
		t.Fatal("expected error from failing writer, got nil")
	}
}

// failingWriter always fails, to exercise writeExportLines' error path.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
