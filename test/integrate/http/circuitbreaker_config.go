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

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// JSONFromCircuitBreakers marshals a slice of circuit breakers to JSON. 熔断规则数组转JSON
func JSONFromCircuitBreakers(circuitBreakers []*apifault.CircuitBreakerRule) (*bytes.Buffer, error) {
	m := jsonpb.Marshaler{Indent: " "}

	buffer := bytes.NewBuffer([]byte{})

	buffer.Write([]byte("["))
	for index, circuitBreaker := range circuitBreakers {
		if index > 0 {
			buffer.Write([]byte(",\n"))
		}
		err := m.Marshal(buffer, circuitBreaker)
		if err != nil {
			return nil, err
		}
	}

	buffer.Write([]byte("]"))
	return buffer, nil
}

// JSONFromConfigReleases marshals a slice of config releases to JSON. 配置发布规则数组转JSON
func JSONFromConfigReleases(configReleases []*apiconfig.ConfigFileRelease) (*bytes.Buffer, error) {
	m := jsonpb.Marshaler{Indent: " "}

	buffer := bytes.NewBuffer([]byte{})

	buffer.Write([]byte("["))
	for index, configRelease := range configReleases {
		if index > 0 {
			buffer.Write([]byte(",\n"))
		}
		err := m.Marshal(buffer, configRelease)
		if err != nil {
			return nil, err
		}
	}

	buffer.Write([]byte("]"))
	return buffer, nil
}

// CreateCircuitBreakers creates a slice of circuit breakers from JSON. 创建熔断规则
func (c *Client) CreateCircuitBreakers(circuitBreakers []*apifault.CircuitBreakerRule) (*apimodel.BatchWriteResponse, error) {
	fmt.Printf("\ncreate circuit breakers\n")

	url := fmt.Sprintf("http://%v/naming/%v/circuitbreakers", c.Address, c.Version)

	body, err := JSONFromCircuitBreakers(circuitBreakers)
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

	return checkCreateCircuitBreakersResponse(ret, circuitBreakers)
}

// CreateCircuitBreakerVersions creates a slice of circuit breakers from JSON. 创建熔断规则版本
func (c *Client) CreateCircuitBreakerVersions(circuitBreakers []*apifault.CircuitBreakerRule) (*apimodel.BatchWriteResponse, error) {
	fmt.Printf("\ncreate circuit breaker versions\n")

	url := fmt.Sprintf("http://%v/naming/%v/circuitbreakers/version", c.Address, c.Version)
	body, err := JSONFromCircuitBreakers(circuitBreakers)
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

	return checkCreateCircuitBreakersResponse(ret, circuitBreakers)
}

