package otel

import (
	"fmt"
	"strconv"
	"time"

	commonotel "github.com/pole-io/pole-server/pkg/common/otel"
)

type options struct {
	commonotel.Config
	SetupSDK bool
}

func loadOptions(raw map[string]interface{}) (options, error) {
	opt := options{
		Config: commonotel.Config{
			Endpoint:           stringOption(raw, "endpoint"),
			Compressor:         stringOption(raw, "compressor"),
			ServiceName:        stringOption(raw, "serviceName"),
			ServiceVersion:     stringOption(raw, "serviceVersion"),
			Environment:        stringOption(raw, "environment"),
			Cluster:            stringOption(raw, "cluster"),
			NodeRole:           stringOption(raw, "nodeRole"),
			Timeout:            5 * time.Second,
			ReconnectionPeriod: 10 * time.Second,
			PushInterval:       5 * time.Second,
		},
		SetupSDK: true,
	}

	var err error
	if opt.Timeout, err = durationOption(raw, "timeout", opt.Timeout); err != nil {
		return options{}, err
	}
	if opt.ReconnectionPeriod, err = durationOption(raw, "reconnectionPeriod", opt.ReconnectionPeriod); err != nil {
		return options{}, err
	}
	if opt.PushInterval, err = durationOption(raw, "pushInterval", opt.PushInterval); err != nil {
		return options{}, err
	}
	if rawSetupSDK, ok := raw["setupSDK"]; ok {
		opt.SetupSDK = boolOption(rawSetupSDK)
	}
	return opt, nil
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

func boolOption(val interface{}) bool {
	switch typed := val.(type) {
	case bool:
		return typed
	case string:
		parsed, err := strconv.ParseBool(typed)
		return err == nil && parsed
	default:
		return fmt.Sprint(typed) == "true"
	}
}
