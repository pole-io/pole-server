package config

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	regexp "github.com/dlclark/regexp2"
	"github.com/google/uuid"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/utils"
	storeapi "github.com/pole-io/pole-server/apis/store"
	commonapi "github.com/pole-io/pole-server/pkg/common/api/v1"
	matchs "github.com/pole-io/pole-server/pkg/common/utils/match"
	configtemplate "github.com/pole-io/pole-server/pkg/config/template"
)

func (s *Server) PublishConfigTemplateRelease(
	ctx context.Context, req *apiconfig.ConfigTemplateRelease) *apimodel.Response {
	if req == nil || req.GetTemplateId() == 0 || req.GetName() == "" {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	draft, err := s.storage.GetConfigFileTemplate(req.GetName())
	if err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if draft == nil || draft.Id != req.GetTemplateId() {
		return commonapi.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	draftSpec := conftypes.ToConfigFileTemplateAPI(draft)
	releaseComment := strings.TrimSpace(req.GetComment())
	if releaseComment == "" {
		releaseComment = draftSpec.GetComment()
	}
	req.Content = draftSpec.GetContent()
	req.Format = draftSpec.GetFormat()
	req.Engine = draftSpec.GetEngine()
	req.ParameterSchema = draftSpec.GetParameterSchema()
	req.Comment = releaseComment
	if !supportedTemplateEngine(req.GetEngine()) {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	if req.GetFormat() == "" {
		req.Format = "text"
	}
	validation := s.PreviewConfigTemplate(ctx, &apiconfig.RenderPreviewRequest{
		Input: &apiconfig.ConfigTemplateRenderInput{
			Content: req.GetContent(), Format: req.GetFormat(), Engine: req.GetEngine(),
			ParameterSchema: req.GetParameterSchema(),
			Values:          templateValidationValues(req.GetParameterSchema()),
		},
	})
	if !validation.GetValid() {
		return commonapi.NewAnyDataResponse(apimodel.Code_BadRequest, validation)
	}
	releases, err := s.storage.ListConfigTemplateReleases(req.GetTemplateId())
	if err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if req.GetVersion() == 0 {
		req.Version = nextTemplateVersion(releases)
	}
	if req.GetId() == "" {
		req.Id = uuid.NewString()
	}
	sum := sha256.Sum256([]byte(req.GetContent()))
	req.ContentSha256 = hex.EncodeToString(sum[:])
	parameterSchema, err := conftypes.EncodeTemplateParameterSchema(req.GetParameterSchema())
	if err != nil {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	data := &conftypes.ConfigTemplateRelease{
		ID:              req.GetId(),
		TemplateID:      req.GetTemplateId(),
		Name:            req.GetName(),
		Content:         req.GetContent(),
		Format:          req.GetFormat(),
		ParameterSchema: parameterSchema,
		Engine:          req.GetEngine().GetName(),
		EngineVersion:   req.GetEngine().GetVersion(),
		Version:         req.GetVersion(),
		ContentSHA256:   req.GetContentSha256(),
		Comment:         req.GetComment(),
		CreateBy:        req.GetCreateBy(),
	}
	if err := s.storage.CreateConfigTemplateRelease(data); err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	return commonapi.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

func (s *Server) SaveNamespaceTemplateValues(
	_ context.Context, req *apiconfig.NamespaceTemplateValues) *apimodel.Response {
	if req == nil || req.GetNamespace() == "" || req.GetTemplateId() == 0 {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	if req.GetId() == "" {
		req.Id = uuid.NewString()
	}
	plainValuesPayload, err := conftypes.EncodeTemplateValues(req.GetValues())
	if err != nil {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	schema, err := s.templateParameterSchema(req.GetTemplateId())
	if err != nil {
		return commonapi.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	valuesPayload, err := s.encodeTemplateValuesForStorage(req.GetValues(), schema)
	if err != nil {
		return commonapi.NewConfigResponseWithInfo(apimodel.Code_EncryptConfigFileException, err.Error())
	}
	req.Revision = stableRevision(plainValuesPayload)
	data := &conftypes.NamespaceTemplateValues{
		ID:         req.GetId(),
		Namespace:  req.GetNamespace(),
		TemplateID: req.GetTemplateId(),
		Values:     valuesPayload,
		Revision:   req.GetRevision(),
		ModifyBy:   req.GetModifyBy(),
	}
	if err := s.storage.SaveNamespaceTemplateValues(data); err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	return commonapi.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

func (s *Server) PublishNamespaceTemplateValueRelease(
	ctx context.Context, req *apiconfig.NamespaceTemplateValueRelease) *apimodel.Response {
	if req == nil || req.GetNamespace() == "" || req.GetTemplateId() == 0 ||
		req.GetTemplateReleaseId() == "" {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	templateRelease, err := s.storage.GetConfigTemplateRelease(req.GetTemplateReleaseId())
	if err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if templateRelease == nil || templateRelease.TemplateID != req.GetTemplateId() {
		return commonapi.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	specTemplate, err := configTemplateReleaseToSpec(templateRelease)
	if err != nil {
		return commonapi.NewConfigResponse(apimodel.Code_ExecuteException)
	}
	preview := s.PreviewConfigTemplate(ctx, &apiconfig.RenderPreviewRequest{
		Input: &apiconfig.ConfigTemplateRenderInput{
			Content:         specTemplate.GetContent(),
			Format:          specTemplate.GetFormat(),
			Engine:          specTemplate.GetEngine(),
			ParameterSchema: specTemplate.GetParameterSchema(),
			Values:          req.GetValues(),
		},
		TemplateReleaseId: templateRelease.ID,
		ValueReleaseId:    req.GetId(),
	})
	if !preview.GetValid() {
		return commonapi.NewAnyDataResponse(apimodel.Code_BadRequest, preview)
	}

	releases, err := s.storage.ListNamespaceTemplateValueReleases(req.GetNamespace(), req.GetTemplateId())
	if err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if req.GetVersion() == 0 {
		req.Version = nextValueVersion(releases)
	}
	if req.GetId() == "" {
		req.Id = uuid.NewString()
	}
	if req.GetValuesId() == "" {
		req.ValuesId = req.GetNamespace() + "@" + strconv.FormatUint(req.GetTemplateId(), 10)
	}
	plainValuesPayload, err := conftypes.EncodeTemplateValues(req.GetValues())
	if err != nil {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	valuesPayload, err := s.encodeTemplateValuesForStorage(req.GetValues(), specTemplate.GetParameterSchema())
	if err != nil {
		return commonapi.NewConfigResponseWithInfo(apimodel.Code_EncryptConfigFileException, err.Error())
	}
	req.Revision = stableRevision(req.GetId(), req.GetTemplateReleaseId(), plainValuesPayload)
	data := &conftypes.NamespaceTemplateValueRelease{
		ID:                req.GetId(),
		ValuesID:          req.GetValuesId(),
		Namespace:         req.GetNamespace(),
		TemplateID:        req.GetTemplateId(),
		TemplateReleaseID: req.GetTemplateReleaseId(),
		Values:            valuesPayload,
		ReleaseType:       valueReleaseTypeFromSpec(req.GetReleaseType()),
		BetaLabels:        req.GetBetaLabels(),
		Priority:          int32(req.GetPriority()),
		Active:            req.GetActive(),
		Version:           req.GetVersion(),
		Revision:          req.GetRevision(),
		Comment:           req.GetComment(),
		CreateBy:          req.GetCreateBy(),
	}
	if err := s.storage.CreateNamespaceTemplateValueRelease(data); err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	return commonapi.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

func (s *Server) BindConfigFileTemplate(
	ctx context.Context, file *apiconfig.ConfigFile) *apimodel.Response {
	binding := file.GetTemplateBinding()
	if file == nil || file.GetNamespace() == "" || file.GetGroup() == "" || file.GetName() == "" ||
		file.GetConfigType() != apiconfig.ConfigFile_CONFIG_TEMPLATE || binding == nil ||
		binding.GetTemplateId() == 0 || binding.GetTemplateReleaseId() == "" {
		return commonapi.NewConfigResponse(apimodel.Code_BadRequest)
	}
	templateRelease, err := s.storage.GetConfigTemplateRelease(binding.GetTemplateReleaseId())
	if err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if templateRelease == nil || templateRelease.TemplateID != binding.GetTemplateId() {
		return commonapi.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	bindings, err := s.storage.ListConfigTemplateBindings(&conftypes.ConfigFileKey{
		Namespace: file.GetNamespace(), Group: file.GetGroup(), Name: file.GetName(),
	})
	if err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if binding.GetBindingReleaseId() == "" {
		binding.BindingReleaseId = uuid.NewString()
	}
	data := &conftypes.ConfigTemplateBinding{
		BindingReleaseID:  binding.GetBindingReleaseId(),
		Namespace:         file.GetNamespace(),
		Group:             file.GetGroup(),
		FileName:          file.GetName(),
		TemplateID:        binding.GetTemplateId(),
		TemplateReleaseID: binding.GetTemplateReleaseId(),
		Active:            false,
		Version:           uint64(len(bindings) + 1),
	}
	tx, err := s.storage.StartTx()
	if err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	defer func() { _ = tx.Rollback() }()
	storedFile, err := s.storage.LockConfigFile(tx, &conftypes.ConfigFileKey{
		Namespace: file.GetNamespace(), Group: file.GetGroup(), Name: file.GetName(),
	})
	if err != nil || storedFile == nil {
		return commonapi.NewConfigResponse(firstCode(err, apimodel.Code_NotFoundResource))
	}
	file.ConfigType = apiconfig.ConfigFile_CONFIG_TEMPLATE
	file.TemplateBinding = binding
	updateData, _ := s.updateConfigFileAttribute(storedFile, conftypes.ToConfigFileStore(file))
	if errResp := s.chains.BeforeUpdateFile(ctx, updateData); errResp != nil {
		return errResp
	}
	if err := s.storage.CreateConfigTemplateBindingTx(tx, data); err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if err := s.storage.UpdateConfigFileTx(tx, updateData); err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if err := tx.Commit(); err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	s.RecordHistory(ctx, configFileRecordEntry(ctx, file, types.OUpdate))
	return commonapi.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, binding)
}

func (s *Server) ListConfigTemplateReleases(
	_ context.Context, templateID uint64) *apimodel.BatchQueryResponse {
	releases, err := s.storage.ListConfigTemplateReleases(templateID)
	if err != nil {
		return commonapi.NewConfigBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}
	out := commonapi.NewConfigBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	for _, release := range releases {
		item, err := configTemplateReleaseToSpec(release)
		if err != nil {
			return commonapi.NewConfigBatchQueryResponse(apimodel.Code_ExecuteException)
		}
		if err := commonapi.AddAnyDataIntoBatchQuery(out, item); err != nil {
			return commonapi.NewConfigBatchQueryResponse(apimodel.Code_ExecuteException)
		}
	}
	out.Amount, out.Size = uint32(len(releases)), uint32(len(releases))
	return out
}

func (s *Server) GetNamespaceTemplateValues(
	_ context.Context, namespace string, templateID uint64) *apimodel.Response {
	values, err := s.storage.GetNamespaceTemplateValues(namespace, templateID)
	if err != nil {
		return commonapi.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if values == nil {
		return commonapi.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	item, err := s.namespaceTemplateValuesToSpec(values)
	if err != nil {
		return commonapi.NewConfigResponse(apimodel.Code_ExecuteException)
	}
	return commonapi.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, item)
}

func (s *Server) ListNamespaceTemplateValueReleases(
	_ context.Context, namespace string, templateID uint64) *apimodel.BatchQueryResponse {
	releases, err := s.storage.ListNamespaceTemplateValueReleases(namespace, templateID)
	if err != nil {
		return commonapi.NewConfigBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}
	out := commonapi.NewConfigBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	for _, release := range releases {
		item, err := s.namespaceTemplateValueReleaseToSpec(release)
		if err != nil {
			return commonapi.NewConfigBatchQueryResponse(apimodel.Code_ExecuteException)
		}
		if err := commonapi.AddAnyDataIntoBatchQuery(out, item); err != nil {
			return commonapi.NewConfigBatchQueryResponse(apimodel.Code_ExecuteException)
		}
	}
	out.Amount, out.Size = uint32(len(releases)), uint32(len(releases))
	return out
}

func (s *Server) ListConfigTemplateBindings(
	_ context.Context, namespace, group, fileName string) *apimodel.BatchQueryResponse {
	bindings, err := s.storage.ListConfigTemplateBindings(&conftypes.ConfigFileKey{
		Namespace: namespace,
		Group:     group,
		Name:      fileName,
	})
	if err != nil {
		return commonapi.NewConfigBatchQueryResponse(storeapi.StoreCode2APICode(err))
	}
	out := commonapi.NewConfigBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	for _, binding := range bindings {
		item := configTemplateBindingToSpec(binding)
		if err := commonapi.AddAnyDataIntoBatchQuery(out, item); err != nil {
			return commonapi.NewConfigBatchQueryResponse(apimodel.Code_ExecuteException)
		}
	}
	out.Amount, out.Size = uint32(len(bindings)), uint32(len(bindings))
	return out
}

func (s *Server) resolveTemplateSnapshot(ctx context.Context, file *conftypes.ConfigFileKey,
	bindingSpec *apiconfig.ConfigTemplateBinding, labels map[string]string) (*apiconfig.RenderSnapshot, error) {
	if bindingSpec == nil {
		return nil, fmt.Errorf("template binding not found in matched config release")
	}
	binding := &conftypes.ConfigTemplateBinding{
		BindingReleaseID:  bindingSpec.GetBindingReleaseId(),
		Namespace:         file.Namespace,
		Group:             file.Group,
		FileName:          file.Name,
		TemplateID:        bindingSpec.GetTemplateId(),
		TemplateReleaseID: bindingSpec.GetTemplateReleaseId(),
	}
	var err error
	templateRelease, err := s.storage.GetConfigTemplateRelease(binding.TemplateReleaseID)
	if err != nil || templateRelease == nil {
		return nil, firstError(err, fmt.Errorf("template release %q not found", binding.TemplateReleaseID))
	}
	releases, err := s.storage.ListNamespaceTemplateValueReleases(file.Namespace, binding.TemplateID)
	if err != nil {
		return nil, err
	}
	valueRelease := selectTemplateValueRelease(releases, binding.TemplateReleaseID, labels)
	if valueRelease == nil {
		return nil, fmt.Errorf("active Namespace template Value release not found")
	}
	specTemplate, err := configTemplateReleaseToSpec(templateRelease)
	if err != nil {
		return nil, err
	}
	specValue, err := s.namespaceTemplateValueReleaseToSpec(valueRelease)
	if err != nil {
		return nil, err
	}
	// Gray rules are a server-side concern. SDKs receive only the selected Value
	// snapshot and never re-run audience matching.
	specValue.BetaLabels = nil
	preview := s.PreviewConfigTemplate(ctx, &apiconfig.RenderPreviewRequest{
		Input: &apiconfig.ConfigTemplateRenderInput{
			Content:         specTemplate.GetContent(),
			Format:          specTemplate.GetFormat(),
			Engine:          specTemplate.GetEngine(),
			ParameterSchema: specTemplate.GetParameterSchema(),
			Values:          specValue.GetValues(),
		},
		TemplateReleaseId: templateRelease.ID,
		ValueReleaseId:    valueRelease.ID,
	})
	if !preview.GetValid() {
		return nil, fmt.Errorf("selected template snapshot failed reference rendering")
	}
	return &apiconfig.RenderSnapshot{
		TemplateBinding: &apiconfig.ConfigTemplateBinding{
			TemplateId:        binding.TemplateID,
			TemplateReleaseId: binding.TemplateReleaseID,
			BindingReleaseId:  binding.BindingReleaseID,
		},
		TemplateRelease: specTemplate,
		ValueRelease:    specValue,
		Revision: stableRevision(binding.BindingReleaseID, templateRelease.ID, valueRelease.ID,
			templateRelease.Engine, templateRelease.EngineVersion),
		ExpectedRenderedSha256: preview.GetRenderedSha256(),
	}, nil
}

func selectTemplateValueRelease(releases []*conftypes.NamespaceTemplateValueRelease,
	templateReleaseID string, labels map[string]string) *conftypes.NamespaceTemplateValueRelease {
	var normal *conftypes.NamespaceTemplateValueRelease
	for _, release := range releases {
		if release == nil || !release.Active || release.TemplateReleaseID != templateReleaseID {
			continue
		}
		if release.ReleaseType == conftypes.TemplateValueReleaseTypeGray {
			if matchTemplateLabels(release.BetaLabels, labels) {
				return release
			}
			continue
		}
		if normal == nil {
			normal = release
		}
	}
	return normal
}

func matchTemplateLabels(rule []*apimodel.ClientLabel, labels map[string]string) bool {
	if len(rule) == 0 {
		return false
	}
	for _, clientLabel := range rule {
		actual, ok := labels[clientLabel.GetKey()]
		if !ok || !matchs.MatchString(actual, clientLabel.GetValue(), func(pattern string) *regexp.Regexp {
			compiled, _ := regexp.Compile(pattern, regexp.RE2)
			return compiled
		}) {
			return false
		}
	}
	return true
}

func configTemplateReleaseToSpec(in *conftypes.ConfigTemplateRelease) (*apiconfig.ConfigTemplateRelease, error) {
	out := &apiconfig.ConfigTemplateRelease{}
	if in.ParameterSchema != "" {
		var err error
		out.ParameterSchema, err = conftypes.DecodeTemplateParameterSchema(in.ParameterSchema)
		if err != nil {
			return nil, err
		}
	}
	out.Id, out.TemplateId, out.Name, out.Content, out.Format = in.ID, in.TemplateID, in.Name, in.Content, in.Format
	out.Version, out.ContentSha256, out.Comment, out.CreateBy = in.Version, in.ContentSHA256, in.Comment, in.CreateBy
	out.Ctime = utils.Time2String(in.CreateTime)
	out.Engine = &apiconfig.ConfigTemplateEngine{Name: in.Engine, Version: in.EngineVersion}
	return out, nil
}

func (s *Server) namespaceTemplateValuesToSpec(
	in *conftypes.NamespaceTemplateValues) (*apiconfig.NamespaceTemplateValues, error) {
	values, err := s.decodeTemplateValuesFromStorage(in.Values)
	if err != nil {
		return nil, err
	}
	return &apiconfig.NamespaceTemplateValues{
		Id:         in.ID,
		Namespace:  in.Namespace,
		TemplateId: in.TemplateID,
		Values:     values,
		Revision:   in.Revision,
		Ctime:      utils.Time2String(in.CreateTime),
		Mtime:      utils.Time2String(in.ModifyTime),
		ModifyBy:   in.ModifyBy,
	}, nil
}

func (s *Server) namespaceTemplateValueReleaseToSpec(
	in *conftypes.NamespaceTemplateValueRelease) (*apiconfig.NamespaceTemplateValueRelease, error) {
	out := &apiconfig.NamespaceTemplateValueRelease{}
	values, err := s.decodeTemplateValuesFromStorage(in.Values)
	if err != nil {
		return nil, err
	}
	out.Values = values
	out.Id, out.ValuesId, out.Namespace = in.ID, in.ValuesID, in.Namespace
	out.TemplateId, out.TemplateReleaseId = in.TemplateID, in.TemplateReleaseID
	out.ReleaseType = valueReleaseTypeToSpec(in.ReleaseType)
	out.BetaLabels, out.Priority, out.Active = in.BetaLabels, uint32(in.Priority), in.Active
	out.Version, out.Revision, out.Comment, out.CreateBy = in.Version, in.Revision, in.Comment, in.CreateBy
	out.Ctime = utils.Time2String(in.CreateTime)
	return out, nil
}

func configTemplateBindingToSpec(in *conftypes.ConfigTemplateBinding) *apiconfig.ConfigTemplateBinding {
	if in == nil {
		return nil
	}
	return &apiconfig.ConfigTemplateBinding{
		TemplateId:        in.TemplateID,
		TemplateReleaseId: in.TemplateReleaseID,
		BindingReleaseId:  in.BindingReleaseID,
	}
}

func supportedTemplateEngine(engine *apiconfig.ConfigTemplateEngine) bool {
	return engine != nil && engine.GetName() == configtemplate.EnginePoleMustache &&
		engine.GetVersion() == configtemplate.EngineVersionV1
}

func templateValidationValues(
	schema []*apiconfig.ConfigTemplateParameterSchema) map[string]*apiconfig.ConfigTemplateValue {
	values := make(map[string]*apiconfig.ConfigTemplateValue, len(schema))
	for _, parameter := range schema {
		if parameter == nil {
			continue
		}
		value := &apiconfig.ConfigTemplateValue{}
		switch parameter.GetType() {
		case apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_STRING:
			value.Value = &apiconfig.ConfigTemplateValue_StringValue{StringValue: "validation"}
		case apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_BOOLEAN:
			value.Value = &apiconfig.ConfigTemplateValue_BooleanValue{BooleanValue: false}
		case apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_INTEGER:
			value.Value = &apiconfig.ConfigTemplateValue_IntegerValue{IntegerValue: 0}
		case apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_DECIMAL:
			value.Value = &apiconfig.ConfigTemplateValue_DecimalValue{DecimalValue: "0"}
		}
		values[parameter.GetName()] = value
	}
	return values
}

func nextTemplateVersion(releases []*conftypes.ConfigTemplateRelease) uint64 {
	var max uint64
	for _, release := range releases {
		if release != nil && release.Version > max {
			max = release.Version
		}
	}
	return max + 1
}

func nextValueVersion(releases []*conftypes.NamespaceTemplateValueRelease) uint64 {
	var max uint64
	for _, release := range releases {
		if release != nil && release.Version > max {
			max = release.Version
		}
	}
	return max + 1
}

func valueReleaseTypeFromSpec(value apiconfig.NamespaceTemplateValueReleaseType) conftypes.TemplateValueReleaseType {
	if value == apiconfig.NamespaceTemplateValueReleaseType_TEMPLATE_VALUE_RELEASE_GRAY {
		return conftypes.TemplateValueReleaseTypeGray
	}
	return conftypes.TemplateValueReleaseTypeNormal
}

func valueReleaseTypeToSpec(value conftypes.TemplateValueReleaseType) apiconfig.NamespaceTemplateValueReleaseType {
	if value == conftypes.TemplateValueReleaseTypeGray {
		return apiconfig.NamespaceTemplateValueReleaseType_TEMPLATE_VALUE_RELEASE_GRAY
	}
	return apiconfig.NamespaceTemplateValueReleaseType_TEMPLATE_VALUE_RELEASE_NORMAL
}

func stableRevision(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte(part))
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}

func firstCode(err error, fallback apimodel.Code) apimodel.Code {
	if err != nil {
		return storeapi.StoreCode2APICode(err)
	}
	return fallback
}
