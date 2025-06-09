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

package httpserver

import (
	"net/http"
	"net/http/pprof"

	"github.com/emicklei/go-restful/v3"
)

// enablePprofAccess 开启pprof接口
func (h *HTTPServer) enablePprofAccess(wsContainer *restful.Container) {
	log.Infof("open http access for pprof")
	wsContainer.Handle("/debug/pprof/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.enablePprof.Load() {
			pprof.Index(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	wsContainer.Handle("/debug/pprof/cmdline", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.enablePprof.Load() {
			pprof.Cmdline(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	wsContainer.Handle("/debug/pprof/profile", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.enablePprof.Load() {
			pprof.Profile(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	wsContainer.Handle("/debug/pprof/symbol", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.enablePprof.Load() {
			pprof.Symbol(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	wsContainer.Handle("/debug/pprof/trace", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.enablePprof.Load() {
			pprof.Trace(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
}
