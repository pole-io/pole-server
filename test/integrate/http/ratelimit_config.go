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

package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/golang/protobuf/jsonpb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

/**
 * @brief 限流规则数组转JSON
 */
func JSONFromRateLimits(rateLimits []*apitraffic.RateLimit) (*bytes.Buffer, error) {
	m := jsonpb.Marshaler{Indent: " "}

	buffer := bytes.NewBuffer([]byte{})

	buffer.Write([]byte("["))
	for index, rateLimit := range rateLimits {
		if index > 0 {
			buffer.Write([]byte(",\n"))
		}
		err := m.Marshal(buffer, rateLimit)
		if err != nil {
			return nil, err
		}
	}

	buffer.Write([]byte("]"))
	return buffer, nil
}

/**
 * @brief 创建限流规则
 */
func (c *Client) CreateRateLimits(rateLimits []*apitraffic.RateLimit) (*apimodel.BatchWriteResponse, error) {
	fmt.Printf("\ncreate rate limits\n")

	url := fmt.Sprintf("http://%v/naming/%v/ratelimits", c.Address, c.Version)

	body, err := JSONFromRateLimits(rateLimits)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	response, err := c.SendRequest("POST", url, body)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	ret, err := GetBatchWriteResponse(response)
	if err != nil {
		fmt.Printf("%v\n", err)
		return ret, err
	}

	return checkCreateRateLimitsResponse(ret, rateLimits)
}

/**
 * @brief 删除限流规则
 */
func (c *Client) DeleteRateLimits(rateLimits []*apitraffic.RateLimit) error {
	fmt.Printf("\ndelete rate limits\n")

	url := fmt.Sprintf("http://%v/naming/%v/ratelimits/delete", c.Address, c.Version)

	body, err := JSONFromRateLimits(rateLimits)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	response, err := c.SendRequest("POST", url, body)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	_, err = GetBatchWriteResponse(response)
	if err != nil {
		if err == io.EOF {
			return nil
		}

		fmt.Printf("%v\n", err)
		return err
	}
	return nil
}

/**
 * @brief 更新限流规则
 */
func (c *Client) UpdateRateLimits(rateLimits []*apitraffic.RateLimit) error {
	fmt.Printf("\nupdate rate limits\n")

	url := fmt.Sprintf("http://%v/naming/%v/ratelimits", c.Address, c.Version)

	body, err := JSONFromRateLimits(rateLimits)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	response, err := c.SendRequest("PUT", url, body)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	_, err = GetBatchWriteResponse(response)
	if err != nil {
		if err == io.EOF {
			return nil
		}

		fmt.Printf("%v\n", err)
		return err
	}
	return nil
}

// EnableRateLimits 启用限流规则
func (c *Client) EnableRateLimits(rateLimits []*apitraffic.RateLimit) error {
	fmt.Printf("\nenable rate limits\n")

	url := fmt.Sprintf("http://%v/naming/%v/ratelimits/enable", c.Address, c.Version)

	rateLimitsEnable := make([]*apitraffic.RateLimit, 0, len(rateLimits))
	for _, rateLimit := range rateLimits {
		rateLimitsEnable = append(rateLimitsEnable, &apitraffic.RateLimit{
			Id:      rateLimit.GetId(),
			Disable: true,
		})
	}
	body, err := JSONFromRateLimits(rateLimitsEnable)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	response, err := c.SendRequest("PUT", url, body)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	_, err = GetBatchWriteResponse(response)
	if err != nil {
		if err == io.EOF {
			return nil
		}

		fmt.Printf("%v\n", err)
		return err
	}
	return nil
}

/**
 * @brief 查询限流规则
 */
