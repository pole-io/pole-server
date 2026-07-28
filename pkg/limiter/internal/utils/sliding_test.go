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

package utils

import (
	"testing"
)

func TestSlidingWindow_SlideFive(t *testing.T) {
	slidingWindow := NewSlidingWindow(5, 1000)

	slidingWindow.AddAndGetCurrent(1000, 1000, 10)
	slidingWindow.AddAndGetCurrent(1200, 1200, 40)
	slidingWindow.AddAndGetCurrent(1700, 1700, 15)
	value := slidingWindow.AddAndGetCurrent(2200, 2200, 30)
	if value != 45 {
		t.Fatalf("value is %d, invalid", value)
	}
}

func TestSlidingWindow_SlideOne(t *testing.T) {
	slidingWindow := NewSlidingWindow(1, 1000)

	slidingWindow.AddAndGetCurrent(1000, 1000, 10)
	slidingWindow.AddAndGetCurrent(1200, 1200, 40)
	slidingWindow.AddAndGetCurrent(1800, 1800, 15)
	value := slidingWindow.AddAndGetCurrent(2000, 2000, 30)
	if value != 30 {
		t.Fatalf("value is %d, invalid", value)
	}
}
