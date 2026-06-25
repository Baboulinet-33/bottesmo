package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"log"
	mathrand "math/rand"
	"net/http"
	"sync"
	"time"

	"tusmo/internal/dictionary"
	"tusmo/internal/game"
)

const maxBodySize = 1 << 20      // 1 MB
const maxSessions = 100000

type session struct {
	game *game.Game
	last time.Time
}

var (
	sessions  = make(map[string]*session)
	mu        sync.Mutex
	templates *template.Template
)

type GameManager struct {
	dailyWord string
	dailyDate string
}

func NewGameManager() *GameManager {
	return &GameManager{}
}

func LoadTemplates(pattern string) error {
	var err error
	templates, err = template.ParseGlob(pattern)
	return err
}

func (m *GameManager) generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *GameManager) ensureDaily() {
	today := time.Now().UTC().Format("2006-01-02")
	if m.dailyDate != today {
		l := dictionary.DailyLength(today)
		word, err := dictionary.DailyWord(l, today)
		if err != nil {
			log.Printf("error generating daily word: %v", err)
			return
		}
		m.dailyWord = word
		m.dailyDate = today
	}
}

func (m *GameManager) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	templates.ExecuteTemplate(w, "layout.html", map[string]any{
		"Page": "home",
	})
}

func (m *GameManager) GamePageHandler(w http.ResponseWriter, r *http.Request) {
	mode := game.GameMode(r.URL.Query().Get("mode"))
	if mode != game.ModeDaily && mode != game.ModeSolo {
		http.Error(w, "invalid mode", http.StatusBadRequest)
		return
	}

	templates.ExecuteTemplate(w, "layout.html", map[string]any{
		"Page": "game",
		"Mode": mode,
	})
}

type newGameRequest struct {
	Mode string `json:"mode"`
}

type newGameResponse struct {
	ID          string `json:"id"`
	WordLength  int    `json:"wordLength"`
	FirstLetter string `json:"firstLetter"`
	MaxTries    int    `json:"maxTries"`
	Mode        string `json:"mode"`
}

func (m *GameManager) NewGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	if len(sessions) >= maxSessions {
		mu.Unlock()
		http.Error(w, "server busy", http.StatusServiceUnavailable)
		return
	}
	mu.Unlock()

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var req newGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	mode := game.GameMode(req.Mode)
	if mode != game.ModeDaily && mode != game.ModeSolo {
		http.Error(w, "invalid mode", http.StatusBadRequest)
		return
	}

	var target string

	if mode == game.ModeDaily {
		m.ensureDaily()
		target = m.dailyWord
	} else {
		minLen := dictionary.MinLength()
		maxLen := dictionary.MaxLength()
		l := minLen + mathrand.Intn(maxLen-minLen+1)
		var err error
		target, err = dictionary.RandomWord(l)
		if err != nil {
			http.Error(w, "no words available", http.StatusInternalServerError)
			return
		}
	}

	g := game.NewGame(target, mode)
	id := m.generateID()

	mu.Lock()
	sessions[id] = &session{game: g, last: time.Now()}
	mu.Unlock()

	resp := newGameResponse{
		ID:          id,
		WordLength:  len(target),
		FirstLetter: string(target[0]),
		MaxTries:    g.MaxTries,
		Mode:        string(mode),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type guessRequest struct {
	GameID string `json:"gameId"`
	Word   string `json:"word"`
}

type guessResponse struct {
	Results  []game.LetterResult `json:"results"`
	Won      bool                `json:"won"`
	GameOver bool                `json:"gameOver"`
	Attempts int                 `json:"attempts"`
	Word     string              `json:"word,omitempty"`
}

func (m *GameManager) GuessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var req guessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	sess, ok := sessions[req.GameID]
	if !ok {
		mu.Unlock()
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	sess.last = time.Now()
	results, err := sess.game.Guess(req.Word)
	won := sess.game.Won
	gameOver := sess.game.GameOver
	attempts := len(sess.game.Attempts)
	target := sess.game.Target
	mu.Unlock()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := guessResponse{
		Results:  results,
		Won:      won,
		GameOver: gameOver,
		Attempts: attempts,
	}

	if gameOver {
		resp.Word = target
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func CleanupSessions() {
	for {
		time.Sleep(5 * time.Minute)
		mu.Lock()
		for id, sess := range sessions {
			if time.Since(sess.last) > time.Hour {
				delete(sessions, id)
			}
		}
		mu.Unlock()
	}
}