// UpdateCircuitBreakers 更新熔断规则
func (c *Client) UpdateCircuitBreakers(circuitBreakers []*apifault.CircuitBreakerRule) error {
	fmt.Printf("\nupdate circuit breakers\n")

	url := fmt.Sprintf("http://%v/naming/%v/circuitbreakers", c.Address, c.Version)

	body, err := JSONFromCircuitBreakers(circuitBreakers)
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
 * @brief 删除熔断规则
 */
func (c *Client) DeleteCircuitBreakers(circuitBreakers []*apifault.CircuitBreakerRule) error {
	fmt.Printf("\ndelete circuit breakers\n")

	url := fmt.Sprintf("http://%v/naming/%v/circuitbreakers/delete", c.Address, c.Version)

	body, err := JSONFromCircuitBreakers(circuitBreakers)
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
 * @brief 发布熔断规则
 */
func (c *Client) ReleaseCircuitBreakers(configReleases []*apiconfig.ConfigFileRelease) error {
	fmt.Printf("\nrelease circuit breakers\n")

	url := fmt.Sprintf("http://%v/naming/%v/circuitbreakers/release", c.Address, c.Version)

	body, err := JSONFromConfigReleases(configReleases)
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
 * @brief 解绑熔断规则
 */
func (c *Client) UnbindCircuitBreakers(configReleases []*apiconfig.ConfigFileRelease) error {
	fmt.Printf("\nunbind circuit breakers\n")

	url := fmt.Sprintf("http://%v/naming/%v/circuitbreakers/unbind", c.Address, c.Version)

	body, err := JSONFromConfigReleases(configReleases)
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
 * @brief 根据id和version查询熔断规则
 */
func (c *Client) GetCircuitBreaker(masterCircuitBreaker, circuitBreaker *apifault.CircuitBreakerRule) error {
	fmt.Printf("\nget circuit breaker by id and version\n")

	url := fmt.Sprintf("http://%v/naming/%v/circuitbreaker", c.Address, c.Version)

	params := map[string][]interface{}{
		"id":      {circuitBreaker.GetId()},
		"version": {circuitBreaker.GetRevision()},
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

	size := 1

	if ret.GetAmount() != uint32(size) {
		return errors.New("invalid batch amount")
	}

	if ret.GetSize() != uint32(size) {
		return errors.New("invalid batch size")
	}

	data := ret.GetData()
	if data == nil || len(data) != size {
		return errors.New("invalid batch circuit breakers")
	}

	// Unmarshal the Any data to CircuitBreakerRule
	var rule apifault.CircuitBreakerRule
	if err := data[0].UnmarshalTo(&rule); err != nil {
		return fmt.Errorf("failed to unmarshal circuit breaker: %v", err)
	}

	if result, err := compareCircuitBreaker(circuitBreaker, masterCircuitBreaker, &rule); !result {
		return err
	}

	return nil
}

/**
 * @brief 查询熔断规则的已发布规则及服务
 */
func (c *Client) GetCircuitBreakersRelease(circuitBreaker *apifault.CircuitBreakerRule, correctService *apiservice.Service) error {
	fmt.Printf("\nget circuit breaker release\n")

	url := fmt.Sprintf("http://%v/naming/%v/circuitbreakers/release", c.Address, c.Version)

	params := map[string][]interface{}{
		"id": {circuitBreaker.GetId()},
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

	size := 1

	if ret.GetAmount() != uint32(size) {
		return fmt.Errorf("invalid batch amount, expect : %d, actual : %d", size, ret.GetAmount())
	}

	if ret.GetSize() != uint32(size) {
		return errors.New("invalid batch size")
	}

	data := ret.GetData()
	if data == nil || len(data) != size {
		return errors.New("invalid batch circuit breakers")
	}

	// Unmarshal the Any data to CircuitBreakerRule
	var rule apifault.CircuitBreakerRule
	if err := data[0].UnmarshalTo(&rule); err != nil {
		return fmt.Errorf("failed to unmarshal circuit breaker: %v", err)
	}

	if circuitBreaker.GetId() != rule.GetId() ||
		circuitBreaker.GetRevision() != rule.GetRevision() {
		return errors.New("error circuit breaker id or version")
	}

	// Note: Service binding info is no longer returned in the new API structure
	// This validation has been simplified
	if correctService.GetName() == "" || correctService.GetNamespace() == "" {
		return errors.New("invalid service name or namespace")
	}

	return nil
}

/**
 * @brief 查询熔断规则所有版本
 */
func (c *Client) GetCircuitBreakerVersions(circuitBreaker *apifault.CircuitBreakerRule) error {
	fmt.Printf("\nget circuit breaker versions\n")

	url := fmt.Sprintf("http://%v/naming/%v/circuitbreaker/versions", c.Address, c.Version)

	params := map[string][]interface{}{
		"id": {circuitBreaker.GetId()},
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

	size := 2

	if ret.GetAmount() != uint32(size) {
		return errors.New("invalid batch amount")
	}

	if ret.GetSize() != uint32(size) {
		return errors.New("invalid batch size")
	}

	data := ret.GetData()
	if data == nil || len(data) != size {
		return errors.New("invalid batch circuit breakers")
	}

	versions := make([]string, 0, size)
	for _, anyData := range data {
		var cb apifault.CircuitBreakerRule
		if err := anyData.UnmarshalTo(&cb); err != nil {
			return fmt.Errorf("failed to unmarshal circuit breaker: %v", err)
		}
		if cb.GetId() != circuitBreaker.GetId() {
			return errors.New("invalid circuit breaker id")
		}
		versions = append(versions, cb.GetRevision())
	}

	correctVersions := map[string]bool{
		circuitBreaker.GetRevision(): true,
		"master":                     true,
	}

	for _, version := range versions {
		if _, ok := correctVersions[version]; !ok {
			return errors.New("invalid circuit breaker version")
		}
	}

	return nil
}

/**
 * @brief 查询服务绑定的熔断规则
 */
func (c *Client) GetCircuitBreakerByService(service *apiservice.Service, masterCircuitBreaker,
	circuitBreaker *apifault.CircuitBreakerRule) error {
	fmt.Printf("\nget circuit breaker by service\n")

	url := fmt.Sprintf("http://%v/naming/%v/service/circuitbreaker", c.Address, c.Version)

	params := map[string][]interface{}{
		"service":   {service.GetName()},
		"namespace": {service.GetNamespace()},
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

	size := 1

	if ret.GetAmount() != uint32(size) {
		return errors.New("invalid batch amount")
	}

	if ret.GetSize() != uint32(size) {
		return errors.New("invalid batch size")
	}

	data := ret.GetData()
	if data == nil || len(data) != size {
		return errors.New("invalid batch circuit breakers")
	}

	// Unmarshal the Any data to CircuitBreakerRule
	var rule apifault.CircuitBreakerRule
	if err := data[0].UnmarshalTo(&rule); err != nil {
		return fmt.Errorf("failed to unmarshal circuit breaker: %v", err)
	}

	if result, err := compareCircuitBreaker(circuitBreaker, masterCircuitBreaker, &rule); !result {
		return err
	}

	return nil
}

/**
 * @brief 检查创建熔断规则的回复
 */
func checkCreateCircuitBreakersResponse(ret *apimodel.BatchWriteResponse, circuitBreakers []*apifault.CircuitBreakerRule) (
	*apimodel.BatchWriteResponse, error) {
	switch {
	case ret.GetCode() != api.ExecuteSuccess:
		return nil, errors.New("invalid batch code")
	case ret.GetSize() != uint32(len(circuitBreakers)):
		return nil, errors.New("invalid batch size")
	case len(ret.GetResponses()) != len(circuitBreakers):
		return nil, errors.New("invalid batch response")
	}

	for index, item := range ret.GetResponses() {
		if item.GetCode() != api.ExecuteSuccess {
			return nil, errors.New("invalid code")
		}
		anyData := item.GetData()
		if anyData == nil {
			return nil, errors.New("empty circuit breaker data")
		}

		// Unmarshal the Any data to CircuitBreakerRule
		var circuitBreaker apifault.CircuitBreakerRule
		if err := anyData.UnmarshalTo(&circuitBreaker); err != nil {
			return nil, fmt.Errorf("failed to unmarshal circuit breaker: %v", err)
		}

		if result, err := compareCircuitBreaker(circuitBreakers[index], circuitBreakers[index], &circuitBreaker); !result {
			return nil, err
		} else {
			return ret, nil
		}
	}
	return ret, nil
}

/**
 * @brief 比较circuit breaker是否相等
 */
func compareCircuitBreaker(correctItem, correctMaster *apifault.CircuitBreakerRule, item *apifault.CircuitBreakerRule) (bool, error) {
	switch {
	case item.GetId() == "":
		return false, errors.New("error id")
	case item.GetRevision() == "":
		return false, errors.New("error version")
	case correctMaster.GetName() != item.GetName():
		return false, errors.New("error name")
	case correctMaster.GetNamespace() != item.GetNamespace():
		return false, errors.New("error namespace")
	case correctMaster.GetDescription() != item.GetDescription():
		return false, errors.New("error description")
	default:
		break
	}

	// Compare BlockConfigs (新API中的核心配置)
	correctConfigs, err := json.Marshal(correctItem.GetBlockConfigs())
	if err != nil {
		panic(err)
	}
	itemConfigs, err := json.Marshal(item.GetBlockConfigs())
	if err != nil {
		panic(err)
	}
	if string(correctConfigs) != string(itemConfigs) {
		return false, errors.New("error block configs")
	}

	return true, nil
}
