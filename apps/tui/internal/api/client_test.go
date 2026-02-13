package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateEbookMapsRequestAndResponse(t *testing.T) {
	t.Parallel()

	checksum := strings.Repeat("a", 64)
	importedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	var gotAuth string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ebooks" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"status": 201,
			"success": true,
			"message": "ok",
			"data": {
				"id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"ownerUserId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"title": "My Book",
				"description": "A description",
				"format": "txt",
				"languageCode": "en",
				"storageKey": "/tmp/books/book.txt",
				"fileSizeBytes": 123,
				"checksumSha256": "` + checksum + `",
				"importedAt": "2026-01-02T03:04:05Z",
				"createdAt": "2026-01-02T03:04:05Z",
				"updatedAt": "2026-01-02T03:04:05Z"
			}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, time.Second)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.SetSession("access-token", "refresh-token", "user-1")

	item, err := client.CreateEbook(context.Background(), CreateEbookInput{
		Title:          "My Book",
		Description:    "A description",
		Format:         "txt",
		LanguageCode:   "en",
		StorageKey:     "/tmp/books/book.txt",
		FileSizeBytes:  123,
		ChecksumSHA256: checksum,
		ImportedAt:     &importedAt,
	})
	if err != nil {
		t.Fatalf("create ebook: %v", err)
	}

	if gotAuth != "Bearer access-token" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if gotBody["title"] != "My Book" {
		t.Fatalf("unexpected title in request body: %#v", gotBody["title"])
	}
	if gotBody["format"] != "txt" {
		t.Fatalf("unexpected format in request body: %#v", gotBody["format"])
	}
	if gotBody["storageKey"] != "/tmp/books/book.txt" {
		t.Fatalf("unexpected storageKey in request body: %#v", gotBody["storageKey"])
	}
	if gotBody["checksumSha256"] != checksum {
		t.Fatalf("unexpected checksum in request body: %#v", gotBody["checksumSha256"])
	}

	if item == nil {
		t.Fatal("expected ebook response")
	}
	if item.Title != "My Book" || item.LanguageCode != "en" || item.FileSizeBytes != 123 || item.ChecksumSHA256 != checksum {
		t.Fatalf("unexpected mapped ebook response: %#v", item)
	}
}

func TestCreateEbookReturnsApiErrorOnFailureStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ebooks" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, time.Second)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.SetSession("access-token", "refresh-token", "user-1")

	_, err = client.CreateEbook(context.Background(), CreateEbookInput{
		Title:          "My Book",
		Format:         "txt",
		StorageKey:     "/tmp/books/book.txt",
		FileSizeBytes:  10,
		ChecksumSHA256: strings.Repeat("b", 64),
	})
	if err == nil {
		t.Fatal("expected create ebook to fail")
	}
	if !strings.Contains(err.Error(), "create ebook failed (500)") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListEbooksMapsChecksumSizeAndLanguage(t *testing.T) {
	t.Parallel()

	checksum := strings.Repeat("c", 64)
	var gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ebooks" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": 200,
			"success": true,
			"page": 1,
			"data": [{
				"id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"ownerUserId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"title": "Mapped Book",
				"description": "Book Description",
				"format": "pdf",
				"languageCode": "fr",
				"storageKey": "/tmp/books/mapped-book.pdf",
				"fileSizeBytes": 321,
				"checksumSha256": "` + checksum + `",
				"importedAt": "2026-01-02T03:04:05Z",
				"createdAt": "2026-01-02T03:04:05Z",
				"updatedAt": "2026-01-02T03:04:05Z"
			}]
		}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, time.Second)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.SetSession("access-token", "refresh-token", "user-1")

	items, err := client.ListEbooks(context.Background(), 10)
	if err != nil {
		t.Fatalf("list ebooks: %v", err)
	}
	if gotAuth != "Bearer access-token" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if len(items) != 1 {
		t.Fatalf("expected one ebook, got %d", len(items))
	}
	got := items[0]
	if got.FileSizeBytes != 321 || got.ChecksumSHA256 != checksum || got.LanguageCode != "fr" {
		t.Fatalf("unexpected mapped ebook fields: %#v", got)
	}
}
