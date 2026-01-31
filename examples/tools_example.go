package main

import (
	"context"
	"clawdlocal/core"
	"encoding/json"
	"fmt"
	"log"
)

// ExampleTool demonstrates how to create a custom tool
type ExampleTool struct{}

func (t *ExampleTool) Name() string {
	return "example_tool"
}

func (t *ExampleTool) Description() string {
	return "An example tool that demonstrates the tool calling system"
}

func (t *ExampleTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"required": []string{"input"},
		"properties": map[string]interface{}{
			"input": map[string]string{
				"type":        "string",
				"description": "Input string to process",
			},
		},
	}
}

func (t *ExampleTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	input, ok := params["input"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required parameter 'input'")
	}
	
	// Process the input (example: just return it with a prefix)
	result := fmt.Sprintf("Processed: %s", input)
	
	return map[string]interface{}{
		"result": result,
		"length": len(input),
	}, nil
}

func main() {
	// Create a new tool registry
	registry := core.NewToolRegistry()
	
	// Register the example tool
	err := registry.Register(&ExampleTool{})
	if err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}
	
	// Create a tool call event
	params := map[string]interface{}{
		"input": "Hello, ClawdLocal!",
	}
	
	event := &core.Event{
		Type: core.EventTypeToolCall,
		Data: core.ToolCallEventData{
			ToolName: "example_tool",
			Params:   params,
		},
	}
	
	// Execute the tool call
	result, err := registry.Execute(context.Background(), event.Data.(core.ToolCallEventData))
	if err != nil {
		log.Fatalf("Tool execution failed: %v", err)
	}
	
	// Print the result
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Tool result:\n%s\n", resultJSON)
}