package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/nasa2468/statusly/internal/api"
	"github.com/nasa2468/statusly/internal/checker"
	"github.com/nasa2468/statusly/internal/config"
	"github.com/nasa2468/statusly/internal/storage"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	store, err := storage.Open(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	// Start monitors
	for _, t := range cfg.Targets {
		go monitor(t, store)
	}

	mux := http.NewServeMux()

	// Static frontend
	mux.Handle("/", http.FileServer(http.Dir("web")))

	// API
	apiServer := &api.Server{Store: store, Config: cfg}
	apiServer.Register(mux)

	log.Printf("Statusly listening on %s", cfg.Server.Address)
	log.Fatal(http.ListenAndServe(cfg.Server.Address, mux))
}

func monitor(t config.Target, store *storage.Store) {
	interval := time.Duration(t.Interval) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}
	timeout := time.Duration(t.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		var r checker.Result

		switch t.Type {
		case "tcp":
			r = checker.TCP(ctx, t.Address, timeout)
		default:
			r = checker.HTTP(ctx, t.Address, timeout)
		}
		cancel()

		c := storage.Check{
			Target:     t.Name,
			Up:         r.Up,
			LatencyMs:  r.LatencyMs,
			StatusCode: r.StatusCode,
			Error:      r.Error,
			CheckedAt:  time.Now().UTC(),
		}

		if err := store.Add(c); err != nil {
			log.Printf("store %s: %v", t.Name, err)
		} else {
			api.RecordMetrics(c)
		}

		time.Sleep(interval)
	}
}
