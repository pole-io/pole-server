package v1

import (
	"testing"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func assertConfigResponseData[T proto.Message](t *testing.T, data *anypb.Any, out T) {
	t.Helper()
	if data == nil {
		t.Fatal("expected config response data to contain the returned object")
	}
	if err := anypb.UnmarshalTo(data, out, proto.UnmarshalOptions{}); err != nil {
		t.Fatalf("unmarshal config response data: %v", err)
	}
}

func TestConfigSingleObjectResponsesIncludeData(t *testing.T) {
	groupResp := NewConfigGroupResponse(apimodel.Code_ExecuteSuccess, &apiconfig.ConfigFileGroup{
		Name:      "group-a",
		Namespace: "default",
	})
	group := &apiconfig.ConfigFileGroup{}
	assertConfigResponseData(t, groupResp.Data, group)
	if group.GetName() != "group-a" {
		t.Fatalf("unexpected group name: %q", group.GetName())
	}

	fileResp := NewConfigFileResponse(apimodel.Code_ExecuteSuccess, &apiconfig.ConfigFile{
		Name:      "app.yaml",
		Namespace: "default",
		Group:     "group-a",
	})
	file := &apiconfig.ConfigFile{}
	assertConfigResponseData(t, fileResp.Data, file)
	if file.GetName() != "app.yaml" {
		t.Fatalf("unexpected config file name: %q", file.GetName())
	}

	releaseResp := NewConfigFileReleaseResponse(apimodel.Code_ExecuteSuccess, &apiconfig.ConfigFileRelease{
		Name:      "release-a",
		Namespace: "default",
		Group:     "group-a",
		FileName:  "app.yaml",
	})
	release := &apiconfig.ConfigFileRelease{}
	assertConfigResponseData(t, releaseResp.Data, release)
	if release.GetFileName() != "app.yaml" {
		t.Fatalf("unexpected release file name: %q", release.GetFileName())
	}
}

func TestConfigEncryptAlgorithmResponseIncludesAlgorithms(t *testing.T) {
	resp := NewConfigEncryptAlgorithmResponse(apimodel.Code_ExecuteSuccess, []string{"AES", "RSA"})

	data := &structpb.Struct{}
	assertConfigResponseData(t, resp.Data, data)
	values := data.GetFields()["algorithms"].GetListValue().GetValues()
	if len(values) != 2 {
		t.Fatalf("expected two algorithms, got %d", len(values))
	}
	if values[0].GetStringValue() != "AES" || values[1].GetStringValue() != "RSA" {
		t.Fatalf("unexpected algorithms: %#v", values)
	}
}
