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
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/golang/protobuf/jsonpb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

/**
 * @brief 服务数组转JSON
 */
func JSONFromServices(services []*apiservice.Service) (*bytes.Buffer, error) {
	m := jsonpb.Marshaler{Indent: " "}

	buffer := bytes.NewBuffer([]byte{})

	buffer.Write([]byte("["))
	for index, service := range services {
		if index > 0 {
			buffer.Write([]byte(",\n"))
		}
		err := m.Marshal(buffer, service)
		if err != nil {
			return nil, err
		}
	}

	buffer.Write([]byte("]"))
	return buffer, nil
}

/**
 * @brief 创建服务
 */
func (c *Client) CreateServices(services []*apiservice.Service) (*apimodel.BatchWriteResponse, error) {
	fmt.Printf("\ncreate services\n")

	url := fmt.Sprintf("http://%v/naming/%v/services", c.Address, c.Version)

	body, err := JSONFromServices(services)
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

	return checkCreateServicesResponse(ret, services)
}

/**
 * @brief 删除服务
 */
func (c *Client) DeleteServices(services []*apiservice.Service) error {
	fmt.Printf("\ndelete services\n")

	url := fmt.Sprintf("http://%v/naming/%v/services/delete", c.Address, c.Version)

	body, err := JSONFromServices(services)
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
 * @brief 更新服务
 */
func (c *Client) UpdateServices(services []*apiservice.Service) error {
	fmt.Printf("\nupdate services\n")

	url := fmt.Sprintf("http://%v/naming/%v/services", c.Address, c.Version)

	body, err := JSONFromServices(services)
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
 * @brief 查询服务
 */
func (c *Client) GetServices(services []*apiservice.Service) error {
	fmt.Printf("\nget services\n")

	url := fmt.Sprintf("http://%v/naming/%v/services", c.Address, c.Version)

	params := map[string][]interface{}{
		"namespace": {services[0].GetNamespace()},
	}
	time.Sleep(2 * time.Second)
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

	servicesSize := len(services)

	if ret.GetAmount() != uint32(servicesSize) {
		return fmt.Errorf("invalid batch amount, expect %d, actual is %v", servicesSize, ret.GetAmount())
	}

	if ret.GetSize() != uint32(servicesSize) {
		return errors.New("invalid batch size")
	}

	collection := make(map[string]*apiservice.Service)
	for _, service := range services {
		collection[service.GetName()] = service
	}

	data := ret.GetData()
	if data == nil || len(data) != servicesSize {
		return errors.New("invalid batch services")
	}

	for _, anyData := range data {
		var item apiservice.Service
		if err := anyData.UnmarshalTo(&item); err != nil {
			return fmt.Errorf("failed to unmarshal service: %v", err)
		}
		if correctItem, ok := collection[item.GetName()]; ok {
			if result := compareService(correctItem, &item); !result {
				return errors.New("invalid service")
			}
		} else {
			return errors.New("invalid service")
		}
	}
	return nil
}

/**
 * @brief 检查创建服务的回复
 */
func checkCreateServicesResponse(ret *apimodel.BatchWriteResponse, services []*apiservice.Service) (
	// #lizard forgives
	*apimodel.BatchWriteResponse, error) {
	if ret.GetCode() != api.ExecuteSuccess {
		return nil, errors.New("invalid batch code")
	}

	servicesSize := len(services)
	if ret.GetSize() != uint32(servicesSize) {
		return nil, errors.New("invalid batch size")
	}

	items := ret.GetResponses()
	if items == nil || len(items) != servicesSize {
		return nil, errors.New("invalid batch response")
	}

	for index, item := range items {
		if item.GetCode() != api.ExecuteSuccess {
			return nil, errors.New("invalid code")
		}

		if item.GetData() == nil {
			return nil, errors.New("empty service")
		}

		var service apiservice.Service
		if err := item.GetData().UnmarshalTo(&service); err != nil {
			return nil, fmt.Errorf("failed to unmarshal service: %v", err)
		}

		name := services[index].GetName()
		if service.GetName() == "" || service.GetName() != name {
			return nil, errors.New("invalid service name")
		}

		namespace := services[index].GetNamespace()
		if service.GetNamespace() == "" || service.GetNamespace() != namespace {
			return nil, errors.New("invalid namespace")
		}

		if service.GetToken() == "" {
			return nil, errors.New("invalid service token")
		}
	}
	return ret, nil
}

/**
 * @brief 比较service是否相等
 */
func compareService(correctItem *apiservice.Service, item *apiservice.Service) bool {
	correctName := correctItem.GetName()
	correctNamespace := correctItem.GetNamespace()
	correctMeta := correctItem.GetMetadata()
	correctPorts := correctItem.GetPorts()
	correctBusiness := correctItem.GetBusiness()
	correctDepartment := correctItem.GetDepartment()
	correctCmdbMod1 := correctItem.GetCmdbMod1()
	correctCmdbMod2 := correctItem.GetCmdbMod2()
	correctCmdbMod3 := correctItem.GetCmdbMod3()
	correctComment := correctItem.GetComment()

	name := item.GetName()
	namespace := item.GetNamespace()
	meta := item.GetMetadata()
	ports := item.GetPorts()
	business := item.GetBusiness()
	department := item.GetDepartment()
	cmdbMod1 := item.GetCmdbMod1()
	cmdbMod2 := item.GetCmdbMod2()
	cmdbMod3 := item.GetCmdbMod3()
	comment := item.GetComment()

	if correctName == name && correctNamespace == namespace && reflect.DeepEqual(correctMeta, meta) &&
		correctPorts == ports && correctBusiness == business && correctDepartment == department &&
		correctCmdbMod1 == cmdbMod1 && correctCmdbMod2 == cmdbMod2 && correctCmdbMod3 == cmdbMod3 &&
		correctComment == comment {
		return true
	}
	return false
}
