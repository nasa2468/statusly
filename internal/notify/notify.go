package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nasa2468/statusly/internal/config"
	"github.com/nasa2468/statusly/internal/storage"
)

type Notifier struct {
	configs []config.Notification
	client  *http.Client
}

func New(cfgs []config.Notification) *Notifier {
	return &Notifier{
		configs: cfgs,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// NotifyStateChange sends alerts when a target goes down or recovers.
func (n *Notifier) NotifyStateChange(prevUp bool, c storage.Check) {
	if len(n.configs) == 0 || prevUp == c.Up {
		return
	}

	var title, body string
	if c.Up {
		title = "Recovered"
		body = fmt.Sprintf("%s is back online\nLatency: %d ms", c.Target, c.LatencyMs)
	} else {
		title = "Down"
		errMsg := c.Error
		if errMsg == "" {
			errMsg = "Service unavailable"
		}
		body = fmt.Sprintf("%s is down\nError: %s", c.Target, errMsg)
	}

	msg := fmt.Sprintf("%s\n%s\nTime: %s", title, body, c.CheckedAt.Format(time.RFC3339))

	for _, cfg := range n.configs {
		if !cfg.Enabled {
			continue
		}
		switch strings.ToLower(cfg.Type) {
		case "telegram":
			go sendTelegram(n.client, cfg, msg)
		case "discord":
			go sendDiscord(n.client, cfg, title, body)
		case "webhook":
			go sendWebhook(n.client, cfg, c)
		}
	}
}

func sendTelegram(client *http.Client, cfg config.Notification, text string) {
	if cfg.Token == "" || cfg.ChatID == "" {
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.Token)
	payload := map[string]string{
		"chat_id":    cfg.ChatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	doJSON(client, url, payload)
}

func sendDiscord(client *http.Client, cfg config.Notification, title, description string) {
	if cfg.URL == "" {
		return
	}
	payload := map[string]any{
		"embeds": []map[string]any{{
			"title":       title,
			"description": description,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		}},
	}
	doJSON(client, cfg.URL, payload)
}

func sendWebhook(client *http.Client, cfg config.Notification, c storage.Check) {
	if cfg.URL == "" {
		return
	}
	payload := map[string]any{
		"target":      c.Target,
		"up":          c.Up,
		"latency_ms":  c.LatencyMs,
		"status_code": c.StatusCode,
		"error":       c.Error,
		"checked_at":  c.CheckedAt.Format(time.RFC3339),
	}
	doJSON(client, cfg.URL, payload)
}

func doJSON(client *http.Client, url string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
