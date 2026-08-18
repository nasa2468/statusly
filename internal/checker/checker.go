package checker

import (
	"context"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
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

	// 2xx-4xx considered up (service is responding)
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

// ICMP uses system ping command for simplicity and fewer permission issues.
// Requires `ping` to be available on the system.
func ICMP(ctx context.Context, address string, timeout time.Duration) Result {
	start := time.Now()

	// Extract host (strip port if present)
	host := address
	if h, _, err := net.SplitHostPort(address); err == nil {
		host = h
	}

	countFlag := "-c"
	timeoutFlag := "-W"
	if runtime.GOOS == "windows" {
		countFlag = "-n"
		timeoutFlag = "-w"
	}

	// timeout in seconds for ping
	sec := int(timeout.Seconds())
	if sec < 1 {
		sec = 1
	}

	args := []string{countFlag, "1", timeoutFlag, strconv.Itoa(sec), host}
	cmd := exec.CommandContext(ctx, "ping", args...)
	out, err := cmd.CombinedOutput()
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return Result{LatencyMs: latency, Error: strings.TrimSpace(string(out))}
	}
	return Result{Up: true, LatencyMs: latency}
}
