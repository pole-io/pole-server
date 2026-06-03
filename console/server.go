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

package console

import (
	"context"
	"fmt"
	"net/http"
	"time"

	console_bootstrap "github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/handlers"
	"github.com/pole-io/pole-server/console/pkg/router"

	_ "github.com/pole-io/pole-server/console/pkg/observer/mysql"
)

// Start initializes and starts the embedded console gateway in the current process.
func Start(ctx context.Context, config *console_bootstrap.Config, errCh chan<- error) (*http.Server, error) {
	handlers.NewAdminGetter(config)
	console_bootstrap.Initialize(config)
	console_bootstrap.SetMode(config)

	address := fmt.Sprintf("%v:%v", config.WebServer.ListenIP, config.WebServer.ListenPort)
	server := &http.Server{
		Addr:    address,
		Handler: router.NewRouter(config),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	go func() {
		err := server.ListenAndServe()
		if err == nil || err == http.ErrServerClosed {
			return
		}
		errCh <- err
	}()

	return server, nil
}
