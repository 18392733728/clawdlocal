package core

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolCallEvent represents a tool call event
type ToolCallEvent struct {
	ToolName string                 `json:"tool_name"`
	Args     map[string]interface{} `json:"args"`
}

// ToolCallHandler handles tool call events
type ToolCallHandler struct {
	tools map[string]Tool
}

// NewToolCallHandler creates a new tool call handler
func NewToolCallHandler(tools map[string]Tool) *ToolCallHandler {
	return &ToolCallHandler{
		tools: tools,
	}
}

// Handle processes a tool call event
func (h *ToolCallHandler) Handle(ctx context.Context, event *Event) error {
	if event.Type != EventTypeToolCall {
		return fmt.Errorf("expected tool call event, got %s", event.Type)
	}

	// Parse tool call data
	var toolCall ToolCallEvent
	if err := json.Unmarshal(event.Data, &toolCall); err != nil {
		return fmt.Errorf("failed to parse tool call data: %w", err)
	}

	// Find the tool
	tool, exists := h.tools[toolCall.ToolName]
	if !exists {
		return fmt.Errorf("tool not found: %s", toolCall.ToolName)
	}

	// Execute the tool
	result, err := tool.Execute(ctx, toolCall.Args)
	if err != nil {
		return fmt.Errorf("tool execution failed: %w", err)
	}

	// Create result event
	resultData, _ := json.Marshal(map[string]interface{}{
		"tool_name": toolCall.ToolName,
		"result":    result,
		"success":   true,
	})

	resultEvent := &Event{
		Type: EventTypeToolResult,
		Data: resultData,
	}

	// Send result back to event loop
	// This would typically be done through the agent's event channel
	fmt.Printf("Tool call result: %+v\n", string(resultData))
	return nil
}

// CanHandle checks if this handler can handle the event type
func (h *ToolCallHandler) CanHandle(eventType EventType) bool {
	return eventType == EventTypeToolCall
}