package skillmarketplace

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	MaxCompressedBundleSize   = 16 << 20
	MaxUncompressedBundleSize = 64 << 20
	MaxBundleEntries          = 2048
)

var ErrInvalidBundle = errors.New("invalid Agent Skills bundle")

var skillNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type BundleInfo struct {
	Bytes       []byte
	Digest      string
	Name        string
	Description string
	Entries     int
	Size        int64
}

type bundleEntry struct {
	name       string
	data       []byte
	executable bool
}

// ValidateAndNormalizeBundle validates an Agent Skills ZIP and returns a
// deterministic ZIP. The digest always describes the stored/downloaded bytes.
func ValidateAndNormalizeBundle(data []byte) (*BundleInfo, error) {
	if len(data) == 0 || len(data) > MaxCompressedBundleSize {
		return nil, fmt.Errorf("%w: compressed size must be in 1..%d bytes", ErrInvalidBundle, MaxCompressedBundleSize)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidBundle, err)
	}
	if len(zr.File) == 0 || len(zr.File) > MaxBundleEntries {
		return nil, fmt.Errorf("%w: entry count must be in 1..%d", ErrInvalidBundle, MaxBundleEntries)
	}
	entries := make([]bundleEntry, 0, len(zr.File))
	seen := make(map[string]struct{}, len(zr.File))
	var total int64
	var manifest []byte
	for _, file := range zr.File {
		name, err := safeBundlePath(file.Name)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("%w: duplicate entry %q", ErrInvalidBundle, name)
		}
		seen[name] = struct{}{}
		mode := file.Mode()
		if mode.IsDir() || strings.HasSuffix(name, "/") {
			continue
		}
		if mode&mode.Type() != 0 || !mode.IsRegular() {
			return nil, fmt.Errorf("%w: non-regular entry %q", ErrInvalidBundle, name)
		}
		if file.Flags&1 != 0 {
			return nil, fmt.Errorf("%w: encrypted entry %q", ErrInvalidBundle, name)
		}
		if file.Method != zip.Store && file.Method != zip.Deflate {
			return nil, fmt.Errorf("%w: unsupported compression method for %q", ErrInvalidBundle, name)
		}
		if file.UncompressedSize64 > MaxUncompressedBundleSize ||
			total+int64(file.UncompressedSize64) > MaxUncompressedBundleSize {
			return nil, fmt.Errorf("%w: uncompressed size exceeds %d bytes", ErrInvalidBundle, MaxUncompressedBundleSize)
		}
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("%w: open %q: %v", ErrInvalidBundle, name, err)
		}
		remaining := MaxUncompressedBundleSize - total
		content, readErr := io.ReadAll(io.LimitReader(rc, remaining+1))
		closeErr := rc.Close()
		if readErr != nil || closeErr != nil {
			return nil, fmt.Errorf("%w: read %q", ErrInvalidBundle, name)
		}
		if int64(len(content)) > remaining {
			return nil, fmt.Errorf("%w: uncompressed size exceeds %d bytes", ErrInvalidBundle, MaxUncompressedBundleSize)
		}
		total += int64(len(content))
		entries = append(entries, bundleEntry{name: name, data: content, executable: mode.Perm()&0o111 != 0})
		if name == "SKILL.md" {
			manifest = content
		}
	}
	name, description, err := validateSkillManifest(manifest)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
	var normalized bytes.Buffer
	zw := zip.NewWriter(&normalized)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if entry.executable {
			header.SetMode(0o755)
		} else {
			header.SetMode(0o644)
		}
		header.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return nil, fmt.Errorf("%w: normalize %q: %v", ErrInvalidBundle, entry.name, err)
		}
		if _, err := writer.Write(entry.data); err != nil {
			return nil, fmt.Errorf("%w: normalize %q: %v", ErrInvalidBundle, entry.name, err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("%w: normalize: %v", ErrInvalidBundle, err)
	}
	if normalized.Len() > MaxCompressedBundleSize {
		return nil, fmt.Errorf("%w: normalized size exceeds %d bytes", ErrInvalidBundle, MaxCompressedBundleSize)
	}
	sum := sha256.Sum256(normalized.Bytes())
	return &BundleInfo{
		Bytes: normalized.Bytes(), Digest: hex.EncodeToString(sum[:]), Name: name,
		Description: description, Entries: len(entries), Size: total,
	}, nil
}

func safeBundlePath(name string) (string, error) {
	if name == "" || strings.ContainsRune(name, 0) || strings.Contains(name, "\\") {
		return "", fmt.Errorf("%w: unsafe path %q", ErrInvalidBundle, name)
	}
	if strings.HasPrefix(name, "/") || path.IsAbs(name) {
		return "", fmt.Errorf("%w: absolute path %q", ErrInvalidBundle, name)
	}
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || cleaned != strings.TrimSuffix(name, "/") {
		return "", fmt.Errorf("%w: non-canonical or traversing path %q", ErrInvalidBundle, name)
	}
	return cleaned, nil
}

func validateSkillManifest(manifest []byte) (string, string, error) {
	if len(manifest) == 0 {
		return "", "", fmt.Errorf("%w: root SKILL.md is required", ErrInvalidBundle)
	}
	text := strings.ReplaceAll(string(manifest), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return "", "", fmt.Errorf("%w: SKILL.md must start with YAML frontmatter", ErrInvalidBundle)
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return "", "", fmt.Errorf("%w: SKILL.md frontmatter is not closed", ErrInvalidBundle)
	}
	frontmatter := text[4 : 4+end]
	var name, description string
	for _, line := range strings.Split(frontmatter, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		switch strings.TrimSpace(key) {
		case "name":
			name = value
		case "description":
			description = value
		}
	}
	if !skillNamePattern.MatchString(name) || len(name) > 64 {
		return "", "", fmt.Errorf("%w: SKILL.md name must be 1..64 lowercase letters, digits, and single hyphens", ErrInvalidBundle)
	}
	if description == "" || len(description) > 1024 {
		return "", "", fmt.Errorf("%w: SKILL.md frontmatter requires name and description", ErrInvalidBundle)
	}
	return name, description, nil
}
