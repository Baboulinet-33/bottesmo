package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"bottesmo/internal/dictionary"
	"bottesmo/internal/handlers"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dictSource := os.Getenv("DICT_WORDS_SOURCE")
	if dictSource == "" {
		dictSource = filepath.Join(wd, "internal", "dictionary", "words.txt")
	}
	if _, err := dictionary.LoadFromSource(dictSource); err != nil {
		log.Fatal(err)
	}

	fullDictSource := os.Getenv("DICT_WORDS_FULL_SOURCE")
	if fullDictSource == "" {
		fullDictSource = filepath.Join(wd, "internal", "dictionary", "words_full.txt")
	}
	if _, err := dictionary.LoadFullFromSource(fullDictSource); err != nil {
		log.Fatal(err)
	}

	tmplPattern := filepath.Join(wd, "web", "templates", "*.html")
	if err := handlers.LoadTemplates(tmplPattern); err != nil {
		log.Fatal(err)
	}

	mgr := handlers.NewGameManager()

	go handlers.CleanupSessions()
	go mgr.MultiplayerManager().CleanupRooms()

	fs := http.FileServer(http.Dir(filepath.Join(wd, "web", "static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", mgr.HomeHandler)
	http.HandleFunc("/game", mgr.GamePageHandler)
	http.HandleFunc("/api/game/new", mgr.NewGameHandler)
	http.HandleFunc("/api/game/guess", mgr.GuessHandler)

	http.HandleFunc("/multiplayer", mgr.MultiplayerPageHandler)
	http.HandleFunc("/api/multiplayer/create", mgr.CreateRoomHandler)
	http.HandleFunc("/api/multiplayer/join", mgr.JoinRoomHandler)
	http.HandleFunc("/api/multiplayer/start", mgr.StartGameHandler)
	http.HandleFunc("/api/multiplayer/guess", mgr.MultiGuessHandler)
	http.HandleFunc("/api/multiplayer/events", mgr.SSEHandler)
	http.HandleFunc("/api/multiplayer/leave", mgr.LeaveRoomHandler)
	http.HandleFunc("/api/multiplayer/restart", mgr.RestartGameHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3102"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Bottesmo starting on :%s (Go %s)", port, runtime.Version())
	log.Fatal(srv.ListenAndServe())
}
