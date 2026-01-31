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
	Memory     MemoryConfig     `yaml:"memory"`
	Web        WebConfig        `yaml:"web"`
	Plugins    PluginsConfig    `yaml:"plugins"`
	Logging    LoggingConfig    `yaml:"logging"`
}

type AgentConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Workspace   string `yaml:"workspace"`
	MaxQueueSize int   `yaml:"max_queue_size"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type MemoryConfig struct {
	ShortTerm ShortTermMemoryConfig `yaml:"short_term"`
	LongTerm  LongTermMemoryConfig  `yaml:"long_term"`
}

type ShortTermMemoryConfig struct {
	Enabled     bool   `yaml:"enabled"`
	MaxEntries  int    `yaml:"max_entries"`
	DefaultTTL  int64  `yaml:"default_ttl"` // seconds
}

type LongTermMemoryConfig struct {
	Enabled    bool   `yaml:"enabled"`
	StorageDir string `yaml:"storage_dir"`
}

type WebConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	StaticDir  string `yaml:"static_dir"`
	APIPrefix  string `yaml:"api_prefix"`
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
			Name:         "ClawdLocal",
			Version:      "0.1.0",
			Description:  "Lightweight local AI agent framework",
			Workspace:    "./workspace",
			MaxQueueSize: 1000,
		},
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Memory: MemoryConfig{
			ShortTerm: ShortTermMemoryConfig{
				Enabled:     true,
				MaxEntries:  1000,
				DefaultTTL:  3600, // 1 hour
			},
			LongTerm: LongTermMemoryConfig{
				Enabled:    true,
				StorageDir: "./memory",
			},
		},
		Web: WebConfig{
			Enabled:   true,
			Host:      "localhost",
			Port:      8080,
			StaticDir: "./web/static",
			APIPrefix: "/api/v1",
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

	// Ensure memory directory exists
	if err := os.MkdirAll(cfg.Memory.LongTerm.StorageDir, 0755); err != nil {
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