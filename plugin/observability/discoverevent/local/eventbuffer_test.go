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

package local

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
)

func TestEventBufferTest(t *testing.T) {

	bufferHolder := newEventBufferHolder(20)

	expectCnt := int64(0)
	for range 10 {
		now := time.Now()
		bufferHolder.Put(&svctypes.InstanceEvent{
			CreateTime: now,
		})

		expectCnt += now.Unix()
	}

	actualCnt := int64(0)

	for bufferHolder.HasNext() {
		event := bufferHolder.Next()
		switch event.(type) {
		case *svctypes.InstanceEvent:
			item := event.(*svctypes.InstanceEvent)
			actualCnt += item.CreateTime.Unix()
		}
	}

	assert.Equal(t, expectCnt, actualCnt, "cnt must be equla")

	bufferHolder.Reset()

	expectCnt = int64(0)
	for i := 20; i < 40; i++ {
		now := time.Now()
		bufferHolder.Put(&svctypes.InstanceEvent{
			CreateTime: now,
		})

		expectCnt += now.Unix()
	}

	actualCnt = int64(0)

	for bufferHolder.HasNext() {
		event := bufferHolder.Next()
		switch event.(type) {
		case *svctypes.InstanceEvent:
			item := event.(*svctypes.InstanceEvent)
			actualCnt += item.CreateTime.Unix()
		}
	}

	assert.Equal(t, expectCnt, actualCnt, "cnt must be equla")
}
