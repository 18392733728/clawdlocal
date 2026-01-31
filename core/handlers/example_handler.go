package handlers

import (
	"clawdlocal/core"
	"log"
)

// ExampleHandler is a sample message handler that demonstrates the handler interface
type ExampleHandler struct{}

// Handle processes incoming messages
func (h *ExampleHandler) Handle(msg *core.Message) error {
	log.Printf("ExampleHandler received message: %+v", msg)
	
	// Process the message based on its type
	switch msg.Type {
	case "text":
		log.Printf("Processing text message: %s", msg.Content)
	case "command":
		log.Printf("Processing command: %s", msg.Content)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
	
	return nil
}

// CanHandle determines if this handler can process the given message
func (h *ExampleHandler) CanHandle(msg *core.Message) bool {
	// This example handler can handle any message
	return true
}