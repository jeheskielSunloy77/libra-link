package reader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTXTFromFixture(t *testing.T) {
	doc, err := loadTXT(filepath.Join("testdata", "sample.txt"))
	if err != nil {
		t.Fatalf("loadTXT returned error: %v", err)
	}
	if doc.Format != "txt" {
		t.Fatalf("expected format txt, got %q", doc.Format)
	}
	if len(doc.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(doc.Lines))
	}
	if doc.Lines[1] != "Beta" {
		t.Fatalf("expected second line to be Beta, got %q", doc.Lines[1])
	}
}

func TestLoadTXTEmptyFileGetsSingleLine(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "empty.txt")
	if err := os.WriteFile(tmp, nil, 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	doc, err := loadTXT(tmp)
	if err != nil {
		t.Fatalf("loadTXT returned error: %v", err)
	}
	if len(doc.Lines) != 1 {
		t.Fatalf("expected 1 line for empty file, got %d", len(doc.Lines))
	}
	if doc.Lines[0] != "" {
		t.Fatalf("expected empty line, got %q", doc.Lines[0])
	}
}
