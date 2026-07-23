package sqldb

import (
	"os"
	"strings"
	"testing"
)

func TestConfigFileGroupIncrementalQueryUsesUnixTimestamp(t *testing.T) {
	source, err := os.ReadFile("config_file_group.go")
	if err != nil {
		t.Fatalf("read config_file_group.go: %v", err)
	}

	content := string(source)
	if !strings.Contains(content, "WHERE UNIX_TIMESTAMP(mtime) >= ?") {
		t.Fatalf("config group incremental query must compare UNIX_TIMESTAMP(mtime), avoiding Go/MySQL timezone mismatch")
	}
	if strings.Contains(content, "WHERE mtime >= ?") {
		t.Fatalf("config group incremental query must not compare DATETIME mtime with Go time.Time directly")
	}
}
