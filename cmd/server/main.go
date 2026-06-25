package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"tusmo/internal/dictionary"
	"tusmo/internal/handlers"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dictPath := filepath.Join(wd, "internal", "dictionary", "words.txt")
	if err := dictionary.Load(dictPath); err != nil {
		log.Fatal(err)
	}

	tmplPattern := filepath.Join(wd, "web", "templates", "*.html")
	if err := handlers.LoadTemplates(tmplPattern); err != nil {
		log.Fatal(err)
	}

	mgr := handlers.NewGameManager()

	go handlers.CleanupSessions()

	fs := http.FileServer(http.Dir(filepath.Join(wd, "web", "static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", mgr.HomeHandler)
	http.HandleFunc("/game", mgr.GamePageHandler)
	http.HandleFunc("/api/game/new", mgr.NewGameHandler)
	http.HandleFunc("/api/game/guess", mgr.GuessHandler)

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

	buildInfo, _ := debug.ReadBuildInfo()
	log.Printf("Tusmo starting on :%s (Go %s)", port, buildInfo.GoVersion)
	log.Fatal(srv.ListenAndServe())
}
