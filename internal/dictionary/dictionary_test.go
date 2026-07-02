package dictionary

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDictionaryLoad(t *testing.T) {
	Reset()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	content := "ABRITE\nACCORD\nACTION\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
}

func TestFullDictionaryLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words_full.txt")
	content := "ABRITE\nACCORD\nACTION\nBONJOUR\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := LoadFull(path); err != nil {
		t.Fatalf("LoadFull failed: %v", err)
	}
}

func TestIsValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words_full.txt")
	content := "ABRITE\nACCORD\nACTION\nBONJOUR\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := LoadFull(path); err != nil {
		t.Fatalf("LoadFull failed: %v", err)
	}

	if !IsValid("ABRITE") {
		t.Error("expected ABRITE to be valid")
	}
	if !IsValid("bonjour") {
		t.Error("expected bonjour (lowercase) to be valid")
	}
	if IsValid("XXXXXX") {
		t.Error("expected XXXXXX to be invalid")
	}
	if IsValid("") {
		t.Error("expected empty string to be invalid")
	}
}

func TestDailyWordDeterministic(t *testing.T) {
	Reset()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	content := "AAAAAA\nBBBBBB\nCCCCCC\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	date := "2025-01-15"
	word1, err := DailyWord(6, date)
	if err != nil {
		t.Fatalf("DailyWord failed: %v", err)
	}

	word2, err := DailyWord(6, date)
	if err != nil {
		t.Fatalf("DailyWord failed: %v", err)
	}

	if word1 != word2 {
		t.Errorf("DailyWord should be deterministic for same date, got %q then %q", word1, word2)
	}

	word3, err := DailyWord(6, "2025-06-20")
	if err != nil {
		t.Fatalf("DailyWord failed: %v", err)
	}

	if word1 == word3 {
		t.Log("same word on different dates (possible collision, not an error)")
	}
}

func TestWordsByLength(t *testing.T) {
	Reset()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	content := "ABRITE\nACCORD\nACTION\nBRUNISS\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	sixLetter := WordsByLength(6)
	if len(sixLetter) != 3 {
		t.Errorf("expected 3 six-letter words, got %d", len(sixLetter))
	}

	sevenLetter := WordsByLength(7)
	if len(sevenLetter) != 1 {
		t.Errorf("expected 1 seven-letter word, got %d", len(sevenLetter))
	}
}

func TestMinMaxLength(t *testing.T) {
	Reset()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	content := "AAAAAA\nBBBBBBB\nCCCCCCCC\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if MinLength() != 6 {
		t.Errorf("expected MinLength = 6, got %d", MinLength())
	}
	if MaxLength() != 8 {
		t.Errorf("expected MaxLength = 8, got %d", MaxLength())
	}
}

func TestRandomWord(t *testing.T) {
	Reset()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	content := "ABRITE\nACCORD\nACTION\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	word, err := RandomWord(6)
	if err != nil {
		t.Fatalf("RandomWord failed: %v", err)
	}

	valid := word == "ABRITE" || word == "ACCORD" || word == "ACTION"
	if !valid {
		t.Errorf("unexpected word: %q", word)
	}
}

func TestVersionLineDetection(t *testing.T) {
	Reset()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	content := "DICT_VERSION=3\nABRITE\nACCORD\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := LoadFromSource(path)
	if err != nil {
		t.Fatalf("LoadFromSource failed: %v", err)
	}

	if result.Version != "3" {
		t.Errorf("expected version=3, got %q", result.Version)
	}

	sixLetter := WordsByLength(6)
	if len(sixLetter) != 2 {
		t.Errorf("expected 2 six-letter words, got %d", len(sixLetter))
	}
}

func TestLoadFromSourceLocal(t *testing.T) {
	Reset()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	content := "ABRITE\nACCORD\nACTION\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	expectedSHA := sha256.Sum256([]byte(content))

	result, err := LoadFromSource(path)
	if err != nil {
		t.Fatalf("LoadFromSource failed: %v", err)
	}

	if result.WordCount != 3 {
		t.Errorf("expected 3 words, got %d", result.WordCount)
	}

	if result.SHA256 != expectedSHA {
		t.Errorf("SHA256 mismatch: got %x, expected %x", result.SHA256, expectedSHA)
	}
}

func TestLoadFromSourceHTTP(t *testing.T) {
	Reset()
	content := "ABRITE\nACCORD\nACTION\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}))
	defer srv.Close()

	result, err := LoadFromSource(srv.URL)
	if err != nil {
		t.Fatalf("LoadFromSource HTTP failed: %v", err)
	}

	if result.WordCount != 3 {
		t.Errorf("expected 3 words, got %d", result.WordCount)
	}

	sixLetter := WordsByLength(6)
	if len(sixLetter) != 3 {
		t.Errorf("expected 3 six-letter words, got %d", len(sixLetter))
	}
}

func TestSHA256Consistency(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	content := "ABRITE\nACCORD\nACTION\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	expectedSHA := sha256.Sum256([]byte(content))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}))
	defer srv.Close()

	Reset()
	resultFile, err := LoadFromSource(path)
	if err != nil {
		t.Fatalf("LoadFromSource local failed: %v", err)
	}

	Reset()
	resultHTTP, err := LoadFromSource(srv.URL)
	if err != nil {
		t.Fatalf("LoadFromSource HTTP failed: %v", err)
	}

	if resultFile.SHA256 != resultHTTP.SHA256 {
		t.Errorf("SHA256 mismatch: local=%x, http=%x", resultFile.SHA256, resultHTTP.SHA256)
	}

	if resultFile.SHA256 != expectedSHA {
		t.Errorf("SHA256 does not match expected: got %x, expected %x", resultFile.SHA256, expectedSHA)
	}
}
