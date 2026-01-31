package main

import (
	"context"
	"fmt"
	"time"
	"clawdlocal/core"
	"clawdlocal/config"
)

func main() {
	fmt.Println("Starting integration test...")

	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// Create agent
	agent, err := core.NewAgent(cfg)
	if err != nil {
		fmt.Printf("Failed to create agent: %v\n", err)
		return
	}

	// Test memory manager
	ctx := context.Background()
	
	// Store short-term memory
	err = agent.MemoryManager.SetShortTermMemory(ctx, "test_key", "test_value", 10*time.Second)
	if err != nil {
		fmt.Printf("Failed to set short-term memory: %v\n", err)
	} else {
		fmt.Println("Successfully stored short-term memory")
	}

	// Retrieve short-term memory
	value, found, err := agent.MemoryManager.GetShortTermMemory(ctx, "test_key")
	if err != nil {
		fmt.Printf("Failed to get short-term memory: %v\n", err)
	} else if found {
		fmt.Printf("Retrieved short-term memory: %v\n", value)
	} else {
		fmt.Println("Short-term memory not found")
	}

	// Store long-term memory
	err = agent.MemoryManager.SetLongTermMemory(ctx, "long_test_key", "long_test_value")
	if err != nil {
		fmt.Printf("Failed to set long-term memory: %v\n", err)
	} else {
		fmt.Println("Successfully stored long-term memory")
	}

	// Retrieve long-term memory
	value, found, err = agent.MemoryManager.GetLongTermMemory(ctx, "long_test_key")
	if err != nil {
		fmt.Printf("Failed to get long-term memory: %v\n", err)
	} else if found {
		fmt.Printf("Retrieved long-term memory: %v\n", value)
	} else {
		fmt.Println("Long-term memory not found")
	}

	// Test tool manager
	echoTool := &core.Tool{
		Name:        "echo",
		Description: "Echo back the input",
		Parameters: map[string]interface{}{
			"message": "string",
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			message, ok := args["message"].(string)
			if !ok {
				return nil, fmt.Errorf("missing message parameter")
			}
			return map[string]interface{}{
				"echo": message,
			}, nil
		},
	}

	err = agent.ToolManager.RegisterTool(echoTool)
	if err != nil {
		fmt.Printf("Failed to register tool: %v\n", err)
	} else {
		fmt.Println("Successfully registered echo tool")
	}

	// Execute tool
	call := &core.ToolCall{
		ID:   "test-call-1",
		Name: "echo",
		Args: map[string]interface{}{
			"message": "Hello, ClawdLocal!",
		},
	}

	result, err := agent.ToolManager.ExecuteTool(ctx, call)
	if err != nil {
		fmt.Printf("Failed to execute tool: %v\n", err)
	} else {
		fmt.Printf("Tool execution result: %+v\n", result)
	}

	fmt.Println("Integration test completed successfully!")
}