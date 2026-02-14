package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPDFExtractsPages(t *testing.T) {
	pdfPath := filepath.Join(t.TempDir(), "sample.pdf")
	if err := os.WriteFile(pdfPath, buildTestPDF([]string{"Page one text", "Page two text"}), 0o644); err != nil {
		t.Fatalf("write pdf fixture: %v", err)
	}

	doc, err := loadPDF(pdfPath)
	if err != nil {
		t.Fatalf("loadPDF returned error: %v", err)
	}
	if doc.Format != "pdf" {
		t.Fatalf("expected pdf format, got %q", doc.Format)
	}
	joined := strings.Join(doc.Lines, "\n")
	if !strings.Contains(joined, "--- Page 1 ---") || !strings.Contains(joined, "--- Page 2 ---") {
		t.Fatalf("expected page separators in output, got: %s", joined)
	}
	if !strings.Contains(joined, "Page one text") || !strings.Contains(joined, "Page two text") {
		t.Fatalf("expected extracted page text, got: %s", joined)
	}
}

func TestLoadPDFNoExtractableTextFails(t *testing.T) {
	pdfPath := filepath.Join(t.TempDir(), "scanned-like.pdf")
	if err := os.WriteFile(pdfPath, buildTestPDF([]string{"", ""}), 0o644); err != nil {
		t.Fatalf("write pdf fixture: %v", err)
	}

	_, err := loadPDF(pdfPath)
	if err == nil {
		t.Fatal("expected error for no-text pdf")
	}
	if !strings.Contains(err.Error(), "no extractable text") {
		t.Fatalf("expected no extractable text error, got: %v", err)
	}
}

func buildTestPDF(pageTexts []string) []byte {
	type pdfObject struct {
		id   int
		body string
	}

	pageCount := len(pageTexts)
	if pageCount == 0 {
		pageTexts = []string{""}
		pageCount = 1
	}

	objects := make([]pdfObject, 0, 2+2*pageCount+1)
	objects = append(objects, pdfObject{id: 1, body: "<< /Type /Catalog /Pages 2 0 R >>"})

	kids := make([]string, 0, pageCount)
	for i := 0; i < pageCount; i++ {
		pageID := 3 + (i * 2)
		kids = append(kids, fmt.Sprintf("%d 0 R", pageID))
	}
	objects = append(objects, pdfObject{id: 2, body: fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), pageCount)})

	for i, text := range pageTexts {
		pageID := 3 + (i * 2)
		contentID := pageID + 1
		escaped := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)").Replace(text)
		stream := fmt.Sprintf("BT /F1 12 Tf 72 720 Td (%s) Tj ET", escaped)
		pageBody := fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R >>", 3+(2*pageCount), contentID)
		contentBody := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream)
		objects = append(objects, pdfObject{id: pageID, body: pageBody})
		objects = append(objects, pdfObject{id: contentID, body: contentBody})
	}

	fontID := 3 + (2 * pageCount)
	objects = append(objects, pdfObject{id: fontID, body: "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"})

	maxID := fontID
	buf := &strings.Builder{}
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, maxID+1)

	for _, obj := range objects {
		offsets[obj.id] = buf.Len()
		buf.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", obj.id, obj.body))
	}

	xrefPos := buf.Len()
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", maxID+1))
	buf.WriteString("0000000000 65535 f \n")
	for id := 1; id <= maxID; id++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[id]))
	}
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\n", maxID+1))
	buf.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrefPos))

	return []byte(buf.String())
}
