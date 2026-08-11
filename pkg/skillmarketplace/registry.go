package skillmarketplace

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	skilltypes "github.com/pole-io/pole-server/apis/pkg/types/skillmarketplace"
	"github.com/pole-io/pole-server/apis/pkg/utils"
)

type RegistryPage struct {
	Items      []skilltypes.RemoteRelease `json:"items"`
	NextCursor string                     `json:"nextCursor,omitempty"`
}

type RegistryAdapter interface {
	List(context.Context, *skilltypes.RegistrySource) (*RegistryPage, error)
}

type HTTPIndexAdapter struct {
	Client *http.Client
}

func (a *HTTPIndexAdapter) List(ctx context.Context, source *skilltypes.RegistrySource) (*RegistryPage, error) {
	endpoint, err := registryURL(source.URL, source.Cursor)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := sameOriginClient(a.client(), source.URL).Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("registry index returned HTTP %d", response.StatusCode)
	}
	var page RegistryPage
	decoder := json.NewDecoder(io.LimitReader(response.Body, 4<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&page); err != nil {
		return nil, fmt.Errorf("decode registry index: %w", err)
	}
	for i := range page.Items {
		item := &page.Items[i]
		if !slugPattern.MatchString(item.Publisher) || !slugPattern.MatchString(item.Name) ||
			ValidateVersion(item.Version) != nil || len(item.Digest) != 64 || item.BundleURL == "" {
			return nil, fmt.Errorf("registry index contains invalid release at index %d", i)
		}
		if err := validateBundleURL(source.URL, item.BundleURL, false); err != nil {
			return nil, fmt.Errorf("registry item %d: %w", i, err)
		}
	}
	return &page, nil
}

func (a *HTTPIndexAdapter) client() *http.Client {
	if a.Client != nil {
		return a.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// PoleRegistryAdapter uses the same stable JSON index envelope, but appends
// Pole's canonical marketplace index path when a server root is configured.
type PoleRegistryAdapter struct{ HTTPIndexAdapter }

func (a *PoleRegistryAdapter) List(ctx context.Context, source *skilltypes.RegistrySource) (*RegistryPage, error) {
	copySource := *source
	parsed, err := url.Parse(copySource.URL)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(parsed.Path, "/api/skill-marketplace/") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/") + "/api/skill-marketplace/v1/index"
		copySource.URL = parsed.String()
	}
	return a.HTTPIndexAdapter.List(ctx, &copySource)
}

type GitRegistryAdapter struct{ Client *http.Client }

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Body    string        `json:"body"`
	Assets  []githubAsset `json:"assets"`
}
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

type githubGitObject struct {
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
	} `json:"object"`
}

type GitSkillCandidate struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Digest      string `json:"digest"`
	Entries     int    `json:"entries"`
	Size        int64  `json:"size"`
	Error       string `json:"error,omitempty"`
	Bundle      []byte `json:"-"`
}

type GitSkillDiscovery struct {
	RepositoryURL string              `json:"repository_url"`
	Reference     string              `json:"reference"`
	Tag           string              `json:"tag"`
	CommitSHA     string              `json:"commit_sha"`
	RootPath      string              `json:"root_path"`
	Version       string              `json:"version"`
	Items         []GitSkillCandidate `json:"items"`
}

type sourceArchiveEntry struct {
	data       []byte
	executable bool
}

const (
	maxGitArchiveSize             = 64 << 20
	maxGitArchiveUncompressedSize = 256 << 20
	maxGitArchiveEntries          = 10000
	maxGitImportSkills            = 256
)

