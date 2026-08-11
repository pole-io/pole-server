package skillmarketplace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (s *RegistrySyncer) ImportGitHubRelease(ctx context.Context, repositoryURL, reference, skillName string) ([]byte, string, error) {
	owner, repo, err := parseGitHubRepository(repositoryURL)
	if err != nil {
		return nil, "", err
	}
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(reference))
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err := s.client.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("GitHub release %q returned HTTP %d", reference, response.StatusCode)
	}
	var release githubRelease
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&release); err != nil {
		return nil, "", err
	}
	var selected *githubAsset
	for i := range release.Assets {
		asset := &release.Assets[i]
		base := strings.TrimSuffix(strings.ToLower(asset.Name), ".zip")
		if strings.HasSuffix(strings.ToLower(asset.Name), ".zip") && (base == skillName || selected == nil) {
			selected = asset
			if base == skillName {
				break
			}
		}
	}
	if selected == nil {
		return nil, "", fmt.Errorf("GitHub release has no ZIP asset")
	}
	assetURL, err := url.Parse(selected.BrowserDownloadURL)
	if err != nil || assetURL.User != nil || assetURL.Scheme != "https" || !strings.EqualFold(assetURL.Hostname(), "github.com") {
		return nil, "", fmt.Errorf("unsafe GitHub release asset URL")
	}
	download, _ := http.NewRequestWithContext(ctx, http.MethodGet, selected.BrowserDownloadURL, nil)
	result, err := s.client.Do(download)
	if err != nil {
		return nil, "", err
	}
	defer result.Body.Close()
	if result.StatusCode < 200 || result.StatusCode >= 300 {
		return nil, "", fmt.Errorf("GitHub asset returned HTTP %d", result.StatusCode)
	}
	bundle, err := io.ReadAll(io.LimitReader(result.Body, MaxCompressedBundleSize+1))
	if err != nil || len(bundle) > MaxCompressedBundleSize {
		return nil, "", fmt.Errorf("read GitHub asset: %w", err)
	}
	version := strings.TrimPrefix(release.TagName, "v")
	if err := ValidateVersion(version); err != nil {
		return nil, "", err
	}
	return bundle, version, nil
}
