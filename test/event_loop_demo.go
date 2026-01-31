package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// EventType 定义事件类型
type EventType string

const (
	EventTypeMessage EventType = "message"
	EventTypeToolCall EventType = "tool_call"
)

// Event 表示一个事件
type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// EventHandler 定义事件处理器接口
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
	CanHandle(eventType EventType) bool
}

// EventLoop 事件循环核心
type EventLoop struct {
	ctx       context.Context
	cancel    context.CancelFunc
	events    chan *Event
	handlers  []EventHandler
	running   bool
	maxQueue  int
	logger    *logrus.Logger
}

// NewEventLoop 创建新的事件循环
func NewEventLoop(ctx context.Context, logger *logrus.Logger, maxQueueSize int) *EventLoop {
	if maxQueueSize <= 0 {
		maxQueueSize = 1000
	}
	
	ctx, cancel := context.WithCancel(ctx)
	
	return &EventLoop{
		ctx:      ctx,
		cancel:   cancel,
		events:   make(chan *Event, maxQueueSize),
		handlers: make([]EventHandler, 0),
		logger:   logger,
		maxQueue: maxQueueSize,
	}
}

// RegisterHandler 注册事件处理器
func (el *EventLoop) RegisterHandler(handler EventHandler) {
	el.handlers = append(el.handlers, handler)
}

// Start 启动事件循环
func (el *EventLoop) Start() error {
	el.running = true
	go el.run()
	return nil
}

// Stop 停止事件循环
func (el *EventLoop) Stop() {
	el.running = false
	el.cancel()
	close(el.events)
}

// Emit 发送事件到事件循环
func (el *EventLoop) Emit(event *Event) error {
	if !el.running {
		return fmt.Errorf("event loop not running")
	}

	select {
	case el.events <- event:
		return nil
	default:
		return fmt.Errorf("event queue full")
	}
}

// run 事件循环主运行函数
func (el *EventLoop) run() {
	for {
		select {
		case event, ok := <-el.events:
			if !ok {
				return
			}

			// 处理事件
			el.handleEvent(event)

		case <-el.ctx.Done():
			return
		}
	}
}

// handleEvent 处理单个事件
func (el *EventLoop) handleEvent(event *Event) {
	var handled bool
	for _, handler := range el.handlers {
		if handler.CanHandle(event.Type) {
			handler.Handle(el.ctx, event)
			handled = true
		}
	}

	if !handled {
		el.logger.WithField("event_type", event.Type).
			Warn("No handler found for event")
	}
}

// 示例处理器
type MessageHandler struct{}

func (h *MessageHandler) CanHandle(eventType EventType) bool {
	return eventType == EventTypeMessage
}

func (h *MessageHandler) Handle(ctx context.Context, event *Event) error {
	fmt.Printf("MessageHandler handling event: %s\n", event.Data)
	return nil
}

type ToolHandler struct{}

func (h *ToolHandler) CanHandle(eventType EventType) bool {
	return eventType == EventTypeToolCall
}

func (h *ToolHandler) Handle(ctx context.Context, event *Event) error {
	fmt.Printf("ToolHandler handling event: %s\n", event.Data)
	return nil
}

func main() {
	logger := logrus.New()
	ctx := context.Background()
	
	// 创建事件循环
	eventLoop := NewEventLoop(ctx, logger, 100)
	
	// 注册处理器
	eventLoop.RegisterHandler(&MessageHandler{})
	eventLoop.RegisterHandler(&ToolHandler{})
	
	// 启动事件循环
	eventLoop.Start()
	
	// 发送测试事件
	eventLoop.Emit(&Event{
		ID:        "1",
		Type:      EventTypeMessage,
		Timestamp: time.Now(),
		Data:      "Hello, World!",
	})
	
	eventLoop.Emit(&Event{
		ID:        "2", 
		Type:      EventTypeToolCall,
		Timestamp: time.Now(),
		Data:      "read_file",
	})
	
	// 等待处理完成
	time.Sleep(1 * time.Second)
	
	// 停止事件循环
	eventLoop.Stop()
	
	fmt.Println("Event loop test completed!")
}