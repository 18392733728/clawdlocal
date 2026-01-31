package main

import (
	"clawdlocal/config"
	"clawdlocal/core"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration
	configPath := "config/default.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create and run agent
	agent, err := core.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	
	log.Println("ClawdLocal agent started successfully!")
	
	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Start the agent with context
	go func() {
		agent.Run(ctx)
	}()
	
	// Wait for shutdown signal
	<-sigChan
	log.Println("Received shutdown signal, stopping agent...")
	cancel()
	
	// Wait for graceful shutdown
	agent.Shutdown()
	log.Println("Agent stopped gracefully")
}