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

package goverrule_chain

import (
	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/goverrule"
	goverrule_auth "github.com/pole-io/pole-server/pkg/goverrule/interceptor/auth"
	"github.com/pole-io/pole-server/pkg/goverrule/interceptor/paramcheck"
)

func init() {
	err := goverrule.RegisterServerProxy("paramcheck", func(pre goverrule.GoverRuleServer,
		s store.Store) (goverrule.GoverRuleServer, error) {
		return paramcheck.NewServer(pre, s), nil
	})
	if err != nil {
		panic(err)
	}

	err = goverrule.RegisterServerProxy("auth", func(pre goverrule.GoverRuleServer,
		s store.Store) (goverrule.GoverRuleServer, error) {
		userSvr, err := authapi.GetUserServer()
		if err != nil {
			return nil, err
		}
		policySvr, err := authapi.GetStrategyServer()
		if err != nil {
			return nil, err
		}

		return goverrule_auth.NewServer(pre, userSvr, policySvr), nil
	})
	if err != nil {
		panic(err)
	}
}
