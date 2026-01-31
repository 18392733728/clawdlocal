package main

import (
	"context"
	"fmt"
	"sync"
)

// Tool represents a callable function that can be executed by the agent
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Handler     func(ctx context.Context, args map[string]interface{}) (interface{}, error)
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

	tr.tools[tool.Name] = tool
	return nil
}

// ExecuteTool executes a tool call
func (tr *ToolRegistry) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	tr.mu.RLock()
	tool, exists := tr.tools[name]
	tr.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool.Handler(ctx, args)
}

// Example tool handlers
func echoTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	message, ok := args["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing message parameter")
	}
	return map[string]interface{}{
		"echo": message,
	}, nil
}

func addTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	a, ok1 := args["a"].(float64)
	b, ok2 := args["b"].(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("missing a or b parameter")
	}
	return map[string]interface{}{
		"sum": a + b,
	}, nil
}

func main() {
	ctx := context.Background()
	registry := NewToolRegistry()

	// Register tools
	registry.RegisterTool(&Tool{
		Name:        "echo",
		Description: "Echo back a message",
		Parameters: map[string]interface{}{
			"message": "string",
		},
		Handler: echoTool,
	})

	registry.RegisterTool(&Tool{
		Name:        "add",
		Description: "Add two numbers",
		Parameters: map[string]interface{}{
			"a": "number",
			"b": "number",
		},
		Handler: addTool,
	})

	// Test echo tool
	result, err := registry.ExecuteTool(ctx, "echo", map[string]interface{}{
		"message": "Hello, World!",
	})
	if err != nil {
		fmt.Printf("Error executing echo tool: %v\n", err)
	} else {
		fmt.Printf("Echo result: %+v\n", result)
	}

	// Test add tool
	result, err = registry.ExecuteTool(ctx, "add", map[string]interface{}{
		"a": 5.0,
		"b": 3.0,
	})
	if err != nil {
		fmt.Printf("Error executing add tool: %v\n", err)
	} else {
		fmt.Printf("Add result: %+v\n", result)
	}

	fmt.Println("Tools test completed!")
}