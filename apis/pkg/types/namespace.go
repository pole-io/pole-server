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

// 注释：ExportToMap函数改动 - 参数类型从[]*wrapperspb.StringValue改为[]string，功能保持不变
func ExportToMap(exportTo []string) map[string]struct{} {
	ret := make(map[string]struct{})
	for _, v := range exportTo {
		// 注释：字段访问简化 - 直接使用string值而非通过.Value访问wrapper
		ret[v] = struct{}{}
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

// 注释：ListServiceExportTo方法改动 - 返回类型从[]*wrapperspb.StringValue改为[]string，简化类型系统
func (n *Namespace) ListServiceExportTo() []string {
	ret := make([]string, 0, len(n.ServiceExportTo))
	for i := range n.ServiceExportTo {
		// 注释：导出列表构建改动 - 直接添加string值而非创建wrapper对象
		ret = append(ret, i)
	}
	return ret
}
