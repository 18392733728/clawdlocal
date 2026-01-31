package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"clawdlocal-test/core"
)

// Example tool handlers
func readFileTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing path argument")
	}
	
	// Simulate file reading
	result := fmt.Sprintf("Content of file: %s", path)
	return result, nil
}

func currentTimeTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return time.Now().Format("2006-01-02 15:04:05"), nil
}

func main() {
	// Create tool registry
	registry := core.NewToolRegistry()

	// Register tools
	readFileToolDef := &core.Tool{
		Name:        "read_file",
		Description: "Read the contents of a file",
		Parameters: map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read",
			},
		},
		Handler: readFileTool,
	}

	currentTimeToolDef := &core.Tool{
		Name:        "current_time",
		Description: "Get the current time",
		Parameters:  map[string]interface{}{},
		Handler:     currentTimeTool,
	}

	err := registry.RegisterTool(readFileToolDef)
	if err != nil {
		log.Fatalf("Failed to register read_file tool: %v", err)
	}

	err = registry.RegisterTool(currentTimeToolDef)
	if err != nil {
		log.Fatalf("Failed to register current_time tool: %v", err)
	}

	fmt.Println("Registered tools:")
	for _, tool := range registry.ListTools() {
		fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
	}

	// Test tool execution
	ctx := context.Background()

	// Test read_file tool
	readCall := &core.ToolCall{
		ID:   "test1",
		Name: "read_file",
		Args: map[string]interface{}{
			"path": "/home/user/test.txt",
		},
	}

	result, err := registry.ExecuteTool(ctx, readCall)
	if err != nil {
		log.Printf("Error executing read_file: %v", err)
	} else {
		fmt.Printf("Read file result: %+v\n", result)
	}

	// Test current_time tool
	timeCall := &core.ToolCall{
		ID:   "test2",
		Name: "current_time",
		Args: map[string]interface{}{},
	}

	result, err = registry.ExecuteTool(ctx, timeCall)
	if err != nil {
		log.Printf("Error executing current_time: %v", err)
	} else {
		fmt.Printf("Current time result: %+v\n", result)
	}

	fmt.Println("Tool system test completed!")
}