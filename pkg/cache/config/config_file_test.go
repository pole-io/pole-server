package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenPebbleCache_DefaultPathUsesPoleData(t *testing.T) {
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

	db, err := openPebbleCache(map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, ".pole_data", "cache", "config", "config_file.pebble")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "data")); !os.IsNotExist(err) {
		t.Fatalf("legacy data directory should not be created, stat err: %v", err)
	}
}
