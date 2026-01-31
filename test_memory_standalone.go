package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"clawdlocal/core"
)

func main() {
	fmt.Println("Testing Memory System...")

	// Create memory manager
	memoryManager, err := core.NewMemoryManager(nil)
	if err != nil {
		log.Fatalf("Failed to create memory manager: %v", err)
	}

	ctx := context.Background()

	// Test short-term memory
	fmt.Println("\n--- Testing Short-term Memory ---")
	
	err = memoryManager.SetShortTermMemory(ctx, "test_key", "test_value", 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to set short-term memory: %v", err)
	}

	value, err := memoryManager.GetShortTermMemory(ctx, "test_key")
	if err != nil {
		log.Fatalf("Failed to get short-term memory: %v", err)
	}
	fmt.Printf("Retrieved short-term memory: %v\n", value)

	// Test long-term memory
	fmt.Println("\n--- Testing Long-term Memory ---")
	
	err = memoryManager.SetLongTermMemory(ctx, "long_test_key", map[string]interface{}{
		"name": "ClawdLocal",
		"version": "0.1.0",
	})
	if err != nil {
		log.Fatalf("Failed to set long-term memory: %v", err)
	}

	longValue, err := memoryManager.GetLongTermMemory(ctx, "long_test_key")
	if err != nil {
		log.Fatalf("Failed to get long-term memory: %v", err)
	}
	fmt.Printf("Retrieved long-term memory: %v\n", longValue)

	// Test cleanup
	fmt.Println("\n--- Testing Cleanup ---")
	memoryManager.Cleanup()
	fmt.Println("Memory cleanup completed")

	fmt.Println("\n✅ Memory system test completed successfully!")
}