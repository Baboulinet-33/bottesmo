package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"tusmo/internal/dictionary"
	"tusmo/internal/game"
)

type SSEEvent struct {
	Event string
	Data  interface{}
}

type MultiplayerHub struct {
	mu    sync.RWMutex
	rooms map[string]map[string]chan SSEEvent
}

func NewMultiplayerHub() *MultiplayerHub {
	return &MultiplayerHub{
		rooms: make(map[string]map[string]chan SSEEvent),
	}
}

func (h *MultiplayerHub) Subscribe(roomCode, playerID string) chan SSEEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[roomCode] == nil {
		h.rooms[roomCode] = make(map[string]chan SSEEvent)
	}

	ch := make(chan SSEEvent, 20)
	h.rooms[roomCode][playerID] = ch
	return ch
}

func (h *MultiplayerHub) Unsubscribe(roomCode, playerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[roomCode] != nil {
		delete(h.rooms[roomCode], playerID)
		if len(h.rooms[roomCode]) == 0 {
			delete(h.rooms, roomCode)
		}
	}
}

func (h *MultiplayerHub) Broadcast(roomCode string, event SSEEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.rooms[roomCode] == nil {
		return
	}

	for _, ch := range h.rooms[roomCode] {
		select {
		case ch <- event:
		default:
		}
	}
}

type playerInfo struct {
	ID             string `json:"id"`
	Nickname       string `json:"nickname"`
	Finished       bool   `json:"finished"`
	Failed         bool   `json:"failed"`
	CurrentWordIdx int    `json:"currentWordIdx"`
	TotalWords     int    `json:"totalWords"`
}

type wordGameState struct {
	Target     string            `json:"target"`
	Attempts   []string          `json:"attempts"`
	MaxTries   int               `json:"maxTries"`
	Won        bool              `json:"won"`
	GameOver   bool              `json:"gameOver"`
	WordLength int               `json:"wordLength"`
	Results    [][]game.LetterResult `json:"results"`
}

type MultiplayerManager struct {
	mu    sync.RWMutex
	rooms map[string]*game.MultiplayerRoom
	hub   *MultiplayerHub
}

func NewMultiplayerManager() *MultiplayerManager {
	return &MultiplayerManager{
		rooms: make(map[string]*game.MultiplayerRoom),
		hub:   NewMultiplayerHub(),
	}
}

func (mm *MultiplayerManager) CleanupRooms() {
	for {
		time.Sleep(5 * time.Minute)
		mm.mu.Lock()
		for code, room := range mm.rooms {
			if time.Since(room.CreatedAt) > time.Hour {
				delete(mm.rooms, code)
			}
		}
		mm.mu.Unlock()
	}
}

func (mm *MultiplayerManager) removePlayerFromRoom(roomCode, playerID string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	room, ok := mm.rooms[roomCode]
	if !ok {
		return
	}
	if _, exists := room.Players[playerID]; !exists {
		return
	}

	room.RemovePlayer(playerID)

	if room.CreatorID == playerID && len(room.Players) > 0 {
		for id := range room.Players {
			room.CreatorID = id
			break
		}
	}

	if len(room.Players) == 0 {
		delete(mm.rooms, roomCode)
		return
	}

	players := playersInfo(room)
	mm.hub.Broadcast(roomCode, SSEEvent{
		Event: "player-left",
		Data: map[string]any{
			"playerID":     playerID,
			"players":      players,
			"newCreatorID": room.CreatorID,
		},
	})
}

func playersInfo(room *game.MultiplayerRoom) []playerInfo {
	var list []playerInfo
	for _, p := range room.Players {
		list = append(list, playerInfo{
			ID:             p.ID,
			Nickname:       p.Nickname,
			Finished:       p.Finished,
			Failed:         p.Failed,
			CurrentWordIdx: p.CurrentWordIdx,
			TotalWords:     room.WordCount,
		})
	}
	return list
}

// --- Handlers on GameManager ---

func (m *GameManager) MultiplayerPageHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "layout.html", map[string]any{
		"Page": "multiplayer",
	})
}

