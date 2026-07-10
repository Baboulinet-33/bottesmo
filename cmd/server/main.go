package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"bottesmo/internal/dictionary"
	"bottesmo/internal/handlers"
	"bottesmo/internal/version"
)

const defaultDrainTimeout = 30 * time.Second

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

	ctx, shutdownCancel := context.WithCancel(context.Background())

	go handlers.CleanupSessions(ctx)
	go mgr.MultiplayerManager().CleanupRooms(ctx)

	fs := http.FileServer(http.Dir(filepath.Join(wd, "web", "static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", mgr.HomeHandler)
	http.HandleFunc("/game", mgr.GamePageHandler)
	http.HandleFunc("/api/status", mgr.StatusHandler)
	http.HandleFunc("/api/game/new", mgr.NewGameHandler)
	http.HandleFunc("/api/game/guess", mgr.GuessHandler)

	http.HandleFunc("/settings", mgr.SettingsPageHandler)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Bottesmo starting on :%s (Go %s)", port, runtime.Version())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	drainTimeout := getDrainTimeout()
	drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
	defer cancel()

	shutdownCancel()

	if err := srv.Shutdown(drainCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}

func getDrainTimeout() time.Duration {
	if v := os.Getenv("DRAIN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultDrainTimeout
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
