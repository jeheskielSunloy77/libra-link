package reader

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEPUBExtractsChaptersAndText(t *testing.T) {
	epubPath := filepath.Join(t.TempDir(), "sample.epub")
	files := map[string]string{
		"mimetype":               "application/epub+zip",
		"META-INF/container.xml": `<?xml version="1.0"?><container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container"><rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles></container>`,
		"OEBPS/content.opf":      `<?xml version="1.0" encoding="UTF-8"?><package version="2.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookId"><manifest><item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/><item id="ch1" href="ch1.xhtml" media-type="application/xhtml+xml"/><item id="ch2" href="ch2.xhtml" media-type="application/xhtml+xml"/></manifest><spine toc="ncx"><itemref idref="ch1"/><itemref idref="ch2"/></spine></package>`,
		"OEBPS/toc.ncx":          `<?xml version="1.0" encoding="UTF-8"?><ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1"><navMap><navPoint id="p1" playOrder="1"><navLabel><text>Intro</text></navLabel><content src="ch1.xhtml"/></navPoint><navPoint id="p2" playOrder="2"><navLabel><text>Deep Dive</text></navLabel><content src="ch2.xhtml"/></navPoint></navMap></ncx>`,
		"OEBPS/ch1.xhtml":        `<html><body><h1>Ignored</h1><p>Hello chapter one.</p></body></html>`,
		"OEBPS/ch2.xhtml":        `<html><body><p>Second chapter body.</p></body></html>`,
	}
	writeEPUBFixture(t, epubPath, files)

	doc, err := loadEPUB(epubPath)
	if err != nil {
		t.Fatalf("loadEPUB returned error: %v", err)
	}
	if doc.Format != "epub" {
		t.Fatalf("expected epub format, got %q", doc.Format)
	}
	joined := strings.Join(doc.Lines, "\n")
	if !strings.Contains(joined, "=== Chapter: Intro ===") {
		t.Fatalf("expected intro chapter heading, got: %s", joined)
	}
	if !strings.Contains(joined, "=== Chapter: Deep Dive ===") {
		t.Fatalf("expected deep dive chapter heading, got: %s", joined)
	}
	if !strings.Contains(joined, "Hello chapter one.") {
		t.Fatalf("expected extracted chapter one text, got: %s", joined)
	}
}

func TestLoadEPUBFallsBackWithoutTOC(t *testing.T) {
	epubPath := filepath.Join(t.TempDir(), "fallback.epub")
	files := map[string]string{
		"mimetype":               "application/epub+zip",
		"META-INF/container.xml": `<?xml version="1.0"?><container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container"><rootfiles><rootfile full-path="OPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles></container>`,
		"OPS/content.opf":        `<?xml version="1.0" encoding="UTF-8"?><package version="2.0" xmlns="http://www.idpf.org/2007/opf"><manifest><item id="ch1" href="part_one.xhtml" media-type="application/xhtml+xml"/></manifest><spine><itemref idref="ch1"/></spine></package>`,
		"OPS/part_one.xhtml":     `<html><body><p>Fallback text.</p></body></html>`,
	}
	writeEPUBFixture(t, epubPath, files)

	doc, err := loadEPUB(epubPath)
	if err != nil {
		t.Fatalf("loadEPUB returned error: %v", err)
	}
	if len(doc.Lines) == 0 {
		t.Fatal("expected lines to be extracted")
	}
	if !strings.Contains(doc.Lines[0], "Chapter") {
		t.Fatalf("expected fallback chapter heading, got %q", doc.Lines[0])
	}
}

func TestLoadEPUBMalformedFileFails(t *testing.T) {
	broken := filepath.Join(t.TempDir(), "broken.epub")
	writeEPUBFixture(t, broken, map[string]string{
		"mimetype": "application/epub+zip",
	})

	if _, err := loadEPUB(broken); err == nil {
		t.Fatal("expected malformed EPUB to return an error")
	}
}

func writeEPUBFixture(t *testing.T, outPath string, files map[string]string) {
	t.Helper()
	f, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
}
