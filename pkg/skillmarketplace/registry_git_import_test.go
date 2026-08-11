package skillmarketplace

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const gitImportCommit = "0123456789abcdef0123456789abcdef01234567"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }

func TestDiscoverGitHubSkillsFindsDirectChildrenAtImmutableTag(t *testing.T) {
	archive := gitSourceArchive(t, map[string]string{
		"skills/alpha/SKILL.md":        "---\nname: alpha\ndescription: Alpha skill\n---\n",
		"skills/alpha/run.sh":          "#!/bin/sh\n",
		"skills/bravo/SKILL.md":        "---\nname: bravo\ndescription: Bravo skill\n---\n",
		"skills/bravo/README.md":       "bravo\n",
		"skills/broken/SKILL.md":       "not valid frontmatter\n",
		"skills/nested/child/SKILL.md": "---\nname: nested-child\ndescription: must not be discovered recursively\n---\n",
	})
	syncer := NewRegistrySyncer(nil, githubImportClient(t, archive, "v1.2.0"))

	discovery, err := syncer.DiscoverGitHubSkills(context.Background(), "https://github.com/pole-io/skills", "v1.2.0", "skills", "")
	require.NoError(t, err)
	require.Equal(t, gitImportCommit, discovery.CommitSHA)
	require.Equal(t, "1.2.0", discovery.Version)
	require.Len(t, discovery.Items, 3)
	require.Equal(t, []string{"alpha", "bravo"}, []string{discovery.Items[0].Name, discovery.Items[1].Name})
	require.Equal(t, []string{"skills/alpha", "skills/bravo"}, []string{discovery.Items[0].Path, discovery.Items[1].Path})
	require.Equal(t, "skills/broken", discovery.Items[2].Path)
	require.Contains(t, discovery.Items[2].Error, "SKILL.md must start with YAML frontmatter")
	for _, item := range discovery.Items[:2] {
		normalized, normalizeErr := ValidateAndNormalizeBundle(item.Bundle)
		require.NoError(t, normalizeErr)
		require.Equal(t, item.Name, normalized.Name)
		require.Equal(t, item.Digest, normalized.Digest)
	}
}

func TestDiscoverGitHubSkillsRejectsTraversingRootPath(t *testing.T) {
	syncer := NewRegistrySyncer(nil, githubImportClient(t, nil, "v1.0.0"))

	_, err := syncer.DiscoverGitHubSkills(context.Background(), "https://github.com/pole-io/skills", "v1.0.0", "../skills", "")
	require.ErrorContains(t, err, "cannot traverse")
}

func TestDiscoverGitHubSkillsTreatsRootManifestAsSingleSkill(t *testing.T) {
	archive := gitSourceArchive(t, map[string]string{
		"SKILL.md":                "---\nname: root-skill\ndescription: Root skill\n---\n",
		"scripts/run.sh":          "#!/bin/sh\n",
		"skills/ignored/SKILL.md": "---\nname: ignored\ndescription: nested content\n---\n",
	})
	syncer := NewRegistrySyncer(nil, githubImportClient(t, archive, "v2.0.0"))

	discovery, err := syncer.DiscoverGitHubSkills(context.Background(), "https://github.com/pole-io/root-skill", "v2.0.0", "", "")
	require.NoError(t, err)
	require.Len(t, discovery.Items, 1)
	require.Equal(t, "root-skill", discovery.Items[0].Name)
	require.Empty(t, discovery.Items[0].Path)
}

func TestDiscoverGitHubSkillsRequiresVersionOverrideForNonSemVerTag(t *testing.T) {
	archive := gitSourceArchive(t, map[string]string{
		"SKILL.md": "---\nname: root-skill\ndescription: Root skill\n---\n",
	})
	syncer := NewRegistrySyncer(nil, githubImportClient(t, archive, "stable"))

	_, err := syncer.DiscoverGitHubSkills(context.Background(), "https://github.com/pole-io/root-skill", "stable", "", "")
	require.ErrorContains(t, err, "provide an explicit version override")
	discovery, err := syncer.DiscoverGitHubSkills(context.Background(), "https://github.com/pole-io/root-skill", "stable", "", "3.1.4")
	require.NoError(t, err)
	require.Equal(t, "3.1.4", discovery.Version)
}

func TestDiscoverGitHubSkillsRejectsBranchWhenNoTagExists(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return response(http.StatusNotFound, `{}`), nil
	})}
	syncer := NewRegistrySyncer(nil, client)

	_, err := syncer.DiscoverGitHubSkills(context.Background(), "https://github.com/pole-io/root-skill", "main", "", "1.0.0")
	require.ErrorContains(t, err, "branches are not accepted")
}

func TestDiscoverGitHubSkillsAgainstPublicTaggedRepository(t *testing.T) {
	if os.Getenv("SKILL_GITHUB_INTEGRATION") == "" {
		t.Skip("set SKILL_GITHUB_INTEGRATION=1 to call the public GitHub API")
	}
	syncer := NewRegistrySyncer(nil, nil)

	discovery, err := syncer.DiscoverGitHubSkills(context.Background(), "https://github.com/psenger/ai-agent-skills", "v2.2.1", "skills", "")
	require.NoError(t, err)
	require.Equal(t, "96716a643138665b4e335f836fb661f126944644", discovery.CommitSHA)
	require.Equal(t, "2.2.1", discovery.Version)
	require.Greater(t, len(discovery.Items), 5)
	valid := 0
	for _, item := range discovery.Items {
		if item.Error == "" {
			valid++
		}
	}
	require.Greater(t, valid, 5)
}

func githubImportClient(t *testing.T, archive []byte, tag string) *http.Client {
	t.Helper()
	return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch {
		case strings.Contains(request.URL.Path, "/releases/tags/"):
			return response(http.StatusOK, fmt.Sprintf(`{"tag_name":%q}`, tag)), nil
		case strings.Contains(request.URL.Path, "/git/ref/tags/"):
			return response(http.StatusOK, fmt.Sprintf(`{"object":{"type":"commit","sha":%q}}`, gitImportCommit)), nil
		case strings.Contains(request.URL.Path, "/zipball/"):
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(archive)), Request: request}, nil
		default:
			t.Fatalf("unexpected GitHub request %s", request.URL)
			return nil, nil
		}
	})}
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func gitSourceArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		header := &zip.FileHeader{Name: "pole-io-skills-" + gitImportCommit[:7] + "/" + name, Method: zip.Deflate}
		header.SetMode(0o644)
		entry, err := writer.CreateHeader(header)
		require.NoError(t, err)
		_, err = entry.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	return buffer.Bytes()
}
