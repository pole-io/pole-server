package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
)

func TestServiceContractOpenPebbleCache_DefaultPathUsesPoleData(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	cache := &ServiceContractCache{}
	db, err := cache.openPebbleCache(map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, ".pole_data", "cache", "service_contract", "service_contract.pebble")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "data")); !os.IsNotExist(err) {
		t.Fatalf("legacy data directory should not be created, stat err: %v", err)
	}
}

func TestServiceContractCacheListReturnsAllVersionsForService(t *testing.T) {
	cache := &ServiceContractCache{}
	if err := cache.Initialize(map[string]interface{}{"cachePath": t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cache.Close()
	})

	cache.setContracts([]*svctypes.EnrichServiceContract{
		{ServiceContract: &svctypes.ServiceContract{
			ID: "http-v1", Namespace: "default", Service: "payments", Type: "openapi",
			Protocol: "http", Version: "v1", Revision: "revision-http", Valid: true,
		}},
		{ServiceContract: &svctypes.ServiceContract{
			ID: "grpc-v2", Namespace: "default", Service: "payments", Type: "protobuf",
			Protocol: "grpc", Version: "v2", Revision: "revision-grpc", Valid: true,
		}},
		{ServiceContract: &svctypes.ServiceContract{
			ID: "other", Namespace: "default", Service: "inventory", Type: "thrift",
			Protocol: "thrift", Version: "v1", Revision: "revision-other", Valid: true,
		}},
	})

	got := cache.List(context.Background(), "default", "payments")
	require.Len(t, got, 2)
	require.Equal(t, "grpc", got[0].Protocol)
	require.Equal(t, "http", got[1].Protocol)
}

func TestServiceContractCacheGetReturnsNilOnMiss(t *testing.T) {
	cache := &ServiceContractCache{}
	if err := cache.Initialize(map[string]interface{}{"cachePath": t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cache.Close()
	})

	got := cache.Get(context.Background(), &svctypes.ServiceContract{
		Namespace: "default",
		Service:   "payments",
		Type:      "openapi",
		Protocol:  "http",
		Version:   "v1",
	})
	if got != nil {
		t.Fatalf("cache miss must return nil, got %#v", got)
	}
}
