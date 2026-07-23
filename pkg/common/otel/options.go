package otel

import (
	"fmt"
	"strconv"
	"time"
)

func ConfigFromOptions(raw map[string]interface{}) (Config, error) {
	config := Config{
		Endpoint:           stringOption(raw, "endpoint"),
		LogsEndpoint:       stringOption(raw, "logsEndpoint"),
		Compressor:         stringOption(raw, "compressor"),
		ServiceName:        stringOption(raw, "serviceName"),
		ServiceVersion:     stringOption(raw, "serviceVersion"),
		Environment:        stringOption(raw, "environment"),
		Cluster:            stringOption(raw, "cluster"),
		NodeRole:           stringOption(raw, "nodeRole"),
		LogSpoolDir:        stringOption(raw, "logSpoolDir"),
		LogSpoolBucket:     stringOption(raw, "logSpoolBucket"),
		Timeout:            5 * time.Second,
		ReconnectionPeriod: 10 * time.Second,
		PushInterval:       5 * time.Second,
	}
	var err error
	if config.Timeout, err = durationOption(raw, "timeout", config.Timeout); err != nil {
		return Config{}, err
	}
	if config.ReconnectionPeriod, err = durationOption(raw, "reconnectionPeriod", config.ReconnectionPeriod); err != nil {
		return Config{}, err
	}
	if config.PushInterval, err = durationOption(raw, "pushInterval", config.PushInterval); err != nil {
		return Config{}, err
	}
	if rawQueueSize, ok := raw["logQueueSize"]; ok {
		config.LogQueueSize = intOption(rawQueueSize)
	}
	if rawSpoolMaxBytes, ok := raw["logSpoolMaxBytes"]; ok {
		config.LogSpoolMaxBytes = int64Option(rawSpoolMaxBytes)
	}
	config.setDefault()
	return config, nil
}

func stringOption(raw map[string]interface{}, key string) string {
	if raw == nil {
		return ""
	}
	val, ok := raw[key]
	if !ok || val == nil {
		return ""
	}
	return fmt.Sprint(val)
}

func durationOption(raw map[string]interface{}, key string, fallback time.Duration) (time.Duration, error) {
	if raw == nil {
		return fallback, nil
	}
	val, ok := raw[key]
	if !ok || val == nil {
		return fallback, nil
	}
	switch typed := val.(type) {
	case time.Duration:
		return typed, nil
	case string:
		if typed == "" {
			return fallback, nil
		}
		duration, err := time.ParseDuration(typed)
		if err != nil {
			return 0, fmt.Errorf("parse %s duration: %w", key, err)
		}
		return duration, nil
	case int:
		return time.Duration(typed) * time.Second, nil
	case int64:
		return time.Duration(typed) * time.Second, nil
	case float64:
		return time.Duration(typed * float64(time.Second)), nil
	default:
		duration, err := time.ParseDuration(fmt.Sprint(typed))
		if err != nil {
			return 0, fmt.Errorf("parse %s duration: %w", key, err)
		}
		return duration, nil
	}
}

func intOption(val interface{}) int {
	switch typed := val.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(typed)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func int64Option(val interface{}) int64 {
	switch typed := val.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}
