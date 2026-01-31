package main

import (
	"clawdlocal/core"
	"clawdlocal/core/handlers"
	"log"
	"time"
)

func main() {
	log.Println("Testing Event Loop...")

	// Create event loop
	loop := core.NewEventLoop()
	
	// Register example handler
	exampleHandler := &handlers.ExampleHandler{}
	loop.RegisterHandler(core.EventTypeMessage, exampleHandler)
	
	// Start the loop in a goroutine
	go func() {
		loop.Run(nil)
	}()
	
	// Send some test messages
	testMsg1 := &core.Message{
		ID:      "test-1",
		Type:    "text",
		Content: "Hello, Event Loop!",
	}
	
	testMsg2 := &core.Message{
		ID:      "test-2", 
		Type:    "command",
		Content: "status",
	}
	
	// Publish messages
	loop.Publish(testMsg1)
	loop.Publish(testMsg2)
	
	// Wait a bit to see processing
	time.Sleep(2 * time.Second)
	
	log.Println("Event Loop test completed!")
}