package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// MemoryEventType 定义记忆相关的事件类型
const (
	EventTypeMemoryStore EventType = "memory_store"
	EventTypeMemoryQuery EventType = "memory_query"
	EventTypeMemoryDelete EventType = "memory_delete"
)

// MemoryEventStoreData 记忆存储事件数据
type MemoryEventStoreData struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	TTL       *time.Duration `json:"ttl,omitempty"`
	Tags      []string    `json:"tags,omitempty"`
	Namespace string      `json:"namespace,omitempty"`
}

// MemoryEventQueryData 记忆查询事件数据
type MemoryEventQueryData struct {
	Key       string `json:"key"`
	Namespace string `json:"namespace,omitempty"`
}

// MemoryEventDeleteData 记忆删除事件数据
type MemoryEventDeleteData struct {
	Key       string `json:"key"`
	Namespace string `json:"namespace,omitempty"`
}

// MemoryHandler 处理记忆相关事件
type MemoryHandler struct {
	memoryManager *MemoryManager
	logger        *logrus.Logger
}

// NewMemoryHandler 创建新的记忆处理器
func NewMemoryHandler(memoryManager *MemoryManager, logger *logrus.Logger) *MemoryHandler {
	return &MemoryHandler{
		memoryManager: memoryManager,
		logger:        logger,
	}
}

// Handle 处理记忆事件
func (h *MemoryHandler) Handle(ctx context.Context, event *Event) error {
	switch event.Type {
	case EventTypeMemoryStore:
		return h.handleStore(ctx, event)
	case EventTypeMemoryQuery:
		return h.handleQuery(ctx, event)
	case EventTypeMemoryDelete:
		return h.handleDelete(ctx, event)
	default:
		return fmt.Errorf("unsupported memory event type: %s", event.Type)
	}
}

// CanHandle 检查是否能处理指定事件类型
func (h *MemoryHandler) CanHandle(eventType EventType) bool {
	return eventType == EventTypeMemoryStore ||
		eventType == EventTypeMemoryQuery ||
		eventType == EventTypeMemoryDelete
}

// handleStore 处理记忆存储事件
func (h *MemoryHandler) handleStore(ctx context.Context, event *Event) error {
	var storeData MemoryEventStoreData
	if err := h.decodeEventData(event.Data, &storeData); err != nil {
		return fmt.Errorf("failed to decode store data: %w", err)
	}

	// 构建存储选项
	options := []MemoryOption{}
	if storeData.TTL != nil {
		options = append(options, WithTTL(*storeData.TTL))
	}
	if len(storeData.Tags) > 0 {
		options = append(options, WithTags(storeData.Tags...))
	}
	if storeData.Namespace != "" {
		options = append(options, WithNamespace(storeData.Namespace))
	}

	// 存储到记忆系统
	if err := h.memoryManager.Store(ctx, storeData.Key, storeData.Value, options...); err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
	}

	h.logger.WithField("key", storeData.Key).
		WithField("namespace", storeData.Namespace).
		Debug("Memory stored successfully")

	return nil
}

// handleQuery 处理记忆查询事件
func (h *MemoryHandler) handleQuery(ctx context.Context, event *Event) error {
	var queryData MemoryEventQueryData
	if err := h.decodeEventData(event.Data, &queryData); err != nil {
		return fmt.Errorf("failed to decode query data: %w", err)
	}

	// 查询记忆
	value, err := h.memoryManager.Get(ctx, queryData.Key, WithNamespace(queryData.Namespace))
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}

	// 将结果添加到事件元数据中，供其他处理器使用
	if event.Metadata == nil {
		event.Metadata = make(map[string]interface{})
	}
	event.Metadata["memory_result"] = value

	h.logger.WithField("key", queryData.Key).
		WithField("namespace", queryData.Namespace).
		Debug("Memory queried successfully")

	return nil
}

// handleDelete 处理记忆删除事件
func (h *MemoryHandler) handleDelete(ctx context.Context, event *Event) error {
	var deleteData MemoryEventDeleteData
	if err := h.decodeEventData(event.Data, &deleteData); err != nil {
		return fmt.Errorf("failed to decode delete data: %w", err)
	}

	// 删除记忆
	if err := h.memoryManager.Delete(ctx, deleteData.Key, WithNamespace(deleteData.Namespace)); err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	h.logger.WithField("key", deleteData.Key).
		WithField("namespace", deleteData.Namespace).
		Debug("Memory deleted successfully")

	return nil
}

// decodeEventData 解码事件数据
func (h *MemoryHandler) decodeEventData(data interface{}, target interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	
	if err := json.Unmarshal(dataBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}
	
	return nil
}