func (a *GitRegistryAdapter) List(ctx context.Context, source *skilltypes.RegistrySource) (*RegistryPage, error) {
	owner, repo, err := parseGitHubRepository(source.URL)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=100", url.PathEscape(owner), url.PathEscape(repo))
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	request.Header.Set("Accept", "application/vnd.github+json")
	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub releases returned HTTP %d", response.StatusCode)
	}
	var releases []githubRelease
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&releases); err != nil {
		return nil, err
	}
	page := &RegistryPage{Items: make([]skilltypes.RemoteRelease, 0)}
	for _, release := range releases {
		version := strings.TrimPrefix(release.TagName, "v")
		if ValidateVersion(version) != nil {
			continue
		}
		for _, asset := range release.Assets {
			if !strings.HasSuffix(strings.ToLower(asset.Name), ".zip") {
				continue
			}
			digest := strings.TrimPrefix(asset.Digest, "sha256:")
			if len(digest) != 64 {
				continue
			}
			name := strings.TrimSuffix(strings.ToLower(asset.Name), ".zip")
			if !skillNamePattern.MatchString(name) {
				continue
			}
			assetURL, parseErr := url.Parse(asset.BrowserDownloadURL)
			if parseErr != nil || assetURL.User != nil || assetURL.Scheme != "https" || !strings.EqualFold(assetURL.Hostname(), "github.com") {
				continue
			}
			page.Items = append(page.Items, skilltypes.RemoteRelease{Publisher: strings.ToLower(owner), Name: name, Description: release.Body, Version: version, Digest: digest, BundleURL: asset.BrowserDownloadURL})
			break
		}
	}
	return page, nil
}

func registryURL(raw, cursor string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", fmt.Errorf("registry URL must be absolute http(s)")
	}
	query := parsed.Query()
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

type RegistrySyncer struct {
	service  *Service
	adapters map[skilltypes.RegistryType]RegistryAdapter
	client   *http.Client
	mu       sync.Mutex
}

func NewRegistrySyncer(service *Service, client *http.Client) *RegistrySyncer {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	httpAdapter := HTTPIndexAdapter{Client: client}
	return &RegistrySyncer{
		service: service, client: client,
		adapters: map[skilltypes.RegistryType]RegistryAdapter{
			skilltypes.RegistryHTTP: &httpAdapter,
			skilltypes.RegistryPole: &PoleRegistryAdapter{HTTPIndexAdapter: httpAdapter},
			skilltypes.RegistryGit:  &GitRegistryAdapter{},
		},
	}
}

func (s *RegistrySyncer) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	_ = s.service.store.RecoverInterruptedRegistrySyncs(ctx, time.Now().Add(-10*time.Minute))
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		cleanup := time.NewTicker(24 * time.Hour)
		defer cleanup.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = s.SyncDue(ctx)
			case <-cleanup.C:
				_, _ = s.service.store.DeleteUnreferencedBundles(ctx, time.Now().Add(-7*24*time.Hour))
			}
		}
	}()
}

func (s *RegistrySyncer) SyncDue(ctx context.Context) (uint32, error) {
	sources, err := s.service.store.ListDueRegistrySources(ctx, time.Now(), 20)
	if err != nil {
		return 0, err
	}
	var synced uint32
	for _, source := range sources {
		if err := s.Sync(ctx, source.ID); err == nil {
			synced++
		}
	}
	return synced, nil
}

func (s *RegistrySyncer) Sync(ctx context.Context, sourceID string) error {
	// The process-local mutex avoids duplicate manual/ticker syncs; the DB state
	// transition is the cross-process recovery and exclusion boundary.
	s.mu.Lock()
	defer s.mu.Unlock()
	source, err := s.service.store.GetRegistrySource(ctx, sourceID)
	if err != nil {
		return err
	}
	started := time.Now().UTC()
	if err := s.service.store.MarkRegistrySyncStarted(ctx, source.ID, started); err != nil {
		return err
	}
	finish := func(cursor string, syncErr error) error {
		finished := time.Now().UTC()
		interval := time.Duration(source.SyncIntervalSeconds) * time.Second
		if interval <= 0 {
			interval = time.Hour
		}
		message := ""
		if syncErr != nil {
			message = syncErr.Error()
		}
		markErr := s.service.store.MarkRegistrySyncFinished(ctx, source.ID, cursor, message, finished, finished.Add(interval))
		if syncErr != nil {
			return syncErr
		}
		return markErr
	}
	adapter := s.adapters[source.Type]
	if adapter == nil {
		return finish(source.Cursor, fmt.Errorf("unsupported registry type %q", source.Type))
	}
	page, err := adapter.List(ctx, source)
	if err != nil {
		return finish(source.Cursor, err)
	}
	for _, item := range page.Items {
		if err := s.importRemoteRelease(ctx, source, item); err != nil {
			return finish(source.Cursor, err)
		}
	}
	return finish(page.NextCursor, nil)
}

