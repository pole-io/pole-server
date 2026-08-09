package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/emicklei/go-restful/v3"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	storeapi "github.com/pole-io/pole-server/apis/store"
)

const environmentPromotionSuccessCode = 200000

type publishEnvironmentPromotionTopologyRequest struct {
	ExpectedRevision uint64 `json:"expected_revision"`
	Comment          string `json:"comment"`
}

func (h *HTTPServer) addEnvironmentPromotionAccess(ws *restful.WebService) {
	ws.Route(ws.GET("/environment-promotion/topology").To(h.GetEnvironmentPromotionTopology))
	ws.Route(ws.POST("/environment-promotion/topology/validate").To(h.ValidateEnvironmentPromotionTopology))
	ws.Route(ws.PUT("/environment-promotion/topology").To(h.SaveEnvironmentPromotionTopology))
	ws.Route(ws.POST("/environment-promotion/topology/publish").To(h.PublishEnvironmentPromotionTopology))
	ws.Route(ws.GET("/environment-promotion/topology/revisions").To(h.ListEnvironmentPromotionTopologyRevisions))
}

func (h *HTTPServer) GetEnvironmentPromotionTopology(req *restful.Request, rsp *restful.Response) {
	ctx := initContext(req)
	topology, err := h.environmentPromotionServer.GetEnvironmentPromotionTopology(ctx)
	if err != nil {
		h.writeEnvironmentPromotionError(rsp, err)
		return
	}
	if _, err := h.authorizeEnvironmentPromotion(ctx, topology, authtypes.Read); err != nil {
		_ = rsp.WriteErrorString(http.StatusForbidden, "无权查看环境晋升拓扑")
		return
	}
	h.writeEnvironmentPromotionData(rsp, topology)
}

