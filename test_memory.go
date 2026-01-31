package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clawdlocal/config"
	"clawdlocal/core"
)

func main() {
	// Load test configuration
	cfg, err := config.Load("config/test.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create agent
	agent, err := core.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the agent in a goroutine
	go func() {
		if err := agent.Run(ctx); err != nil {
			log.Printf("Agent error: %v", err)
		}
	}()

	// Test memory functionality after a short delay
	time.Sleep(2 * time.Second)

	// Test short-term memory
	fmt.Println("Testing short-term memory...")
	err = agent.MemoryManager.SetShortTermMemory(ctx, "test_key", "test_value", 30*time.Second)
	if err != nil {
		log.Printf("Failed to set short-term memory: %v", err)
	} else {
		fmt.Println("Short-term memory set successfully")
	}

	value, err := agent.MemoryManager.GetShortTermMemory(ctx, "test_key")
	if err != nil {
		log.Printf("Failed to get short-term memory: %v", err)
	} else {
		fmt.Printf("Retrieved short-term memory: %v\n", value)
	}

	// Test long-term memory
	fmt.Println("Testing long-term memory...")
	err = agent.MemoryManager.SetLongTermMemory(ctx, "test_long_key", "test_long_value")
	if err != nil {
		log.Printf("Failed to set long-term memory: %v", err)
	} else {
		fmt.Println("Long-term memory set successfully")
	}

	value, err = agent.MemoryManager.GetLongTermMemory(ctx, "test_long_key")
	if err != nil {
		log.Printf("Failed to get long-term memory: %v", err)
	} else {
		fmt.Printf("Retrieved long-term memory: %v\n", value)
	}

	// Wait for shutdown signal
	fmt.Println("Test completed. Press Ctrl+C to exit...")
	<-sigChan
	fmt.Println("Shutting down...")
	cancel()
	agent.Shutdown()
}