func (s *RegistrySyncer) importRemoteRelease(ctx context.Context, source *skilltypes.RegistrySource,
	item skilltypes.RemoteRelease) error {
	if source.Type == skilltypes.RegistryGit {
		assetURL, err := url.Parse(item.BundleURL)
		if err != nil || assetURL.User != nil || assetURL.Scheme != "https" || !strings.EqualFold(assetURL.Hostname(), "github.com") {
			return fmt.Errorf("unsafe GitHub release asset URL")
		}
	} else {
		if err := validateBundleURL(source.URL, item.BundleURL, false); err != nil {
			return err
		}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, item.BundleURL, nil)
	if err != nil {
		return err
	}
	client := s.client
	if source.Type != skilltypes.RegistryGit {
		client = sameOriginClient(s.client, source.URL)
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("download %s/%s@%s returned HTTP %d", item.Publisher, item.Name, item.Version, response.StatusCode)
	}
	bundle, err := io.ReadAll(io.LimitReader(response.Body, MaxCompressedBundleSize+1))
	if err != nil {
		return err
	}
	if len(bundle) > MaxCompressedBundleSize {
		return fmt.Errorf("remote bundle exceeds compressed size limit")
	}
	publisher, err := s.service.store.GetSkillPublisher(ctx, item.Publisher)
	if errors.Is(err, ErrNotFound) {
		err = s.service.store.UpsertSkillPublisher(ctx, &skilltypes.Publisher{
			ID: utils.NewUUID(), Name: item.Publisher, DisplayName: item.Publisher,
			OwnerID: "registry:" + source.ID, Trusted: source.TrustLevel == skilltypes.TrustTrusted,
		})
	} else if err == nil && publisher.OwnerID != "registry:"+source.ID {
		return fmt.Errorf("registry publisher %q conflicts with an existing local or other-source publisher", item.Publisher)
	}
	if err != nil {
		return err
	}
	_, err = s.service.Publish(ctx, PublishRequest{
		Publisher: item.Publisher, Name: item.Name, Description: item.Description,
		Version: item.Version, Visibility: skilltypes.VisibilityPublic, Bundle: bundle,
		Actor: Principal{ID: "registry:" + source.ID, Admin: true}, SourceID: source.ID,
		SourceDigest: item.Digest, SourceTrust: source.TrustLevel,
	})
	if errors.Is(err, ErrVersionExists) {
		return nil
	}
	return err
}

func parseGitHubRepository(raw string) (string, string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), "github.com") || parsed.User != nil {
		return "", "", fmt.Errorf("Git registry currently supports only https://github.com/{owner}/{repo}")
	}
	parts := strings.Split(strings.Trim(strings.TrimSuffix(parsed.Path, ".git"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid GitHub repository URL")
	}
	return parts[0], parts[1], nil
}

func validateBundleURL(sourceURL, bundleURL string, allowCrossOrigin bool) error {
	source, err := url.Parse(sourceURL)
	if err != nil {
		return err
	}
	bundle, err := url.Parse(bundleURL)
	if err != nil || (bundle.Scheme != "http" && bundle.Scheme != "https") || bundle.Host == "" || bundle.User != nil {
		return fmt.Errorf("bundleUrl must be absolute http(s) without userinfo")
	}
	if !allowCrossOrigin && !strings.EqualFold(source.Host, bundle.Host) {
		return fmt.Errorf("cross-origin bundleUrl is not allowed")
	}
	return nil
}

func sameOriginClient(base *http.Client, origin string) *http.Client {
	copyClient := *base
	previous := copyClient.CheckRedirect
	copyClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if err := validateBundleURL(origin, req.URL.String(), false); err != nil {
			return err
		}
		if previous != nil {
			return previous(req, via)
		}
		return nil
	}
	return &copyClient
}

