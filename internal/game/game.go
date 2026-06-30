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

func computeResults(target, word string) ([]LetterResult, error) {
	word = strings.ToUpper(word)
	if len(word) != len(target) {
		return nil, fmt.Errorf("word must be exactly %d letters", len(target))
	}
	if word[0] != target[0] {
		return nil, fmt.Errorf("word must start with %c", target[0])
	}

	targetRunes := []rune(target)
	guessRunes := []rune(word)

	remaining := make(map[rune]int)
	for _, r := range targetRunes {
		remaining[r]++
	}

	results := make([]LetterResult, len(target))
	for i := 0; i < len(target); i++ {
		if guessRunes[i] == targetRunes[i] {
			results[i] = LetterResult{Letter: guessRunes[i], Status: StatusCorrect}
			remaining[guessRunes[i]]--
		} else {
			results[i] = LetterResult{Letter: guessRunes[i], Status: StatusAbsent}
		}
	}
	for i := 0; i < len(target); i++ {
		if results[i].Status == StatusCorrect {
			continue
		}
		if remaining[guessRunes[i]] > 0 {
			results[i].Status = StatusPresent
			remaining[guessRunes[i]]--
		}
	}
	return results, nil
}

func (g *Game) Guess(word string) ([]LetterResult, error) {
	if g.GameOver {
		return nil, fmt.Errorf("game is over")
	}

	word = strings.ToUpper(word)
	results, err := computeResults(g.Target, word)
	if err != nil {
		return nil, err
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
