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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveStartMode_DefaultsToAll(t *testing.T) {
	mode, err := ResolveStartMode("", "")

	require.NoError(t, err)
	assert.Equal(t, StartModeAll, mode)
}

func TestResolveStartMode_UsesConfigMode(t *testing.T) {
	mode, err := ResolveStartMode(StartModeServer, "")

	require.NoError(t, err)
	assert.Equal(t, StartModeControlPlane, mode)
}

func TestResolveStartMode_CLIOverridesConfig(t *testing.T) {
	mode, err := ResolveStartMode(StartModeServer, StartModeConsole)

	require.NoError(t, err)
	assert.Equal(t, StartModeConsole, mode)
}

func TestResolveStartMode_RejectsInvalidMode(t *testing.T) {
	_, err := ResolveStartMode("", "invalid")

	require.Error(t, err)
}

func TestResolveStartMode_AcceptsCanonicalModes(t *testing.T) {
	tests := []string{
		StartModeAll,
		StartModeConsole,
		StartModeControlPlane,
		StartModeLimiterServer,
		StartModeFull,
	}

	for _, expected := range tests {
		t.Run(expected, func(t *testing.T) {
			mode, err := ResolveStartMode("", expected)

			require.NoError(t, err)
			assert.Equal(t, expected, mode)
		})
	}
}

func TestResolveStartProfile_AllKeepsCompatibilityModules(t *testing.T) {
	profile, err := ResolveStartProfile("", StartModeAll)

	require.NoError(t, err)
	assert.Equal(t, StartProfile{
		Mode:         StartModeAll,
		ControlPlane: true,
		Console:      true,
	}, profile)
}

func TestResolveStartProfile_FullIncludesAllModules(t *testing.T) {
	profile, err := ResolveStartProfile("", StartModeFull)

	require.NoError(t, err)
	assert.Equal(t, StartProfile{
		Mode:          StartModeFull,
		ControlPlane:  true,
		Console:       true,
		LimiterServer: true,
	}, profile)
}

func TestResolveStartProfile_ModuleMatrix(t *testing.T) {
	tests := []struct {
		mode     string
		expected StartProfile
	}{
		{
			mode: StartModeConsole,
			expected: StartProfile{
				Mode: StartModeConsole, Console: true,
			},
		},
		{
			mode: StartModeControlPlane,
			expected: StartProfile{
				Mode: StartModeControlPlane, ControlPlane: true,
			},
		},
		{
			mode: StartModeLimiterServer,
			expected: StartProfile{
				Mode: StartModeLimiterServer, LimiterServer: true,
			},
		},
		{
			mode: StartModeServer,
			expected: StartProfile{
				Mode: StartModeControlPlane, ControlPlane: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.mode, func(t *testing.T) {
			profile, err := ResolveStartProfile(test.mode, "")

			require.NoError(t, err)
			assert.Equal(t, test.expected, profile)
		})
	}
}
