package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nasa2468/statusly/internal/config"
	"github.com/nasa2468/statusly/internal/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	Store  *storage.Store
	Config *config.Config
}

var (
	checksTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "statusly_checks_total",
		Help: "Total checks recorded by Statusly",
	}, []string{"target", "status"})

	latency = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "statusly_latency_ms",
		Help: "Latest latency in milliseconds",
	}, []string{"target"})
)

func init() {
	prometheus.MustRegister(checksTotal, latency)
}

func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/api/status", s.status)
	mux.HandleFunc("/api/incidents", s.incidents)
	mux.HandleFunc("/api/recent", s.recent)
	mux.HandleFunc("/api/history", s.history)
	mux.Handle("/metrics", promhttp.Handler())
}

func (s *Server) status(w http.ResponseWriter, _ *http.Request) {
	summaries, err := s.Store.Summaries()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	overall := "operational"
	for _, sum := range summaries {
		if !sum.Up {
			overall = "degraded"
			break
		}
	}

	resp := map[string]any{
		"title":       s.Config.Title,
		"description": s.Config.Description,
		"status":      overall,
		"targets":     summaries,
	}
	writeJSON(w, resp)
}

func (s *Server) incidents(w http.ResponseWriter, _ *http.Request) {
	items, err := s.Store.Incidents(50)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, items)
}

func (s *Server) recent(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if n <= 0 {
		n = 100
	}
	if n > 1000 {
		n = 1000
	}
	items, err := s.Store.Recent(n)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, items)
}

func (s *Server) history(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "target is required", 400)
		return
	}
	hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))
	if hours <= 0 {
		hours = 24
	}
	points, err := s.Store.History(target, hours)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, points)
}

func RecordMetrics(c storage.Check) {
	status := "down"
	if c.Up {
		status = "up"
	}
	checksTotal.WithLabelValues(c.Target, status).Inc()
	latency.WithLabelValues(c.Target).Set(float64(c.LatencyMs))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(v)
}
