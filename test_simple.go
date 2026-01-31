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
	memoryManager, err := core.NewMemoryManager(nil)
	if err != nil {
		log.Fatalf("Failed to create memory manager: %v", err)
	}
	
	ctx := context.Background()
	
	// Test short-term memory
	fmt.Println("Testing short-term memory...")
	err = memoryManager.SetShortTermMemory(ctx, "test_key", "test_value", 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to set short-term memory: %v", err)
	}
	
	value, err := memoryManager.GetShortTermMemory(ctx, "test_key")
	if err != nil {
		log.Fatalf("Failed to get short-term memory: %v", err)
	}
	
	fmt.Printf("Short-term memory value: %v\n", value)
	
	// Test long-term memory
	fmt.Println("Testing long-term memory...")
	err = memoryManager.SetLongTermMemory(ctx, "test_long_key", "test_long_value")
	if err != nil {
		log.Fatalf("Failed to set long-term memory: %v", err)
	}
	
	longValue, err := memoryManager.GetLongTermMemory(ctx, "test_long_key")
	if err != nil {
		log.Fatalf("Failed to get long-term memory: %v", err)
	}
	
	fmt.Printf("Long-term memory value: %v\n", longValue)
	
	// Test event loop
	fmt.Println("Testing event loop...")
	eventLoop := core.NewEventLoop(ctx, nil, 100)
	
	// Register a simple handler
	handler := &SimpleHandler{}
	eventLoop.RegisterHandler(handler)
	
	err = eventLoop.Start()
	if err != nil {
		log.Fatalf("Failed to start event loop: %v", err)
	}
	
	// Emit a test event
	event := &core.Event{
		ID:        "test-1",
		Type:      core.EventTypeMessage,
		Timestamp: time.Now(),
		Data:      "Hello, World!",
	}
	
	err = eventLoop.Emit(event)
	if err != nil {
		log.Fatalf("Failed to emit event: %v", err)
	}
	
	time.Sleep(1 * time.Second)
	
	eventLoop.Stop()
	fmt.Println("All tests passed!")
}

type SimpleHandler struct{}

func (h *SimpleHandler) Handle(ctx context.Context, event *core.Event) error {
	fmt.Printf("Handled event: %s - %v\n", event.Type, event.Data)
	return nil
}

func (h *SimpleHandler) CanHandle(eventType core.EventType) bool {
	return true
}