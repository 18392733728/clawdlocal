package core

import (
	"clawdlocal/config"
	"github.com/sirupsen/logrus"
)

type Agent struct {
	logger *logrus.Logger
	config *config.Config
}

func NewAgent(configPath string) (*Agent, error) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	// Setup logger based on config
	logger := logrus.New()
	
	// TODO: Implement logging configuration
	
	return &Agent{
		logger: logger,
		config: cfg,
	}, nil
}

func (a *Agent) Run() {
	a.logger.Info("ClawdLocal agent started successfully!")
	a.logger.Infof("Agent: %s v%s", a.config.Agent.Name, a.config.Agent.Version)
	a.logger.Infof("Workspace: %s", a.config.Agent.Workspace)
	a.logger.Infof("Server: %s:%d", a.config.Server.Host, a.config.Server.Port)
	
	// Main event loop will go here
	select {} // Keep running
}