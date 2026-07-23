package discover

import (
	"testing"

	"google.golang.org/grpc"

	v1 "github.com/pole-io/pole-server/plugin/apiserver/grpcserver/discover/v1"
)

func TestGRPCServerRegistersClientServices(t *testing.T) {
	server := grpc.NewServer()
	t.Cleanup(server.Stop)

	g := &GRPCServer{
		dsvr: v1.NewDiscoverGRPCServer(),
		csvr: v1.NewConfigGRPCServer(),
	}
	g.registerServices(server)

	services := server.GetServiceInfo()
	if _, ok := services["v1.DiscoverGRPC"]; !ok {
		t.Fatalf("v1.DiscoverGRPC not registered: %v", services)
	}
	if _, ok := services["v1.PoleHeartbeatGRPC"]; !ok {
		t.Fatalf("v1.PoleHeartbeatGRPC not registered: %v", services)
	}
	if _, ok := services["v1.ConfigGRPC"]; !ok {
		t.Fatalf("v1.ConfigGRPC not registered: %v", services)
	}
}
