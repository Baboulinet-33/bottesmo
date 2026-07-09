package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bottesmo/internal/dictionary"
)

func setupTestDicts(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()

	wordsPath := filepath.Join(dir, "words.txt")
	wordsContent := "DICT_VERSION=2\nABRITE\nACCORD\nACTION\n"
	if err := os.WriteFile(wordsPath, []byte(wordsContent), 0644); err != nil {
		t.Fatal(err)
	}

	fullPath := filepath.Join(dir, "words_full.txt")
	fullContent := "DICT_VERSION=2\nABRITE\nACCORD\nACTION\nBONJOUR\n"
	if err := os.WriteFile(fullPath, []byte(fullContent), 0644); err != nil {
		t.Fatal(err)
	}

	return wordsPath, fullPath
}

func TestStatusHandler_Returns200(t *testing.T) {
	dictionary.Reset()
	wordsPath, fullPath := setupTestDicts(t)

	if _, err := dictionary.LoadFromSource(wordsPath); err != nil {
		t.Fatalf("LoadFromSource failed: %v", err)
	}
	if _, err := dictionary.LoadFullFromSource(fullPath); err != nil {
		t.Fatalf("LoadFullFromSource failed: %v", err)
	}

	mgr := NewGameManager()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	mgr.StatusHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", body["status"])
	}
	if body["version"] == "" {
		t.Error("expected non-empty version")
	}

	dicts, ok := body["dictionaries"].([]any)
	if !ok {
		t.Fatal("expected dictionaries to be an array")
	}
	if len(dicts) != 2 {
		t.Fatalf("expected 2 dictionaries, got %d", len(dicts))
	}
}

func TestSettingsPageHandler(t *testing.T) {
	if err := LoadTemplates("../../web/templates/*.html"); err != nil {
		t.Fatalf("LoadTemplates failed: %v", err)
	}

	mgr := NewGameManager()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rec := httptest.NewRecorder()

	mgr.SettingsPageHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Param") {
		t.Error("expected body to contain 'Param' (Paramètres)")
	}
	if !strings.Contains(body, "palette-grid") {
		t.Error("expected body to contain 'palette-grid'")
	}
}

func TestStatusHandler_DictionaryEntries(t *testing.T) {
	dictionary.Reset()
	wordsPath, fullPath := setupTestDicts(t)

	wordsResult, err := dictionary.LoadFromSource(wordsPath)
	if err != nil {
		t.Fatalf("LoadFromSource failed: %v", err)
	}
	fullResult, err := dictionary.LoadFullFromSource(fullPath)
	if err != nil {
		t.Fatalf("LoadFullFromSource failed: %v", err)
	}

	mgr := NewGameManager()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	mgr.StatusHandler(rec, req)

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	dicts := body["dictionaries"].([]any)

	wordsEntry := dicts[0].(map[string]any)
	if wordsEntry["name"] != "words" {
		t.Errorf("expected name 'words', got %v", wordsEntry["name"])
	}
	if wordsEntry["word_count"] != float64(wordsResult.WordCount) {
		t.Errorf("expected word_count %d, got %v", wordsResult.WordCount, wordsEntry["word_count"])
	}
	if wordsEntry["sha256"] == "" {
		t.Error("expected non-empty sha256")
	}

	fullEntry := dicts[1].(map[string]any)
	if fullEntry["name"] != "words_full" {
		t.Errorf("expected name 'words_full', got %v", fullEntry["name"])
	}
	if fullEntry["word_count"] != float64(fullResult.WordCount) {
		t.Errorf("expected word_count %d, got %v", fullResult.WordCount, fullEntry["word_count"])
	}
	if fullEntry["sha256"] == "" {
		t.Error("expected non-empty sha256")
	}
}

func TestStatusHandler_NonGetReturns405(t *testing.T) {
	dictionary.Reset()
	wordsPath, fullPath := setupTestDicts(t)

	if _, err := dictionary.LoadFromSource(wordsPath); err != nil {
		t.Fatalf("LoadFromSource failed: %v", err)
	}
	if _, err := dictionary.LoadFullFromSource(fullPath); err != nil {
		t.Fatalf("LoadFullFromSource failed: %v", err)
	}

	mgr := NewGameManager()

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/status", nil)
		rec := httptest.NewRecorder()
		mgr.StatusHandler(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405 for %s, got %d", method, rec.Code)
		}
	}
}

func TestStatusHandler_FieldsPresent(t *testing.T) {
	dictionary.Reset()
	wordsPath, fullPath := setupTestDicts(t)

	if _, err := dictionary.LoadFromSource(wordsPath); err != nil {
		t.Fatalf("LoadFromSource failed: %v", err)
	}
	if _, err := dictionary.LoadFullFromSource(fullPath); err != nil {
		t.Fatalf("LoadFullFromSource failed: %v", err)
	}

	mgr := NewGameManager()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	mgr.StatusHandler(rec, req)

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if _, ok := body["status"]; !ok {
		t.Error("missing 'status' key")
	}
	if _, ok := body["version"]; !ok {
		t.Error("missing 'version' key")
	}
	if _, ok := body["dictionaries"]; !ok {
		t.Fatal("missing 'dictionaries' key")
	}

	dicts := body["dictionaries"].([]any)
	for i, entry := range dicts {
		e := entry.(map[string]any)
		for _, key := range []string{"name", "word_count", "sha256"} {
			if _, ok := e[key]; !ok {
				t.Errorf("dictionary[%d] missing '%s' key", i, key)
			}
		}
	}
}
