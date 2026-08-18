package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/nasa2468/statusly/internal/api"
	"github.com/nasa2468/statusly/internal/checker"
	"github.com/nasa2468/statusly/internal/config"
	"github.com/nasa2468/statusly/internal/notify"
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

	notifier := notify.New(cfg.Notifications)

	// Track previous state for change detection
	var stateMu sync.Mutex
	prevState := make(map[string]bool)

	// Start monitors
	for _, t := range cfg.Targets {
		go monitor(t, store, notifier, &stateMu, prevState)
	}

	mux := http.NewServeMux()

	// Static frontend
	mux.Handle("/", http.FileServer(http.Dir("web")))

	// API + Badge
	apiServer := &api.Server{Store: store, Config: cfg}
	apiServer.Register(mux)

	log.Printf("Statusly listening on %s", cfg.Server.Address)
	log.Fatal(http.ListenAndServe(cfg.Server.Address, mux))
}

func monitor(t config.Target, store *storage.Store, notifier *notify.Notifier, mu *sync.Mutex, prev map[string]bool) {
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
		case "icmp", "ping":
			r = checker.ICMP(ctx, t.Address, timeout)
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

			// Notify on state change
			mu.Lock()
			wasUp, seen := prev[t.Name]
			if !seen {
				// first check, just record state
				prev[t.Name] = c.Up
			} else {
				notifier.NotifyStateChange(wasUp, c)
				prev[t.Name] = c.Up
			}
			mu.Unlock()
		}

		time.Sleep(interval)
	}
}
