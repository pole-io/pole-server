package discover

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	restful "github.com/emicklei/go-restful/v3"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/goverrule"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

type fakeFaultDetectReleaseRuleServer struct {
	goverrule.GoverRuleServer

	releaseCalled     bool
	faultDetectCalled bool
	releaseFilter     map[string]string
}

func (f *fakeFaultDetectReleaseRuleServer) GetRuleReleases(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse {
	f.releaseCalled = true
	f.releaseFilter = filter
	return api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
}

func (f *fakeFaultDetectReleaseRuleServer) GetFaultDetectRules(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse {
	f.faultDetectCalled = true
	return api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
}

func TestGetPublishFaultDetectRulesUsesRuleReleases(t *testing.T) {
	ruleServer := &fakeFaultDetectReleaseRuleServer{}
	server := &HTTPServer{ruleServer: ruleServer}

	ws := new(restful.WebService)
	ws.Path("/naming/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	server.addFaultDetectRuleAccess(ws)

	container := restful.NewContainer()
	container.Add(ws)

	req := httptest.NewRequest(http.MethodGet, "/naming/v1/faultdetectors/releases?id=fault-id&offset=0&limit=10", nil)
	rsp := httptest.NewRecorder()
	container.ServeHTTP(rsp, req)

	if rsp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rsp.Code)
	}
	if !ruleServer.releaseCalled {
		t.Fatalf("expected fault detector releases to call GetRuleReleases")
	}
	if ruleServer.faultDetectCalled {
		t.Fatalf("expected fault detector releases not to call GetFaultDetectRules")
	}
	if got := ruleServer.releaseFilter["rule_id"]; got != "fault-id" {
		t.Fatalf("expected rule_id fault-id, got %q", got)
	}
	if got := ruleServer.releaseFilter["resource"]; got != apimodel.RuleRelease_FaultDetectRules.String() {
		t.Fatalf("expected FaultDetectRules resource, got %q", got)
	}
}