func (c *Client) GetRateLimits(rateLimits []*apitraffic.RateLimit) error {
	fmt.Printf("\nget rate limits\n")

	url := fmt.Sprintf("http://%v/naming/%v/ratelimits", c.Address, c.Version)

	params := map[string][]interface{}{
		"namespace": {rateLimits[0].GetNamespace()},
	}

	url = c.CompleteURL(url, params)
	response, err := c.SendRequest("GET", url, nil)
	if err != nil {
		return err
	}

	ret, err := GetBatchQueryResponse(response)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	if ret.GetCode() != api.ExecuteSuccess {
		return errors.New("invalid batch code")
	}

	rateLimitsSize := len(rateLimits)

	if ret.GetAmount() != uint32(rateLimitsSize) {
		return errors.New("invalid batch amount")
	}

	if ret.GetSize() != uint32(rateLimitsSize) {
		return errors.New("invalid batch size")
	}

	collection := make(map[string]*apitraffic.RateLimit)
	for _, rateLimit := range rateLimits {
		collection[rateLimit.GetService()] = rateLimit
	}

	data := ret.GetData()
	if data == nil || len(data) != rateLimitsSize {
		return errors.New("invalid batch rate limits")
	}

	for _, anyData := range data {
		var item apitraffic.RateLimit
		if err := anyData.UnmarshalTo(&item); err != nil {
			return fmt.Errorf("failed to unmarshal rate limit: %v", err)
		}
		if correctItem, ok := collection[item.GetService()]; ok {
			if result, err := compareRateLimit(correctItem, &item); !result {
				return fmt.Errorf("invalid rate limit. namespace is %v, service is %v, err is %s",
					item.GetNamespace(), item.GetService(), err.Error())
			}
		} else {
			return fmt.Errorf("rate limit not found. namespace is %v, service is %v",
				item.GetNamespace(), item.GetService())
		}
	}
	return nil
}

/**
 * @brief 检查创建限流规则的回复
 */
func checkCreateRateLimitsResponse(ret *apimodel.BatchWriteResponse, rateLimits []*apitraffic.RateLimit) (
	*apimodel.BatchWriteResponse, error) {
	switch {
	case ret.GetCode() != api.ExecuteSuccess:
		return nil, errors.New("invalid batch code")
	case ret.GetSize() != uint32(len(rateLimits)):
		return nil, errors.New("invalid batch size")
	case len(ret.GetResponses()) != len(rateLimits):
		return nil, errors.New("invalid batch response")
	}

	for index, item := range ret.GetResponses() {
		if item.GetCode() != api.ExecuteSuccess {
			return nil, errors.New("invalid code")
		}
		if item.GetData() == nil {
			return nil, errors.New("empty rate limit")
		}
		var rateLimit apitraffic.RateLimit
		if err := item.GetData().UnmarshalTo(&rateLimit); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rate limit: %v", err)
		}
		if result, err := compareRateLimit(rateLimits[index], &rateLimit); !result {
			return nil, err
		}
	}
	return ret, nil
}

/**
 * @brief 比较rate limit是否相等
 */
func compareRateLimit(correctItem *apitraffic.RateLimit, item *apitraffic.RateLimit) (bool, error) {
	switch {
	case (correctItem.GetId()) != "" && (correctItem.GetId() != item.GetId()):
		return false, fmt.Errorf(
			"invalid id, expect %s, actual %s", correctItem.GetId(), item.GetId())
	case correctItem.GetService() != item.GetService():
		return false, fmt.Errorf("error service, expect %s, actual %s",
			correctItem.GetService(), item.GetService())
	case correctItem.GetNamespace() != item.GetNamespace():
		return false, fmt.Errorf("error namespace, expect %s, actual %s",
			correctItem.GetNamespace(), item.GetNamespace())
	case correctItem.GetPriority() != item.GetPriority():
		return false, fmt.Errorf("invalid priority, expect %v, actual %v",
			correctItem.GetPriority(), item.GetPriority())
	case correctItem.GetType() != item.GetType():
		return false, fmt.Errorf("error type, exepct %v, actual %v", correctItem.GetType(), item.GetType())
	case correctItem.GetDisable() != item.GetDisable():
		return false, fmt.Errorf("error disable, expect %v, actual %v",
			correctItem.GetDisable(), item.GetDisable())
	default:
		break
	}

	if equal, err := checkField(correctItem.GetRules(), item.GetRules(), "rules"); !equal {
		return equal, err
	}

	if equal, err := checkField(correctItem.GetReport(), item.GetReport(), "report"); !equal {
		return equal, err
	}

	return checkField(correctItem.GetCluster(), item.GetCluster(), "cluster")
}

/**
 * @brief 检查字段是否一致
 */
func checkField(correctItem, actualItem interface{}, name string) (bool, error) {
	expect, err := json.Marshal(correctItem)
	if err != nil {
		panic(err)
	}
	actual, err := json.Marshal(actualItem)
	if err != nil {
		panic(err)
	}

	if string(expect) != string(actual) {
		return false, fmt.Errorf("error %s, expect %s ,actual %s", name, expect, actual)
	}
	return true, nil
}
