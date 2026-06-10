/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, Tencent. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 */

package sqldb

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
)

type jsonContainsFields map[string]interface{}

func (m jsonContainsFields) Match(v driver.Value) bool {
	raw, ok := v.(string)
	if !ok {
		data, ok := v.([]byte)
		if !ok {
			return false
		}
		raw = string(data)
	}
	var actual map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &actual); err != nil {
		return false
	}
	for k, expected := range m {
		if !reflect.DeepEqual(actual[k], expected) {
			return false
		}
	}
	return true
}
