package core

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateMessageID 生成唯一的消息ID
func GenerateMessageID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备
		return time.Now().Format("20060102150405.999999999")
	}
	return hex.EncodeToString(bytes)
}

// GetCurrentTimestamp 获取当前时间戳（Unix纳秒）
func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}