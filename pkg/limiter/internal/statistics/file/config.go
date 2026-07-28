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

package file

import (
	"errors"
	"fmt"
)

const (
	minLogInterval             = 1
	minPrecisionLogInterval    = 1
	defaultLogReportLogSize    = 100
	defaultLogReportLogBackups = 20
	defaultLogReportLogMaxAge  = 10
)

// ReportConfig 智研上报配置项
type ReportConfig struct {
	RateLimitAppName          string `json:"rate-limit-app-name"`
	RateLimitReportLogPath    string `json:"rate-limit-report-log-path"`
	RateLimitPrecisionLogPath string `json:"rate-limit-precision-log-path"`
	RateLimitEventLogPath     string `json:"rate-limit-event-log-path"`
	ServerAppName             string `json:"server-app-name"`
	ServerReportLogPath       string `json:"server-report-log-path"`
	PrecisionLogInterval      int    `json:"precision-log-interval"`
	LogInterval               int    `json:"log-interval"`
	LogSize                   int    `json:"log-size"`
	LogBackups                int    `json:"log-backups"`
	LogMaxAge                 int    `json:"log-max-age"`
}

// Validate 校验
func (r *ReportConfig) Validate() error {
	if len(r.RateLimitAppName) == 0 {
		return errors.New("rate-limit-app-name is empty")
	}
	if len(r.ServerAppName) == 0 {
		return errors.New("server-app-name is empty")
	}
	if r.LogInterval < minLogInterval {
		return fmt.Errorf("log-interval must be greater than %d", minLogInterval)
	}
	if r.PrecisionLogInterval < minPrecisionLogInterval {
		return fmt.Errorf("precision-log-interval must be greater than %d", minPrecisionLogInterval)
	}
	if len(r.RateLimitReportLogPath) == 0 {
		return errors.New("rate-limit-report-log-path is empty")
	}
	if len(r.RateLimitPrecisionLogPath) == 0 {
		return errors.New("rate-limit-precision-log-path is empty")
	}
	if len(r.RateLimitEventLogPath) == 0 {
		return errors.New("rate-limit-event-log-path is empty")
	}
	if len(r.ServerReportLogPath) == 0 {
		return errors.New("server-report-log-path is empty")
	}
	if r.LogSize == 0 {
		r.LogSize = defaultLogReportLogSize
	}
	if r.LogBackups == 0 {
		r.LogBackups = defaultLogReportLogBackups
	}
	if r.LogMaxAge == 0 {
		r.LogMaxAge = defaultLogReportLogMaxAge
	}
	return nil
}
