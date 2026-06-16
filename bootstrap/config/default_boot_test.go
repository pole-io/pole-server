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

package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultLoggerOptionsUseLogsDirectory(t *testing.T) {
	options := defaultLoggerOptions()
	require.NotEmpty(t, options)

	for name, option := range options {
		require.Truef(t, strings.HasPrefix(option.RotateOutputPath, "logs/"),
			"%s rotate output should use logs directory, got %q", name, option.RotateOutputPath)
		require.Truef(t, strings.HasPrefix(option.ErrorRotateOutputPath, "logs/"),
			"%s error rotate output should use logs directory, got %q", name, option.ErrorRotateOutputPath)
	}
}

func TestDefaultBootstrapConsoleLoggerUsesLogsDirectory(t *testing.T) {
	bootstrap := defaultBootstrap()

	require.Equal(t, "logs/runtime/pole-console.log", bootstrap.Console.Logger.RotateOutputPath)
	require.Contains(t, bootstrap.Console.Logger.ErrorOutputPaths, "logs/runtime/pole-console-error.log")
}
