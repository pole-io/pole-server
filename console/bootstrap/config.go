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

package bootstrap

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pole-io/pole-server/console/pkg/common/log"
	store "github.com/pole-io/pole-server/console/pkg/observer"
	"gopkg.in/yaml.v2"
)

var (
	_globalConfig *Config
)

// PoleServer polaris server配置
type PoleServer struct {
	Address string `yaml:"address"`
}

type MonitorServer struct {
	Address string `yaml:"address"`
}

// Config 配置
type Config struct {
	Logger        log.Options   `yaml:"logger"`
	WebServer     WebServer     `yaml:"webServer"`
	PoleServer    PoleServer    `yaml:"poleServer"`
	MonitorServer MonitorServer `yaml:"monitorServer"`
	Futures       string        `yaml:"futures"`
	Store         store.Config  `yaml:"store"`
}

func (c *Config) HasFutures(s string) bool {
	return strings.Contains(c.Futures, s)
}

// WebServer web server配置
type WebServer struct {
	Mode       string `yaml:"mode"`
	ListenIP   string `yaml:"listenIP"`
	ListenPort int    `yaml:"listenPort"`
	CoreURL    string `yaml:"coreURL"`
	NamingURL  string `yaml:"namingURL"`
	AuthURL    string `yaml:"authURL"`
	MonitorURL string `yaml:"monitorURL"`
	ConfigURL  string `yaml:"configURL"`
	LogURL     string `yaml:"logURL"`
	WebPath    string `yaml:"webPath"`
	JWT        JWT    `yaml:"jwt"`
	MainUser   string `yaml:"mainUser"`
}

// JWT jwtToken 相关的配置
type JWT struct {
	// 参与jwt运算的key
	SecretKey string `yaml:"secretKey"`
	// 过期时间, 单位为秒
	Expired int `yaml:"expired"`
}

// LoadConfig 加载配置文件
func LoadConfig(filePath string) (*Config, error) {
	if filePath == "" {
		err := errors.New("invalid config file path")
		fmt.Printf("[ERR0R] %v\n", err)
		return nil, err
	}

	fmt.Printf("[INFO] load config from %v\n", filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return nil, err
	}

	finalContent := os.ExpandEnv(string(content))

	config := &Config{
		Logger: DefaultLoggerOptions(),
	}
	config.WebServer.JWT.Expired = 1800 // 默认30分钟
	config.WebServer.JWT.SecretKey = "polarismesh@2021"
	err = yaml.NewDecoder(bytes.NewBuffer([]byte(finalContent))).Decode(config)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
	}

	_globalConfig = config
	return config, nil
}

func GetConfig() *Config {
	return _globalConfig
}

func DefaultLoggerOptions() log.Options {
	return log.Options{
		ErrorOutputPaths:   []string{"logs/runtime/pole-console-error.log"},
		RotateOutputPath:   "logs/runtime/pole-console.log",
		RotationMaxSize:    500,
		RotationMaxAge:     30,
		RotationMaxBackups: 100,
		Level:              "info",
	}
}
