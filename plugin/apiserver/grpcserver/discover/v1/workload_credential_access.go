package v1

import (
	"context"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

func (g *DiscoverGRPCServer) Issue(ctx context.Context,
	req *apisecurity.WorkloadCredentialIssueRequest) (*apisecurity.WorkloadCredentialResponse, error) {
	if g.workloadCredentialServer == nil {
		return unavailableCredentialResponse(), nil
	}
	return g.workloadCredentialServer.Issue(utils.ConvertGRPCContext(ctx), req), nil
}

func (g *DiscoverGRPCServer) Renew(ctx context.Context,
	req *apisecurity.WorkloadCredentialRenewRequest) (*apisecurity.WorkloadCredentialResponse, error) {
	if g.workloadCredentialServer == nil {
		return unavailableCredentialResponse(), nil
	}
	return g.workloadCredentialServer.Renew(utils.ConvertGRPCContext(ctx), req), nil
}

func unavailableCredentialResponse() *apisecurity.WorkloadCredentialResponse {
	resp := api.NewResponse(apimodel.Code_WorkloadCredentialIssuerUnavailable)
	return &apisecurity.WorkloadCredentialResponse{Code: resp.Code, Info: resp.Info}
}
