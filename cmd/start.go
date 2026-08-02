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

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/pole-io/pole-server/bootstrap"
	"github.com/pole-io/pole-server/pluginapi"
)

var (
	configFilePath = ""
	startMode      = ""
	pluginRegistry *pluginapi.Registry

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "start running",
		Long:  "start running",
		RunE: func(c *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(
				context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return bootstrap.Run(ctx, bootstrap.Options{
				ConfigPath:     configFilePath,
				ModeOverride:   startMode,
				PluginRegistry: pluginRegistry,
			})
		},
	}
)

// init 解析命令参数
func init() {
	startCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "conf/pole-server.yaml", "config file path")
	startCmd.PersistentFlags().StringVar(&startMode, "mode", "",
		"启动模式：all、console、control-plane、limiter-server、full（server 是 control-plane 的弃用别名）")
}
