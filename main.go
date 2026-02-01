package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"clawdlocal/config"
	"clawdlocal/core"
	"clawdlocal/tools"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create agent
	agent, err := core.NewAgent(cfg)
	if err != nil {
		fmt.Printf("Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	// Register built-in tools
	tools.RegisterAllTools(agent)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nShutting down...")
		agent.Shutdown()
		cancel()
	}()

	// Run the agent
	if err := agent.Run(ctx); err != nil {
		fmt.Printf("Agent error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Agent stopped gracefully")
}