func (s *RegistrySyncer) DiscoverGitHubSkills(ctx context.Context, repositoryURL, reference, rootPath, versionOverride string) (*GitSkillDiscovery, error) {
	owner, repo, err := parseGitHubRepository(repositoryURL)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(reference) == "" {
		return nil, fmt.Errorf("Git reference must be an explicit tag or release tag")
	}
	cleanRoot, err := cleanGitRootPath(rootPath)
	if err != nil {
		return nil, err
	}
	tag, commitSHA, err := s.resolveGitHubTag(ctx, owner, repo, reference)
	if err != nil {
		return nil, err
	}
	version := strings.TrimSpace(versionOverride)
	if version == "" {
		version = strings.TrimPrefix(tag, "v")
	}
	if err := ValidateVersion(version); err != nil {
		return nil, fmt.Errorf("Git tag %q is not SemVer; provide an explicit version override: %w", tag, err)
	}
	archive, err := s.downloadGitHubArchive(ctx, owner, repo, commitSHA)
	if err != nil {
		return nil, err
	}
	files, err := readGitHubSourceArchive(archive)
	if err != nil {
		return nil, err
	}
	items, err := discoverGitSkills(files, cleanRoot)
	if err != nil {
		return nil, err
	}
	return &GitSkillDiscovery{
		RepositoryURL: repositoryURL,
		Reference:     reference,
		Tag:           tag,
		CommitSHA:     commitSHA,
		RootPath:      cleanRoot,
		Version:       version,
		Items:         items,
	}, nil
}

func (s *RegistrySyncer) resolveGitHubTag(ctx context.Context, owner, repo, reference string) (string, string, error) {
	tag := reference
	releaseEndpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(reference))
	response, err := s.githubJSON(ctx, releaseEndpoint)
	if err != nil {
		return "", "", err
	}
	if response.StatusCode == http.StatusOK {
		var release githubRelease
		if err := decodeGitHubJSON(response, &release); err != nil {
			return "", "", err
		}
		if release.TagName == "" {
			return "", "", fmt.Errorf("GitHub release %q has no tag", reference)
		}
		tag = release.TagName
	} else {
		status := response.StatusCode
		response.Body.Close()
		if status != http.StatusNotFound {
			return "", "", fmt.Errorf("GitHub release %q returned HTTP %d", reference, status)
		}
	}
	refEndpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/tags/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(tag))
	response, err = s.githubJSON(ctx, refEndpoint)
	if err != nil {
		return "", "", err
	}
	if response.StatusCode != http.StatusOK {
		status := response.StatusCode
		response.Body.Close()
		return "", "", fmt.Errorf("GitHub tag %q returned HTTP %d; branches are not accepted", tag, status)
	}
	var object githubGitObject
	if err := decodeGitHubJSON(response, &object); err != nil {
		return "", "", err
	}
	for depth := 0; object.Object.Type == "tag" && depth < 5; depth++ {
		endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/tags/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(object.Object.SHA))
		response, err = s.githubJSON(ctx, endpoint)
		if err != nil {
			return "", "", err
		}
		if response.StatusCode != http.StatusOK {
			status := response.StatusCode
			response.Body.Close()
			return "", "", fmt.Errorf("resolve annotated GitHub tag %q returned HTTP %d", tag, status)
		}
		if err := decodeGitHubJSON(response, &object); err != nil {
			return "", "", err
		}
	}
	if object.Object.Type != "commit" || !isGitCommitSHA(object.Object.SHA) {
		return "", "", fmt.Errorf("GitHub tag %q does not resolve to a commit", tag)
	}
	return tag, strings.ToLower(object.Object.SHA), nil
}

