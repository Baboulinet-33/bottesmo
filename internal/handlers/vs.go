package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"tusmo/internal/dictionary"
	"tusmo/internal/game"
)

type SSEEvent struct {
	Event string
	Data  any
}

type SSEHub struct {
	mu      sync.Mutex
	clients map[string]chan SSEEvent
}

type VSBroker struct {
	mu   sync.Mutex
	hubs map[string]*SSEHub
}

var (
	vsRooms  = make(map[string]*game.VSRoom)
	vsRoomMu sync.Mutex
	broker   = &VSBroker{hubs: make(map[string]*SSEHub)}
)

const roomCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateRoomCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = roomCodeChars[rand.Intn(len(roomCodeChars))]
	}
	return string(b)
}

func (b *VSBroker) Subscribe(roomID, playerID string) chan SSEEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	hub, ok := b.hubs[roomID]
	if !ok {
		hub = &SSEHub{clients: make(map[string]chan SSEEvent)}
		b.hubs[roomID] = hub
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()

	if ch, ok := hub.clients[playerID]; ok {
		close(ch)
	}

	ch := make(chan SSEEvent, 64)
	hub.clients[playerID] = ch
	return ch
}

func (b *VSBroker) Unsubscribe(roomID, playerID string) {
	b.mu.Lock()
	hub, ok := b.hubs[roomID]
	b.mu.Unlock()

	if !ok {
		return
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()

	if ch, ok := hub.clients[playerID]; ok {
		close(ch)
		delete(hub.clients, playerID)
	}

	if len(hub.clients) == 0 {
		b.mu.Lock()
		delete(b.hubs, roomID)
		b.mu.Unlock()
	}
}

func (b *VSBroker) Broadcast(roomID string, event SSEEvent) {
	b.mu.Lock()
	hub, ok := b.hubs[roomID]
	b.mu.Unlock()

	if !ok {
		return
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()

	for _, ch := range hub.clients {
		select {
		case ch <- event:
		default:
		}
	}
}

func (m *GameManager) VSPageHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "layout.html", map[string]any{
		"Page": "vs",
	})
}

func (m *GameManager) VSJoinPageHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	templates.ExecuteTemplate(w, "layout.html", map[string]any{
		"Page": "vs",
		"Code": code,
	})
}

type createRoomRequest struct {
	LengthMode string `json:"lengthMode"`
	WordCount  int    `json:"wordCount"`
}

type createRoomResponse struct {
	RoomID   string `json:"roomId"`
	Code     string `json:"code"`
	ShareURL string `json:"shareUrl"`
}

func (m *GameManager) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	mode := game.VSWordLengthMode(req.LengthMode)
	if mode != game.VSModeProgressive && mode != game.VSModeRandom {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid length mode"})
		return
	}

	if req.WordCount < 3 || req.WordCount > 10 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "word count must be between 3 and 10"})
		return
	}

	roomID := m.generateID()
	code := generateRoomCode()
	creatorID := m.generateID()

	room := game.NewVSRoom(roomID, code, creatorID, req.WordCount, mode)

	vsRoomMu.Lock()
	vsRooms[roomID] = room
	vsRoomMu.Unlock()

	writeJSON(w, http.StatusOK, createRoomResponse{
		RoomID:   roomID,
		Code:     code,
		ShareURL: fmt.Sprintf("/vs/join?code=%s", code),
	})
}

type joinRoomRequest struct {
	Code     string `json:"code"`
	Nickname string `json:"nickname"`
}

type joinRoomResponse struct {
	RoomID    string       `json:"roomId"`
	PlayerID  string       `json:"playerId"`
	IsCreator bool         `json:"isCreator"`
	Players   []playerInfo `json:"players"`
}

type playerInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

func (m *GameManager) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if req.Nickname == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nickname required"})
		return
	}

	if len(req.Nickname) > 20 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nickname too long"})
		return
	}

	vsRoomMu.Lock()
	var room *game.VSRoom
	for _, r := range vsRooms {
		if r.Code == req.Code && !r.Started {
			room = r
			break
		}
	}
	vsRoomMu.Unlock()

	if room == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found or already started"})
		return
	}

	playerID := m.generateID()
	if err := game.AddPlayer(room, playerID, req.Nickname); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	broker.Broadcast(room.ID, SSEEvent{
		Event: "player-joined",
		Data: map[string]any{
			"playerId":    playerID,
			"nickname":    req.Nickname,
			"playerCount": len(room.Players),
		},
	})

	players := make([]playerInfo, 0, len(room.PlayerOrder))
	for _, pid := range room.PlayerOrder {
		p := room.Players[pid]
		players = append(players, playerInfo{ID: p.ID, Nickname: p.Nickname})
	}

	isCreator := playerID == room.CreatorID

	writeJSON(w, http.StatusOK, joinRoomResponse{
		RoomID:    room.ID,
		PlayerID:  playerID,
		IsCreator: isCreator,
		Players:   players,
	})
}

type startGameRequest struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
}

type startGameResponse struct {
	Words   []string     `json:"words"`
	Players []playerInfo `json:"players"`
}

func (m *GameManager) StartGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req startGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	vsRoomMu.Lock()
	room, ok := vsRooms[req.RoomID]
	vsRoomMu.Unlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	if req.PlayerID != room.CreatorID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only the creator can start the game"})
		return
	}

	lengths := game.GenerateWordSequence(room.LengthMode, room.WordCount)
	words := make([]string, len(lengths))
	for i, l := range lengths {
		word, err := dictionary.RandomWord(l)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate words"})
			return
		}
		words[i] = word
	}

	if err := game.StartGame(room, words); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	players := make([]playerInfo, 0, len(room.PlayerOrder))
	for _, pid := range room.PlayerOrder {
		p := room.Players[pid]
		players = append(players, playerInfo{ID: p.ID, Nickname: p.Nickname})
	}

	broker.Broadcast(room.ID, SSEEvent{
		Event: "game-started",
		Data: map[string]any{
			"words":   words,
			"players": players,
		},
	})

	writeJSON(w, http.StatusOK, startGameResponse{
		Words:   words,
		Players: players,
	})
}

