package types

import (
	"errors"
	"time"
)

var (
	// ErrorNoNamespace 没有找到对应的命名空间
	ErrorNoNamespace error = errors.New("no such namespace")
	// ErrorNoService 没有找到对应的服务
	ErrorNoService error = errors.New("no such service")
)

// Namespace 命名空间结构体
type Namespace struct {
	Name       string
	Comment    string
	Token      string
	Owner      string
	Valid      bool
	CreateTime time.Time
	ModifyTime time.Time
	// ServiceExportTo 服务可见性设置
	ServiceExportTo map[string]struct{}
	// Metadata 命名空间元数据
	Metadata map[string]string
}
