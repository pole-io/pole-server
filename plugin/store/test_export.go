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
	"errors"
	"fmt"

	storeapi "github.com/pole-io/pole-server/apis/store"
)

// TestGetStore 获取Store
func TestGetStore() (storeapi.Store, error) {
	name := storeapi.GetStoreConfig().Name
	if name == "" {
		return nil, errors.New("store name is empty")
	}

	s, ok := storeapi.StoreSlots[name]
	if !ok {
		return nil, fmt.Errorf("store `%s` not found", name)
	}
	_ = s.Destroy()
	fmt.Printf("[Store][Info] current use store plugin : %s\n", s.Name())

	if err := s.Initialize(storeapi.GetStoreConfig()); err != nil {
		fmt.Printf("[ERROR] initialize store `%s` fail: %v", s.Name(), err)
		panic(err)
	}
	return s, nil
}
