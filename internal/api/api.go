package api

import (
	"encoding/json"
	"fmt"
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
	mux.HandleFunc("/badge.svg", s.badge)
	mux.HandleFunc("/badge/", s.badgeTarget)
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

// Overall status badge: /badge.svg
func (s *Server) badge(w http.ResponseWriter, _ *http.Request) {
	summaries, err := s.Store.Summaries()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	overall := "operational"
	color := "#10b981" // green
	for _, sum := range summaries {
		if !sum.Up {
			overall = "degraded"
			color = "#f59e0b" // amber
			break
		}
	}
	if len(summaries) == 0 {
		overall = "unknown"
		color = "#6b7280"
	}
	renderBadge(w, "status", overall, color)
}

// Per-target badge: /badge/Website.svg or /badge/Website
func (s *Server) badgeTarget(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/badge/"):]
	name = stringsTrimSuffix(name, ".svg")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	summaries, err := s.Store.Summaries()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var found *storage.Summary
	for i := range summaries {
		if summaries[i].Target == name {
			found = &summaries[i]
			break
		}
	}
	if found == nil {
		renderBadge(w, name, "unknown", "#6b7280")
		return
	}

	label := "up"
	color := "#10b981"
	if !found.Up {
		label = "down"
		color = "#ef4444"
	}
	renderBadge(w, name, label, color)
}

func renderBadge(w http.ResponseWriter, label, status, color string) {
	// Simple shields.io style SVG
	labelW := 6*len(label) + 12
	statusW := 6*len(status) + 12
	totalW := labelW + statusW

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img">
  <linearGradient id="s" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="m"><rect width="%d" height="20" rx="3" fill="#fff"/></mask>
  <g mask="url(#m)">
    <rect width="%d" height="20" fill="#555"/>
    <rect x="%d" width="%d" height="20" fill="%s"/>
    <rect width="%d" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11">
    <text x="%d" y="14">%s</text>
    <text x="%d" y="14">%s</text>
  </g>
</svg>`, totalW, totalW, labelW, labelW, statusW, color, totalW, labelW/2, label, labelW+statusW/2, status)

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(svg))
}

func stringsTrimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
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
