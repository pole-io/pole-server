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
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
)

// InstanceArgs 用于通过服务实例查询服务的参数
type InstanceArgs struct {
	Hosts []string
	Ports []uint32
	Meta  map[string]string
}

type InstanceMetadataRequest struct {
	InstanceID string
	Revision   string
	Keys       []string
	Metadata   map[string]string
}

var storeCodeAPICodeMap = map[StatusCode]apimodel.Code{
	EmptyParamsErr:             apimodel.Code_InvalidParameter,
	OutOfRangeErr:              apimodel.Code_InvalidParameter,
	DataConflictErr:            apimodel.Code_DataConflict,
	NotFoundNamespace:          apimodel.Code_NotFoundNamespace,
	NotFoundService:            apimodel.Code_NotFoundService,
	NotFoundMasterConfig:       apimodel.Code_NotFoundMasterConfig,
	NotFoundTagConfigOrService: apimodel.Code_NotFoundTagConfigOrService,
	ExistReleasedConfig:        apimodel.Code_ExistReleasedConfig,
	DuplicateEntryErr:          apimodel.Code_ExistedResource,
}

// StoreCode2APICode store code to api code
func StoreCode2APICode(err error) apimodel.Code {
	code := Code(err)
	apiCode, ok := storeCodeAPICodeMap[code]
	if ok {
		return apiCode
	}

	return apimodel.Code_StoreLayerException
}
