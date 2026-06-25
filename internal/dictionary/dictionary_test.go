package dictionary

import (
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
