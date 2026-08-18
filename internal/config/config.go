package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        Server         `yaml:"server"`
	Database      string         `yaml:"database"`
	Title         string         `yaml:"title"`
	Description   string         `yaml:"description"`
	Targets       []Target       `yaml:"targets"`
	Notifications []Notification `yaml:"notifications"`
}

type Server struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
}

type Target struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // http, tcp, icmp
	Address  string `yaml:"address"`
	Interval int    `yaml:"interval"` // seconds
	Timeout  int    `yaml:"timeout"`  // seconds
}

type Notification struct {
	Type    string `yaml:"type"` // telegram, discord, webhook
	URL     string `yaml:"url"`  // webhook url or discord webhook
	Token   string `yaml:"token"`
	ChatID  string `yaml:"chat_id"` // for telegram
	Enabled bool   `yaml:"enabled"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Address == "" {
		cfg.Server.Address = ":8080"
	}
	if cfg.Database == "" {
		cfg.Database = "data/statusly.db"
	}
	if cfg.Title == "" {
		cfg.Title = "Statusly"
	}
	return &cfg, nil
}
