package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      Server   `yaml:"server"`
	Database    string   `yaml:"database"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Targets     []Target `yaml:"targets"`
}

type Server struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
}

type Target struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // http or tcp
	Address  string `yaml:"address"`
	Interval int    `yaml:"interval"` // seconds
	Timeout  int    `yaml:"timeout"`  // seconds
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
