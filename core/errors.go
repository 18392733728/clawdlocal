package core

import "errors"

var (
	ErrEventLoopNotRunning = errors.New("event loop is not running")
	ErrEventQueueFull      = errors.New("event queue is full")
	ErrInvalidEvent        = errors.New("invalid event")
	ErrHandlerNotFound     = errors.New("no handler found for event type")
	ErrMessageQueueClosed  = errors.New("message queue is closed")
	ErrInvalidMessageType  = errors.New("invalid message type")
)