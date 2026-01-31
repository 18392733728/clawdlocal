package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Tool represents a callable function that can be executed by the agent
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Handler     ToolHandler            `json:"-"`
}

// ToolHandler is the function signature for tool handlers
type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// ToolCall represents a request to call a tool
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Args     map[string]interface{} `json:"args"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Result  interface{} `json:"result"`
	Error   string      `json:"error,omitempty"`
}

// ToolRegistry manages registered tools
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]*Tool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*Tool),
	}
}

// RegisterTool registers a new tool
func (tr *ToolRegistry) RegisterTool(tool *Tool) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if _, exists := tr.tools[tool.Name]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name)
	}

	tr.tools[tool.Name] = tool
	return nil
}

// GetTool retrieves a tool by name
func (tr *ToolRegistry) GetTool(name string) (*Tool, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	tool, exists := tr.tools[name]
	return tool, exists
}

// ListTools returns all registered tools
func (tr *ToolRegistry) ListTools() []*Tool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	tools := make([]*Tool, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ExecuteTool executes a tool call
func (tr *ToolRegistry) ExecuteTool(ctx context.Context, call *ToolCall) (*ToolResult, error) {
	tool, exists := tr.GetTool(call.Name)
	if !exists {
		return &ToolResult{
			ID:   call.ID,
			Name: call.Name,
			Error: fmt.Sprintf("tool %s not found", call.Name),
		}, nil
	}

	result, err := tool.Handler(ctx, call.Args)
	if err != nil {
		return &ToolResult{
			ID:   call.ID,
			Name: call.Name,
			Error: err.Error(),
		}, nil
	}

	return &ToolResult{
		ID:     call.ID,
		Name:   call.Name,
		Result: result,
	}, nil
}

// ToolCallEvent represents a tool call event
type ToolCallEvent struct {
	Call   *ToolCall `json:"call"`
	Result chan *ToolResult `json:"-"`
}

// NewToolCallEvent creates a new tool call event
func NewToolCallEvent(call *ToolCall) *ToolCallEvent {
	return &ToolCallEvent{
		Call:   call,
		Result: make(chan *ToolResult, 1),
	}
}

// MarshalJSON implements JSON marshaling for ToolCallEvent
func (tce *ToolCallEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(tce.Call)
}

// UnmarshalJSON implements JSON unmarshaling for ToolCallEvent
func (tce *ToolCallEvent) UnmarshalJSON(data []byte) error {
	call := &ToolCall{}
	if err := json.Unmarshal(data, call); err != nil {
		return err
	}
	tce.Call = call
	tce.Result = make(chan *ToolResult, 1)
	return nil
}