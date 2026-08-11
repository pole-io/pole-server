package skillmarketplace

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type ScanEvidence struct {
	Scanner  string   `json:"scanner"`
	Files    int      `json:"files"`
	Scripts  int      `json:"scripts"`
	Findings []string `json:"findings,omitempty"`
}

func ScanBundle(bundle []byte) (string, string, error) {
	zr, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		return "failed", "", err
	}
	evidence := ScanEvidence{Scanner: "pole-static-v1", Files: len(zr.File)}
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			continue
		}
		extension := strings.ToLower(filepath.Ext(file.Name))
		if extension == ".sh" || extension == ".py" || extension == ".js" || extension == ".ts" || extension == ".ps1" {
			evidence.Scripts++
		}
		rc, err := file.Open()
		if err != nil {
			return "failed", "", err
		}
		content, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			return "failed", "", readErr
		}
		if bytes.Contains(content, []byte("-----BEGIN PRIVATE KEY-----")) ||
			bytes.Contains(content, []byte("-----BEGIN OPENSSH PRIVATE KEY-----")) {
			evidence.Findings = append(evidence.Findings, fmt.Sprintf("embedded private key in %s", file.Name))
		}
		if len(content) >= 4 && (bytes.Equal(content[:4], []byte{0x7f, 'E', 'L', 'F'}) ||
			bytes.Equal(content[:4], []byte{'M', 'Z', 0x90, 0x00}) ||
			bytes.Equal(content[:4], []byte{0xcf, 0xfa, 0xed, 0xfe})) {
			evidence.Findings = append(evidence.Findings, fmt.Sprintf("native executable in %s", file.Name))
		}
	}
	payload, err := json.Marshal(evidence)
	if err != nil {
		return "failed", "", err
	}
	if len(evidence.Findings) > 0 {
		return "failed", string(payload), nil
	}
	return "passed", string(payload), nil
}
