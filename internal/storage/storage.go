package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type Check struct {
	ID         int64     `json:"id"`
	Target     string    `json:"target"`
	Up         bool      `json:"up"`
	LatencyMs  int64     `json:"latency_ms"`
	StatusCode int       `json:"status_code"`
	Error      string    `json:"error,omitempty"`
	CheckedAt  time.Time `json:"checked_at"`
}

type Summary struct {
	Target     string  `json:"target"`
	Up         bool    `json:"up"`
	Uptime     float64 `json:"uptime"`
	AvgLatency float64 `json:"avg_latency"`
	LastCheck  string  `json:"last_check"`
	LastError  string  `json:"last_error,omitempty"`
	StatusCode int     `json:"status_code,omitempty"`
	LatencyMs  int64   `json:"latency_ms,omitempty"`
}

type HistoryPoint struct {
	Time      string `json:"time"`
	LatencyMs int64  `json:"latency_ms"`
	Up        bool   `json:"up"`
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS checks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			target TEXT NOT NULL,
			up INTEGER NOT NULL,
			latency_ms INTEGER,
			status_code INTEGER,
			error TEXT,
			checked_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_checks_target_time ON checks(target, checked_at DESC);
	`)
	return err
}

func (s *Store) Add(c Check) error {
	_, err := s.db.Exec(
		`INSERT INTO checks (target, up, latency_ms, status_code, error, checked_at) VALUES (?, ?, ?, ?, ?, ?)`,
		c.Target, boolToInt(c.Up), c.LatencyMs, c.StatusCode, c.Error, c.CheckedAt,
	)
	return err
}

func (s *Store) Summaries() ([]Summary, error) {
	rows, err := s.db.Query(`
		SELECT target, MAX(checked_at) as last_time
		FROM checks
		GROUP BY target
		ORDER BY target
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Summary
	for rows.Next() {
		var target string
		var lastTime time.Time
		if err := rows.Scan(&target, &lastTime); err != nil {
			return nil, err
		}

		sum := Summary{Target: target, LastCheck: lastTime.UTC().Format(time.RFC3339)}

		// Latest status
		var up int
		var latency int64
		var code int
		var errMsg sql.NullString
		err := s.db.QueryRow(`
			SELECT up, latency_ms, status_code, error FROM checks
			WHERE target = ? ORDER BY checked_at DESC LIMIT 1
		`, target).Scan(&up, &latency, &code, &errMsg)
		if err == nil {
			sum.Up = up == 1
			sum.StatusCode = code
			sum.LatencyMs = latency
			if errMsg.Valid {
				sum.LastError = errMsg.String
			}
		}

		// Uptime last 24h
		var total, ups sql.NullInt64
		s.db.QueryRow(`
			SELECT COUNT(*), COALESCE(SUM(up), 0) FROM checks
			WHERE target = ? AND checked_at > datetime('now', '-24 hours')
		`, target).Scan(&total, &ups)
		if total.Valid && total.Int64 > 0 {
			sum.Uptime = float64(ups.Int64) / float64(total.Int64) * 100
		}

		// Avg latency last 24h (only successful checks)
		var avg sql.NullFloat64
		s.db.QueryRow(`
			SELECT AVG(latency_ms) FROM checks
			WHERE target = ? AND up = 1 AND checked_at > datetime('now', '-24 hours')
		`, target).Scan(&avg)
		if avg.Valid {
			sum.AvgLatency = avg.Float64
		}

		results = append(results, sum)
	}
	return results, nil
}

func (s *Store) Recent(limit int) ([]Check, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`
		SELECT id, target, up, latency_ms, status_code, error, checked_at
		FROM checks ORDER BY checked_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []Check
	for rows.Next() {
		var c Check
		var up int
		var errMsg sql.NullString
		if err := rows.Scan(&c.ID, &c.Target, &up, &c.LatencyMs, &c.StatusCode, &errMsg, &c.CheckedAt); err != nil {
			return nil, err
		}
		c.Up = up == 1
		if errMsg.Valid {
			c.Error = errMsg.String
		}
		checks = append(checks, c)
	}
	return checks, nil
}

func (s *Store) Incidents(limit int) ([]Check, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, target, up, latency_ms, status_code, error, checked_at
		FROM checks WHERE up = 0 ORDER BY checked_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []Check
	for rows.Next() {
		var c Check
		var up int
		var errMsg sql.NullString
		if err := rows.Scan(&c.ID, &c.Target, &up, &c.LatencyMs, &c.StatusCode, &errMsg, &c.CheckedAt); err != nil {
			return nil, err
		}
		c.Up = false
		if errMsg.Valid {
			c.Error = errMsg.String
		}
		checks = append(checks, c)
	}
	return checks, nil
}

// History returns recent latency points for a target (for charts)
func (s *Store) History(target string, hours int) ([]HistoryPoint, error) {
	if hours <= 0 {
		hours = 24
	}
	if hours > 168 {
		hours = 168 // max 7 days
	}
	modifier := fmt.Sprintf("-%d hours", hours)
	rows, err := s.db.Query(`
		SELECT checked_at, latency_ms, up
		FROM checks
		WHERE target = ? AND checked_at > datetime('now', ?)
		ORDER BY checked_at ASC
	`, target, modifier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []HistoryPoint
	for rows.Next() {
		var t time.Time
		var latency int64
		var up int
		if err := rows.Scan(&t, &latency, &up); err != nil {
			return nil, err
		}
		points = append(points, HistoryPoint{
			Time:      t.UTC().Format(time.RFC3339),
			LatencyMs: latency,
			Up:        up == 1,
		})
	}
	return points, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