type vsGuessRequest struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
	Word     string `json:"word"`
}

type vsGuessResponse struct {
	Results      []game.LetterResult `json:"results"`
	WordComplete bool                `json:"wordComplete"`
	AllComplete  bool                `json:"allComplete"`
	Won          bool                `json:"won"`
	GameOver     bool                `json:"gameOver"`
	CurrentWord  int                 `json:"currentWord"`
	Failed       bool                `json:"failed"`
	Word         string              `json:"word,omitempty"`
}

func (m *GameManager) VSGuessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req vsGuessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	vsRoomMu.Lock()
	room, ok := vsRooms[req.RoomID]
	vsRoomMu.Unlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	if !dictionary.IsValid(req.Word) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Mot invalide"})
		return
	}

	results, wordComplete, allComplete, currentWord, err := game.ProcessGuess(room, req.PlayerID, req.Word)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	won := allComplete && wordComplete
	failed := allComplete && !wordComplete

	var targetWord string
	if wordComplete {
		idx := currentWord - 1
		if idx >= 0 && idx < len(room.Words) {
			targetWord = room.Words[idx]
		}
	}

	progressData := map[string]any{
		"playerId":    req.PlayerID,
		"currentWord": currentWord,
		"wordComplete": wordComplete,
		"allComplete":  allComplete,
		"won":          won,
		"failed":       failed,
	}
	broker.Broadcast(room.ID, SSEEvent{Event: "progress", Data: progressData})

	if allComplete && game.IsRoomFinished(room) {
		game.SetRoomFinished(room)
		rankings := game.GetRankings(room)
		broker.Broadcast(room.ID, SSEEvent{Event: "game-over", Data: map[string]any{"rankings": rankings}})
	}

	resp := vsGuessResponse{
		Results:      results,
		WordComplete: wordComplete,
		AllComplete:  allComplete,
		Won:          won,
		Failed:       failed,
		CurrentWord:  currentWord,
		Word:         targetWord,
	}

	if allComplete {
		resp.GameOver = true
	}

	writeJSON(w, http.StatusOK, resp)
}

func (m *GameManager) SSEHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomId")
	playerID := r.URL.Query().Get("playerId")

	if roomID == "" || playerID == "" {
		http.Error(w, "roomId and playerId required", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := broker.Subscribe(roomID, playerID)
	defer broker.Unsubscribe(roomID, playerID)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(evt.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Event, string(data))
			flusher.Flush()
		}
	}
}

type rematchRequest struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
}

type rematchResponse struct {
	RoomID    string            `json:"roomId"`
	Code      string            `json:"code"`
	ShareURL  string            `json:"shareUrl"`
	PlayerIDs map[string]string `json:"playerIds"`
}

func (m *GameManager) RematchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req rematchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	vsRoomMu.Lock()
	oldRoom, ok := vsRooms[req.RoomID]
	vsRoomMu.Unlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	if req.PlayerID != oldRoom.CreatorID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only the creator can rematch"})
		return
	}

	newRoomID := m.generateID()
	newCode := generateRoomCode()
	creatorID := m.generateID()

	wordCount := oldRoom.WordCount
	lengthMode := oldRoom.LengthMode

	lengths := game.GenerateWordSequence(lengthMode, wordCount)
	words := make([]string, len(lengths))
	for i, l := range lengths {
		word, err := dictionary.RandomWord(l)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate words"})
			return
		}
		words[i] = word
	}

	newRoom := game.NewVSRoom(newRoomID, newCode, creatorID, wordCount, lengthMode)

	oldToNewIDs := make(map[string]string)

	for _, pid := range oldRoom.PlayerOrder {
		p := oldRoom.Players[pid]
		newPID := m.generateID()
		isCreator := pid == oldRoom.CreatorID
		if isCreator {
			newRoom.CreatorID = newPID
			oldToNewIDs[pid] = newPID
		} else {
			oldToNewIDs[pid] = newPID
		}
		game.AddPlayer(newRoom, newPID, p.Nickname)
	}

	game.StartGame(newRoom, words)

	vsRoomMu.Lock()
	vsRooms[newRoomID] = newRoom
	vsRoomMu.Unlock()

	broker.Broadcast(req.RoomID, SSEEvent{
		Event: "rematch",
		Data: map[string]any{
			"roomId":    newRoomID,
			"code":      newCode,
			"shareUrl":  fmt.Sprintf("/vs/join?code=%s", newCode),
			"playerIds": oldToNewIDs,
		},
	})

	writeJSON(w, http.StatusOK, rematchResponse{
		RoomID:    newRoomID,
		Code:      newCode,
		ShareURL:  fmt.Sprintf("/vs/join?code=%s", newCode),
		PlayerIDs: oldToNewIDs,
	})
}

func CleanupVSRooms() {
	for {
		time.Sleep(5 * time.Minute)
		vsRoomMu.Lock()
		for id, room := range vsRooms {
			idle := time.Since(room.CreatedAt) > time.Hour
			finished := room.Finished
			if idle || finished {
				delete(vsRooms, id)
			}
		}
		vsRoomMu.Unlock()
	}
}
