package game

import (
	"fmt"
	"strings"
)

func NewGame(target string, mode GameMode) *Game {
	return &Game{
		Target:     strings.ToUpper(target),
		MaxTries:   6,
		Mode:       mode,
		WordLength: len(target),
	}
}

func (g *Game) Guess(word string) ([]LetterResult, error) {
	word = strings.ToUpper(word)

	if g.GameOver {
		return nil, fmt.Errorf("game is over")
	}

	if len(word) != g.WordLength {
		return nil, fmt.Errorf("word must be exactly %d letters", g.WordLength)
	}

	targetRunes := []rune(g.Target)
	guessRunes := []rune(word)

	remaining := make(map[rune]int)
	for _, r := range targetRunes {
		remaining[r]++
	}

	results := make([]LetterResult, g.WordLength)

	for i := 0; i < g.WordLength; i++ {
		if guessRunes[i] == targetRunes[i] {
			results[i] = LetterResult{Letter: guessRunes[i], Status: StatusCorrect}
			remaining[guessRunes[i]]--
		} else {
			results[i] = LetterResult{Letter: guessRunes[i], Status: StatusAbsent}
		}
	}

	for i := 0; i < g.WordLength; i++ {
		if results[i].Status == StatusCorrect {
			continue
		}
		if remaining[guessRunes[i]] > 0 {
			results[i].Status = StatusPresent
			remaining[guessRunes[i]]--
		}
	}

	g.Attempts = append(g.Attempts, word)

	if word == g.Target {
		g.Won = true
		g.GameOver = true
	} else if len(g.Attempts) >= g.MaxTries {
		g.GameOver = true
	}

	return results, nil
}
