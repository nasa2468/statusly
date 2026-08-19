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

	// Track previous state for change detection. Notification delivery is
	// asynchronous, so monitoring never waits on a remote webhook.
	var stateMu sync.Mutex
	prevState := make(map[string]bool)

	for _, t := range cfg.Targets {
		go monitor(t, store, notifier, &stateMu, prevState)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("web")))

	apiServer := &api.Server{Store: store, Config: cfg}
	apiServer.Register(mux)

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Statusly listening on %s", cfg.Server.Address)
	log.Fatal(server.ListenAndServe())
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

			mu.Lock()
			wasUp, seen := prev[t.Name]
			if !seen {
				prev[t.Name] = c.Up
				mu.Unlock()
			} else {
				prev[t.Name] = c.Up
				mu.Unlock()
				// Never hold the shared state lock while a notifier is running.
				notifier.NotifyStateChange(wasUp, c)
			}
		}

		time.Sleep(interval)
	}
}
