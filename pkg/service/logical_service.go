package service

import (
	"context"
	"sort"
	"strings"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
)

func (s *Server) CreateLogicalServices(
	_ context.Context, reqs []*apiservice.LogicalService) *apimodel.BatchWriteResponse {
	out := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range reqs {
		response := s.createLogicalService(req)
		api.Collect(out, response)
	}
	return api.FormatBatchWriteResponse(out)
}

func (s *Server) createLogicalService(req *apiservice.LogicalService) *apimodel.Response {
	_, existing, err := s.storage.ListLogicalServices(req.GetName(), 0, 100)
	if err != nil {
		return api.NewAnyDataResponse(storeapi.StoreCode2APICode(err), req)
	}
	for _, item := range existing {
		if item.Name == req.GetName() {
			return api.NewAnyDataResponse(apimodel.Code_ExistedResource, req)
		}
	}
	if req.GetId() == "" {
		req.Id = utils.NewUUID()
	}
	req.Revision = utils.NewUUID()
	model := &svctypes.LogicalService{
		ID: req.GetId(), Name: req.GetName(), Comment: req.GetComment(), Owner: req.GetOwners(),
		Business: req.GetBusiness(), Department: req.GetDepartment(), Revision: req.GetRevision(), Valid: true,
	}
	if err := s.storage.CreateLogicalService(model); err != nil {
		return api.NewAnyDataResponse(storeapi.StoreCode2APICode(err), req)
	}
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

func (s *Server) UpdateLogicalServices(
	_ context.Context, reqs []*apiservice.LogicalService) *apimodel.BatchWriteResponse {
	out := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range reqs {
		previous := req.GetRevision()
		model := &svctypes.LogicalService{
			ID: req.GetId(), Name: req.GetName(), Comment: req.GetComment(), Owner: req.GetOwners(),
			Business: req.GetBusiness(), Department: req.GetDepartment(), Revision: utils.NewUUID(), Valid: true,
		}
		if err := s.storage.UpdateLogicalService(model, previous); err != nil {
			api.Collect(out, api.NewAnyDataResponse(storeapi.StoreCode2APICode(err), req))
			continue
		}
		req.Revision = model.Revision
		api.Collect(out, api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req))
	}
	return api.FormatBatchWriteResponse(out)
}

func (s *Server) DeleteLogicalServices(
	_ context.Context, reqs []*apiservice.LogicalService) *apimodel.BatchWriteResponse {
	out := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range reqs {
		code := apimodel.Code_ExecuteSuccess
		if err := s.storage.DeleteLogicalService(req.GetId()); err != nil {
			code = storeapi.StoreCode2APICode(err)
		}
		api.Collect(out, api.NewAnyDataResponse(code, req))
	}
	return api.FormatBatchWriteResponse(out)
}

func (s *Server) GetLogicalServices(
	ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	offset, limit, err := valid.ParseOffsetAndLimit(query)
	if err != nil {
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}
	var total uint32
	var items []*svctypes.LogicalService
	if id := query["id"]; id != "" {
		item, getErr := s.storage.GetLogicalService(id)
		if getErr != nil {
			return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(getErr))
		}
		if item != nil {
			total = 1
			items = append(items, item)
		}
	} else {
		total, items, err = s.storage.ListLogicalServices(query["name"], offset, limit)
		if err != nil {
			return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
		}
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	bindings, err := s.storage.ListServiceEnvironmentBindingsByLogicalIDs(ids)
	if err != nil {
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}
	out := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	out.Amount = total
	for _, item := range items {
		spec := item.ToSpec()
		for _, binding := range bindings[item.ID] {
			service := s.caches.Service().GetServiceByID(binding.ServiceID)
			if service == nil || !service.Valid || s.isSystemNamespace(service.Namespace) ||
				!logicalServiceVisible(ctx, service) {
				continue
			}
			count := s.caches.Instance().GetInstancesCountByServiceID(binding.ServiceID)
			spec.EnvironmentCount++
			spec.TotalInstanceCount += count.TotalInstanceCount
			spec.HealthyInstanceCount += count.HealthyInstanceCount
		}
		if err := api.AddAnyDataIntoBatchQuery(out, spec); err != nil {
			return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
		}
	}
	out.Size = uint32(len(out.Data))
	return out
}

func (s *Server) GetLogicalServiceEnvironments(
	ctx context.Context, logicalServiceID string) *apimodel.BatchQueryResponse {
	bindings, err := s.storage.ListServiceEnvironmentBindings(logicalServiceID)
	if err != nil {
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}
	out := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	for _, binding := range bindings {
		service := s.caches.Service().GetServiceByID(binding.ServiceID)
		if service != nil {
			if !service.Valid {
				service = nil
			} else if !logicalServiceVisible(ctx, service) {
				continue
			}
		}
		spec := binding.ToSpec(service)
		if service != nil {
			count := s.caches.Instance().GetInstancesCountByServiceID(binding.ServiceID)
			spec.Service.TotalInstanceCount = count.TotalInstanceCount
			spec.Service.HealthyInstanceCount = count.HealthyInstanceCount
		}
		if err := api.AddAnyDataIntoBatchQuery(out, spec); err != nil {
			return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
		}
	}
	out.Amount = uint32(len(out.Data))
	out.Size = uint32(len(out.Data))
	return out
}

