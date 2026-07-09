package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"bottesmo/internal/dictionary"
	"bottesmo/internal/game"
)

func setupMultiTestDicts(t *testing.T) {
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

	dictionary.Reset()
	if _, err := dictionary.LoadFromSource(wordsPath); err != nil {
		t.Fatalf("LoadFromSource failed: %v", err)
	}
	if _, err := dictionary.LoadFullFromSource(fullPath); err != nil {
		t.Fatalf("LoadFullFromSource failed: %v", err)
	}
}

func newTestGameManager() *GameManager {
	return &GameManager{
		multi: NewMultiplayerManager(),
	}
}

// createAndStartRoom creates a room, starts the game, and returns (roomCode, creatorPlayerID).
func createAndStartRoom(t *testing.T, gm *GameManager, nickname string) (string, string) {
	t.Helper()

	// Create room
	body := map[string]any{
		"nickname":  nickname,
		"mode":      "progressif",
		"wordCount": 3,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/multiplayer/create", bytes.NewReader(b))
	w := httptest.NewRecorder()
	gm.CreateRoomHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create room failed: %s", w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	roomCode := resp["roomCode"].(string)
	creatorID := resp["playerID"].(string)

	// Start game
	startBody, _ := json.Marshal(map[string]any{"roomCode": roomCode, "playerID": creatorID})
	req2 := httptest.NewRequest(http.MethodPost, "/api/multiplayer/start", bytes.NewReader(startBody))
	w2 := httptest.NewRecorder()
	gm.StartGameHandler(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("start game failed: %s", w2.Body.String())
	}

	return roomCode, creatorID
}

func TestJoinRoomHandler_MidGame_BroadcastsRankings(t *testing.T) {
	setupMultiTestDicts(t)
	gm := newTestGameManager()

	roomCode, _ := createAndStartRoom(t, gm, "Alice")

	// Subscribe a listener to the hub
	ch := gm.multi.hub.Subscribe(roomCode, "listener")
	defer gm.multi.hub.Unsubscribe(roomCode, "listener")

	// Drain any previous events (game-started etc.)
	drainCh:
	for {
		select {
		case <-ch:
		default:
			break drainCh
		}
	}

	// New player joins mid-game
	joinBody, _ := json.Marshal(map[string]any{
		"roomCode": roomCode,
		"playerID": "joiner-1",
		"nickname": "Bob",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/multiplayer/join", bytes.NewReader(joinBody))
	w := httptest.NewRecorder()
	gm.JoinRoomHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("join room failed: %s", w.Body.String())
	}

	// Receive the player-joined event
	evt := <-ch
	if evt.Event != "player-joined" {
		t.Fatalf("expected player-joined event, got %s", evt.Event)
	}

	data, ok := evt.Data.(map[string]any)
	if !ok {
		t.Fatal("event Data is not map[string]any")
	}

	rankingsRaw, ok := data["rankings"]
	if !ok {
		t.Fatal("player-joined broadcast missing 'rankings' field when room is playing")
	}

	rankings, ok := rankingsRaw.([]game.RankingEntry)
	if !ok {
		t.Fatalf("rankings field has unexpected type %T", rankingsRaw)
	}

	// Find the joiner in rankings
	found := false
	for _, r := range rankings {
		if r.PlayerID == "joiner-1" {
			found = true
			if r.Finished {
				t.Error("late joiner should not be Finished")
			}
			if r.Failed {
				t.Error("late joiner should not be Failed")
			}
			if r.WordResults != nil {
				t.Error("WordResults should be nil (stripped)")
			}
		}
	}
	if !found {
		t.Error("late joiner not found in rankings")
	}
}

func TestJoinRoomHandler_Lobby_NoRankingsField(t *testing.T) {
	setupMultiTestDicts(t)
	gm := newTestGameManager()

	// Create room (don't start it — stays in lobby)
	body := map[string]any{
		"nickname":  "Alice",
		"mode":      "progressif",
		"wordCount": 3,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/multiplayer/create", bytes.NewReader(b))
	w := httptest.NewRecorder()
	gm.CreateRoomHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create room failed: %s", w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	roomCode := resp["roomCode"].(string)

	ch := gm.multi.hub.Subscribe(roomCode, "listener")
	defer gm.multi.hub.Unsubscribe(roomCode, "listener")

	// Drain prior events
	drainLobby:
	for {
		select {
		case <-ch:
		default:
			break drainLobby
		}
	}

	joinBody, _ := json.Marshal(map[string]any{
		"roomCode": roomCode,
		"playerID": "joiner-2",
		"nickname": "Bob",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/api/multiplayer/join", bytes.NewReader(joinBody))
	w2 := httptest.NewRecorder()
	gm.JoinRoomHandler(w2, req2)

	evt := <-ch
	if evt.Event != "player-joined" {
		t.Fatalf("expected player-joined, got %s", evt.Event)
	}

	data := evt.Data.(map[string]any)
	if _, ok := data["rankings"]; ok {
		t.Error("rankings should not be present in lobby player-joined broadcast")
	}
}

func TestProgressBroadcast_IncludesRankings(t *testing.T) {
	setupMultiTestDicts(t)
	gm := newTestGameManager()

	roomCode, creatorID := createAndStartRoom(t, gm, "Alice")

	// Subscribe listener
	ch := gm.multi.hub.Subscribe(roomCode, "listener")
	defer gm.multi.hub.Unsubscribe(roomCode, "listener")

	// Drain prior events
	drainProgress:
	for {
		select {
		case <-ch:
		default:
			break drainProgress
		}
	}

	// Make a guess to trigger progress broadcast
	// Get the word sequence from the room
	gm.multi.mu.RLock()
	room := gm.multi.rooms[roomCode]
	firstWord := room.WordSequence[0]
	gm.multi.mu.RUnlock()

	// Submit a valid wrong guess — use a word from the dict that isn't the target
	candidates := []string{"ABRITE", "ACCORD", "ACTION"}
	var wrongGuess string
	for _, c := range candidates {
		if c != firstWord {
			wrongGuess = c
			break
		}
	}
	if wrongGuess == "" {
		t.Fatal("could not find a valid wrong guess")
	}

	guessBody, _ := json.Marshal(map[string]any{
		"roomCode": roomCode,
		"playerID": creatorID,
		"word":     wrongGuess,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/multiplayer/guess", bytes.NewReader(guessBody))
	w := httptest.NewRecorder()
	gm.MultiGuessHandler(w, req)
	// We don't care about the response status for invalid guess

	// Receive events until we find the progress event (or exhaust buffered ones)
	var progressEvt *SSEEvent
	for len(ch) > 0 || progressEvt == nil {
		select {
		case evt := <-ch:
			if evt.Event == "progress" {
				e := evt
				progressEvt = &e
			}
		default:
			break
		}
		if progressEvt != nil {
			break
		}
		if len(ch) == 0 {
			break
		}
	}

	if progressEvt == nil {
		t.Fatal("no progress event received")
	}

	data, ok := progressEvt.Data.(map[string]any)
	if !ok {
		t.Fatal("progress event Data is not map[string]any")
	}

	rankingsRaw, ok := data["rankings"]
	if !ok {
		t.Fatal("progress broadcast missing 'rankings' field")
	}

	rankings, ok := rankingsRaw.([]game.RankingEntry)
	if !ok {
		t.Fatalf("rankings has unexpected type %T", rankingsRaw)
	}

	for _, r := range rankings {
		if r.WordResults != nil {
			t.Errorf("WordResults should be nil on progress broadcast for player %s", r.PlayerID)
		}
	}
}
