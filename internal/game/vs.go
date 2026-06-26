package game

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type VSWordLengthMode string

const (
	VSModeProgressive VSWordLengthMode = "progressive"
	VSModeRandom      VSWordLengthMode = "random"
)

type VSPlayer struct {
	ID            string
	Nickname      string
	CurrentWord   int
	WordGames     []*Game
	StartTime     time.Time
	CompletedTime *time.Time
	Failed        bool
}

type VSRoom struct {
	mu          sync.Mutex
	ID          string
	Code        string
	CreatorID   string
	Players     map[string]*VSPlayer
	PlayerOrder []string
	Words       []string
	WordCount   int
	LengthMode  VSWordLengthMode
	Started     bool
	Finished    bool
	CreatedAt   time.Time
}

type RankingEntry struct {
	PlayerID      string
	Nickname      string
	CompletedTime *time.Duration
	Failed        bool
	Rank          int
}

func GenerateWordSequence(mode VSWordLengthMode, count int) []int {
	seq := make([]int, count)
	progressiveLengths := []int{6, 7, 8, 9, 10}
	for i := 0; i < count; i++ {
		if mode == VSModeProgressive {
			seq[i] = progressiveLengths[i%len(progressiveLengths)]
		} else {
			seq[i] = 6 + rand.Intn(5)
		}
	}
	return seq
}

func NewVSRoom(id, code, creatorID string, wordCount int, lengthMode VSWordLengthMode) *VSRoom {
	return &VSRoom{
		ID:         id,
		Code:       code,
		CreatorID:  creatorID,
		Players:    make(map[string]*VSPlayer),
		WordCount:  wordCount,
		LengthMode: lengthMode,
		CreatedAt:  time.Now(),
	}
}

func AddPlayer(room *VSRoom, playerID, nickname string) error {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Started {
		return fmt.Errorf("game already started")
	}

	if len(room.Players) >= 20 {
		return fmt.Errorf("room is full (max 20 players)")
	}

	if _, ok := room.Players[playerID]; ok {
		return fmt.Errorf("player already in room")
	}

	room.Players[playerID] = &VSPlayer{
		ID:        playerID,
		Nickname:  nickname,
		WordGames: make([]*Game, room.WordCount),
	}
	room.PlayerOrder = append(room.PlayerOrder, playerID)
	return nil
}

func StartGame(room *VSRoom, words []string) error {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Started {
		return fmt.Errorf("game already started")
	}

	if len(room.Players) < 2 {
		return fmt.Errorf("need at least 2 players")
	}

	room.Words = words
	room.Started = true
	now := time.Now()
	for _, p := range room.Players {
		p.StartTime = now
	}
	return nil
}

func ProcessGuess(room *VSRoom, playerID, word string) (results []LetterResult, wordComplete, allComplete bool, currentWord int, err error) {
	room.mu.Lock()
	defer room.mu.Unlock()

	player, ok := room.Players[playerID]
	if !ok {
		return nil, false, false, 0, fmt.Errorf("player not found")
	}

	if player.Failed {
		return nil, false, false, 0, fmt.Errorf("player has already failed")
	}

	if player.CompletedTime != nil {
		return nil, false, false, 0, fmt.Errorf("player has already completed all words")
	}

	wordIdx := player.CurrentWord
	if wordIdx >= len(room.Words) {
		return nil, false, true, wordIdx, nil
	}

	g := player.WordGames[wordIdx]
	if g == nil {
		target := room.Words[wordIdx]
		g = NewGame(target, ModeSolo)
		player.WordGames[wordIdx] = g
	}

	results, err = g.Guess(word)
	if err != nil {
		return nil, false, false, wordIdx, err
	}

	wordComplete = g.Won
	allComplete = false

	if wordComplete {
		player.CurrentWord++
		if player.CurrentWord >= len(room.Words) {
			now := time.Now()
			player.CompletedTime = &now
			allComplete = true
		}
	} else if g.GameOver {
		player.Failed = true
		allComplete = true
	}

	return results, wordComplete, allComplete, player.CurrentWord, nil
}

func GetRankings(room *VSRoom) []RankingEntry {
	room.mu.Lock()
	defer room.mu.Unlock()

	var winners []RankingEntry
	var failed []RankingEntry

	for _, pid := range room.PlayerOrder {
		p := room.Players[pid]
		entry := RankingEntry{
			PlayerID: p.ID,
			Nickname: p.Nickname,
			Failed:   p.Failed,
		}
		if !p.Failed && p.CompletedTime != nil {
			duration := p.CompletedTime.Sub(p.StartTime)
			entry.CompletedTime = &duration
			winners = append(winners, entry)
		} else {
			failed = append(failed, entry)
		}
	}

	sort.Slice(winners, func(i, j int) bool {
		return *winners[i].CompletedTime < *winners[j].CompletedTime
	})

	entries := append(winners, failed...)
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries
}

func SetRoomFinished(room *VSRoom) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.Finished = true
}

func IsRoomFinished(room *VSRoom) bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	if !room.Started {
		return false
	}

	for _, p := range room.Players {
		if !p.Failed && p.CompletedTime == nil {
			return false
		}
	}
	return true
}
