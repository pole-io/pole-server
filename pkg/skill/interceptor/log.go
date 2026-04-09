/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package interceptor

import (
	"context"

	"go.uber.org/zap"

	"github.com/pole-io/pole-server/pkg/common/log"
)

// skillInterceptorLog 拦截器日志记录器
type skillInterceptorLog struct {
	logger *log.Scope
}

// Info 记录信息日志
func (l *skillInterceptorLog) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Error 记录错误日志
func (l *skillInterceptorLog) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Debug 记录调试日志
func (l *skillInterceptorLog) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Warn 记录警告日志
func (l *skillInterceptorLog) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// WithContext 添加上下文
func (l *skillInterceptorLog) WithContext(ctx context.Context) *skillInterceptorLog {
	return &skillInterceptorLog{
		logger: l.logger,
	}
}

// NewSkillInterceptorLog 创建新的日志记录器
func NewSkillInterceptorLog() *skillInterceptorLog {
	return &skillInterceptorLog{
		logger: log.RegisterScope("skill_interceptor", "Skill interceptor logging", 0),
	}
}