func (s *RegistrySyncer) githubJSON(ctx context.Context, endpoint string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	return s.client.Do(request)
}

func decodeGitHubJSON(response *http.Response, target any) error {
	defer response.Body.Close()
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(target); err != nil {
		return fmt.Errorf("decode GitHub response: %w", err)
	}
	return nil
}

func (s *RegistrySyncer) downloadGitHubArchive(ctx context.Context, owner, repo, commitSHA string) ([]byte, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(commitSHA))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err := trustedGitHubClient(s.client).Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub source archive returned HTTP %d", response.StatusCode)
	}
	archive, err := io.ReadAll(io.LimitReader(response.Body, maxGitArchiveSize+1))
	if err != nil {
		return nil, fmt.Errorf("read GitHub source archive: %w", err)
	}
	if len(archive) > maxGitArchiveSize {
		return nil, fmt.Errorf("GitHub source archive exceeds %d bytes", maxGitArchiveSize)
	}
	return archive, nil
}

func trustedGitHubClient(base *http.Client) *http.Client {
	copyClient := *base
	previous := copyClient.CheckRedirect
	copyClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" || req.URL.User != nil || !trustedGitHubHost(req.URL.Hostname()) {
			return fmt.Errorf("unsafe GitHub archive redirect")
		}
		if previous != nil {
			return previous(req, via)
		}
		return nil
	}
	return &copyClient
}

func trustedGitHubHost(host string) bool {
	return strings.EqualFold(host, "api.github.com") || strings.EqualFold(host, "github.com") || strings.EqualFold(host, "codeload.github.com")
}

func isGitCommitSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

func cleanGitRootPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "." {
		return "", nil
	}
	if strings.Contains(value, "\\") || strings.ContainsRune(value, 0) || strings.HasPrefix(value, "/") {
		return "", fmt.Errorf("Git root path must be a safe relative path")
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || cleaned != strings.TrimSuffix(value, "/") {
		return "", fmt.Errorf("Git root path must be canonical and cannot traverse")
	}
	return cleaned, nil
}

func readGitHubSourceArchive(data []byte) (map[string]sourceArchiveEntry, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open GitHub source archive: %w", err)
	}
	if len(zr.File) == 0 || len(zr.File) > maxGitArchiveEntries {
		return nil, fmt.Errorf("GitHub source archive entry count must be in 1..%d", maxGitArchiveEntries)
	}
	files := make(map[string]sourceArchiveEntry)
	var prefix string
	var total int64
	for _, file := range zr.File {
		name := strings.TrimSuffix(file.Name, "/")
		if name == "" || strings.Contains(name, "\\") || strings.ContainsRune(name, 0) || strings.HasPrefix(name, "/") || path.Clean(name) != name || strings.HasPrefix(name, "../") {
			return nil, fmt.Errorf("GitHub source archive contains unsafe path %q", file.Name)
		}
		parts := strings.SplitN(name, "/", 2)
		if prefix == "" {
			prefix = parts[0]
		} else if prefix != parts[0] {
			return nil, fmt.Errorf("GitHub source archive must have one repository root")
		}
		if len(parts) == 1 || file.FileInfo().IsDir() {
			continue
		}
		mode := file.Mode()
		if mode&mode.Type() != 0 || !mode.IsRegular() {
			return nil, fmt.Errorf("GitHub source archive contains non-regular entry %q", file.Name)
		}
		if total+int64(file.UncompressedSize64) > maxGitArchiveUncompressedSize {
			return nil, fmt.Errorf("GitHub source archive exceeds %d uncompressed bytes", maxGitArchiveUncompressedSize)
		}
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		remaining := maxGitArchiveUncompressedSize - total
		content, readErr := io.ReadAll(io.LimitReader(rc, remaining+1))
		closeErr := rc.Close()
		if readErr != nil || closeErr != nil || int64(len(content)) > remaining {
			return nil, fmt.Errorf("read GitHub source archive entry %q", file.Name)
		}
		total += int64(len(content))
		relative := parts[1]
		if _, exists := files[relative]; exists {
			return nil, fmt.Errorf("GitHub source archive contains duplicate path %q", relative)
		}
		files[relative] = sourceArchiveEntry{data: content, executable: mode.Perm()&0o111 != 0}
	}
	return files, nil
}

