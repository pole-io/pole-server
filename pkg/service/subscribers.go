package service

import (
	"context"

	"github.com/pole-io/specification/source/go/api/v1/service_manage"
)

// GetServiceSubscribers implements DiscoverServer.
func (s *Server) GetServiceSubscribers(ctx context.Context, query map[string]string) *service_manage.BatchQueryResponse {
	panic("unimplemented")
}
