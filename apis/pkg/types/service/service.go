package service

import (
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	"google.golang.org/protobuf/types/known/wrapperspb"

	commontime "github.com/pole-io/pole-server/pkg/common/time"
)

type ServicePort struct {
	Port     uint32
	Protocol string
}

// Service 服务数据
type Service struct {
	ID           string
	Name         string
	Namespace    string
	Business     string
	Ports        string
	Meta         map[string]string
	Comment      string
	Department   string
	CmdbMod1     string
	CmdbMod2     string
	CmdbMod3     string
	Token        string
	Owner        string
	Revision     string
	Reference    string
	ReferFilter  string
	PlatformID   string
	Valid        bool
	CreateTime   time.Time
	ModifyTime   time.Time
	Mtime        int64
	Ctime        int64
	ServicePorts []*ServicePort
	// ExportTo 服务可见性暴露设置
	ExportTo    map[string]struct{}
	OldExportTo map[string]struct{}
}

func (s *Service) ToSpec() *apiservice.Service {
	return &apiservice.Service{
		Name:       wrapperspb.String(s.Name),
		Namespace:  wrapperspb.String(s.Namespace),
		Metadata:   s.CopyMeta(),
		Ports:      wrapperspb.String(s.Ports),
		Business:   wrapperspb.String(s.Business),
		Department: wrapperspb.String(s.Department),
		CmdbMod1:   wrapperspb.String(s.CmdbMod1),
		CmdbMod2:   wrapperspb.String(s.CmdbMod2),
		CmdbMod3:   wrapperspb.String(s.CmdbMod3),
		Comment:    wrapperspb.String(s.Comment),
		Owners:     wrapperspb.String(s.Owner),
		Token:      wrapperspb.String(s.Token),
		Ctime:      wrapperspb.String(commontime.Time2String(s.CreateTime)),
		Mtime:      wrapperspb.String(commontime.Time2String(s.ModifyTime)),
		Revision:   wrapperspb.String(s.Revision),
		Id:         wrapperspb.String(s.ID),
		ExportTo:   s.ListExportTo(),
	}
}

func (s *Service) CopyMeta() map[string]string {
	ret := make(map[string]string)
	for k, v := range s.Meta {
		ret[k] = v
	}
	return ret
}

func (s *Service) ProtectThreshold() float32 {
	if len(s.Meta) == 0 {
		return 0
	}
	val := s.Meta[MetadataServiceProtectThreshold]
	threshold, _ := strconv.ParseFloat(val, 32)
	return float32(threshold)
}

func (s *Service) ListExportTo() []*wrappers.StringValue {
	ret := make([]*wrappers.StringValue, 0, len(s.ExportTo))
	for i := range s.ExportTo {
		ret = append(ret, &wrappers.StringValue{Value: i})
	}
	return ret
}

// EnhancedService 服务增强数据
type EnhancedService struct {
	*Service
	TotalInstanceCount   uint32
	HealthyInstanceCount uint32
}

// ServiceKey 服务名
type ServiceKey struct {
	Namespace string
	Name      string
}

func (s *ServiceKey) Equal(o *ServiceKey) bool {
	if s == nil {
		return false
	}
	if o == nil {
		return false
	}
	return s.Name == o.Name && s.Namespace == o.Namespace
}

func (s *ServiceKey) Domain() string {
	return s.Name + "." + s.Namespace
}

// IsAlias 便捷函数封装
func (s *Service) IsAlias() bool {
	return s.Reference != ""
}

// ServiceAlias 服务别名结构体
type ServiceAlias struct {
	ID             string
	Alias          string
	AliasNamespace string
	ServiceID      string
	Service        string
	Namespace      string
	Owner          string
	Comment        string
	CreateTime     time.Time
	ModifyTime     time.Time
	ExportTo       map[string]struct{}
}

func (s *ServiceAlias) ListExportTo() []*wrappers.StringValue {
	ret := make([]*wrappers.StringValue, 0, len(s.ExportTo))
	for i := range s.ExportTo {
		ret = append(ret, &wrappers.StringValue{Value: i})
	}
	return ret
}
