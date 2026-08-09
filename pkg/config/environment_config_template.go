package config

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	storeapi "github.com/pole-io/pole-server/apis/store"
	commonapi "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	configtemplate "github.com/pole-io/pole-server/pkg/config/template"
)

func (s *Server) GetNamespaceConfigTemplateDraft(
	_ context.Context, namespace string, templateID uint64) *apimodel.Response {
	if namespace == "" || templateID == 0 {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	draft, err := s.getOrInitializeNamespaceConfigTemplateDraft(namespace, templateID)
	if err != nil {
		return commonapi.NewConfigResponseWithInfo(apimodel.Code_NotFoundResource, err.Error())
	}
	return namespaceConfigTemplateDraftResponse(draft)
}

func (s *Server) SaveNamespaceConfigTemplateDraft(ctx context.Context,
	draft *conftypes.NamespaceConfigTemplateDraft) *apimodel.Response {
	if draft == nil || draft.Namespace == "" || draft.TemplateID == 0 || draft.Content == "" {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	if _, err := s.findConfigTemplateByID(draft.TemplateID); err != nil {
		return commonapi.NewConfigResponseWithInfo(apimodel.Code_NotFoundResource, err.Error())
	}
	namespace, err := s.storage.GetNamespace(draft.Namespace)
	if err != nil || namespace == nil {
		return commonapi.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	if namespace.Kind == apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM {
		return commonapi.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "系统空间不能维护配置模板草稿")
	}
	if draft.Format == "" {
		draft.Format = "text"
	}
	if draft.Engine == "" {
		draft.Engine = configtemplate.EnginePoleMustache
	}
	if draft.EngineVersion == "" {
		draft.EngineVersion = configtemplate.EngineVersionV1
	}
	if draft.Engine != configtemplate.EnginePoleMustache || draft.EngineVersion != configtemplate.EngineVersionV1 {
		return commonapi.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "unsupported template engine")
	}
	if _, err := conftypes.DecodeTemplateParameterSchema(draft.ParameterSchema); err != nil {
		return commonapi.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "invalid parameter schema")
	}
	expectedVersion := draft.DraftVersion
	draft.Revision = stableRevision(draft.Content, draft.Format, draft.ParameterSchema,
		draft.Engine, draft.EngineVersion)
	draft.InitializedFrom = strings.TrimSpace(draft.InitializedFrom)
	if draft.InitializedFrom == "" {
		draft.InitializedFrom = "environment-draft"
	}
	draft.ModifyBy = utils.ParseOwnerID(ctx)
	storage, err := s.namespaceConfigTemplateDraftStorage()
	if err != nil {
		return commonapi.NewConfigResponseWithInfo(apimodel.Code_ExecuteException, err.Error())
	}
	if err := storage.SaveNamespaceConfigTemplateDraft(draft, expectedVersion); err != nil {
		if err == storeapi.ErrNamespaceConfigTemplateDraftConflict {
			return commonapi.NewConfigResponseWithInfo(apimodel.Code_ExecuteException,
				"草稿已被其他操作更新，请刷新后重试")
		}
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	updated, err := storage.GetNamespaceConfigTemplateDraft(draft.Namespace, draft.TemplateID)
	if err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	return namespaceConfigTemplateDraftResponse(updated)
}

func (s *Server) getOrInitializeNamespaceConfigTemplateDraft(
	namespace string, templateID uint64) (*conftypes.NamespaceConfigTemplateDraft, error) {
	storage, err := s.namespaceConfigTemplateDraftStorage()
	if err != nil {
		// Compatibility for third-party stores and old test doubles during the
		// rolling upgrade. Production MySQL implements the scoped draft store.
		identity, identityErr := s.findConfigTemplateByID(templateID)
		if identityErr != nil {
			return nil, identityErr
		}
		return &conftypes.NamespaceConfigTemplateDraft{
			Namespace: namespace, TemplateID: templateID, Name: identity.Name,
			Comment: identity.Comment, Content: identity.Content,
			Format: identity.Format, ParameterSchema: identity.ParameterSchema,
			Engine: identity.Engine, EngineVersion: identity.EngineVersion,
			Revision: identity.Revision, InitializedFrom: "legacy-global",
		}, nil
	}
	draft, err := storage.GetNamespaceConfigTemplateDraft(namespace, templateID)
	if err != nil {
		return nil, err
	}
	if draft != nil {
		identity, identityErr := s.findConfigTemplateByID(templateID)
		if identityErr != nil {
			return nil, identityErr
		}
		draft.Name = identity.Name
		draft.Comment = identity.Comment
		return draft, nil
	}
	identity, err := s.findConfigTemplateByID(templateID)
	if err != nil {
		return nil, err
	}

	var snapshot *conftypes.ConfigTemplateRelease
	valueReleases, err := s.storage.ListNamespaceTemplateValueReleases(namespace, templateID)
	if err != nil {
		return nil, err
	}
	var selected *conftypes.NamespaceTemplateValueRelease
	for _, release := range valueReleases {
		if release.ReleaseType == conftypes.TemplateValueReleaseTypeNormal && release.Active &&
			(selected == nil || release.Version > selected.Version) {
			selected = release
		}
	}
	if selected == nil {
		for _, release := range valueReleases {
			if selected == nil || release.Version > selected.Version {
				selected = release
			}
		}
	}
	initializedFrom := "legacy-global"
	if selected != nil {
		snapshot, err = s.storage.GetConfigTemplateRelease(selected.TemplateReleaseID)
		if err != nil {
			return nil, err
		}
		if snapshot != nil {
			initializedFrom = "environment-release:" + selected.ID
		}
	}
	if snapshot != nil {
		return &conftypes.NamespaceConfigTemplateDraft{
			Namespace: namespace, TemplateID: templateID, Name: identity.Name,
			Comment: identity.Comment, Content: snapshot.Content,
			Format: snapshot.Format, ParameterSchema: snapshot.ParameterSchema,
			Engine: snapshot.Engine, EngineVersion: snapshot.EngineVersion,
			Revision: stableRevision(snapshot.Content, snapshot.Format, snapshot.ParameterSchema,
				snapshot.Engine, snapshot.EngineVersion),
			InitializedFrom: initializedFrom,
		}, nil
	}
	return &conftypes.NamespaceConfigTemplateDraft{
		Namespace: namespace, TemplateID: templateID, Name: identity.Name,
		Comment: identity.Comment, Content: identity.Content,
		Format: identity.Format, ParameterSchema: identity.ParameterSchema,
		Engine: identity.Engine, EngineVersion: identity.EngineVersion,
		Revision: identity.Revision, InitializedFrom: initializedFrom,
	}, nil
}

func (s *Server) namespaceConfigTemplateDraftStorage() (storeapi.NamespaceConfigTemplateDraftStore, error) {
	storage, ok := s.storage.(storeapi.NamespaceConfigTemplateDraftStore)
	if !ok || storage == nil {
		return nil, fmt.Errorf("namespace config template draft store is unavailable")
	}
	return storage, nil
}

func namespaceConfigTemplateDraftResponse(draft *conftypes.NamespaceConfigTemplateDraft) *apimodel.Response {
	if draft == nil {
		return commonapi.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	data, err := structpb.NewStruct(map[string]any{
		"namespace": draft.Namespace, "template_id": float64(draft.TemplateID),
		"name": draft.Name, "comment": draft.Comment,
		"content": draft.Content, "format": draft.Format,
		"parameter_schema": draft.ParameterSchema, "engine": draft.Engine,
		"engine_version": draft.EngineVersion, "revision": draft.Revision,
		"draft_version": float64(draft.DraftVersion), "initialized_from": draft.InitializedFrom,
	})
	if err != nil {
		return commonapi.NewConfigResponse(apimodel.Code_ExecuteException)
	}
	return commonapi.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, data)
}
