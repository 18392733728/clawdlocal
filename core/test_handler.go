package core

import (
	"context"
	"fmt"
)

// TestHandler is a simple test message handler for demonstration
type TestHandler struct{}

// Handle processes the message
func (h *TestHandler) Handle(ctx context.Context, msg *Message) error {
	fmt.Printf("TestHandler received message: %+v\n", msg)
	return nil
}

// CanHandle checks if this handler can handle the message type
func (h *TestHandler) CanHandle(msgType MessageType) bool {
	return true // Handle all message types for testing
}

// Priority returns the handler priority (lower number = higher priority)
func (h *TestHandler) Priority() int {
	return 100
}