func (h *HTTPServer) ValidateEnvironmentPromotionTopology(req *restful.Request, rsp *restful.Response) {
	ctx := initContext(req)
	var topology types.EnvironmentPromotionTopology
	if err := req.ReadEntity(&topology); err != nil {
		_ = rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if _, err := h.authorizeEnvironmentPromotion(ctx, &topology, authtypes.Read); err != nil {
		_ = rsp.WriteErrorString(http.StatusForbidden, "无权校验环境晋升拓扑")
		return
	}
	issues := h.environmentPromotionServer.ValidateEnvironmentPromotionTopology(&topology)
	h.writeEnvironmentPromotionData(rsp, map[string]any{"valid": len(issues) == 0, "issues": issues})
}

func (h *HTTPServer) SaveEnvironmentPromotionTopology(req *restful.Request, rsp *restful.Response) {
	ctx := initContext(req)
	var topology types.EnvironmentPromotionTopology
	if err := req.ReadEntity(&topology); err != nil {
		_ = rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	authorizedCtx, err := h.authorizeEnvironmentPromotion(ctx, &topology, authtypes.Modify)
	if err != nil {
		_ = rsp.WriteErrorString(http.StatusForbidden, "无权修改环境晋升拓扑")
		return
	}
	issues, err := h.environmentPromotionServer.SaveEnvironmentPromotionTopology(authorizedCtx, &topology)
	if err != nil {
		h.writeEnvironmentPromotionError(rsp, err)
		return
	}
	if len(issues) != 0 {
		rsp.WriteHeader(http.StatusUnprocessableEntity)
		h.writeEnvironmentPromotionData(rsp, map[string]any{"valid": false, "issues": issues})
		return
	}
	updated, err := h.environmentPromotionServer.GetEnvironmentPromotionTopology(authorizedCtx)
	if err != nil {
		h.writeEnvironmentPromotionError(rsp, err)
		return
	}
	h.writeEnvironmentPromotionData(rsp, updated)
}

func (h *HTTPServer) PublishEnvironmentPromotionTopology(req *restful.Request, rsp *restful.Response) {
	ctx := initContext(req)
	var body publishEnvironmentPromotionTopologyRequest
	if err := req.ReadEntity(&body); err != nil {
		_ = rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	topology, err := h.environmentPromotionServer.GetEnvironmentPromotionTopology(ctx)
	if err != nil {
		h.writeEnvironmentPromotionError(rsp, err)
		return
	}
	authorizedCtx, err := h.authorizeEnvironmentPromotion(ctx, topology, authtypes.Modify)
	if err != nil {
		_ = rsp.WriteErrorString(http.StatusForbidden, "无权发布环境晋升拓扑")
		return
	}
	revision, issues, err := h.environmentPromotionServer.PublishEnvironmentPromotionTopology(
		authorizedCtx, body.ExpectedRevision, body.Comment)
	if err != nil {
		h.writeEnvironmentPromotionError(rsp, err)
		return
	}
	if len(issues) != 0 {
		rsp.WriteHeader(http.StatusUnprocessableEntity)
		h.writeEnvironmentPromotionData(rsp, map[string]any{"valid": false, "issues": issues})
		return
	}
	h.writeEnvironmentPromotionData(rsp, revision)
}

func (h *HTTPServer) ListEnvironmentPromotionTopologyRevisions(req *restful.Request, rsp *restful.Response) {
	ctx := initContext(req)
	topology, err := h.environmentPromotionServer.GetEnvironmentPromotionTopology(ctx)
	if err != nil {
		h.writeEnvironmentPromotionError(rsp, err)
		return
	}
	if _, err := h.authorizeEnvironmentPromotion(ctx, topology, authtypes.Read); err != nil {
		_ = rsp.WriteErrorString(http.StatusForbidden, "无权查看环境晋升拓扑历史")
		return
	}
	revisions, err := h.environmentPromotionServer.ListEnvironmentPromotionTopologyRevisions(ctx)
	if err != nil {
		h.writeEnvironmentPromotionError(rsp, err)
		return
	}
	h.writeEnvironmentPromotionData(rsp, revisions)
}

func (h *HTTPServer) authorizeEnvironmentPromotion(ctx context.Context,
	topology *types.EnvironmentPromotionTopology, operation authtypes.ResourceOperation) (context.Context, error) {
	if h.systemConfigUser == nil || h.systemConfigAuth == nil {
		return ctx, authtypes.ErrorTokenInvalid
	}
	method := authtypes.DescribeNamespaces
	if operation != authtypes.Read {
		method = authtypes.UpdateNamespaces
	}
	resources := make([]authtypes.ResourceEntry, 0)
	seen := map[string]struct{}{}
	appendResource := func(name string) {
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		resources = append(resources, authtypes.ResourceEntry{Type: apisecurity.ResourceType_Namespaces, ID: name})
	}
	if topology != nil {
		for _, edge := range topology.Edges {
			appendResource(edge.Source)
			appendResource(edge.Target)
		}
		for _, binding := range topology.LaneBaseBindings {
			appendResource(binding.Lane)
			appendResource(binding.Base)
		}
	}
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithModule(authtypes.CoreModule),
		authtypes.WithOperation(operation),
		authtypes.WithMethod(method),
		authtypes.WithAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
			apisecurity.ResourceType_Namespaces: resources,
		}),
	)
	if err := h.systemConfigUser.CheckCredential(authCtx); err != nil {
		return ctx, err
	}
	pass, err := h.systemConfigAuth.CheckConsolePermission(authCtx)
	if err != nil {
		return ctx, err
	}
	if !pass {
		return ctx, errors.New("environment promotion permission denied")
	}
	return authCtx.GetRequestContext(), nil
}

func (h *HTTPServer) writeEnvironmentPromotionData(rsp *restful.Response, data any) {
	_ = rsp.WriteAsJson(map[string]any{
		"code": environmentPromotionSuccessCode,
		"info": "execute success",
		"data": data,
	})
}

func (h *HTTPServer) writeEnvironmentPromotionError(rsp *restful.Response, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, storeapi.ErrEnvironmentPromotionRevisionConflict) {
		status = http.StatusConflict
	}
	_ = rsp.WriteErrorString(status, err.Error())
}
