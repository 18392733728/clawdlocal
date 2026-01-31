package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"clawdlocal/core"
)

// CustomMessageHandler 自定义消息处理器
type CustomMessageHandler struct {
	name string
}

func NewCustomMessageHandler(name string) *CustomMessageHandler {
	return &CustomMessageHandler{name: name}
}

func (h *CustomMessageHandler) Handle(ctx context.Context, msg *core.Message) error {
	fmt.Printf("[%s] Handling message: %+v\n", h.name, msg)
	
	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

func (h *CustomMessageHandler) CanHandle(msgType core.MessageType) bool {
	// 处理所有类型的消息
	return true
}

func (h *CustomMessageHandler) Priority() int {
	return 10 // 默认优先级
}

func main() {
	// 创建上下文
	ctx := context.Background()
	
	// 创建消息路由器
	router := core.NewMessageRouter()
	
	// 注册处理器
	handler1 := NewCustomMessageHandler("Handler-1")
	handler2 := NewCustomMessageHandler("Handler-2")
	
	router.RegisterHandler(handler1)
	router.RegisterHandler(handler2)
	
	// 创建并启动事件循环
	eventLoop := core.NewEventLoop(ctx, nil, 100)
	
	// 启动事件循环
	if err := eventLoop.Start(); err != nil {
		log.Fatalf("Failed to start event loop: %v", err)
	}
	
	// 发送一些测试消息
	messages := []core.Message{
		{
			ID:      "msg-1",
			Type:    core.MessageTypeUserInput,
			Source:  "user",
			Target:  "agent",
			Payload: "Hello, world!",
			Timestamp: time.Now(),
		},
		{
			ID:      "msg-2",
			Type:    core.MessageTypeSystemEvent,
			Source:  "system",
			Target:  "agent",
			Payload: "System started",
			Timestamp: time.Now(),
		},
		{
			ID:      "msg-3",
			Type:    core.MessageTypeToolResponse,
			Source:  "tool",
			Target:  "agent",
			Payload: map[string]interface{}{
				"tool": "web_search",
				"result": "Search completed",
			},
			Timestamp: time.Now(),
		},
	}
	
	// 将消息发送到事件循环
	for _, msg := range messages {
		event := &core.Event{
			ID:        msg.ID,
			Type:      core.EventTypeMessage,
			Timestamp: msg.Timestamp,
			Data:      msg,
		}
		
		if err := eventLoop.Emit(event); err != nil {
			log.Printf("Failed to emit event: %v", err)
		}
		
		// 稍微延迟以观察处理顺序
		time.Sleep(50 * time.Millisecond)
	}
	
	// 等待一段时间让所有消息被处理
	time.Sleep(2 * time.Second)
	
	// 停止事件循环
	eventLoop.Stop()
	
	fmt.Println("Example completed!")
}