func discoverGitSkills(files map[string]sourceArchiveEntry, rootPath string) ([]GitSkillCandidate, error) {
	manifestPath := "SKILL.md"
	if rootPath != "" {
		manifestPath = rootPath + "/SKILL.md"
	}
	candidatePaths := make([]string, 0)
	if _, ok := files[manifestPath]; ok {
		candidatePaths = append(candidatePaths, rootPath)
	} else {
		prefix := rootPath
		if prefix != "" {
			prefix += "/"
		}
		seen := map[string]struct{}{}
		for name := range files {
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			relative := strings.TrimPrefix(name, prefix)
			parts := strings.Split(relative, "/")
			if len(parts) == 2 && parts[1] == "SKILL.md" {
				candidate := prefix + parts[0]
				seen[candidate] = struct{}{}
			}
		}
		for candidate := range seen {
			candidatePaths = append(candidatePaths, candidate)
		}
		sort.Strings(candidatePaths)
	}
	if len(candidatePaths) == 0 {
		location := rootPath
		if location == "" {
			location = "repository root"
		}
		return nil, fmt.Errorf("no Skill found at %s or its direct children", location)
	}
	if len(candidatePaths) > maxGitImportSkills {
		return nil, fmt.Errorf("Git import found %d Skills; maximum is %d", len(candidatePaths), maxGitImportSkills)
	}
	items := make([]GitSkillCandidate, 0, len(candidatePaths))
	for _, candidatePath := range candidatePaths {
		bundle, err := bundleGitSkill(files, candidatePath)
		if err != nil {
			items = append(items, GitSkillCandidate{Path: candidatePath, Error: err.Error()})
			continue
		}
		items = append(items, GitSkillCandidate{
			Path: candidatePath, Name: bundle.Name, Description: bundle.Description,
			Digest: bundle.Digest, Entries: bundle.Entries, Size: bundle.Size, Bundle: bundle.Bytes,
		})
	}
	pathsByName := make(map[string][]int)
	for index := range items {
		if items[index].Name != "" {
			pathsByName[items[index].Name] = append(pathsByName[items[index].Name], index)
		}
	}
	for name, indexes := range pathsByName {
		if len(indexes) < 2 {
			continue
		}
		for _, index := range indexes {
			items[index].Error = fmt.Sprintf("duplicate Skill name %q in the same Git snapshot", name)
			items[index].Bundle = nil
		}
	}
	return items, nil
}

func bundleGitSkill(files map[string]sourceArchiveEntry, candidatePath string) (*BundleInfo, error) {
	prefix := candidatePath
	if prefix != "" {
		prefix += "/"
	}
	names := make([]string, 0)
	for name := range files {
		if strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, name := range names {
		entry := files[name]
		relative := strings.TrimPrefix(name, prefix)
		header := &zip.FileHeader{Name: relative, Method: zip.Deflate}
		if entry.executable {
			header.SetMode(0o755)
		} else {
			header.SetMode(0o644)
		}
		fileWriter, err := writer.CreateHeader(header)
		if err != nil {
			return nil, err
		}
		if _, err := fileWriter.Write(entry.data); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return ValidateAndNormalizeBundle(buffer.Bytes())
}
