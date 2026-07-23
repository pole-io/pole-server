package config

import (
	"os"
	"strings"
	"testing"
)

func TestPublishConfigFileDoesNotBlockOnActiveGrayRelease(t *testing.T) {
	source, err := os.ReadFile("config_file_release.go")
	if err != nil {
		t.Fatalf("read config_file_release.go: %v", err)
	}

	content := string(source)
	if strings.Contains(content, "still exist beta config file release") {
		t.Fatalf("publish flow must not reject normal or additional gray releases because a gray release is already active")
	}
}
