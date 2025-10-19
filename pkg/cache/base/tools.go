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

package base

import (
	"sort"
	"strings"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
)

const mtimeLogIntervalSec = 120

// LogLastMtime 定时打印mtime更新结果
func LogLastMtime(lastMtimeLogged int64, lastMtime int64, prefix string) int64 {
	curTimeSec := time.Now().Unix()
	if lastMtimeLogged == 0 || curTimeSec-lastMtimeLogged >= mtimeLogIntervalSec {
		lastMtimeLogged = curTimeSec
		log.Infof("[Cache][%s] current lastMtime is %s", prefix, time.Unix(lastMtime, 0))
	}
	return lastMtimeLogged
}

func NewExpireEntry[T any](t T, maxAlive time.Duration) *ExpireEntry[T] {
	return &ExpireEntry[T]{
		data:     t,
		maxAlive: maxAlive,
	}
}

func EmptyExpireEntry[T any](t T, maxAlive time.Duration) *ExpireEntry[T] {
	return &ExpireEntry[T]{
		empty:    true,
		maxAlive: maxAlive,
	}
}

type ExpireEntry[T any] struct {
	empty      bool
	data       T
	lastAccess time.Time
	maxAlive   time.Duration
}

func (e *ExpireEntry[T]) Get() T {
	if e.empty {
		return e.data
	}
	e.lastAccess = time.Now()
	return e.data
}

func (e *ExpireEntry[T]) IsExpire() bool {
	return time.Since(e.lastAccess) > e.maxAlive
}

func SortBeforeTrim[T rules.IRule](rules []T, order string, offset, limit uint32) (uint32, []T) {
	amount := uint32(len(rules))
	if offset >= amount || limit == 0 {
		return amount, nil
	}
	sort.Slice(rules, func(i, j int) bool {
		asc := strings.ToLower(order) == "asc" || order == ""
		return OrderByMtime(rules[i], rules[j], asc)
	})
	endIdx := offset + limit
	if endIdx > amount {
		endIdx = amount
	}
	return amount, rules[offset:endIdx]
}

func OrderByMtime(a, b rules.IRule, asc bool) bool {
	if a.GetMtime().After(b.GetMtime()) {
		return asc
	}
	if a.GetMtime().Before(b.GetMtime()) {
		return false
	}
	return strings.Compare(a.GetId(), b.GetId()) < 0 && asc
}
