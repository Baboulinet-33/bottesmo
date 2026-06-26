package game

import (
	"testing"
	"time"
)

func TestGenerateWordSequence_Progressive(t *testing.T) {
	tests := []struct {
		count int
		want  []int
	}{
		{3, []int{6, 7, 8}},
		{5, []int{6, 7, 8, 9, 10}},
		{7, []int{6, 7, 8, 9, 10, 6, 7}},
		{10, []int{6, 7, 8, 9, 10, 6, 7, 8, 9, 10}},
	}
	for _, tt := range tests {
		got := GenerateWordSequence(VSModeProgressive, tt.count)
		if len(got) != tt.count {
			t.Errorf("count=%d: got length %d, want %d", tt.count, len(got), tt.count)
		}
		for i, v := range got {
			if v != tt.want[i] {
				t.Errorf("count=%d, seq[%d] = %d, want %d", tt.count, i, v, tt.want[i])
			}
		}
	}
}

func TestGenerateWordSequence_Random(t *testing.T) {
	count := 50
	seq := GenerateWordSequence(VSModeRandom, count)
	if len(seq) != count {
		t.Errorf("got length %d, want %d", len(seq), count)
	}
	for i, v := range seq {
		if v < 6 || v > 10 {
			t.Errorf("seq[%d] = %d, want between 6 and 10", i, v)
		}
	}
}

func TestAddPlayer(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 5, VSModeProgressive)

	err := AddPlayer(room, "p1", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(room.Players))
	}

	err = AddPlayer(room, "p2", "Bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(room.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(room.Players))
	}

	err = AddPlayer(room, "p1", "AliceDuplicate")
	if err == nil {
		t.Error("expected error for duplicate player ID")
	}

	for i := 0; i < 30; i++ {
		pid := string(rune('z' + i))
		err := AddPlayer(room, pid, "Extra")
		if i < 18 {
			if err != nil {
				t.Fatalf("unexpected error on player %d: %v", i+3, err)
			}
		} else {
			if err == nil {
				t.Error("expected error when room is full (max 20)")
			}
		}
	}

	if len(room.Players) > 20 {
		t.Errorf("players exceeded 20: got %d", len(room.Players))
	}
}

func TestAddPlayerAfterStart(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 3, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")
	StartGame(room, []string{"ABRITE", "ACCORD", "ACTION"})

	err := AddPlayer(room, "p3", "Charlie")
	if err == nil {
		t.Error("expected error when adding player after game started")
	}
}

func TestStartGame(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 3, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")

	words := []string{"ABRITE", "ACCORD", "ACTION"}
	err := StartGame(room, words)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !room.Started {
		t.Error("expected room to be started")
	}

	if len(room.Words) != 3 {
		t.Errorf("expected 3 words, got %d", len(room.Words))
	}

	for i, w := range words {
		if room.Words[i] != w {
			t.Errorf("word[%d] = %s, want %s", i, room.Words[i], w)
		}
	}

	for _, p := range room.Players {
		if p.StartTime.IsZero() {
			t.Error("expected StartTime to be set")
		}
	}
}

func TestStartGameTooFewPlayers(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 3, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")

	err := StartGame(room, []string{"ABRITE", "ACCORD", "ACTION"})
	if err == nil {
		t.Error("expected error when starting with < 2 players")
	}
}

func TestStartGameTwice(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 3, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")

	StartGame(room, []string{"ABRITE", "ACCORD", "ACTION"})

	err := StartGame(room, []string{"BEAUTE", "BIJOUX", "BATARD"})
	if err == nil {
		t.Error("expected error when starting twice")
	}
}

func TestProcessGuessCorrect(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 2, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")
	StartGame(room, []string{"ABRITE", "ACCORD"})

	results, wordComplete, allComplete, currentWord, err := ProcessGuess(room, "p1", "ABRITE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wordComplete {
		t.Error("expected wordComplete to be true")
	}
	if allComplete {
		t.Error("expected allComplete to be false (more words remain)")
	}
	if currentWord != 1 {
		t.Errorf("expected currentWord=1, got %d", currentWord)
	}
	for _, r := range results {
		if r.Status != StatusCorrect {
			t.Errorf("expected all correct, got %v", r.Status)
		}
	}

	results, wordComplete, allComplete, currentWord, err = ProcessGuess(room, "p1", "ACCORD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wordComplete {
		t.Error("expected wordComplete to be true")
	}
	if !allComplete {
		t.Error("expected allComplete to be true (last word)")
	}
	if currentWord != 2 {
		t.Errorf("expected currentWord=2, got %d", currentWord)
	}

	_, _, _, _, err = ProcessGuess(room, "p1", "ACTION")
	if err == nil {
		t.Error("expected error when player already completed all words")
	}
}

func TestProcessGuessWrong(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 1, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")
	StartGame(room, []string{"ABRITE"})

	results, wordComplete, allComplete, currentWord, err := ProcessGuess(room, "p1", "ABIDES")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wordComplete {
		t.Error("expected wordComplete to be false")
	}
	if allComplete {
		t.Error("expected allComplete to be false")
	}
	if currentWord != 0 {
		t.Errorf("expected currentWord=0, got %d", currentWord)
	}
	_ = results
}

