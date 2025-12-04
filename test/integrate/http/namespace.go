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

	"github.com/golang/protobuf/jsonpb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

// JSONFromNamespaces 将命名空间数组转换为JSON
func JSONFromNamespaces(namespaces []*apimodel.Namespace) (*bytes.Buffer, error) {
	m := jsonpb.Marshaler{Indent: " "}

	buffer := bytes.NewBuffer([]byte{})

	buffer.Write([]byte("["))
	for index, namespace := range namespaces {
		if index > 0 {
			buffer.Write([]byte(",\n"))
		}
		err := m.Marshal(buffer, namespace)
		if err != nil {
			return nil, err
		}
	}

	buffer.Write([]byte("]"))

	return buffer, nil
}

// CreateNamespaces 创建命名空间
func (c *Client) CreateNamespaces(namespaces []*apimodel.Namespace) (*apimodel.BatchWriteResponse, error) {
	fmt.Printf("\ncreate namespaces\n")

	url := fmt.Sprintf("http://%v/naming/%v/namespaces", c.Address, c.Version)

	body, err := JSONFromNamespaces(namespaces)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	response, err := c.SendRequestWithRequestID("CreateNamespaces", "POST", url, body)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	ret, err := GetBatchWriteResponse(response)
	if err != nil {
		fmt.Printf("%v\n", err)
		return ret, err
	}

	return checkCreateNamespacesResponse(ret, namespaces)
}

// DeleteNamespaces 删除命名空间
func (c *Client) DeleteNamespaces(namespaces []*apimodel.Namespace) error {
	fmt.Printf("\ndelete namespaces\n")

	url := fmt.Sprintf("http://%v/naming/%v/namespaces/delete", c.Address, c.Version)

	body, err := JSONFromNamespaces(namespaces)
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

// DeleteNamespaces 删除命名空间
func (c *Client) DeleteNamespacesGetResp(namespaces []*apimodel.Namespace) (*apimodel.BatchWriteResponse, error) {
	fmt.Printf("\ndelete namespaces\n")

	url := fmt.Sprintf("http://%v/naming/%v/namespaces/delete", c.Address, c.Version)

	body, err := JSONFromNamespaces(namespaces)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	response, err := c.SendRequest("POST", url, body)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	resp, err := GetBatchWriteResponse(response)
	if err != nil {
		fmt.Printf("%v\n", err)
		return resp, err
	}

	return resp, nil
}

// UpdateNamesapces 更新命名空间
func (c *Client) UpdateNamesapces(namespaces []*apimodel.Namespace) error {
	fmt.Printf("\nupdate namespaces\n")

	url := fmt.Sprintf("http://%v/naming/%v/namespaces", c.Address, c.Version)

	body, err := JSONFromNamespaces(namespaces)
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

// GetNamespaces 查询命名空间
func (c *Client) GetNamespaces(namespaces []*apimodel.Namespace) ([]*apimodel.Namespace, error) {
	fmt.Printf("\nget namespaces\n")

	url := fmt.Sprintf("http://%v/naming/%v/namespaces", c.Address, c.Version)

	params := map[string][]interface{}{
		"name": {namespaces[0].GetName(), namespaces[1].GetName()},
	}

	url = c.CompleteURL(url, params)
	response, err := c.SendRequestWithRequestID("GetNamespaces", "GET", url, nil)
	if err != nil {
		return nil, err
	}

	ret, err := GetBatchQueryResponse(response)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil, err
	}

	if ret.GetCode() != api.ExecuteSuccess {
		return nil, errors.New("invalid batch code")
	}

	namespacesSize := len(namespaces)

	if ret.GetAmount() != uint32(namespacesSize) {
		return nil, fmt.Errorf("invalid batch amount: %d %d", ret.GetAmount(), namespacesSize)
	}

	if ret.GetSize() != uint32(namespacesSize) {
		return nil, errors.New("invalid batch size")
	}

	collection := make(map[string]*apimodel.Namespace)
	for _, namespace := range namespaces {
		collection[namespace.GetName()] = namespace
	}

	data := ret.GetData()
	if data == nil || len(data) != namespacesSize {
		return nil, errors.New("invalid batch namespaces")
	}

	items := make([]*apimodel.Namespace, 0, len(data))
	for _, anyData := range data {
		var item apimodel.Namespace
		if err := anyData.UnmarshalTo(&item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal namespace: %v", err)
		}
		if correctItem, ok := collection[item.GetName()]; ok {
			if result := compareNamespace(correctItem, &item); !result {
				return nil, errors.New("invalid namespace")
			}
		} else {
			return nil, errors.New("invalid namespace")
		}
		items = append(items, &item)
	}
	return items, nil
}

/**
 * @brief 检查创建命名空间的回复
 */
func checkCreateNamespacesResponse(ret *apimodel.BatchWriteResponse, namespaces []*apimodel.Namespace) (
	*apimodel.BatchWriteResponse, error) {
	if ret.GetCode() != api.ExecuteSuccess {
		return nil, errors.New("invalid batch code")
	}

	namespacesSize := len(namespaces)
	if ret.GetSize() != uint32(namespacesSize) {
		return nil, errors.New("invalid batch size")
	}

	items := ret.GetResponses()
	if items == nil || len(items) != namespacesSize {
		return nil, errors.New("invalid batch response")
	}

	for index, item := range items {
		if item.GetCode() != api.ExecuteSuccess {
			return nil, errors.New("invalid code")
		}

		if item.GetData() == nil {
			return nil, errors.New("empty namespace")
		}

		var namespace apimodel.Namespace
		if err := item.GetData().UnmarshalTo(&namespace); err != nil {
			return nil, fmt.Errorf("failed to unmarshal namespace: %v", err)
		}

		name := namespaces[index].GetName()
		if namespace.GetName() == "" || namespace.GetName() != name {
			return nil, errors.New("invalid namespace name")
		}

		if namespace.GetToken() == "" {
			return nil, errors.New("invalid namespace token")
		}
	}
	return ret, nil
}

/**
 * @brief 比较namespace是否相等
 */
func compareNamespace(correctItem *apimodel.Namespace, item *apimodel.Namespace) bool {
	correctName := correctItem.GetName()
	correctComment := correctItem.GetComment()
	correctOwners := correctItem.GetOwners()

	name := item.GetName()
	comment := item.GetComment()
	owners := item.GetOwners()

	if correctName == name && correctComment == comment && correctOwners == owners {
		return true
	}
	return false
}
