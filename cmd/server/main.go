package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"bottesmo/internal/dictionary"
	"bottesmo/internal/handlers"
	"bottesmo/internal/version"
)

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("bottesmo v%s\n", version.Version)
		os.Exit(0)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	loadDict("DICT_WORDS_SOURCE", filepath.Join(wd, "internal", "dictionary", "words.txt"), dictionary.LoadFromSource)
	loadDict("DICT_WORDS_FULL_SOURCE", filepath.Join(wd, "internal", "dictionary", "words_full.txt"), dictionary.LoadFullFromSource)

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
	http.HandleFunc("/api/status", mgr.StatusHandler)
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

func loadDict(envName, defaultPath string, loader func(string) (dictionary.LoadResult, error)) {
	source := os.Getenv(envName)
	if source == "" {
		source = defaultPath
	}
	if _, err := loader(source); err != nil {
		log.Fatal(err)
	}
}
