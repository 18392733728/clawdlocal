package core

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EventType 定义事件类型
type EventType string

const (
	EventTypeMessage   EventType = "message"
	EventTypeToolCall  EventType = "tool_call"
	EventTypeSystem    EventType = "system"
	EventTypeHeartbeat EventType = "heartbeat"
	EventTypeCron      EventType = "cron"
)

// Event 表示一个事件
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
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
	wg        sync.WaitGroup
	logger    *logrus.Logger
	mu        sync.RWMutex
	running   bool
	maxQueue  int
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
	el.mu.Lock()
	defer el.mu.Unlock()
	el.handlers = append(el.handlers, handler)
}

// Start 启动事件循环
func (el *EventLoop) Start() error {
	el.mu.Lock()
	if el.running {
		el.mu.Unlock()
		return nil
	}
	el.running = true
	el.mu.Unlock()

	el.logger.Info("Starting event loop")
	el.wg.Add(1)
	go el.run()

	return nil
}

// Stop 停止事件循环
func (el *EventLoop) Stop() {
	el.mu.Lock()
	if !el.running {
		el.mu.Unlock()
		return
	}
	el.running = false
	el.mu.Unlock()

	el.logger.Info("Stopping event loop")
	el.cancel()
	close(el.events)
	el.wg.Wait()
}

// Emit 发送事件到事件循环
func (el *EventLoop) Emit(event *Event) error {
	el.mu.RLock()
	if !el.running {
		el.mu.RUnlock()
		return ErrEventLoopNotRunning
	}
	el.mu.RUnlock()

	select {
	case el.events <- event:
		return nil
	case <-el.ctx.Done():
		return el.ctx.Err()
	default:
		return ErrEventQueueFull
	}
}

// run 事件循环主运行函数
func (el *EventLoop) run() {
	defer el.wg.Done()

	for {
		select {
		case event, ok := <-el.events:
			if !ok {
				el.logger.Info("Event loop channel closed")
				return
			}

			// 处理事件
			if err := el.handleEvent(event); err != nil {
				el.logger.WithError(err).WithField("event_id", event.ID).
					Error("Failed to handle event")
			}

		case <-el.ctx.Done():
			el.logger.Info("Event loop context cancelled")
			return
		}
	}
}

// handleEvent 处理单个事件
func (el *EventLoop) handleEvent(event *Event) error {
	el.mu.RLock()
	handlers := make([]EventHandler, len(el.handlers))
	copy(handlers, el.handlers)
	el.mu.RUnlock()

	var handled bool
	for _, handler := range handlers {
		if handler.CanHandle(event.Type) {
			if err := handler.Handle(el.ctx, event); err != nil {
				el.logger.WithError(err).WithField("handler", handler).
					WithField("event_type", event.Type).
					Error("Handler failed")
				continue
			}
			handled = true
		}
	}

	if !handled {
		el.logger.WithField("event_type", event.Type).
			Warn("No handler found for event")
	}

	return nil
}

// GetQueueLength 获取当前队列长度
func (el *EventLoop) GetQueueLength() int {
	el.mu.RLock()
	defer el.mu.RUnlock()
	return len(el.events)
}

// IsRunning 检查事件循环是否正在运行
func (el *EventLoop) IsRunning() bool {
	el.mu.RLock()
	defer el.mu.RUnlock()
	return el.running
}