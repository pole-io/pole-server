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

package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter_ProxiesAIMCPRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const (
		userID = "user1"
		token  = "token1"
		secret = "polarismesh@2021"
	)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/v1/user/token" {
			assert.Equal(t, userID, r.URL.Query().Get("id"))
			assert.Equal(t, token, r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":200000,"data":{"auth_token":"token1","token_enable":true}}`))
			return
		}
		assert.Equal(t, "/ai/mcp/v1/servers", r.URL.Path)
		assert.Equal(t, "0", r.URL.Query().Get("offset"))
		assert.Equal(t, token, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200000,"data":[{"name":"demo-mcp"}]}`))
	}))
	defer backend.Close()

	webPath := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(webPath, "assets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(webPath, "index.html"), []byte("<html></html>"), 0o644))

	router := NewRouter(&bootstrap.Config{
		WebServer: bootstrap.WebServer{
			WebPath:    webPath + string(os.PathSeparator),
			CoreURL:    "/core/v1",
			NamingURL:  "/naming/v1",
			AuthURL:    "/auth/v1",
			ConfigURL:  "/config/v1",
			MonitorURL: "/api/v1",
			JWT: bootstrap.JWT{
				SecretKey: secret,
				Expired:   1800,
			},
		},
		PoleServer: bootstrap.PoleServer{
			Address: strings.TrimPrefix(backend.URL, "http://"),
		},
	})

	console := httptest.NewServer(router)
	defer console.Close()

	req, err := http.NewRequest(http.MethodGet, console.URL+"/ai/mcp/v1/servers?offset=0", nil)
	require.NoError(t, err)
	jwtCookie := newTestJWTCookie(t, userID, token, secret)
	req.AddCookie(jwtCookie)
	req.Header.Set("x-pole-user", userID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "demo-mcp")
}

func TestNewRouter_ProxiesAIA2ARequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const (
		userID = "user1"
		token  = "token1"
		secret = "polarismesh@2021"
	)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/v1/user/token" {
			assert.Equal(t, userID, r.URL.Query().Get("id"))
			assert.Equal(t, token, r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":200000,"data":{"auth_token":"token1","token_enable":true}}`))
			return
		}
		assert.Equal(t, "/ai/a2a/v1/agents", r.URL.Path)
		assert.Equal(t, "0", r.URL.Query().Get("offset"))
		assert.Equal(t, token, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200000,"amount":1,"data":[{"name":"demo-agent"}]}`))
	}))
	defer backend.Close()

	webPath := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(webPath, "assets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(webPath, "index.html"), []byte("<html></html>"), 0o644))

	router := NewRouter(&bootstrap.Config{
		WebServer: bootstrap.WebServer{
			WebPath:    webPath + string(os.PathSeparator),
			CoreURL:    "/core/v1",
			NamingURL:  "/naming/v1",
			AuthURL:    "/auth/v1",
			ConfigURL:  "/config/v1",
			MonitorURL: "/api/v1",
			JWT: bootstrap.JWT{
				SecretKey: secret,
				Expired:   1800,
			},
		},
		PoleServer: bootstrap.PoleServer{
			Address: strings.TrimPrefix(backend.URL, "http://"),
		},
	})

	console := httptest.NewServer(router)
	defer console.Close()

	req, err := http.NewRequest(http.MethodGet, console.URL+"/ai/a2a/v1/agents?offset=0", nil)
	require.NoError(t, err)
	jwtCookie := newTestJWTCookie(t, userID, token, secret)
	req.AddCookie(jwtCookie)
	req.Header.Set("x-pole-user", userID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "demo-agent")
}

func newTestJWTCookie(t *testing.T, userID, token, secret string) *http.Cookie {
	t.Helper()

	now := time.Now()
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"UserID": userID,
		"Token":  token,
		"exp":    now.Add(time.Hour).Unix(),
		"nbf":    now.Unix(),
		"iat":    now.Unix(),
	}).SignedString([]byte(secret))
	require.NoError(t, err)
	return &http.Cookie{Name: "jwt", Value: signed}
}
