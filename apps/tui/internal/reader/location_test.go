package reader

import "testing"

func TestEncodeDecodeTXTLocation(t *testing.T) {
	doc := &Document{Format: "txt", Lines: []string{"a", "b", "c"}, LineIndex: []LineAnchor{{Line: 0, Page: -1, Spine: -1, Offset: 0}, {Line: 1, Page: -1, Spine: -1, Offset: 1}, {Line: 2, Page: -1, Spine: -1, Offset: 2}}}
	token := EncodeLocation(doc, 2)
	if token != "fmt=txt;line=2" {
		t.Fatalf("unexpected token: %q", token)
	}
	line, ok := DecodeLocation(doc, token)
	if !ok || line != 2 {
		t.Fatalf("expected decode to return line 2, got ok=%v line=%d", ok, line)
	}
}

func TestEncodeDecodePDFLocation(t *testing.T) {
	doc := &Document{
		Format: "pdf",
		Lines:  []string{"--- Page 1 ---", "hello", "--- Page 2 ---", "world"},
		LineIndex: []LineAnchor{
			{Line: 0, Page: 1, Spine: -1, Offset: 0},
			{Line: 1, Page: 1, Spine: -1, Offset: 1},
			{Line: 2, Page: 2, Spine: -1, Offset: 0},
			{Line: 3, Page: 2, Spine: -1, Offset: 1},
		},
	}

	token := EncodeLocation(doc, 3)
	if token != "fmt=pdf;page=2;line=3" {
		t.Fatalf("unexpected token: %q", token)
	}
	line, ok := DecodeLocation(doc, token)
	if !ok || line != 3 {
		t.Fatalf("expected decode to return line 3, got ok=%v line=%d", ok, line)
	}
}

func TestEncodeDecodeEPUBLocation(t *testing.T) {
	doc := &Document{
		Format: "epub",
		Lines:  []string{"=== Chapter: One ===", "alpha", "=== Chapter: Two ===", "beta"},
		LineIndex: []LineAnchor{
			{Line: 0, Page: -1, Spine: 0, Offset: 0},
			{Line: 1, Page: -1, Spine: 0, Offset: 1},
			{Line: 2, Page: -1, Spine: 1, Offset: 0},
			{Line: 3, Page: -1, Spine: 1, Offset: 1},
		},
	}

	token := EncodeLocation(doc, 3)
	if token != "fmt=epub;spine=1;offset=1;line=3" {
		t.Fatalf("unexpected token: %q", token)
	}
	line, ok := DecodeLocation(doc, token)
	if !ok || line != 3 {
		t.Fatalf("expected decode to return line 3, got ok=%v line=%d", ok, line)
	}
}

func TestDecodeInvalidTokenFails(t *testing.T) {
	doc := &Document{Format: "txt", Lines: []string{"a", "b"}, LineIndex: []LineAnchor{{Line: 0, Page: -1, Spine: -1, Offset: 0}, {Line: 1, Page: -1, Spine: -1, Offset: 1}}}
	if _, ok := DecodeLocation(doc, "invalid"); ok {
		t.Fatal("expected decode to fail for invalid token")
	}
}

func TestDecodeLegacyLineToken(t *testing.T) {
	doc := &Document{Format: "txt", Lines: []string{"a", "b", "c"}, LineIndex: []LineAnchor{{Line: 0, Page: -1, Spine: -1, Offset: 0}, {Line: 1, Page: -1, Spine: -1, Offset: 1}, {Line: 2, Page: -1, Spine: -1, Offset: 2}}}
	line, ok := DecodeLocation(doc, "line:1")
	if !ok || line != 1 {
		t.Fatalf("expected legacy line decode to return 1, got ok=%v line=%d", ok, line)
	}
}
