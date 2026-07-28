package store

import "github.com/pole-io/pole-server/pkg/console/internal/common/model"

// Store 通用存储接口
type ObserverStore interface {
	// Name 存储层的名字
	Name() string
	// Initialize 存储的初始化函数
	Initialize(c *Config) error
	// Destroy 存储的析构函数
	Destroy() error
	HistoryFetcher
	EventFetcher
	SystemSettingsRepository
}

type HistoryFetcher interface {
	// GetHistory 获取历史记录
	GetHistory(filter map[string]string) ([]*model.OperationRecord, error)
}

type EventFetcher interface {
	// GetEvents 获取事件
	GetEvents(filter map[string]string) ([]*model.EventRecord, error)
}
