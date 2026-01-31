package main

import (
	"clawdlocal/config"
	"clawdlocal/core"
	"clawdlocal/core/handlers"
	"log"
	"time"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config/event_loop.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create agent with event loop
	agent, err := core.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Register example handler
	agent.RegisterHandler("example", &handlers.ExampleHandler{})

	// Start the agent
	go agent.Run()

	// Send some test messages
	for i := 0; i < 5; i++ {
		msg := &core.Message{
			ID:      core.GenerateMessageID(),
			Type:    "example",
			Payload: map[string]interface{}{"counter": i, "timestamp": time.Now().Unix()},
			Metadata: map[string]string{
				"source": "example_client",
			},
		}
		
		if err := agent.SendMessage(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
		
		time.Sleep(1 * time.Second)
	}

	// Keep the program running
	select {}
}