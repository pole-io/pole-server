package service

import (
	"os"
	"path/filepath"
	"testing"
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
