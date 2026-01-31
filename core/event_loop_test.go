package core

import (
	"context"
	"testing"
	"time"
)

func TestEventLoop(t *testing.T) {
	// 创建事件循环
	loop := NewEventLoop()

	// 启动事件循环
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go loop.Run(ctx)

	// 测试消息处理
	testMessage := &Message{
		ID:      "test-1",
		Type:    "test",
		Payload: map[string]interface{}{"data": "hello"},
	}

	processed := make(chan bool, 1)
	
	// 注册处理器
	loop.RegisterHandler("test", func(msg *Message) error {
		if msg.Payload.(map[string]interface{})["data"] == "hello" {
			processed <- true
		}
		return nil
	})

	// 发送消息
	err := loop.SendMessage(testMessage)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待处理完成
	select {
	case <-processed:
		// 成功处理
	case <-time.After(2 * time.Second):
		t.Fatal("Message was not processed in time")
	}
}

func TestEventLoopShutdown(t *testing.T) {
	loop := NewEventLoop()
	
	ctx, cancel := context.WithCancel(context.Background())
	
	go loop.Run(ctx)
	
	// 立即取消上下文
	cancel()
	
	// 给一些时间让循环停止
	time.Sleep(100 * time.Millisecond)
	
	// 验证循环已停止（这里主要是验证不会panic）
	// 实际的停止验证需要更复杂的机制
}

func TestConcurrentMessageProcessing(t *testing.T) {
	loop := NewEventLoop()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go loop.Run(ctx)
	
	// 注册处理器
	processedCount := 0
	loop.RegisterHandler("concurrent", func(msg *Message) error {
		processedCount++
		return nil
	})
	
	// 并发发送多条消息
	const messageCount = 100
	for i := 0; i < messageCount; i++ {
		msg := &Message{
			ID:   "msg-" + string(rune(i)),
			Type: "concurrent",
			Payload: map[string]interface{}{
				"index": i,
			},
		}
		err := loop.SendMessage(msg)
		if err != nil {
			t.Fatalf("Failed to send message %d: %v", i, err)
		}
	}
	
	// 等待所有消息处理完成
	time.Sleep(2 * time.Second)
	
	if processedCount != messageCount {
		t.Errorf("Expected %d messages processed, got %d", messageCount, processedCount)
	}
}