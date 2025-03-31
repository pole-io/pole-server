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

package rsa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"

	"github.com/pole-io/pole-server/plugin"
)

const (
	// PluginName plugin name
	PluginName = "RSA"
)

func init() {
	plugin.RegisterPlugin(PluginName, &RSACrypto{})
}

// AESCrypto AES crypto
type RSACrypto struct {
}

// Name 返回插件名字
func (h *RSACrypto) Name() string {
	return PluginName
}

// Destroy 销毁插件
func (h *RSACrypto) Destroy() error {
	return nil
}

// Initialize 插件初始化
func (h *RSACrypto) Initialize(c *plugin.ConfigEntry) error {
	return nil
}

// GenerateKey generate key
func (c *RSACrypto) GenerateKey() ([]byte, error) {
	key := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Encrypt AES encrypt plaintext and base64 encode ciphertext
func (c *RSACrypto) Encrypt(plaintext string, key []byte) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	oriKey, err := base64.StdEncoding.DecodeString(string(key))
	if err != nil {
		return "", err
	}
	pub, err := x509.ParsePKCS1PublicKey(oriKey)
	if err != nil {
		return "", err
	}
	totalLen := len(plaintext)
	segLen := pub.Size() - 11
	start := 0
	buffer := bytes.Buffer{}
	for start < totalLen {
		end := start + segLen
		if end > totalLen {
			end = totalLen
		}
		seg, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(plaintext)[start:end])
		if err != nil {
			return "", err
		}
		buffer.Write(seg)
		start = end
	}
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

// Decrypt base64 decode ciphertext and AES decrypt
func (c *RSACrypto) Decrypt(ciphertext string, key []byte) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	oriKey, err := base64.StdEncoding.DecodeString(string(key))
	if err != nil {
		return "", err
	}
	oritext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	priv, err := x509.ParsePKCS1PrivateKey(oriKey)
	if err != nil {
		return "", err
	}
	keySize := priv.Size()
	totalLen := len(oritext)
	start := 0
	buffer := bytes.Buffer{}
	for start < totalLen {
		end := start + keySize
		if end > totalLen {
			end = totalLen
		}
		seg, err := rsa.DecryptPKCS1v15(rand.Reader, priv, oritext[start:end])
		if err != nil {
			return "", err
		}
		buffer.Write(seg)
		start = end
	}
	return buffer.String(), nil
}

// RSAKey RSA key pair
type RSAKey struct {
	PrivateKey string
	PublicKey  string
}

// GenerateKey generate RSA key pair
func GenerateRSAKey() (*RSAKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	rsaKey := &RSAKey{
		PrivateKey: base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(privateKey)),
		PublicKey:  base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)),
	}
	return rsaKey, nil
}
