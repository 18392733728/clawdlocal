package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MessageType 定义消息类型
type MessageType string

const (
	MessageTypeUserInput     MessageType = "user_input"
	MessageTypeSystemEvent   MessageType = "system_event"
	MessageTypeToolResponse  MessageType = "tool_response"
	MessageTypeAgentMessage  MessageType = "agent_message"
	MessageTypeExternalEvent MessageType = "external_event"
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
	Priority() int // 优先级，数字越小优先级越高
}

// MessageRouter 路由消息到合适的处理器
type MessageRouter struct {
	handlers map[MessageType][]MessageHandler
	mu       sync.RWMutex
}

// NewMessageRouter 创建新的消息路由器
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		handlers: make(map[MessageType][]MessageHandler),
	}
}

// RegisterHandler 注册消息处理器
func (r *MessageRouter) RegisterHandler(handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, msgType := range []MessageType{
		MessageTypeUserInput,
		MessageTypeSystemEvent,
		MessageTypeToolResponse,
		MessageTypeAgentMessage,
		MessageTypeExternalEvent,
	} {
		if handler.CanHandle(msgType) {
			r.handlers[msgType] = append(r.handlers[msgType], handler)
			// 按优先级排序
			handlers := r.handlers[msgType]
			for i := len(handlers) - 1; i > 0; i-- {
				if handlers[i].Priority() < handlers[i-1].Priority() {
					handlers[i], handlers[i-1] = handlers[i-1], handlers[i]
				} else {
					break
				}
			}
		}
	}
}

// Route 路由消息到所有匹配的处理器
func (r *MessageRouter) Route(ctx context.Context, msg *Message) error {
	r.mu.RLock()
	handlers, exists := r.handlers[msg.Type]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handlers registered for message type: %s", msg.Type)
	}

	var firstErr error
	for _, handler := range handlers {
		if err := handler.Handle(ctx, msg); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			// 继续处理其他处理器，即使有错误
		}
	}

	return firstErr
}

// GetHandlers 获取指定类型的消息处理器
func (r *MessageRouter) GetHandlers(msgType MessageType) []MessageHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handlers, exists := r.handlers[msgType]
	if !exists {
		return nil
	}

	// 返回副本以避免外部修改
	result := make([]MessageHandler, len(handlers))
	copy(result, handlers)
	return result
}

// MessageQueue 消息队列
type MessageQueue struct {
	queue chan *Message
	ctx   context.Context
	cancel context.CancelFunc
	mu    sync.Mutex
	closed bool
}

// NewMessageQueue 创建新的消息队列
func NewMessageQueue(bufferSize int) *MessageQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &MessageQueue{
		queue:  make(chan *Message, bufferSize),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Enqueue 添加消息到队列
func (q *MessageQueue) Enqueue(msg *Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("message queue is closed")
	}

	select {
	case q.queue <- msg:
		return nil
	case <-q.ctx.Done():
		return fmt.Errorf("message queue context cancelled")
	}
}

// Dequeue 从队列获取消息
func (q *MessageQueue) Dequeue() (*Message, error) {
	select {
	case msg := <-q.queue:
		return msg, nil
	case <-q.ctx.Done():
		return nil, fmt.Errorf("message queue context cancelled")
	}
}

// Close 关闭消息队列
func (q *MessageQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.closed {
		q.closed = true
		q.cancel()
		close(q.queue)
	}
}

// IsClosed 检查队列是否已关闭
func (q *MessageQueue) IsClosed() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.closed
}

// Size 获取队列当前大小
func (q *MessageQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return 0
	}
	return len(q.queue)
}