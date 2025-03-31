package types

import (
	"errors"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
)

var (
	// ErrorNoNamespace 没有找到对应的命名空间
	ErrorNoNamespace error = errors.New("no such namespace")
	// ErrorNoService 没有找到对应的服务
	ErrorNoService error = errors.New("no such service")
)

func ExportToMap(exportTo []*wrappers.StringValue) map[string]struct{} {
	ret := make(map[string]struct{})
	for _, v := range exportTo {
		ret[v.Value] = struct{}{}
	}
	return ret
}

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

func (n *Namespace) ListServiceExportTo() []*wrappers.StringValue {
	ret := make([]*wrappers.StringValue, 0, len(n.ServiceExportTo))
	for i := range n.ServiceExportTo {
		ret = append(ret, &wrappers.StringValue{Value: i})
	}
	return ret
}
