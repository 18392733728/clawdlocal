package config

import (
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration for ClawdLocal agent
type Config struct {
	Agent      AgentConfig      `yaml:"agent"`
	Server     ServerConfig     `yaml:"server"`
	Plugins    PluginsConfig    `yaml:"plugins"`
	Logging    LoggingConfig    `yaml:"logging"`
}

type AgentConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Workspace   string `yaml:"workspace"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type PluginsConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Paths       []string `yaml:"paths"`
	AutoReload  bool     `yaml:"auto_reload"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Load loads configuration from file or creates default config
func Load(configPath string) (*Config, error) {
	// Create default config
	cfg := &Config{
		Agent: AgentConfig{
			Name:        "ClawdLocal",
			Version:     "0.1.0",
			Description: "Lightweight local AI agent framework",
			Workspace:   "./workspace",
		},
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Plugins: PluginsConfig{
			Enabled:    true,
			Paths:      []string{"./plugins"},
			AutoReload: false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}

	// Try to load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return cfg, err
		}
		
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return cfg, err
		}
	}

	// Ensure workspace directory exists
	if err := os.MkdirAll(cfg.Agent.Workspace, 0755); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Save saves the current configuration to file
func (c *Config) Save(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(configPath, data, 0644)
}