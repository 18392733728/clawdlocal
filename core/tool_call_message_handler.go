package core

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolCallMessage represents a tool call message
type ToolCallMessage struct {
	ToolName string                 `json:"tool_name"`
	Args     map[string]interface{} `json:"args"`
}

// ToolCallMessageHandler handles tool call messages
type ToolCallMessageHandler struct {
	ToolManager *ToolManager
}

// Handle processes a tool call message
func (h *ToolCallMessageHandler) Handle(ctx context.Context, msg *Message) error {
	if msg.Type != MessageTypeToolResponse {
		return nil // Only handle tool response messages
	}

	// Parse tool call data from message payload
	var toolCall ToolCallMessage
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}
	
	if err := json.Unmarshal(payloadBytes, &toolCall); err != nil {
		return fmt.Errorf("failed to parse tool call data: %w", err)
	}

	// Execute the tool
	call := &ToolCall{
		ID:   msg.ID,
		Name: toolCall.ToolName,
		Args: toolCall.Args,
	}
	
	result, err := h.ToolManager.ExecuteTool(ctx, call)
	if err != nil {
		return fmt.Errorf("tool execution failed: %w", err)
	}

	// Log the result
	resultPayload := map[string]interface{}{
		"tool_name": toolCall.ToolName,
		"result":    result.Result,
		"error":     result.Error,
		"success":   result.Error == "",
	}

	// In a real implementation, this would be sent back through the message router
	// For now, just log it
	fmt.Printf("Tool call result: %+v\n", resultPayload)
	return nil
}

// CanHandle checks if this handler can handle the message type
func (h *ToolCallMessageHandler) CanHandle(msgType MessageType) bool {
	return msgType == MessageTypeToolResponse
}

// Priority returns the handler priority
func (h *ToolCallMessageHandler) Priority() int {
	return 200
}