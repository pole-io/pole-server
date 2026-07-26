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

package resource

import (
	"reflect"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

func Test_parseNodeID(t *testing.T) {
	type args struct {
		nodeID string
	}
	tests := []struct {
		name                 string
		args                 args
		wantRunType          string
		wantPolarisNamespace string
		wantUuid             string
		wantHostIP           string
	}{
		{
			name: "test-1",
			args: args{
				nodeID: "default/12345~127.0.0.1",
			},
			wantRunType:          "sidecar",
			wantPolarisNamespace: "default",
			wantUuid:             "12345",
			wantHostIP:           "127.0.0.1",
		},
		{
			name: "test-1",
			args: args{
				nodeID: "sidecar~bookinfo/productpage-v1-7bc8c4cb99-c69gh~10.0.0.208",
			},
			wantRunType:          string(RunTypeSidecar),
			wantPolarisNamespace: "bookinfo",
			wantUuid:             "productpage-v1-7bc8c4cb99-c69gh",
			wantHostIP:           "10.0.0.208",
		},
		{
			name: "test-1",
			args: args{
				nodeID: "gateway~default/12345~127.0.0.1",
			},
			wantRunType:          "gateway",
			wantPolarisNamespace: "default",
			wantUuid:             "12345",
			wantHostIP:           "127.0.0.1",
		},
		{
			name: "test-1",
			args: args{
				nodeID: "12345~127.0.0.1",
			},
			wantRunType:          "",
			wantPolarisNamespace: "",
			wantUuid:             "",
			wantHostIP:           "",
		},
		{
			name: "test-1",
			args: args{
				nodeID: "default/127.0.0.1",
			},
			wantRunType:          "",
			wantPolarisNamespace: "",
			wantUuid:             "",
			wantHostIP:           "",
		},
		{
			name: "test-1",
			args: args{
				nodeID: "sidecar~default/127.0.0.1",
			},
			wantRunType:          "",
			wantPolarisNamespace: "",
			wantUuid:             "",
			wantHostIP:           "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRunType, gotPolarisNamespace, gotUuid, gotHostIP := ParseNodeID(tt.args.nodeID)
			if gotRunType != tt.wantRunType {
				t.Errorf("parseNodeID() gotRunType = %v, want %v", gotRunType, tt.wantRunType)
			}
			if gotPolarisNamespace != tt.wantPolarisNamespace {
				t.Errorf("parseNodeID() gotPolarisNamespace = %v, want %v", gotPolarisNamespace, tt.wantPolarisNamespace)
			}
			if gotUuid != tt.wantUuid {
				t.Errorf("parseNodeID() gotUuid = %v, want %v", gotUuid, tt.wantUuid)
			}
			if gotHostIP != tt.wantHostIP {
				t.Errorf("parseNodeID() gotHostIP = %v, want %v", gotHostIP, tt.wantHostIP)
			}
		})
	}
}

func TestXDSNodeManagerRemovesIndexesOnlyAfterLastStreamCloses(t *testing.T) {
	manager := NewXDSNodeManager()
	node := &corev3.Node{Id: "sidecar~prod/checkout~127.0.0.1"}
	manager.AddNodeIfAbsent(1, node)
	manager.AddNodeIfAbsent(2, node)

	if removed := manager.DelNode(1); removed {
		t.Fatal("first stream close must retain node")
	}
	if got := manager.ListEnvoyNodes(); len(got) != 1 {
		t.Fatalf("ListEnvoyNodes() after first close = %d, want 1", len(got))
	}
	if removed := manager.DelNode(2); !removed {
		t.Fatal("last stream close must remove node")
	}
	if got := manager.ListEnvoyNodes(); len(got) != 0 {
		t.Fatalf("ListEnvoyNodes() after last close = %d, want 0", len(got))
	}
}

func TestXDSClientGovernanceLabelsOnlyUsesExplicitPrefix(t *testing.T) {
	client := &XDSClient{Metadata: map[string]string{
		"pole.io/governance-label.env":            "canary",
		"pole.io/governance-label.region":         "ap-shanghai",
		"sidecar.polarismesh.cn/serviceName":      "checkout",
		"sidecar.polarismesh.cn/serviceNamespace": "prod",
		"sidecar.polarismesh.cn/openOnDemand":     "true",
	}}
	want := map[string]string{
		"env":    "canary",
		"region": "ap-shanghai",
	}
	if got := client.GovernanceLabels(); !reflect.DeepEqual(got, want) {
		t.Fatalf("GovernanceLabels() = %#v, want %#v", got, want)
	}
}
