package service

import (
	"context"
	aipmodel "github.com/pole-io/specification/source/go/api/v1/model"
)

// GetServiceSubscribers implements DiscoverServer.
func (s *Server) GetServiceSubscribers(ctx context.Context, query map[string]string) *aipmodel.BatchQueryResponse {
	panic("unimplemented")
}
