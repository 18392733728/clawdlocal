package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MessageType 定义消息类型
type MessageType string

const (
	MessageTypeUserInput    MessageType = "user_input"
	MessageTypeSystemEvent  MessageType = "system_event"
	MessageTypeToolResponse MessageType = "tool_response"
)

// Message 表示系统中的消息
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType           `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Payload   interface{}           `json:"payload"`
	Timestamp time.Time             `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageHandler 处理消息的接口
type MessageHandler interface {
	Handle(ctx context.Context, msg *Message) error
	CanHandle(msgType MessageType) bool
	Priority() int
}

// UserMessageHandler 处理用户输入消息
type UserMessageHandler struct{}

func (h *UserMessageHandler) Handle(ctx context.Context, msg *Message) error {
	fmt.Printf("UserMessageHandler handling message: %v\n", msg.Payload)
	return nil
}

func (h *UserMessageHandler) CanHandle(msgType MessageType) bool {
	return msgType == MessageTypeUserInput
}

func (h *UserMessageHandler) Priority() int {
	return 1
}

// SystemMessageHandler 处理系统事件消息
type SystemMessageHandler struct{}

func (h *SystemMessageHandler) Handle(ctx context.Context, msg *Message) error {
	fmt.Printf("SystemMessageHandler handling message: %v\n", msg.Payload)
	return nil
}

func (h *SystemMessageHandler) CanHandle(msgType MessageType) bool {
	return msgType == MessageTypeSystemEvent
}

func (h *SystemMessageHandler) Priority() int {
	return 2
}

// MessageRouter 路由消息到合适的处理器
type MessageRouter struct {
	handlers map[MessageType][]MessageHandler
}

func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		handlers: make(map[MessageType][]MessageHandler),
	}
}

func (r *MessageRouter) RegisterHandler(handler MessageHandler) {
	for _, msgType := range []MessageType{MessageTypeUserInput, MessageTypeSystemEvent, MessageTypeToolResponse} {
		if handler.CanHandle(msgType) {
			r.handlers[msgType] = append(r.handlers[msgType], handler)
		}
	}
}

func (r *MessageRouter) Route(ctx context.Context, msg *Message) error {
	handlers, exists := r.handlers[msg.Type]
	if !exists {
		return fmt.Errorf("no handlers registered for message type: %s", msg.Type)
	}

	var firstErr error
	for _, handler := range handlers {
		if err := handler.Handle(ctx, msg); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func main() {
	ctx := context.Background()
	router := NewMessageRouter()

	// 注册处理器
	router.RegisterHandler(&UserMessageHandler{})
	router.RegisterHandler(&SystemMessageHandler{})

	// 创建测试消息
	userMsg := &Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeUserInput,
		Source:    "user",
		Target:    "agent",
		Payload:   "Hello, Agent!",
		Timestamp: time.Now(),
	}

	systemMsg := &Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeSystemEvent,
		Source:    "system",
		Target:    "agent",
		Payload:   "System started",
		Timestamp: time.Now(),
	}

	// 路由消息
	fmt.Println("Testing user message routing...")
	err := router.Route(ctx, userMsg)
	if err != nil {
		fmt.Printf("Error routing user message: %v\n", err)
	}

	fmt.Println("Testing system message routing...")
	err = router.Route(ctx, systemMsg)
	if err != nil {
		fmt.Printf("Error routing system message: %v\n", err)
	}

	fmt.Println("Message handler test completed!")
}