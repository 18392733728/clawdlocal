package core

import (
	"context"
	"clawdlocal/config"
	"time"
	"github.com/sirupsen/logrus"
)

// Agent represents the main ClawdLocal agent
type Agent struct {
	logger        *logrus.Logger
	config        *config.Config
	eventLoop     *EventLoop
	messageRouter *MessageRouter
	ToolManager   *ToolManager
	MemoryManager *MemoryManager
}

// NewAgent creates a new agent instance
func NewAgent(cfg *config.Config) (*Agent, error) {
	// Setup logger based on config
	logger := logrus.New()
	
	// Create tool manager
	toolManager, err := NewToolManager(logger)
	if err != nil {
		return nil, err
	}
	
	// Create memory manager
	memoryManager, err := NewMemoryManager(logger, &MemoryConfig{
		ShortTermCapacity: 1000,
		LongTermFile:      cfg.Memory.LongTerm.StorageDir + "/long_term.json",
		CleanupInterval:   5 * time.Minute,
	})
	if err != nil {
		return nil, err
	}
	
	return &Agent{
		logger:        logger,
		config:        cfg,
		ToolManager:   toolManager,
		MemoryManager: memoryManager,
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
	a.eventLoop = NewEventLoop(ctx, a.logger, a.config.Agent.MaxQueueSize)
	
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
	a.messageRouter.RegisterHandler(&EchoHandler{})
	
	// Register test handler  
	a.messageRouter.RegisterHandler(&TestHandler{})
	
	// Register tool call handler
	a.messageRouter.RegisterHandler(&ToolCallMessageHandler{
		ToolManager: a.ToolManager,
	})
	
	// Register memory handler
	a.messageRouter.RegisterHandler(&MemoryMessageHandler{
		MemoryManager: a.MemoryManager,
	})
}

// Shutdown gracefully stops the agent
func (a *Agent) Shutdown() {
	if a.eventLoop != nil {
		a.eventLoop.Stop()
	}
	a.logger.Info("Agent shutdown complete")
}