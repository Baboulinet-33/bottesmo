package game

import (
	"testing"
)

func TestGuessCorrect(t *testing.T) {
	g := NewGame("ABRITE", ModeSolo)
	results, err := g.Guess("ABRITE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !g.Won {
		t.Error("expected won = true")
	}
	if !g.GameOver {
		t.Error("expected gameOver = true")
	}
	for i, r := range results {
		if r.Status != StatusCorrect {
			t.Errorf("result[%d] expected StatusCorrect, got %v", i, r.Status)
		}
	}
}

func TestGuessAllWrong(t *testing.T) {
	g := NewGame("ABRITE", ModeSolo)
	results, err := g.Guess("XYZXYZ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Won {
		t.Error("expected won = false")
	}
	for i, r := range results {
		if r.Status != StatusAbsent {
			t.Errorf("result[%d] expected StatusAbsent, got %v", i, r.Status)
		}
	}
}

func TestGuessMixed(t *testing.T) {
	g := NewGame("ABRITE", ModeSolo)
	results, err := g.Guess("ABIDES")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != StatusCorrect {
		t.Errorf("result[0] (A) expected StatusCorrect, got %v", results[0].Status)
	}
	if results[1].Status != StatusCorrect {
		t.Errorf("result[1] (B) expected StatusCorrect, got %v", results[1].Status)
	}
	if results[2].Status != StatusPresent {
		t.Errorf("result[2] (I) expected StatusPresent, got %v", results[2].Status)
	}
	if results[4].Status != StatusPresent {
		t.Errorf("result[4] (E) expected StatusPresent, got %v", results[4].Status)
	}
}

func TestGuessDuplicateLetters(t *testing.T) {
	g := NewGame("ACCORD", ModeSolo)
	results, err := g.Guess("CACTUS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []LetterStatus{StatusPresent, StatusPresent, StatusCorrect, StatusAbsent, StatusAbsent, StatusAbsent}
	for i, e := range expected {
		if results[i].Status != e {
			t.Errorf("result[%d] (%c) expected %v, got %v", i, results[i].Letter, e, results[i].Status)
		}
	}
}

func TestGuessOverMaxAttempts(t *testing.T) {
	g := NewGame("ABRITE", ModeSolo)
	for i := 0; i < 6; i++ {
		g.Guess("XYZXYZ")
	}
	if !g.GameOver {
		t.Error("expected gameOver after 6 attempts")
	}
	if g.Won {
		t.Error("expected won = false after 6 failed attempts")
	}
}

func TestGuessInvalidLength(t *testing.T) {
	g := NewGame("ABRITE", ModeSolo)
	_, err := g.Guess("ABC")
	if err == nil {
		t.Error("expected error for wrong length")
	}
	_, err = g.Guess("ABCDEFG")
	if err == nil {
		t.Error("expected error for wrong length")
	}
}

func TestGuessAfterGameOver(t *testing.T) {
	g := NewGame("ABRITE", ModeSolo)
	g.GameOver = true
	_, err := g.Guess("ABRITE")
	if err == nil {
		t.Error("expected error when game is over")
	}
}

func TestNewGameInitialState(t *testing.T) {
	g := NewGame("ABRITE", ModeDaily)
	if g.Won {
		t.Error("expected Won = false")
	}
	if g.GameOver {
		t.Error("expected GameOver = false")
	}
	if g.MaxTries != 6 {
		t.Errorf("expected MaxTries = 6, got %d", g.MaxTries)
	}
	if g.Mode != ModeDaily {
		t.Errorf("expected Mode = ModeDaily, got %v", g.Mode)
	}
	if len(g.Attempts) != 0 {
		t.Errorf("expected 0 attempts, got %d", len(g.Attempts))
	}
}
