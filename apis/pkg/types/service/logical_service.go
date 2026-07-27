package service

import (
	"time"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/utils"
)

// LogicalService is a control-plane aggregate. Runtime clients never use this
// identity for registration or discovery.
type LogicalService struct {
	ID         string
	Name       string
	Comment    string
	Owner      string
	Business   string
	Department string
	Revision   string
	Valid      bool
	CreateTime time.Time
	ModifyTime time.Time
}

func (s *LogicalService) ToSpec() *apiservice.LogicalService {
	if s == nil {
		return nil
	}
	return &apiservice.LogicalService{
		Id:         s.ID,
		Name:       s.Name,
		Comment:    s.Comment,
		Owners:     s.Owner,
		Business:   s.Business,
		Department: s.Department,
		Revision:   s.Revision,
		Ctime:      utils.Time2String(s.CreateTime),
		Mtime:      utils.Time2String(s.ModifyTime),
		Editable:   true,
		Deleteable: true,
	}
}

type ServiceEnvironmentBinding struct {
	LogicalServiceID string
	ServiceID        string
	Namespace        string
	ServiceName      string
	CreateTime       time.Time
	ModifyTime       time.Time
}

func (b *ServiceEnvironmentBinding) ToSpec(service *Service) *apiservice.ServiceEnvironmentBinding {
	if b == nil {
		return nil
	}
	result := &apiservice.ServiceEnvironmentBinding{
		LogicalServiceId: b.LogicalServiceID,
		ServiceId:        b.ServiceID,
		Namespace:        b.Namespace,
		ServiceName:      b.ServiceName,
		Ctime:            utils.Time2String(b.CreateTime),
		Mtime:            utils.Time2String(b.ModifyTime),
	}
	if service != nil {
		result.Service = service.ToSpec()
		result.Namespace = service.Namespace
		result.ServiceName = service.Name
	}
	return result
}
