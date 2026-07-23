package config_auth

import (
	"context"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	"google.golang.org/protobuf/types/known/anypb"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

func (s *Server) filterConfigGroupQuery(
	ctx context.Context,
	authCtx *authtypes.AcquireContext,
	resp *apimodel.BatchQueryResponse,
) *apimodel.BatchQueryResponse {
	return filterConfigQueryResponse(resp, func(item *anypb.Any) bool {
		group := &apiconfig.ConfigFileGroup{}
		if err := item.UnmarshalTo(group); err != nil {
			return false
		}
		return s.canReadConfigGroup(ctx, authCtx, group.GetNamespace(), group.GetName(), group.GetId(), group.GetMetadata())
	})
}

func (s *Server) filterConfigFileQuery(
	ctx context.Context,
	authCtx *authtypes.AcquireContext,
	resp *apimodel.BatchQueryResponse,
) *apimodel.BatchQueryResponse {
	return filterConfigQueryResponse(resp, func(item *anypb.Any) bool {
		file := &apiconfig.ConfigFile{}
		if err := item.UnmarshalTo(file); err != nil {
			return false
		}
		return s.canReadConfigGroup(ctx, authCtx, file.GetNamespace(), file.GetGroup(), "", nil)
	})
}

func filterConfigQueryResponse(
	resp *apimodel.BatchQueryResponse,
	allowed func(*anypb.Any) bool,
) *apimodel.BatchQueryResponse {
	if resp == nil || resp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		return resp
	}

	visible := make([]*anypb.Any, 0, len(resp.GetData()))
	for _, item := range resp.GetData() {
		if item != nil && allowed(item) {
			visible = append(visible, item)
		}
	}
	resp.Data = visible
	resp.Amount = uint32(len(visible))
	resp.Size = uint32(len(visible))
	return resp
}

func (s *Server) canReadConfigGroup(
	ctx context.Context,
	authCtx *authtypes.AcquireContext,
	namespace, groupName, groupID string,
	metadata map[string]string,
) bool {
	checker := s.policySvr.GetAuthChecker()
	if groupID == "" {
		group := s.cacheMgr.ConfigGroup().GetGroupByName(namespace, groupName)
		if group != nil {
			groupID = group.Id
			metadata = group.Metadata
		}
	}
	if groupID != "" && checker.ResourcePredicate(authCtx, &authtypes.ResourceEntry{
		Type:     apisecurity.ResourceType_ConfigGroups,
		ID:       groupID,
		Metadata: metadata,
	}) {
		return true
	}

	ns := s.cacheMgr.Namespace().GetNamespace(namespace)
	if ns == nil {
		return false
	}
	return checker.ResourcePredicate(authCtx, &authtypes.ResourceEntry{
		Type:     apisecurity.ResourceType_Namespaces,
		ID:       ns.Name,
		Metadata: ns.Metadata,
	})
}
