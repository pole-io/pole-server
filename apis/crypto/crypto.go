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

package crypto

import (
	"fmt"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/pluginapi"
)

var cryptoRuntime pluginapi.RuntimeValue[CryptoManager]

// Crypto Crypto interface
type Crypto interface {
	apis.Plugin
	GenerateKey() ([]byte, error)
	Encrypt(plaintext string, key []byte) (cryptotext string, err error)
	Decrypt(cryptotext string, key []byte) (string, error)
}

// GetCrypto get the crypto plugin
func GetCryptoManager() CryptoManager {
	manager, err := cryptoRuntime.Load(func() (CryptoManager, error) {
		var (
			entries []apis.ConfigEntry
		)
		entries = append(entries, apis.GetPluginConfig().Crypto.Entries...)
		manager := &defaultCryptoManager{
			cryptos: make(map[string]Crypto),
			options: entries,
		}

		if err := manager.Initialize(); err != nil {
			return nil, err
		}
		return manager, nil
	})
	if err != nil {
		panic(fmt.Errorf("Crypto plugin init err: %s", err.Error()))
	}
	return manager
}

// CryptoManager crypto algorithm manager
type CryptoManager interface {
	Name() string
	Initialize() error
	Destroy() error
	GetCryptoAlgoNames() []string
	GetCrypto(algo string) (Crypto, error)
}

// defaultCryptoManager crypto algorithm manager
type defaultCryptoManager struct {
	cryptos map[string]Crypto
	options []apis.ConfigEntry
}

func (c *defaultCryptoManager) Name() string {
	return "CryptoManager"
}

func (c *defaultCryptoManager) Initialize() error {
	for i := range c.options {
		entry := c.options[i]
		item, err := apis.ResolvePlugin(apis.PluginTypeCrypto, entry.Name)
		if err != nil {
			return fmt.Errorf("resolve Crypto plugin %q: %w", entry.Name, err)
		}
		crypto, ok := item.(Crypto)
		if !ok {
			return fmt.Errorf("plugin target: %s not Crypto", entry.Name)
		}
		if err := crypto.Initialize(&entry); err != nil {
			return err
		}
		c.cryptos[entry.Name] = crypto
	}
	return nil
}

func (c *defaultCryptoManager) Destroy() error {
	for i := range c.cryptos {
		if err := c.cryptos[i].Destroy(); err != nil {
			return err
		}
	}
	return nil
}

func (c *defaultCryptoManager) GetCryptoAlgoNames() []string {
	var names []string
	for name := range c.cryptos {
		names = append(names, name)
	}
	return names
}

func (c *defaultCryptoManager) GetCrypto(algo string) (Crypto, error) {
	crypto, ok := c.cryptos[algo]
	if !ok {
		return nil, fmt.Errorf("plugin Crypto not found target: %s", algo)
	}
	return crypto, nil
}
