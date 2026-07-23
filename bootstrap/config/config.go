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
	"errors"
	"fmt"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/apis/apiserver"
	storeapi "github.com/pole-io/pole-server/apis/store"
	console_bootstrap "github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/pkg/admin"
	"github.com/pole-io/pole-server/pkg/cache"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/config"
	"github.com/pole-io/pole-server/pkg/goverrule"
	"github.com/pole-io/pole-server/pkg/namespace"
	"github.com/pole-io/pole-server/pkg/service"
	"github.com/pole-io/pole-server/pkg/systemconfig"
	"github.com/pole-io/pole-server/pkg/workloadcredential"
)

// Config 配置
type Config struct {
	Bootstrap           Bootstrap                 `yaml:"bootstrap"`
	APIServers          string                    `yaml:"apiservers"`
	Cache               cache.Config              `yaml:"cache"`
	Namespace           namespace.Config          `yaml:"namespace"`
	Naming              service.Config            `yaml:"naming"`
	GoverRule           goverrule.Config          `yaml:"goverrule"`
	Config              config.Config             `yaml:"config"`
	Maintain            admin.Config              `yaml:"maintain"`
	Store               storeapi.Config           `yaml:"store"`
	Auth                auth.Config               `yaml:"auth"`
	Plugin              apis.Config               `yaml:"plugin"`
	WorkloadCredential  workloadcredential.Config `yaml:"workloadCredential"`
	SystemConfigSources systemconfig.SourceIndex  `yaml:"-" json:"-"`
}

// Bootstrap 启动引导配置
type Bootstrap struct {
	Logger         string                   `yaml:"logger"`
	Mode           string                   `yaml:"mode"`
	Console        console_bootstrap.Config `yaml:"console"`
	StartInOrder   map[string]interface{}   `yaml:"startInOrder"`
	PolarisService PolarisService           `yaml:"polaris_service"`
}

// PolarisService sergo-server的自注册配置
type PolarisService struct {
	EnableRegister    bool       `yaml:"enable_register"`
	ProbeAddress      string     `yaml:"probe_address"`
	SelfAddress       string     `yaml:"self_address"`
	NetworkInter      string     `yaml:"network_inter"`
	Isolated          bool       `yaml:"isolated"`
	DisableHeartbeat  bool       `yaml:"disable_heartbeat"`
	HeartbeatInterval int        `yaml:"heartbeat_interval"`
	Services          []*Service `yaml:"services"`
}

// Service 服务的自注册的配置
type Service struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Protocols []string          `yaml:"protocols"`
	Metadata  map[string]string `yaml:"metadata"`
}

// APIEntries 对外提供的apiServers
type APIEntries struct {
	Name      string   `yaml:"name"`
	Protocols []string `yaml:"protocols"`
}

const (
	// DefaultPolarisName default polaris name
	DefaultPolarisName = "pole-server"
	// DefaultPolarisNamespace default namespace
	DefaultPolarisNamespace = "pole-system"
	// DefaultFilePath default file path
	DefaultFilePath = "pole-server.yaml"
	// DefaultHeartbeatInterval default interval second for heartbeat
	DefaultHeartbeatInterval = 5
)

// Load 加载配置
func Load(filePath string) (*Config, error) {
	if filePath == "" {
		err := errors.New("invalid config file path")
		fmt.Printf("[ERROR] %v\n", err)
		return nil, err
	}

	fmt.Printf("[INFO] load config from %v\n", filePath)
	conf := &Config{
		Bootstrap: defaultBootstrap(),
		Maintain:  *admin.DefaultConfig(),
	}
	sources, err := systemconfig.SourceIndexFromFile(filePath)
	if err != nil {
		return nil, err
	}
	if _, err := utils.LoadYAML(filePath, conf); err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return nil, err
	}
	conf.SystemConfigSources = sources
	conf.Bootstrap.Console.SystemConfigSources = sources.Subtree("bootstrap.console")
	return conf, nil
}

func LoadAPIEntries(f string) ([]apiserver.Config, error) {
	if f == "" {
		return nil, fmt.Errorf("invalid api server config file path")
	}

	ret := make([]apiserver.Config, 0)
	if _, err := utils.LoadYAML(f, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
