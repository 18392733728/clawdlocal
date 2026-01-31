package main

import (
	"clawdlocal/config"
	"clawdlocal/core"
	"log"
	"os"
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
	agent.Run()
}