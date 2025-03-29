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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pole-io/pole-server/common/version"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "print version",
		Long:  "print version",
		Run: func(c *cobra.Command, args []string) {
			fmt.Printf("version: %v\n", version.Get())
		},
	}

	revisionCmd = &cobra.Command{
		Use:   "revision",
		Short: "print revision with building date",
		Long:  "print revision with building date",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("revision: %v\n", version.GetRevision())
		},
	}
)
