package skillmarketplace

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	maxPreviewFileSize  = 256 << 10
	maxPreviewTotalSize = 1 << 20
)

type BundleManifestEntry struct {
	Path       string `json:"path"`
	Type       string `json:"type"`
	Size       uint64 `json:"size"`
	Mode       string `json:"mode"`
	Executable bool   `json:"executable"`
	Text       string `json:"text,omitempty"`
	Encoding   string `json:"encoding,omitempty"`
}

func InspectBundle(bundle []byte) ([]BundleManifestEntry, error) {
	zr, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		return nil, err
	}
	entries := make([]BundleManifestEntry, 0, len(zr.File))
	previewed := 0
	for _, file := range zr.File {
		mode := file.Mode()
		entryType := "file"
		if mode.IsDir() {
			entryType = "directory"
		}
		entry := BundleManifestEntry{Path: file.Name, Type: entryType, Size: file.UncompressedSize64,
			Mode: fmt.Sprintf("%04o", mode.Perm()), Executable: mode.Perm()&0o111 != 0}
		if entryType == "file" && file.UncompressedSize64 <= maxPreviewFileSize &&
			previewed+int(file.UncompressedSize64) <= maxPreviewTotalSize {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			content, readErr := io.ReadAll(io.LimitReader(rc, maxPreviewFileSize+1))
			_ = rc.Close()
			if readErr != nil {
				return nil, readErr
			}
			if len(content) <= maxPreviewFileSize && utf8.Valid(content) && !strings.ContainsRune(string(content), 0) {
				entry.Text, entry.Encoding = string(content), "utf-8"
				previewed += len(content)
			}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
