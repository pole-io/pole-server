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

package store

import (
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
)

// Store 通用存储接口
type Store interface {
	// Name 存储层的名字
	Name() string
	// Initialize 存储的初始化函数
	Initialize(c *Config) error
	// Destroy 存储的析构函数
	Destroy() error
	// CreateTransaction 创建事务对象
	CreateTransaction() (Transaction, error)
	// StartTx 开启一个原子事务
	StartTx() (Tx, error)
	// StartReadTx 开启一个原子事务
	StartReadTx() (Tx, error)
	// NamespaceStore Service namespace interface
	NamespaceStore
	// NamingModuleStore Service Registration Discovery Module Storage Interface
	NamingModuleStore
	// ConfigFileModuleStore Configure the central module storage interface
	ConfigFileModuleStore
	// ClientStore Client the central module storage interface
	ClientStore
	// AdminStore Maintain inteface
	AdminStore
	// GrayStore mgr gray resource
	GrayStore
	// AuthStore Auth storage interface
	AuthStore
}

// NamespaceStore Namespace storage interface
type NamespaceStore interface {
	// AddNamespace Save a namespace
	AddNamespace(namespace *types.Namespace) error
	// UpdateNamespace Update namespace
	UpdateNamespace(namespace *types.Namespace) error
	// GetNamespace Get the details of the namespace according to Name
	GetNamespace(name string) (*types.Namespace, error)
	// GetNamespaces Query Namespace from the database
	GetNamespaces(filter map[string][]string, offset, limit int) ([]*types.Namespace, uint32, error)
	// GetMoreNamespaces Get incremental data
	// 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
	GetMoreNamespaces(mtime time.Time) ([]*types.Namespace, error)
}

// GrayStore Gray storage interface
type GrayStore interface {
	// CleanGrayResource .
	CleanGrayResource(tx Tx, data *rules.GrayResource) error
	// CreateGrayResourceTx .
	CreateGrayResourceTx(tx Tx, data *rules.GrayResource) error
	// GetMoreGrayResouces .
	GetMoreGrayResouces(firstUpdate bool, mtime time.Time) ([]*rules.GrayResource, error)
}

// Transaction Transaction interface, does not support multi-level concurrency operation,
// currently only support a single price serial operation
type Transaction interface {
	// Commit Transaction
	Commit() error
	// LockBootstrap Start the lock, limit the concurrent number of Server boot
	LockBootstrap(key string, server string) error
	// LockNamespace Row it locks Namespace
	LockNamespace(name string) (*types.Namespace, error)
	// DeleteNamespace Delete Namespace
	DeleteNamespace(name string) error
	// LockService Row it locks service
	LockService(name string, namespace string) (*svctypes.Service, error)
	// RLockService Shared lock service
	RLockService(name string, namespace string) (*svctypes.Service, error)
}

// Tx Atomic matters without any business attributes.Abstraction of different storage type transactions
type Tx interface {
	// Commit Transaction
	Commit() error
	// Rollback transaction
	Rollback() error
	// GetDelegateTx Get the original proxy transaction object.Different storage types have no business implementation
	GetDelegateTx() interface{}
	// CreateReadView create a snapshot read view
	CreateReadView() error
}

// ToolStore Storage related functions and tool interfaces
type ToolStore interface {
	// GetUnixSecond Get the current time
	GetUnixSecond(maxWait time.Duration) (int64, error)
}