func (s *Server) GetUnboundServiceEnvironments(
	ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	offset, limit, err := valid.ParseOffsetAndLimit(query)
	if err != nil {
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}
	allBindings, err := s.storage.ListServiceEnvironmentBindings("")
	if err != nil {
		return api.NewBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}
	bound := make(map[string]struct{}, len(allBindings))
	for _, binding := range allBindings {
		bound[binding.ServiceID] = struct{}{}
	}
	items := make([]*apiservice.Service, 0)
	_ = s.caches.Service().IteratorServices(func(_ string, service *svctypes.Service) (bool, error) {
		if service == nil || !service.Valid || service.IsAlias() ||
			s.isSystemNamespace(service.Namespace) || !logicalServiceVisible(ctx, service) {
			return true, nil
		}
		if _, ok := bound[service.ID]; ok {
			return true, nil
		}
		if namespace := query["namespace"]; namespace != "" && service.Namespace != namespace {
			return true, nil
		}
		if name := query["name"]; name != "" && !strings.Contains(service.Name, name) {
			return true, nil
		}
		spec := service.ToSpec()
		count := s.caches.Instance().GetInstancesCountByServiceID(service.ID)
		spec.TotalInstanceCount = count.TotalInstanceCount
		spec.HealthyInstanceCount = count.HealthyInstanceCount
		items = append(items, spec)
		return true, nil
	})
	sort.Slice(items, func(i, j int) bool {
		if items[i].GetNamespace() != items[j].GetNamespace() {
			return items[i].GetNamespace() < items[j].GetNamespace()
		}
		if items[i].GetName() != items[j].GetName() {
			return items[i].GetName() < items[j].GetName()
		}
		return items[i].GetId() < items[j].GetId()
	})
	out := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	out.Amount = uint32(len(items))
	start := min(int(offset), len(items))
	end := min(start+int(limit), len(items))
	for _, item := range items[start:end] {
		if err := api.AddAnyDataIntoBatchQuery(out, item); err != nil {
			return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
		}
	}
	out.Size = uint32(len(out.Data))
	return out
}

func (s *Server) GetServiceEnvironmentBinding(
	ctx context.Context, serviceID string) *apimodel.Response {
	binding, err := s.storage.GetServiceEnvironmentBinding(serviceID)
	if err != nil {
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if binding == nil {
		return api.NewResponse(apimodel.Code_ExecuteSuccess)
	}
	service := s.caches.Service().GetServiceByID(binding.ServiceID)
	if service != nil {
		if !service.Valid {
			service = nil
		} else if !logicalServiceVisible(ctx, service) {
			return api.NewResponse(apimodel.Code_NotAllowedAccess)
		}
	}
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, binding.ToSpec(service))
}

func logicalServiceVisible(ctx context.Context, service *svctypes.Service) bool {
	for _, predicate := range cacheapi.LoadServicePredicates(ctx) {
		if !predicate(ctx, service) {
			return false
		}
	}
	return true
}

func (s *Server) isSystemNamespace(name string) bool {
	if name == SystemNamespace {
		return true
	}
	namespace := s.caches.Namespace().GetNamespace(name)
	if namespace != nil {
		return namespace.Kind == apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM
	}
	return false
}

func (s *Server) BindServiceEnvironment(
	_ context.Context, req *apiservice.BindServiceEnvironmentRequest) *apimodel.Response {
	logical, err := s.storage.GetLogicalService(req.GetLogicalServiceId())
	if err != nil {
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	if logical == nil {
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	service := s.caches.Service().GetServiceByID(req.GetServiceId())
	if service == nil || !service.Valid || service.IsAlias() {
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	if s.isSystemNamespace(service.Namespace) {
		return api.NewResponseWithMsg(apimodel.Code_NotAllowedAccess,
			"system namespace services cannot join business logical services")
	}
	binding := &svctypes.ServiceEnvironmentBinding{
		LogicalServiceID: logical.ID, ServiceID: service.ID,
		Namespace: service.Namespace, ServiceName: service.Name,
	}
	if err := s.storage.BindServiceEnvironment(binding, utils.NewUUID()); err != nil {
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, binding.ToSpec(service))
}

func (s *Server) UnbindServiceEnvironment(
	_ context.Context, req *apiservice.UnbindServiceEnvironmentRequest) *apimodel.Response {
	if err := s.storage.UnbindServiceEnvironment(
		req.GetLogicalServiceId(), req.GetServiceId(), utils.NewUUID()); err != nil {
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	return api.NewResponse(apimodel.Code_ExecuteSuccess)
}
