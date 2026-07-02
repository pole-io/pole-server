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

package sqldb

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
)

func TestServiceSubscriberStore_UpsertAndQuery(t *testing.T) {
	db := openServiceSubscriberTestDB(t)
	t.Cleanup(func() {
		db.Close()
	})
	require.NoError(t, ensureServiceSubscribeGraphSchema(db))

	store := &serviceStore{master: db, slave: db}
	cleanupServiceSubscriberTestRows(t, db)
	t.Cleanup(func() {
		cleanupServiceSubscriberTestRows(t, db)
	})

	sub := &svctypes.ServiceSubscriber{
		Caller: &svctypes.ServiceKey{Name: "subscriber-client", Namespace: "default"},
		Callee: []*svctypes.ServiceKey{{Name: "pole.checker", Namespace: SystemNamespace}},
	}

	require.NoError(t, store.AddServiceSubscibes([]*svctypes.ServiceSubscriber{sub}))
	require.NoError(t, store.AddServiceSubscibes([]*svctypes.ServiceSubscriber{sub}))

	total, list, err := store.BatchGetServiceSubscribers(context.Background(), map[string]string{
		"callee_name":      "pole.checker",
		"callee_namespace": SystemNamespace,
	}, 0, 10)
	require.NoError(t, err)
	require.Equal(t, uint32(1), total)
	require.Len(t, list, 1)
	require.Equal(t, "subscriber-client", list[0].Caller.Name)
	require.Equal(t, "default", list[0].Caller.Namespace)
	require.Len(t, list[0].Callee, 1)
	require.Equal(t, "pole.checker", list[0].Callee[0].Name)
	require.Equal(t, SystemNamespace, list[0].Callee[0].Namespace)

	incremental, err := store.GetMoreServiceSubscibes(time.Now().Add(-time.Minute), false)
	require.NoError(t, err)
	require.Contains(t, incremental, "default::subscriber-client")
}

func TestServiceStoreDeleteServiceCleansSubscribeGraph(t *testing.T) {
	db := openServiceSubscriberTestDB(t)
	t.Cleanup(func() {
		db.Close()
	})
	require.NoError(t, ensureServiceSubscribeGraphSchema(db))

	store := &serviceStore{master: db, slave: db}
	cleanupServiceDeleteSubscribeTestRows(t, db)
	t.Cleanup(func() {
		cleanupServiceDeleteSubscribeTestRows(t, db)
	})

	caller := &svctypes.Service{
		ID:        "svcsubcleancaller000000000001",
		Name:      "subscriber-clean-caller",
		Namespace: "default",
		Token:     "token",
		Revision:  "revision",
		Owner:     "pole",
	}
	target := &svctypes.Service{
		ID:        "svcsubcleantarget000000000001",
		Name:      "subscriber-clean-target",
		Namespace: "default",
		Token:     "token",
		Revision:  "revision",
		Owner:     "pole",
	}
	observer := &svctypes.Service{
		ID:        "svcsubcleanobserver0000000001",
		Name:      "subscriber-clean-observer",
		Namespace: "default",
		Token:     "token",
		Revision:  "revision",
		Owner:     "pole",
	}

	require.NoError(t, store.AddService(caller))
	require.NoError(t, store.AddService(target))
	require.NoError(t, store.AddService(observer))
	require.NoError(t, store.AddServiceSubscibes([]*svctypes.ServiceSubscriber{
		{
			Caller: &svctypes.ServiceKey{Name: caller.Name, Namespace: caller.Namespace},
			Callee: []*svctypes.ServiceKey{{Name: target.Name, Namespace: target.Namespace}},
		},
		{
			Caller: &svctypes.ServiceKey{Name: target.Name, Namespace: target.Namespace},
			Callee: []*svctypes.ServiceKey{{Name: observer.Name, Namespace: observer.Namespace}},
		},
	}))

	require.NoError(t, store.DeleteService(target.ID, target.Name, target.Namespace))

	total, list, err := store.BatchGetServiceSubscribers(context.Background(), map[string]string{
		"callee_name":      target.Name,
		"callee_namespace": target.Namespace,
	}, 0, 10)
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, list)

	total, list, err = store.BatchGetServiceSubscribers(context.Background(), map[string]string{
		"caller_name":      target.Name,
		"caller_namespace": target.Namespace,
	}, 0, 10)
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, list)
}

func openServiceSubscriberTestDB(t *testing.T) *BaseDB {
	t.Helper()

	user := getenv("MYSQL_DB_USER", getenv("MYSQL_USER", "root"))
	password := getenv("MYSQL_DB_PWD", getenv("MYSQL_PWD", "123456"))
	host := getenv("MYSQL_HOST", "127.0.0.1:3306")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/pole_server?loc=Local", user, password, host)

	db, err := NewBaseDB(&dbConfig{dbType: "mysql", dns: dsn})
	if err != nil {
		t.Skipf("skip MySQL integration test: %v", err)
	}
	return db
}

func cleanupServiceSubscriberTestRows(t *testing.T, db *BaseDB) {
	t.Helper()

	_, err := db.Exec(`
		delete from service_subscribe_graph
		where (caller_name = ? and caller_namespace = ?)
		   or (callee_name = ? and callee_namespace = ?)`,
		"subscriber-client", "default", "pole.checker", SystemNamespace)
	require.NoError(t, err)
}

func cleanupServiceDeleteSubscribeTestRows(t *testing.T, db *BaseDB) {
	t.Helper()

	names := []string{
		"subscriber-clean-caller",
		"subscriber-clean-target",
		"subscriber-clean-observer",
	}
	_, err := db.Exec(`
		delete from service_subscribe_graph
		where caller_name in (?, ?, ?)
		   or callee_name in (?, ?, ?)`,
		names[0], names[1], names[2], names[0], names[1], names[2])
	require.NoError(t, err)

	_, err = db.Exec(`
		delete from service_metadata
		where id in (
			select id from service
			where namespace = 'default'
			  and name in (?, ?, ?)
		)`,
		names[0], names[1], names[2])
	require.NoError(t, err)

	_, err = db.Exec(`
		delete from service
		where namespace = 'default'
		  and name in (?, ?, ?)`,
		names[0], names[1], names[2])
	require.NoError(t, err)
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
