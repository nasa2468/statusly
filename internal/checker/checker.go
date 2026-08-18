package checker

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Result struct {
	Up         bool
	LatencyMs  int64
	StatusCode int
	Error      string
}

func HTTP(ctx context.Context, address string, timeout time.Duration) Result {
	client := &http.Client{Timeout: timeout}
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return Result{Error: err.Error()}
	}

	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return Result{LatencyMs: latency, Error: err.Error()}
	}
	defer resp.Body.Close()

	// Consider 2xx-4xx as "up" (service is responding)
	up := resp.StatusCode >= 200 && resp.StatusCode < 500
	return Result{
		Up:         up,
		LatencyMs:  latency,
		StatusCode: resp.StatusCode,
	}
}

func TCP(ctx context.Context, address string, timeout time.Duration) Result {
	start := time.Now()
	d := net.Dialer{Timeout: timeout}

	conn, err := d.DialContext(ctx, "tcp", address)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return Result{LatencyMs: latency, Error: err.Error()}
	}
	conn.Close()
	return Result{Up: true, LatencyMs: latency}
}
