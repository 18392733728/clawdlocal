package core

import (
	"context"
	"clawdlocal/config"
	"github.com/sirupsen/logrus"
)

// Agent represents the main ClawdLocal agent
type Agent struct {
	logger        *logrus.Logger
	config        *config.Config
	eventLoop     *EventLoop
	messageRouter *MessageRouter
	toolManager   *ToolManager
}

// NewAgent creates a new agent instance
func NewAgent(cfg *config.Config) (*Agent, error) {
	// Setup logger based on config
	logger := logrus.New()
	
	// TODO: Implement logging configuration from cfg
	
	// Create tool manager
	toolManager, err := NewToolManager(logger)
	if err != nil {
		return nil, err
	}
	
	return &Agent{
		logger:      logger,
		config:      cfg,
		toolManager: toolManager,
	}, nil
}

// Run starts the agent's main event loop
func (a *Agent) Run(ctx context.Context) error {
	a.logger.Info("ClawdLocal agent starting...")
	a.logger.Infof("Agent: %s v%s", a.config.Agent.Name, a.config.Agent.Version)
	a.logger.Infof("Workspace: %s", a.config.Agent.Workspace)
	a.logger.Infof("Server: %s:%d", a.config.Server.Host, a.config.Server.Port)
	
	// Initialize message router
	a.messageRouter = NewMessageRouter()
	
	// Register default handlers
	a.registerDefaultHandlers()
	
	// Initialize event loop
	a.eventLoop = NewEventLoop(ctx, a.logger, a.config.Agent.MaxQueueSize, a.toolManager)
	
	// Start the event loop
	if err := a.eventLoop.Start(); err != nil {
		return err
	}
	
	a.logger.Info("ClawdLocal agent started successfully!")
	
	// Wait for context cancellation
	<-ctx.Done()
	
	return ctx.Err()
}

// registerDefaultHandlers registers built-in message handlers
func (a *Agent) registerDefaultHandlers() {
	// Register echo handler
	a.messageRouter.RegisterHandler(&EchoHandler{}, 50)
	
	// Register test handler  
	a.messageRouter.RegisterHandler(&TestHandler{}, 100)
	
	// Register tool call handler
	a.messageRouter.RegisterHandler(&ToolCallHandler{
		toolManager: a.toolManager,
	}, 200)
}

// Shutdown gracefully stops the agent
func (a *Agent) Shutdown() {
	if a.eventLoop != nil {
		a.eventLoop.Stop()
	}
	a.logger.Info("Agent shutdown complete")
}