package core

import (
	"context"
	"encoding/json"
	"fmt"
)

// MemoryMessageType defines memory-related message types
const (
	MessageTypeMemoryStore MessageType = "memory_store"
	MessageTypeMemoryGet   MessageType = "memory_get"
	MessageTypeMemoryDelete MessageType = "memory_delete"
)

// MemoryMessagePayload represents memory operation payload
type MemoryMessagePayload struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value,omitempty"`
}

// MemoryMessageHandler handles memory-related messages
type MemoryMessageHandler struct {
	MemoryManager *MemoryManager
}

// Handle processes memory messages
func (h *MemoryMessageHandler) Handle(ctx context.Context, msg *Message) error {
	// Convert payload to JSON bytes first
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}
	
	payload := &MemoryMessagePayload{}
	if err := json.Unmarshal(payloadBytes, payload); err != nil {
		return fmt.Errorf("failed to unmarshal memory payload: %w", err)
	}

	switch msg.Type {
	case MessageTypeMemoryStore:
		return h.handleStore(ctx, payload)
	case MessageTypeMemoryGet:
		return h.handleGet(ctx, payload)
	case MessageTypeMemoryDelete:
		return h.handleDelete(ctx, payload)
	default:
		return fmt.Errorf("unsupported memory message type: %s", msg.Type)
	}
}

// CanHandle checks if this handler can handle the message type
func (h *MemoryMessageHandler) CanHandle(msgType MessageType) bool {
	return msgType == MessageTypeMemoryStore ||
		msgType == MessageTypeMemoryGet ||
		msgType == MessageTypeMemoryDelete
}

// Priority returns the handler priority
func (h *MemoryMessageHandler) Priority() int {
	return 150
}

// handleStore stores a value in long-term memory
func (h *MemoryMessageHandler) handleStore(ctx context.Context, payload *MemoryMessagePayload) error {
	return h.MemoryManager.SetLongTermMemory(ctx, payload.Key, payload.Value)
}

// handleGet retrieves a value from long-term memory
func (h *MemoryMessageHandler) handleGet(ctx context.Context, payload *MemoryMessagePayload) error {
	value, found, err := h.MemoryManager.GetLongTermMemory(ctx, payload.Key)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}
	if !found {
		return fmt.Errorf("memory key not found: %s", payload.Key)
	}
	
	// Add result to message metadata for other handlers
	if msg, ok := ctx.Value("current_message").(*Message); ok {
		if msg.Metadata == nil {
			msg.Metadata = make(map[string]interface{})
		}
		msg.Metadata["memory_result"] = value
	}
	
	return nil
}

// handleDelete removes a value from long-term memory
func (h *MemoryMessageHandler) handleDelete(ctx context.Context, payload *MemoryMessagePayload) error {
	// Note: Current MemoryManager doesn't have Delete method
	// For now, we'll just log that deletion is requested
	fmt.Printf("Memory deletion requested for key: %s (not implemented)\n", payload.Key)
	return nil
}