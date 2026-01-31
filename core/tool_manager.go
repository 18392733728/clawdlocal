package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
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
	ID   string                 `json:"id"`
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

// ToolManager manages registered tools
type ToolManager struct {
	mu    sync.RWMutex
	tools map[string]*Tool
	logger *logrus.Logger
}

// NewToolManager creates a new tool manager
func NewToolManager(logger *logrus.Logger) (*ToolManager, error) {
	return &ToolManager{
		tools:  make(map[string]*Tool),
		logger: logger,
	}, nil
}

// RegisterTool registers a new tool
func (tm *ToolManager) RegisterTool(tool *Tool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tools[tool.Name]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name)
	}

	tm.tools[tool.Name] = tool
	tm.logger.Infof("Registered tool: %s", tool.Name)
	return nil
}

// GetTool retrieves a tool by name
func (tm *ToolManager) GetTool(name string) (*Tool, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tool, exists := tm.tools[name]
	return tool, exists
}

// ListTools returns all registered tools
func (tm *ToolManager) ListTools() []*Tool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tools := make([]*Tool, 0, len(tm.tools))
	for _, tool := range tm.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ExecuteTool executes a tool call
func (tm *ToolManager) ExecuteTool(ctx context.Context, call *ToolCall) (*ToolResult, error) {
	tool, exists := tm.GetTool(call.Name)
	if !exists {
		return &ToolResult{
			ID:    call.ID,
			Name:  call.Name,
			Error: fmt.Sprintf("tool %s not found", call.Name),
		}, nil
	}

	result, err := tool.Handler(ctx, call.Args)
	if err != nil {
		return &ToolResult{
			ID:    call.ID,
			Name:  call.Name,
			Error: err.Error(),
		}, nil
	}

	return &ToolResult{
		ID:     call.ID,
		Name:   call.Name,
		Result: result,
	}, nil
}