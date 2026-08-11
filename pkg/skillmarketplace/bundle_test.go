package skillmarketplace

import (
	"archive/zip"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type zipTestEntry struct {
	name, content string
	mode          uint32
}

func testZIP(t *testing.T, entries ...zipTestEntry) []byte {
	t.Helper()
	var out bytes.Buffer
	writer := zip.NewWriter(&out)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		header.SetMode(os.FileMode(entry.mode))
		target, err := writer.CreateHeader(header)
		require.NoError(t, err)
		_, err = target.Write([]byte(entry.content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	return out.Bytes()
}

func TestValidateAndNormalizeBundleIsDeterministicAndPreservesSafeExecBit(t *testing.T) {
	manifest := "---\nname: demo-skill\ndescription: deterministic demo\n---\nbody\n"
	first := testZIP(t, zipTestEntry{name: "scripts/run.sh", content: "#!/bin/sh\n", mode: 0o755}, zipTestEntry{name: "SKILL.md", content: manifest, mode: 0o600})
	second := testZIP(t, zipTestEntry{name: "SKILL.md", content: manifest, mode: 0o644}, zipTestEntry{name: "scripts/run.sh", content: "#!/bin/sh\n", mode: 0o777})
	one, err := ValidateAndNormalizeBundle(first)
	require.NoError(t, err)
	two, err := ValidateAndNormalizeBundle(second)
	require.NoError(t, err)
	require.Equal(t, one.Digest, two.Digest)
	require.Equal(t, one.Bytes, two.Bytes)
	entries, err := InspectBundle(one.Bytes)
	require.NoError(t, err)
	require.Equal(t, "0644", entries[0].Mode)
	require.False(t, entries[0].Executable)
	require.Equal(t, "0755", entries[1].Mode)
	require.True(t, entries[1].Executable)
}

func TestValidateAndNormalizeBundleRejectsTraversalAndSymlink(t *testing.T) {
	manifest := "---\nname: demo-skill\ndescription: demo\n---\n"
	_, err := ValidateAndNormalizeBundle(testZIP(t, zipTestEntry{name: "../SKILL.md", content: manifest, mode: 0o644}))
	require.ErrorIs(t, err, ErrInvalidBundle)
	var out bytes.Buffer
	writer := zip.NewWriter(&out)
	header := &zip.FileHeader{Name: "SKILL.md"}
	header.SetMode(os.ModeSymlink | 0o777)
	target, err := writer.CreateHeader(header)
	require.NoError(t, err)
	_, _ = target.Write([]byte("target"))
	require.NoError(t, writer.Close())
	_, err = ValidateAndNormalizeBundle(out.Bytes())
	require.ErrorIs(t, err, ErrInvalidBundle)
}

func TestManifestValidationUsesAgentSkillsNameAndDescriptionLimits(t *testing.T) {
	_, err := ValidateAndNormalizeBundle(testZIP(t, zipTestEntry{name: "SKILL.md", content: "---\nname: bad_name\ndescription: demo\n---\n", mode: 0o644}))
	require.ErrorIs(t, err, ErrInvalidBundle)
	_, err = ValidateAndNormalizeBundle(testZIP(t, zipTestEntry{name: "SKILL.md", content: "---\nname: good-name\ndescription: " + strings.Repeat("x", 1025) + "\n---\n", mode: 0o644}))
	require.ErrorIs(t, err, ErrInvalidBundle)
}