type createRoomRequest struct {
	Mode      string `json:"mode"`
	WordCount int    `json:"wordCount"`
	Nickname  string `json:"nickname"`
}

type createRoomResponse struct {
	RoomCode string `json:"roomCode"`
	ShareURL string `json:"shareURL"`
	PlayerID string `json:"playerID"`
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

	if req.Mode != "progressif" && req.Mode != "aleatoire" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mode must be progressif or aleatoire"})
		return
	}

	if req.WordCount < 1 || req.WordCount > 10 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "wordCount must be between 1 and 10"})
		return
	}

	if len(req.Nickname) == 0 || len(req.Nickname) > 20 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nickname must be between 1 and 20 characters"})
		return
	}

	mm := m.multi

	var code string
	mm.mu.Lock()
	for {
		code = game.GenerateRoomCode()
		if _, exists := mm.rooms[code]; !exists {
			break
		}
	}

	playerID := m.generateID()
	room := game.NewMultiplayerRoom(code, req.Mode, req.WordCount, playerID, req.Nickname)
	mm.rooms[code] = room
	mm.mu.Unlock()

	writeJSON(w, http.StatusOK, createRoomResponse{
		RoomCode: code,
		ShareURL: "/multiplayer?join=" + code,
		PlayerID: playerID,
	})
}

type joinRoomRequest struct {
	RoomCode string `json:"roomCode"`
	Nickname string `json:"nickname"`
	PlayerID string `json:"playerID,omitempty"`
}

