package config

import (
	"testing"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
)

func TestConfigGrayReleaseActiveKeyIncludesReleaseName(t *testing.T) {
	normal := ConfigFileReleaseKey{
		Namespace:   "default",
		Group:       "group-a",
		FileName:    "app.yaml",
		Name:        "normal-v1",
		ReleaseType: ReleaseTypeNormal,
	}
	if got, want := normal.ActiveKey(), "default@group-a@app.yaml@normal"; got != want {
		t.Fatalf("normal active key changed: got %q want %q", got, want)
	}

	grayA := normal
	grayA.Name = "gray-a"
	grayA.ReleaseType = ReleaseTypeGray
	grayB := grayA
	grayB.Name = "gray-b"

	if grayA.ActiveKey() == grayB.ActiveKey() {
		t.Fatalf("gray active keys must include release name, got %q for both releases", grayA.ActiveKey())
	}
	if got, want := grayA.ActiveKey(), "default@group-a@app.yaml@gray@gray-a"; got != want {
		t.Fatalf("gray active key mismatch: got %q want %q", got, want)
	}
}

func TestConfigGrayResourceIncludesReleaseName(t *testing.T) {
	release := &SimpleConfigFileRelease{
		ConfigFileReleaseKey: &ConfigFileReleaseKey{
			Namespace:   "default",
			Group:       "group-a",
			FileName:    "app.yaml",
			Name:        "gray-a",
			ReleaseType: ReleaseTypeGray,
		},
	}

	if got, want := release.GetGrayResource(), string(rules.GrayModuleConfig)+"@default@group-a@app.yaml@gray-a"; got != want {
		t.Fatalf("gray resource key mismatch: got %q want %q", got, want)
	}
}

func TestConfigFileReleaseAPIIncludesAuditFields(t *testing.T) {
	createTime := time.Unix(1720000000, 0)
	modifyTime := createTime.Add(2 * time.Minute)
	apiRelease := ToConfiogFileReleaseApi(&ConfigFileRelease{
		SimpleConfigFileRelease: &SimpleConfigFileRelease{
			ConfigFileReleaseKey: &ConfigFileReleaseKey{
				Name:      "gray-a",
				Namespace: "default",
				Group:     "group-a",
				FileName:  "app.yaml",
			},
			CreateTime: createTime,
			CreateBy:   "admin",
			ModifyTime: modifyTime,
			ModifyBy:   "admin",
		},
	})

	if apiRelease.GetCtime() == "" || apiRelease.GetMtime() == "" {
		t.Fatalf("release api must expose ctime/mtime, got ctime=%q mtime=%q", apiRelease.GetCtime(), apiRelease.GetMtime())
	}
	if apiRelease.GetCreateBy() != "admin" || apiRelease.GetModifyBy() != "admin" {
		t.Fatalf("release api must expose create_by/modify_by, got create_by=%q modify_by=%q", apiRelease.GetCreateBy(), apiRelease.GetModifyBy())
	}
}
