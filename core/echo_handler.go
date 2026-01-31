package core

import (
	"context"
	"fmt"
)

// EchoHandler echoes back messages
type EchoHandler struct{}

// Handle processes the message
func (h *EchoHandler) Handle(ctx context.Context, msg *Message) error {
	fmt.Printf("Echo: %+v\n", msg)
	return nil
}

// CanHandle checks if this handler can handle the message type
func (h *EchoHandler) CanHandle(msgType MessageType) bool {
	return true // Handle all message types
}

// Priority returns the handler priority
func (h *EchoHandler) Priority() int {
	return 50
}