func TestProcessGuessFailsAfterSixAttempts(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 1, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")
	StartGame(room, []string{"ABRITE"})

	for i := 0; i < 5; i++ {
		_, _, allComplete, _, err := ProcessGuess(room, "p1", "ABIDES")
		if err != nil {
			t.Fatalf("unexpected error on attempt %d: %v", i, err)
		}
		if allComplete {
			t.Errorf("unexpected allComplete on attempt %d", i)
		}
	}

	_, _, allComplete, currentWord, err := ProcessGuess(room, "p1", "ABIDES")
	if err != nil {
		t.Fatalf("unexpected error on final attempt: %v", err)
	}
	if !allComplete {
		t.Error("expected allComplete to be true after 6 failed attempts")
	}
	if currentWord != 0 {
		t.Errorf("expected currentWord=0 (failed), got %d", currentWord)
	}

	p := room.Players["p1"]
	if !p.Failed {
		t.Error("expected player to be marked as failed")
	}
}

func TestProcessGuessAfterFailure(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 1, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")
	StartGame(room, []string{"ABRITE"})

	for i := 0; i < 6; i++ {
		ProcessGuess(room, "p1", "ABIDES")
	}

	_, _, _, _, err := ProcessGuess(room, "p1", "ABRITE")
	if err == nil {
		t.Error("expected error when player has already failed")
	}
}

func TestGetRankings(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 2, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")
	AddPlayer(room, "p3", "Charlie")
	StartGame(room, []string{"ABRITE", "ACCORD"})

	ProcessGuess(room, "p1", "ABRITE")
	time.Sleep(10 * time.Millisecond)
	ProcessGuess(room, "p2", "ABRITE")
	ProcessGuess(room, "p1", "ACCORD")
	ProcessGuess(room, "p2", "ACCORD")

	for i := 0; i < 6; i++ {
		ProcessGuess(room, "p3", "ABIDES")
	}

	rankings := GetRankings(room)

	if len(rankings) != 3 {
		t.Fatalf("expected 3 rankings, got %d", len(rankings))
	}

	if rankings[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", rankings[0].Rank)
	}
	if rankings[0].Failed {
		t.Error("expected rank 1 to not be failed")
	}

	if rankings[1].Rank != 2 {
		t.Errorf("expected rank 2, got %d", rankings[1].Rank)
	}
	if rankings[1].Failed {
		t.Error("expected rank 2 to not be failed")
	}

	if rankings[2].Rank != 3 {
		t.Errorf("expected rank 3, got %d", rankings[2].Rank)
	}
	if !rankings[2].Failed {
		t.Error("expected rank 3 to be failed")
	}
	if rankings[2].CompletedTime != nil {
		t.Error("expected failed player to have nil CompletedTime")
	}

	if *rankings[0].CompletedTime > *rankings[1].CompletedTime {
		t.Error("expected rank 1 faster than rank 2")
	}
}

func TestIsRoomFinished(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 1, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")

	if IsRoomFinished(room) {
		t.Error("expected false before game starts")
	}

	StartGame(room, []string{"ABRITE"})

	if IsRoomFinished(room) {
		t.Error("expected false when nobody has finished")
	}

	ProcessGuess(room, "p1", "ABRITE")

	if IsRoomFinished(room) {
		t.Error("expected false when not all players finished")
	}

	ProcessGuess(room, "p2", "ABRITE")

	if !IsRoomFinished(room) {
		t.Error("expected true when all players have finished")
	}
}

func TestRematch(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 3, VSModeRandom)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")
	StartGame(room, []string{"ABRITE", "ACCORD", "ACTION"})

	newRoom := NewVSRoom("r2", "XYZ789", "creator", room.WordCount, room.LengthMode)

	for _, pid := range room.PlayerOrder {
		p := room.Players[pid]
		AddPlayer(newRoom, pid+"-new", p.Nickname)
	}

	StartGame(newRoom, []string{"BEAUTE", "BIJOUX", "BATARD"})

	if newRoom.WordCount != room.WordCount {
		t.Errorf("expected WordCount %d, got %d", room.WordCount, newRoom.WordCount)
	}
	if newRoom.LengthMode != room.LengthMode {
		t.Errorf("expected LengthMode %s, got %s", room.LengthMode, newRoom.LengthMode)
	}
	if len(newRoom.Players) != len(room.Players) {
		t.Errorf("expected %d players, got %d", len(room.Players), len(newRoom.Players))
	}
	if !newRoom.Started {
		t.Error("expected new room to be started")
	}
}

func TestSetRoomFinished(t *testing.T) {
	room := NewVSRoom("r1", "ABC123", "creator", 1, VSModeProgressive)
	AddPlayer(room, "p1", "Alice")
	AddPlayer(room, "p2", "Bob")
	StartGame(room, []string{"ABRITE"})

	if room.Finished {
		t.Error("expected Finished to be false initially")
	}

	SetRoomFinished(room)

	if !room.Finished {
		t.Error("expected Finished to be true after SetRoomFinished")
	}
}
