package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx := context.Background()
	mm := NewSimpleMemoryManager()

	// Test short-term memory with TTL
	fmt.Println("Testing short-term memory...")
	err := mm.SetShortTermMemory(ctx, "test_key", "test_value", 10*time.Second)
	if err != nil {
		fmt.Printf("Error setting short-term memory: %v\n", err)
		return
	}

	value, err := mm.GetShortTermMemory(ctx, "test_key")
	if err != nil {
		fmt.Printf("Error getting short-term memory: %v\n", err)
		return
	}
	fmt.Printf("Retrieved short-term memory: %v\n", value)

	// Test long-term memory
	fmt.Println("Testing long-term memory...")
	err = mm.SetLongTermMemory(ctx, "long_key", "long_value")
	if err != nil {
		fmt.Printf("Error setting long-term memory: %v\n", err)
		return
	}

	value, err = mm.GetLongTermMemory(ctx, "long_key")
	if err != nil {
		fmt.Printf("Error getting long-term memory: %v\n", err)
		return
	}
	fmt.Printf("Retrieved long-term memory: %v\n", value)

	// Test saving and loading
	fmt.Println("Testing save/load...")
	err = mm.SaveToFile("test_memory.json")
	if err != nil {
		fmt.Printf("Error saving memory: %v\n", err)
		return
	}

	newMM := NewSimpleMemoryManager()
	err = newMM.LoadFromFile("test_memory.json")
	if err != nil {
		fmt.Printf("Error loading memory: %v\n", err)
		return
	}

	value, err = newMM.GetLongTermMemory(ctx, "long_key")
	if err != nil {
		fmt.Printf("Error getting loaded long-term memory: %v\n", err)
		return
	}
	fmt.Printf("Retrieved loaded long-term memory: %v\n", value)

	fmt.Println("All tests passed!")
}