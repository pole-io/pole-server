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

// Package utils contains common utility functions
package container

import (
	"sync"
)

// NewSet returns a new Set
func NewSet[K comparable]() *Set[K] {
	return &Set[K]{
		container: make(map[K]struct{}),
	}
}

type Set[K comparable] struct {
	container map[K]struct{}
}

// Add adds a string to the set
func (set *Set[K]) Add(val K) {
	set.container[val] = struct{}{}
}

// Remove removes a string from the set
func (set *Set[K]) Remove(val K) {
	delete(set.container, val)
}

func (set *Set[K]) ToSlice() []K {
	ret := make([]K, 0, len(set.container))
	for k := range set.container {
		ret = append(ret, k)
	}
	return ret
}

func (set *Set[K]) Range(fn func(val K)) {
	for k := range set.container {
		fn(k)
	}
}

type Reference[K, R comparable] struct {
	Key        K
	Referencer R
}

// NewRefSyncSet returns a new Set
func NewRefSyncSet[K, R comparable]() *RefSyncSet[K, R] {
	return &RefSyncSet[K, R]{
		container: map[K]map[R]struct{}{},
	}
}

type RefSyncSet[K, R comparable] struct {
	container map[K]map[R]struct{}
	lock      sync.RWMutex
}

// Add adds a string to the set
func (set *RefSyncSet[K, R]) Add(val Reference[K, R]) {
	set.lock.Lock()
	defer set.lock.Unlock()

	if _, ok := set.container[val.Key]; !ok {
		set.container[val.Key] = map[R]struct{}{}
	}
	refs := set.container[val.Key]
	refs[val.Referencer] = struct{}{}
}

// Remove removes a string from the set
func (set *RefSyncSet[K, R]) Remove(val Reference[K, R]) {
	set.lock.Lock()
	defer set.lock.Unlock()
	if _, ok := set.container[val.Key]; !ok {
		return
	}
	refs := set.container[val.Key]
	delete(refs, val.Referencer)
	if len(refs) == 0 {
		delete(set.container, val.Key)
	} else {
		set.container[val.Key] = refs
	}
}

func (set *RefSyncSet[K, R]) ToSlice() []K {
	set.lock.RLock()
	defer set.lock.RUnlock()

	ret := make([]K, 0, len(set.container))
	for k := range set.container {
		ret = append(ret, k)
	}
	return ret
}

func (set *RefSyncSet[K, R]) Range(fn func(val K)) {
	set.lock.RLock()
	snapshot := map[K]struct{}{}
	for k := range set.container {
		snapshot[k] = struct{}{}
	}
	set.lock.RUnlock()

	for k := range snapshot {
		fn(k)
	}
}

func (set *RefSyncSet[K, R]) Len() int {
	set.lock.RLock()
	defer set.lock.RUnlock()

	return len(set.container)
}

// Contains contains target value
func (set *RefSyncSet[K, R]) Contains(val K) bool {
	set.lock.Lock()
	defer set.lock.Unlock()

	_, exist := set.container[val]
	return exist
}

// NewSyncSet returns a new Set
func NewSyncSet[K comparable]() *SyncSet[K] {
	return &SyncSet[K]{
		container: make(map[K]struct{}),
	}
}

type SyncSet[K comparable] struct {
	container map[K]struct{}
	lock      sync.RWMutex
}

// Add adds a string to the set
func (set *SyncSet[K]) Add(val K) {
	set.lock.Lock()
	defer set.lock.Unlock()

	set.container[val] = struct{}{}
}

// Add adds a string to the set
func (set *SyncSet[K]) AddAll(vals *SyncSet[K]) {
	vals.Range(func(val K) {
		set.lock.Lock()
		defer set.lock.Unlock()
		set.container[val] = struct{}{}
	})
}

// Remove removes a string from the set
func (set *SyncSet[K]) Remove(val K) {
	set.lock.Lock()
	defer set.lock.Unlock()

	delete(set.container, val)
}

func (set *SyncSet[K]) ToSlice() []K {
	set.lock.RLock()
	defer set.lock.RUnlock()

	ret := make([]K, 0, len(set.container))
	for k := range set.container {
		ret = append(ret, k)
	}
	return ret
}

func (set *SyncSet[K]) Range(fn func(val K)) {
	set.lock.RLock()
	snapshot := map[K]struct{}{}
	for k := range set.container {
		snapshot[k] = struct{}{}
	}
	set.lock.RUnlock()

	for k := range snapshot {
		fn(k)
	}
}

func (set *SyncSet[K]) Len() int {
	set.lock.RLock()
	defer set.lock.RUnlock()

	return len(set.container)
}

// Contains contains target value
func (set *SyncSet[K]) Contains(val K) bool {
	set.lock.Lock()
	defer set.lock.Unlock()

	_, exist := set.container[val]
	return exist
}
