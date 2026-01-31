package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"clawdlocal/core"
)

func main() {
	fmt.Println("Testing ClawdLocal Memory System...")

	// Create memory manager
	memoryManager, err := core.NewMemoryManager(&core.MemoryConfig{
		ShortTerm: &core.ShortTermMemoryConfig{
			MaxSize: 100,
			DefaultTTL: 300, // 5 minutes
		},
		LongTerm: &core.LongTermMemoryConfig{
			StoragePath: "./workspace/memory",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create memory manager: %v", err)
	}

	ctx := context.Background()

	// Test short-term memory
	fmt.Println("\n--- Testing Short-Term Memory ---")
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
	fmt.Println("\n--- Testing Long-Term Memory ---")
	err = memoryManager.SetLongTermMemory(ctx, "long_test_key", "long_test_value")
	if err != nil {
		log.Fatalf("Failed to set long-term memory: %v", err)
	}

	longValue, err := memoryManager.GetLongTermMemory(ctx, "long_test_key")
	if err != nil {
		log.Fatalf("Failed to get long-term memory: %v", err)
	}
	fmt.Printf("Retrieved long-term memory: %v\n", longValue)

	// Test expiration
	fmt.Println("\n--- Testing Expiration ---")
	time.Sleep(11 * time.Second)
	expiredValue, err := memoryManager.GetShortTermMemory(ctx, "test_key")
	if err != nil {
		fmt.Printf("Expected: short-term memory expired, got error: %v\n", err)
	} else {
		fmt.Printf("Unexpected: short-term memory still exists: %v\n", expiredValue)
	}

	fmt.Println("\nMemory system test completed successfully!")
}