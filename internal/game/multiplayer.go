package game

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"tusmo/internal/dictionary"
)

type MultiplayerRoom struct {
	Code       string
	Mode       string
	WordCount  int
	CreatorID  string
	Players    map[string]*MultiplayerPlayer
	Started    bool
	StartTime  time.Time
	Finished   bool
	State      string
	CreatedAt  time.Time
	MaxPlayers int
}

type MultiplayerPlayer struct {
	ID             string
	Nickname       string
	WordSequence   []string
	WordGames      []*Game
	CurrentWordIdx int
	StartTime      time.Time
	CompletedTime  time.Time
	Failed         bool
	Finished       bool
	JoinedAt       time.Time
}

type RankingEntry struct {
	PlayerID string
	Nickname string
	Time     time.Duration
	Failed   bool
	Finished bool
}

type MultiplayerGuessResult struct {
	Results        []LetterResult `json:"results"`
	WordFinished   bool           `json:"wordFinished"`
	WordWon        bool           `json:"wordWon"`
	PlayerFinished bool           `json:"playerFinished"`
	PlayerFailed   bool           `json:"playerFailed"`
	CurrentWordIdx int            `json:"currentWordIdx"`
	TotalWords     int            `json:"totalWords"`
}

const roomCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func GenerateRoomCode() string {
	code := make([]byte, 6)
	for i := range code {
		code[i] = roomCodeChars[rand.Intn(len(roomCodeChars))]
	}
	return string(code)
}

func GenerateWordSequence(mode string, count int) []string {
	sequence := make([]string, count)
	lengths := []int{6, 7, 8, 9, 10}

	for i := 0; i < count; i++ {
		var l int
		switch mode {
		case "aleatoire":
			l = 6 + rand.Intn(5)
		default:
			l = lengths[i%len(lengths)]
		}

		word, err := dictionary.RandomWord(l)
		if err != nil {
			for attempt := 0; attempt < 10; attempt++ {
				l = 6 + rand.Intn(5)
				word, err = dictionary.RandomWord(l)
				if err == nil {
					break
				}
			}
			if err != nil {
				sequence[i] = "AAAAAA"
				continue
			}
		}
		sequence[i] = word
	}

	return sequence
}

func NewMultiplayerRoom(code, mode string, wordCount int, creatorID, creatorNickname string) *MultiplayerRoom {
	room := &MultiplayerRoom{
		Code:       code,
		Mode:       mode,
		WordCount:  wordCount,
		CreatorID:  creatorID,
		Players:    make(map[string]*MultiplayerPlayer),
		State:      "lobby",
		CreatedAt:  time.Now(),
		MaxPlayers: 20,
	}

	creator := &MultiplayerPlayer{
		ID:       creatorID,
		Nickname: creatorNickname,
		JoinedAt: time.Now(),
	}
	room.Players[creatorID] = creator

	return room
}

func (r *MultiplayerRoom) AddPlayer(id, nickname string) error {
	if len(r.Players) >= r.MaxPlayers {
		return fmt.Errorf("room is full")
	}
	if _, exists := r.Players[id]; exists {
		return nil
	}

	player := &MultiplayerPlayer{
		ID:       id,
		Nickname: nickname,
		JoinedAt: time.Now(),
	}
	r.Players[id] = player
	return nil
}

func (r *MultiplayerRoom) StartGame() error {
	if r.State != "lobby" {
		return fmt.Errorf("game already started")
	}
	if len(r.Players) < 1 {
		return fmt.Errorf("no players in room")
	}

	r.State = "playing"
	r.StartTime = time.Now()

	for _, player := range r.Players {
		player.WordSequence = GenerateWordSequence(r.Mode, r.WordCount)
		player.WordGames = make([]*Game, r.WordCount)
		for i, word := range player.WordSequence {
			player.WordGames[i] = NewGame(word, ModeSolo)
		}
		player.CurrentWordIdx = 0
		player.StartTime = time.Now()
	}

	return nil
}

func (r *MultiplayerRoom) ProcessGuess(playerID, guess string) (*MultiplayerGuessResult, error) {
	player, ok := r.Players[playerID]
	if !ok {
		return nil, fmt.Errorf("player not found")
	}
	if player.Finished {
		return nil, fmt.Errorf("player already finished")
	}

	currentGame := player.WordGames[player.CurrentWordIdx]
	results, err := currentGame.Guess(guess)
	if err != nil {
		return nil, err
	}

	result := &MultiplayerGuessResult{
		Results:        results,
		CurrentWordIdx: player.CurrentWordIdx,
		TotalWords:     r.WordCount,
	}

	if currentGame.GameOver {
		result.WordFinished = true
		if currentGame.Won {
			player.CurrentWordIdx++
			result.WordWon = true

			if player.CurrentWordIdx >= r.WordCount {
				player.Finished = true
				player.CompletedTime = time.Now()
				result.PlayerFinished = true
			}
		} else {
			player.Failed = true
			player.Finished = true
			player.CompletedTime = time.Now()
			result.PlayerFinished = true
			result.PlayerFailed = true
		}
	}

	return result, nil
}

func (r *MultiplayerRoom) GetRankings() []RankingEntry {
	var entries []RankingEntry
	for _, p := range r.Players {
		entry := RankingEntry{
			PlayerID: p.ID,
			Nickname: p.Nickname,
			Failed:   p.Failed,
			Finished: p.Finished,
		}
		if p.Finished {
			entry.Time = p.CompletedTime.Sub(r.StartTime)
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Finished != entries[j].Finished {
			return entries[i].Finished
		}
		if entries[i].Failed != entries[j].Failed {
			return !entries[i].Failed
		}
		return entries[i].Time < entries[j].Time
	})

	return entries
}

func (r *MultiplayerRoom) IsRoomFinished() bool {
	for _, p := range r.Players {
		if !p.Finished {
			return false
		}
	}
	return true
}

func (r *MultiplayerRoom) RemovePlayer(playerID string) {
	delete(r.Players, playerID)
}

func ComputeAttemptResults(target string, attempts []string) ([][]LetterResult, error) {
	results := make([][]LetterResult, len(attempts))
	for i, word := range attempts {
		rowResults, err := computeResults(target, word)
		if err != nil {
			return nil, err
		}
		results[i] = rowResults
	}
	return results, nil
}
