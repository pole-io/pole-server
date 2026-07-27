package file

import (
	"strings"
	"testing"
)

func validReportOption() map[string]interface{} {
	return map[string]interface{}{
		"rate-limit-app-name":           "pole-limiter-stat",
		"rate-limit-report-log-path":    "log/rate-limit-report.log",
		"rate-limit-precision-log-path": "log/rate-limit-precision.log",
		"rate-limit-event-log-path":     "log/rate-limit-event.log",
		"server-app-name":               "pole-limiter-server",
		"server-report-log-path":        "log/server-report.log",
		"precision-log-interval":        1,
		"log-interval":                  60,
	}
}

func TestDecodeReportConfigAcceptsKebabCase(t *testing.T) {
	conf, err := decodeReportConfig(validReportOption())
	if err != nil {
		t.Fatalf("decode canonical report config: %v", err)
	}
	if conf.RateLimitAppName != "pole-limiter-stat" {
		t.Fatalf("unexpected rate limit app name: %q", conf.RateLimitAppName)
	}
}

func TestDecodeReportConfigRejectsUnknownKey(t *testing.T) {
	option := validReportOption()
	option["log_size"] = 100

	_, err := decodeReportConfig(option)
	if err == nil {
		t.Fatal("expected unknown plugin option to be rejected")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got: %v", err)
	}
}
