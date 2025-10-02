package batchctrl

import (
	"errors"
	"time"
)

// CtrlConfig CtrlConfig .
type CtrlConfig struct {
	// Label 批任务执行器标签
	Label string
	// QueueSize 注册请求队列的长度
	QueueSize uint32
	// WaitTime 最长多久一次批量操作
	WaitTime time.Duration
	// MaxBatchCount 每次操作最大的批量数
	MaxBatchCount uint32
	// Concurrency 任务工作协程数量
	Concurrency uint32
	// Handler 任务处理函数
	Handler func(tasks []Future)
}

func (c CtrlConfig) Verify() error {
	if c.Handler == nil {
		return errors.New("Handler is nil")
	}
	if c.Label == "" {
		return errors.New("Label is empty")
	}
	return nil
}
