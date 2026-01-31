package handlers

import (
	"context"
	"clawdlocal/core"
	"log"
)

// ExampleHandler is a simple handler for demonstration purposes
type ExampleHandler struct{}

// CanHandle returns true if this handler can handle the given message type
func (h *ExampleHandler) CanHandle(msgType core.MessageType) bool {
	return msgType == core.MessageType("example")
}

// Handle processes example messages
func (h *ExampleHandler) Handle(ctx context.Context, msg *core.Message) error {
	// Convert payload to map for safe access
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		log.Printf("ExampleHandler received invalid payload type: %T", msg.Payload)
		return nil
	}
	
	counter, _ := payload["counter"]
	timestamp, _ := payload["timestamp"]
	
	log.Printf("ExampleHandler received message: ID=%s, Counter=%v, Timestamp=%v",
		msg.ID, counter, timestamp)
	
	// Echo back a response
	response := &core.Message{
		ID:      core.GenerateMessageID(),
		Type:    core.MessageType("example_response"),
		Payload: map[string]interface{}{"original_id": msg.ID, "processed_at": core.GetCurrentTimestamp()},
		Metadata: map[string]interface{}{
			"handler": "example",
		},
	}
	
	// In a real implementation, this would be sent through the agent's message system
	log.Printf("ExampleHandler sending response: ID=%s", response.ID)
	return nil
}