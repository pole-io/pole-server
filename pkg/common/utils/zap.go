package utils

import (
	"context"

	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/pkg/types"
)

// ZapRequestID 生成X-Request-Id的日志描述
func ZapRequestID(id string) zap.Field {
	return zap.String(types.PoleRequestID, id)
}

// RequestID 从ctx中获取X-Request-Id
func RequestID(ctx context.Context) zap.Field {
	return zap.String(types.PoleRequestID, ParseRequestID(ctx))
}

// ZapPlatformID 生成Platform-ID的日志描述
func ZapPlatformID(id string) zap.Field {
	return zap.String("platform-id", id)
}

// ZapInstanceID 生成instanceID的日志描述
func ZapInstanceID(id string) zap.Field {
	return zap.String("instance-id", id)
}

// ZapNamespace 生成namespace的日志描述
func ZapNamespace(namespace string) zap.Field {
	return zap.String("namesapce", namespace)
}

// ZapGroup 生成group的日志描述
func ZapGroup(group string) zap.Field {
	return zap.String("group", group)
}

// ZapFileName 生成fileName的日志描述
func ZapFileName(fileName string) zap.Field {
	return zap.String("file-name", fileName)
}

// ZapReleaseName 生成fileName的日志描述
func ZapReleaseName(fileName string) zap.Field {
	return zap.String("release-name", fileName)
}

// ZapVersion 生成 version 的日志描述
func ZapVersion(version uint64) zap.Field {
	return zap.Uint64("version", version)
}