type joinRoomResponse struct {
	PlayerID       string          `json:"playerID"`
	RoomCode       string          `json:"roomCode"`
	Mode           string          `json:"mode"`
	WordCount      int             `json:"wordCount"`
	State          string          `json:"state"`
	CreatorID      string          `json:"creatorID"`
	Players        []playerInfo    `json:"players"`
	WordSequence   []string        `json:"wordSequence,omitempty"`
	CurrentWordIdx int             `json:"currentWordIdx,omitempty"`
	WordGames      []wordGameState `json:"wordGames,omitempty"`
	StartTime      string          `json:"startTime,omitempty"`
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

	if len(req.RoomCode) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "roomCode required"})
		return
	}

	mm := m.multi
	mm.mu.Lock()
	room, ok := mm.rooms[req.RoomCode]
	if !ok {
		mm.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	playerID := req.PlayerID
	if playerID == "" {
		playerID = m.generateID()
	}

	isNewPlayer := false
	if _, exists := room.Players[playerID]; !exists {
		isNewPlayer = true
		if err := room.AddPlayer(playerID, req.Nickname); err != nil {
			mm.mu.Unlock()
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	resp := joinRoomResponse{
		PlayerID:  playerID,
		RoomCode:  room.Code,
		Mode:      room.Mode,
		WordCount: room.WordCount,
		State:     room.State,
		CreatorID: room.CreatorID,
		Players:   playersInfo(room),
	}

	if room.State == "playing" {
		player := room.Players[playerID]
		resp.WordSequence = room.WordSequence
		resp.CurrentWordIdx = player.CurrentWordIdx
		resp.StartTime = room.StartTime.Format(time.RFC3339Nano)

		resp.WordGames = make([]wordGameState, len(player.WordGames))
		for i, g := range player.WordGames {
			wgs := wordGameState{
				Target:     g.Target,
				Attempts:   g.Attempts,
				MaxTries:   g.MaxTries,
				Won:        g.Won,
				GameOver:   g.GameOver,
				WordLength: g.WordLength,
			}
			if len(g.Attempts) > 0 {
				results, err := game.ComputeAttemptResults(g.Target, g.Attempts)
				if err == nil {
					wgs.Results = results
				}
			}
			resp.WordGames[i] = wgs
		}
	}

	mm.mu.Unlock()

	if isNewPlayer {
		mm.hub.Broadcast(room.Code, SSEEvent{
			Event: "player-joined",
			Data:  map[string]any{"players": playersInfo(room)},
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

type startGameRequest struct {
	RoomCode string `json:"roomCode"`
	PlayerID string `json:"playerID"`
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

	mm := m.multi
	mm.mu.Lock()
	room, ok := mm.rooms[req.RoomCode]
	if !ok {
		mm.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	if room.CreatorID != req.PlayerID {
		mm.mu.Unlock()
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only the creator can start the game"})
		return
	}

	if err := room.StartGame(); err != nil {
		mm.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	players := playersInfo(room)
	mm.mu.Unlock()

	mm.hub.Broadcast(room.Code, SSEEvent{
		Event: "game-started",
		Data:  map[string]any{"players": players},
	})

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

type multiGuessRequest struct {
	RoomCode string `json:"roomCode"`
	PlayerID string `json:"playerID"`
	Word     string `json:"word"`
}

type multiGuessResponse struct {
	game.MultiplayerGuessResult
	Rankings []game.RankingEntry `json:"rankings,omitempty"`
}

func (m *GameManager) MultiGuessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var req multiGuessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if !dictionary.IsValid(req.Word) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Mot invalide"})
		return
	}

	mm := m.multi
	mm.mu.Lock()
	room, ok := mm.rooms[req.RoomCode]
	if !ok {
		mm.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	multiGuessResult, err := room.ProcessGuess(req.PlayerID, req.Word)
	if err != nil {
		mm.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp := multiGuessResponse{
		MultiplayerGuessResult: *multiGuessResult,
	}

	players := playersInfo(room)
	roomFinished := room.IsRoomFinished()

	mm.mu.Unlock()

	mm.hub.Broadcast(room.Code, SSEEvent{
		Event: "progress",
		Data:  map[string]any{"players": players},
	})

	if resp.PlayerFinished {
		rankings := room.GetRankings()
		resp.Rankings = rankings
		mm.hub.Broadcast(room.Code, SSEEvent{
			Event: "player-finished",
			Data: map[string]any{
				"playerID": req.PlayerID,
				"rankings": rankings,
			},
		})
	}

	if roomFinished {
		rankings := room.GetRankings()
		mm.hub.Broadcast(room.Code, SSEEvent{
			Event: "game-over",
			Data:  map[string]any{"rankings": rankings},
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (m *GameManager) SSEHandler(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room")
	playerID := r.URL.Query().Get("player")

	if roomCode == "" || playerID == "" {
		http.Error(w, "room and player parameters required", http.StatusBadRequest)
		return
	}

	mm := m.multi
	mm.mu.RLock()
	room, ok := mm.rooms[roomCode]
	if !ok {
		mm.mu.RUnlock()
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	if _, exists := room.Players[playerID]; !exists {
		mm.mu.RUnlock()
		http.Error(w, "not a player in this room", http.StatusForbidden)
		return
	}
	mm.mu.RUnlock()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	http.NewResponseController(w).SetWriteDeadline(time.Time{})

	ch := mm.hub.Subscribe(roomCode, playerID)
	defer mm.hub.Unsubscribe(roomCode, playerID)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			mm.removePlayerFromRoom(roomCode, playerID)
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(event.Data)
			if err != nil {
				log.Printf("SSE marshal error: %v", err)
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, string(data))
			flusher.Flush()
		}
	}
}

type leaveRoomRequest struct {
	RoomCode string `json:"roomCode"`
	PlayerID string `json:"playerID"`
}

type restartGameRequest struct {
	RoomCode string `json:"roomCode"`
	PlayerID string `json:"playerID"`
}

func (m *GameManager) RestartGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var req restartGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	mm := m.multi
	mm.mu.Lock()
	room, ok := mm.rooms[req.RoomCode]
	if !ok {
		mm.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	if room.CreatorID != req.PlayerID {
		mm.mu.Unlock()
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only the creator can restart the game"})
		return
	}

	if err := room.RestartGame(); err != nil {
		mm.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	players := playersInfo(room)
	mm.mu.Unlock()

	mm.hub.Broadcast(room.Code, SSEEvent{
		Event: "game-restarted",
		Data:  map[string]any{"players": players},
	})

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (m *GameManager) LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var req leaveRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	mm := m.multi
	mm.removePlayerFromRoom(req.RoomCode, req.PlayerID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}
