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

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
	bootstrap "github.com/pole-io/pole-server/pkg/console/config"
)

func ReverseHandleBootstrap(poleServer *bootstrap.PoleServer, conf *bootstrap.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Header.Del("Cookie")

		director := func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = poleServer.Address
			req.Host = poleServer.Address
		}
		proxy := &httputil.ReverseProxy{Director: director, ModifyResponse: func(resp *http.Response) error {
			if resp.StatusCode != http.StatusNotFound {
				return nil
			}
			if err := resp.Body.Close(); err != nil {
				return err
			}
			body := []byte(`{"code": 200000,"info": "success"}`)
			resp.StatusCode = http.StatusOK
			resp.Header["Content-Length"] = []string{fmt.Sprint(len(body))}
			resp.Body = io.NopCloser(bytes.NewReader(body))
			return nil
		}}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// ReverseHandleAdminUserExist 反向代理
func ReverseHandleAdminUserExist(polarisServer *bootstrap.PoleServer, conf *bootstrap.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Header.Del("Cookie")

		director := func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = polarisServer.Address
			req.Host = polarisServer.Address
		}
		proxy := &httputil.ReverseProxy{Director: director,
			ErrorHandler: func(resp http.ResponseWriter, _ *http.Request, _ error) {
				owner := "polaris"
				if conf.WebServer.MainUser != "" {
					owner = conf.WebServer.MainUser
				}
				body := []byte(`{"code":200000,"info":"success","user":{"name":"` + owner + `"}}`)
				resp.Header().Set("Content-Type", "application/json")
				resp.WriteHeader(http.StatusOK)
				_, _ = resp.Write(body)
			},
			ModifyResponse: func(resp *http.Response) error {
				if resp.StatusCode != http.StatusOK {
					if err := resp.Body.Close(); err != nil {
						return err
					}
					owner := "polaris"
					if conf.WebServer.MainUser != "" {
						owner = conf.WebServer.MainUser
					}
					body := []byte(`{"code":200000,"info":"success","user":{"name":"` + owner + `"}}`)
					resp.StatusCode = http.StatusOK
					resp.Header["Content-Length"] = []string{fmt.Sprint(len(body))}
					resp.Body = io.NopCloser(bytes.NewReader(body))
				}
				// 读取到当前的 main user
				ret, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				if err := resp.Body.Close(); err != nil {
					return err
				}
				userRsp := struct {
					Code int    `json:"code"`
					Info string `json:"info"`
					User struct {
						Name string `json:"name"`
					} `json:"user"`
				}{}
				if err := json.Unmarshal(ret, &userRsp); err != nil {
					return err
				}
				if userRsp.User.Name != "" {
					conf.WebServer.MainUser = userRsp.User.Name
				}
				resp.Body = io.NopCloser(bytes.NewReader(ret))
				return nil
			}}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
