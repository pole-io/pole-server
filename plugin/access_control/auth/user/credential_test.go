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

package defaultuser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/pkg/types"
)

func TestAuthTokenFromContextPrefersParsedToken(t *testing.T) {
	ctx := types.AppendRequestHeader(context.Background(), map[string][]string{
		types.HeaderAuthorizationKey: {"header-token"},
	})
	ctx = context.WithValue(ctx, types.ContextAuthTokenKey, "context-token")

	require.Equal(t, "context-token", authTokenFromContext(ctx))
}

func TestAuthTokenFromContextReadsAuthorizationCaseInsensitive(t *testing.T) {
	tests := map[string]map[string][]string{
		"canonical": {
			types.HeaderAuthorizationKey: {"canonical-token"},
		},
		"lowercase": {
			"authorization": {"lower-token"},
		},
		"alternate token header": {
			"x-polaris-token": {"polaris-token"},
		},
	}

	for name, headers := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := types.AppendRequestHeader(context.Background(), headers)

			require.NotEmpty(t, authTokenFromContext(ctx))
		})
	}
}

func TestIsUserTokenUsable(t *testing.T) {
	const (
		userID = "49dba3c69bca4b668903901d85c61528"
		salt   = "polarismesh@2021"
	)
	svr := &Server{authOpt: &AuthConfig{Salt: salt}}

	validToken, err := createUserToken(userID, salt)
	require.NoError(t, err)
	require.True(t, svr.isUserTokenUsable(userID, validToken))

	otherUserToken, err := createUserToken("other-user", salt)
	require.NoError(t, err)
	require.False(t, svr.isUserTokenUsable(userID, otherUserToken))

	oldSaltToken, err := createUserToken(userID, "pole@@1234567890")
	require.NoError(t, err)
	require.False(t, svr.isUserTokenUsable(userID, oldSaltToken))
}
