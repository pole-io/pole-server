package skillmarketplace

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	skilltypes "github.com/pole-io/pole-server/apis/pkg/types/skillmarketplace"
	"github.com/pole-io/pole-server/apis/pkg/utils"
	"github.com/pole-io/pole-server/apis/store"
	market "github.com/pole-io/pole-server/pkg/skillmarketplace"
	httpcommon "github.com/pole-io/pole-server/plugin/apiserver/httpserver/utils"
)

const basePath = "/api/skill-marketplace"

type credentialChecker interface {
	CheckCredential(*authtypes.AcquireContext) error
}
type permissionChecker interface {
	CheckConsolePermission(*authtypes.AcquireContext) (bool, error)
}

type HTTPServer struct {
	service     *market.Service
	syncer      *market.RegistrySyncer
	credentials credentialChecker
	permissions permissionChecker
	cache       cacheapi.CacheManager
}

func NewServer(ctx context.Context, storage store.Store, credentials credentialChecker,
	permissions permissionChecker, cacheMgr cacheapi.CacheManager) (*HTTPServer, error) {
	marketStore, ok := storage.(store.SkillMarketplaceStore)
	if !ok {
		return nil, fmt.Errorf("configured store does not implement SkillMarketplaceStore")
	}
	service := market.NewService(marketStore)
	server := &HTTPServer{service: service, syncer: market.NewRegistrySyncer(service, nil),
		credentials: credentials, permissions: permissions, cache: cacheMgr}
	server.syncer.Run(ctx, time.Minute)
	return server, nil
}

func (h *HTTPServer) GetAccessServer() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(basePath).Consumes(restful.MIME_JSON, "multipart/form-data").Produces(restful.MIME_JSON, "application/zip")
	ws.Route(ws.GET("/v1/skills").To(h.searchSkills))
	ws.Route(ws.GET("/v1/skills/{publisher}/{name}").To(h.getSkill))
	ws.Route(ws.GET("/v1/skills/{publisher}/{name}/releases/{version}/bundle").To(h.downloadBundle))
	ws.Route(ws.GET("/v1/skills/{publisher}/{name}/releases/{version}/bundle-manifest").To(h.bundleManifest))
	ws.Route(ws.GET("/v1/index").To(h.registryIndex))
	ws.Route(ws.POST("/v1/publishers").To(h.upsertPublisher))
	ws.Route(ws.GET("/v1/publishers/{publisher}/members").To(h.listPublisherMembers))
	ws.Route(ws.PUT("/v1/publishers/{publisher}/members").To(h.replacePublisherMembers))
	ws.Route(ws.POST("/v1/publishers/{publisher}/keys/{version}/revoke").To(h.revokePublisherKey))
	ws.Route(ws.POST("/v1/releases").To(h.publishRelease))
	ws.Route(ws.POST("/v1/skills/releases/upload").To(h.publishRelease))
	ws.Route(ws.POST("/v1/skills/releases/import-git/discover").To(h.discoverGitRelease))
	ws.Route(ws.POST("/v1/skills/releases/import-git").To(h.importGitRelease))
	ws.Route(ws.POST("/v1/releases/{releaseId}/review").To(h.reviewRelease))
	ws.Route(ws.GET("/v1/reviews").To(h.listReviews))
	ws.Route(ws.PUT("/v1/reviews/{releaseId}").To(h.reviewRelease))
	ws.Route(ws.POST("/v1/skills/{publisher}/{name}/releases/{version}/lifecycle").To(h.setLifecycle))
	ws.Route(ws.GET("/v1/skills/{publisher}/{name}/grants").To(h.listGrants))
	ws.Route(ws.PUT("/v1/skills/{publisher}/{name}/grants").To(h.replaceGrants))
	ws.Route(ws.GET("/v1/registry-sources").To(h.listRegistrySources))
	ws.Route(ws.POST("/v1/registry-sources").To(h.createRegistrySource))
	ws.Route(ws.PUT("/v1/registry-sources/{id}").To(h.updateRegistrySource))
	ws.Route(ws.DELETE("/v1/registry-sources/{id}").To(h.deleteRegistrySource))
	ws.Route(ws.POST("/v1/registry-sources/{id}/sync").To(h.syncRegistrySource))
	return ws
}

func (h *HTTPServer) searchSkills(req *restful.Request, rsp *restful.Response) {
	viewer := h.optionalPrincipal(req)
	offset, _ := strconv.ParseUint(req.QueryParameter("offset"), 10, 32)
	limit, _ := strconv.ParseUint(req.QueryParameter("limit"), 10, 32)
	items, total, err := h.service.Search(req.Request.Context(), req.QueryParameter("query"), viewer, uint32(offset), uint32(limit))
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, map[string]any{"items": items, "total": total}, restful.MIME_JSON)
}

func (h *HTTPServer) getSkill(req *restful.Request, rsp *restful.Response) {
	skill, err := h.service.Get(req.Request.Context(), req.PathParameter("publisher"), req.PathParameter("name"), h.optionalPrincipal(req))
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, skill, restful.MIME_JSON)
}

func (h *HTTPServer) downloadBundle(req *restful.Request, rsp *restful.Response) {
	name, version := req.PathParameter("name"), req.PathParameter("version")
	bundle, release, err := h.service.Download(req.Request.Context(), req.PathParameter("publisher"), name, version, h.optionalPrincipal(req))
	if err != nil {
		writeError(rsp, err)
		return
	}
	rsp.AddHeader("X-Pole-Bundle-Sha256", release.Digest)
	rsp.AddHeader("Content-Type", "application/zip")
	rsp.AddHeader("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.zip"`, name, version))
	rsp.WriteHeader(http.StatusOK)
	_, _ = rsp.Write(bundle)
}

func (h *HTTPServer) bundleManifest(req *restful.Request, rsp *restful.Response) {
	publisher, name, version := req.PathParameter("publisher"), req.PathParameter("name"), req.PathParameter("version")
	bundle, release, err := h.service.Download(req.Request.Context(), publisher, name, version, h.optionalPrincipal(req))
	if err != nil {
		writeError(rsp, err)
		return
	}
	entries, err := market.InspectBundle(bundle)
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, map[string]any{"publisher": publisher, "name": name,
		"version": version, "digest": release.Digest, "entries": entries}, restful.MIME_JSON)
}

func (h *HTTPServer) registryIndex(req *restful.Request, rsp *restful.Response) {
	offset64, _ := strconv.ParseUint(req.QueryParameter("cursor"), 10, 32)
	offset := uint32(offset64)
	items, total, err := h.service.Search(req.Request.Context(), "", market.Principal{}, offset, 100)
	if err != nil {
		writeError(rsp, err)
		return
	}
	base := requestBaseURL(req.Request)
	remote := make([]skilltypes.RemoteRelease, 0)
	for _, skill := range items {
		detail, err := h.service.Get(req.Request.Context(), skill.Publisher, skill.Name, market.Principal{})
		if err != nil {
			continue
		}
		for _, release := range detail.Releases {
			if release.Status != skilltypes.ReleasePublished {
				continue
			}
			remote = append(remote,
				skilltypes.RemoteRelease{Publisher: skill.Publisher, Name: skill.Name, Description: skill.Description,
					Version: release.Version, Digest: release.Digest, BundleURL: fmt.Sprintf("%s%s/v1/skills/%s/%s/releases/%s/bundle",
						base, basePath, url.PathEscape(skill.Publisher), url.PathEscape(skill.Name), url.PathEscape(release.Version))})
		}
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, market.RegistryPage{Items: remote,
		NextCursor: nextIndexCursor(offset, uint32(len(items)), total)}, restful.MIME_JSON)
}

func nextIndexCursor(offset, count, total uint32) string {
	if count == 0 || offset+count >= total {
		return ""
	}
	return strconv.FormatUint(uint64(offset+count), 10)
}

type publisherRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	PublicKey   string `json:"publicKey"`
	Trusted     bool   `json:"trusted"`
}

func (h *HTTPServer) upsertPublisher(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, false)
	if err != nil {
		writeError(rsp, err)
		return
	}
	var body publisherRequest
	if err := req.ReadEntity(&body); err != nil {
		writeError(rsp, err)
		return
	}
	key, err := base64.StdEncoding.DecodeString(body.PublicKey)
	if err != nil {
		writeError(rsp, fmt.Errorf("publicKey must be base64: %w", err))
		return
	}
	publisher := &skilltypes.Publisher{Name: body.Name, DisplayName: body.DisplayName, PublicKey: key, Trusted: body.Trusted && actor.Admin}
	if err := h.service.UpsertPublisher(req.Request.Context(), publisher, actor); err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, publisher, restful.MIME_JSON)
}

func (h *HTTPServer) publishRelease(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, false)
	if err != nil {
		writeError(rsp, err)
		return
	}
	req.Request.Body = http.MaxBytesReader(rsp.ResponseWriter, req.Request.Body, market.MaxCompressedBundleSize+(1<<20))
	if err := req.Request.ParseMultipartForm(market.MaxCompressedBundleSize + (1 << 20)); err != nil {
		writeError(rsp, err)
		return
	}
	file, _, err := req.Request.FormFile("bundle")
	if err != nil {
		writeError(rsp, err)
		return
	}
	defer file.Close()
	bundle, err := io.ReadAll(io.LimitReader(file, market.MaxCompressedBundleSize+1))
	if err != nil {
		writeError(rsp, err)
		return
	}
	signature, signedAt, err := parseSignatureEnvelope(req)
	if err != nil {
		writeError(rsp, err)
		return
	}
	release, err := h.service.Publish(req.Request.Context(), market.PublishRequest{Publisher: req.Request.FormValue("publisher"),
		Name: req.Request.FormValue("name"), Description: req.Request.FormValue("description"), Version: req.Request.FormValue("version"),
		Visibility: skilltypes.Visibility(req.Request.FormValue("visibility")), Bundle: bundle, Signature: signature, SignedAt: signedAt, Actor: actor})
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusCreated, release, restful.MIME_JSON)
}

type detachedSignatureEnvelope struct {
	Algorithm string `json:"algorithm"`
	SignedAt  string `json:"signedAt"`
	Signature string `json:"signature"`
}

func parseSignatureEnvelope(req *restful.Request) ([]byte, time.Time, error) {
	file, _, fileErr := req.Request.FormFile("signature")
	if fileErr == nil {
		defer file.Close()
		payload, err := io.ReadAll(io.LimitReader(file, 16<<10))
		if err != nil {
			return nil, time.Time{}, err
		}
		var envelope detachedSignatureEnvelope
		if json.Unmarshal(payload, &envelope) == nil && envelope.Signature != "" {
			if envelope.Algorithm != "Ed25519" {
				return nil, time.Time{}, fmt.Errorf("signature algorithm must be Ed25519")
			}
			signature, err := base64.StdEncoding.DecodeString(envelope.Signature)
			if err != nil {
				return nil, time.Time{}, fmt.Errorf("invalid detached signature base64: %w", err)
			}
			signedAt, err := time.Parse(time.RFC3339, envelope.SignedAt)
			if err != nil {
				return nil, time.Time{}, fmt.Errorf("invalid signature signedAt: %w", err)
			}
			return signature, signedAt, nil
		}
		signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(payload)))
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("signature file must be JSON envelope or base64")
		}
		signedAt, err := time.Parse(time.RFC3339, req.Request.FormValue("signedAt"))
		return signature, signedAt, err
	}
	value := req.Request.FormValue("signature")
	if value == "" {
		return nil, time.Time{}, nil
	}
	signature, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, time.Time{}, err
	}
	signedAt, err := time.Parse(time.RFC3339, req.Request.FormValue("signedAt"))
	return signature, signedAt, err
}

type gitImportRequest struct {
	RepositoryURL string                `json:"repository_url"`
	Reference     string                `json:"reference"`
	RootPath      string                `json:"root_path"`
	Version       string                `json:"version"`
	CommitSHA     string                `json:"commit_sha"`
	Publisher     string                `json:"publisher"`
	Visibility    skilltypes.Visibility `json:"visibility"`
}

func (h *HTTPServer) discoverGitRelease(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, false)
	if err != nil {
		writeError(rsp, err)
		return
	}
	var body gitImportRequest
	if err := req.ReadEntity(&body); err != nil {
		writeError(rsp, err)
		return
	}
	if !h.service.CanPublishPublisher(req.Request.Context(), body.Publisher, actor) {
		writeError(rsp, market.ErrForbidden)
		return
	}
	discovery, err := h.syncer.DiscoverGitHubSkills(req.Request.Context(), body.RepositoryURL, body.Reference, body.RootPath, body.Version)
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, discovery, restful.MIME_JSON)
}

func (h *HTTPServer) importGitRelease(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, false)
	if err != nil {
		writeError(rsp, err)
		return
	}
	var body gitImportRequest
	if err := req.ReadEntity(&body); err != nil {
		writeError(rsp, err)
		return
	}
	if !h.service.CanPublishPublisher(req.Request.Context(), body.Publisher, actor) {
		writeError(rsp, market.ErrForbidden)
		return
	}
	if body.Visibility == "" {
		body.Visibility = skilltypes.VisibilityPrivate
	}
	if body.Visibility == skilltypes.VisibilityPublic {
		writeError(rsp, market.ErrManualPublicImport)
		return
	}
	if body.Visibility != skilltypes.VisibilityPrivate {
		writeError(rsp, fmt.Errorf("Git import visibility must be private"))
		return
	}
	if body.CommitSHA == "" {
		writeError(rsp, fmt.Errorf("Git import requires commit_sha from the discovery preview"))
		return
	}
	discovery, err := h.syncer.DiscoverGitHubSkills(req.Request.Context(), body.RepositoryURL, body.Reference, body.RootPath, body.Version)
	if err != nil {
		writeError(rsp, err)
		return
	}
	if !strings.EqualFold(discovery.CommitSHA, body.CommitSHA) {
		writeError(rsp, fmt.Errorf("Git tag moved after preview; discover the repository again before importing"))
		return
	}
	result := h.service.ImportGitDiscovery(req.Request.Context(), discovery, body.Publisher, body.Visibility, actor)
	_ = rsp.WriteHeaderAndJson(http.StatusCreated, result, restful.MIME_JSON)
}

type reviewRequest struct {
	Decision skilltypes.ReviewDecision `json:"decision"`
	Comment  string                    `json:"comment"`
}

func (h *HTTPServer) reviewRelease(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, true)
	if err != nil {
		writeError(rsp, err)
		return
	}
	var body reviewRequest
	if err := req.ReadEntity(&body); err != nil {
		writeError(rsp, err)
		return
	}
	if err := h.service.Review(req.Request.Context(), req.PathParameter("releaseId"), body.Decision, body.Comment, actor); err != nil {
		writeError(rsp, err)
		return
	}
	rsp.WriteHeader(http.StatusNoContent)
}

func (h *HTTPServer) listReviews(req *restful.Request, rsp *restful.Response) {
	if _, err := h.principal(req, true); err != nil {
		writeError(rsp, err)
		return
	}
	status := skilltypes.ReleaseStatus(req.QueryParameter("status"))
	items, err := h.service.Store().ListSkillReviews(req.Request.Context(), status, 100)
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, map[string]any{"items": items}, restful.MIME_JSON)
}

type lifecycleRequest struct {
	Yanked     bool   `json:"yanked"`
	Deprecated string `json:"deprecated"`
}

func (h *HTTPServer) setLifecycle(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, false)
	if err != nil {
		writeError(rsp, err)
		return
	}
	var body lifecycleRequest
	if err := req.ReadEntity(&body); err != nil {
		writeError(rsp, err)
		return
	}
	if err := h.service.SetLifecycle(req.Request.Context(), req.PathParameter("publisher"), req.PathParameter("name"), req.PathParameter("version"), body.Yanked, body.Deprecated, actor); err != nil {
		writeError(rsp, err)
		return
	}
	rsp.WriteHeader(http.StatusNoContent)
}

func (h *HTTPServer) listGrants(req *restful.Request, rsp *restful.Response) {
	actor, skill, err := h.authorizeSkillOwner(req)
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = actor
	items, err := h.service.Store().ListSkillGrants(req.Request.Context(), skill.ID)
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, map[string]any{"items": items}, restful.MIME_JSON)
}
func (h *HTTPServer) replaceGrants(req *restful.Request, rsp *restful.Response) {
	_, skill, err := h.authorizeSkillOwner(req)
	if err != nil {
		writeError(rsp, err)
		return
	}
	var body struct {
		Items []*skilltypes.Grant `json:"items"`
	}
	if err := req.ReadEntity(&body); err != nil {
		writeError(rsp, err)
		return
	}
	for _, grant := range body.Items {
		if grant == nil || (grant.PrincipalType != "user" && grant.PrincipalType != "group" && grant.PrincipalType != "role") {
			writeError(rsp, fmt.Errorf("grant principalType must be user, group, or role"))
			return
		}
		grant.SkillID = skill.ID
	}
	if err := h.service.Store().ReplaceSkillGrants(req.Request.Context(), skill.ID, body.Items); err != nil {
		writeError(rsp, err)
		return
	}
	rsp.WriteHeader(http.StatusNoContent)
}
func (h *HTTPServer) authorizeSkillOwner(req *restful.Request) (market.Principal, *skilltypes.Skill, error) {
	actor, err := h.principal(req, false)
	if err != nil {
		return actor, nil, err
	}
	skill, err := h.service.Store().GetSkill(req.Request.Context(), req.PathParameter("publisher"), req.PathParameter("name"))
	if err != nil {
		return actor, nil, err
	}
	if !actor.Admin && actor.ID != skill.OwnerID {
		return actor, nil, market.ErrForbidden
	}
	return actor, skill, nil
}

func (h *HTTPServer) listPublisherMembers(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, false)
	if err != nil {
		writeError(rsp, err)
		return
	}
	publisher, err := h.service.Store().GetSkillPublisher(req.Request.Context(), req.PathParameter("publisher"))
	if err != nil || (!actor.Admin && actor.ID != publisher.OwnerID) {
		writeError(rsp, market.ErrForbidden)
		return
	}
	items, err := h.service.Store().ListSkillPublisherMembers(req.Request.Context(), publisher.ID)
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, map[string]any{"items": items}, restful.MIME_JSON)
}
func (h *HTTPServer) replacePublisherMembers(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, false)
	if err != nil {
		writeError(rsp, err)
		return
	}
	publisher, err := h.service.Store().GetSkillPublisher(req.Request.Context(), req.PathParameter("publisher"))
	if err != nil || (!actor.Admin && actor.ID != publisher.OwnerID) {
		writeError(rsp, market.ErrForbidden)
		return
	}
	var body struct {
		Items []*skilltypes.PublisherMember `json:"items"`
	}
	if err := req.ReadEntity(&body); err != nil {
		writeError(rsp, err)
		return
	}
	for _, item := range body.Items {
		if item == nil || (item.PrincipalType != "user" && item.PrincipalType != "group" && item.PrincipalType != "role") ||
			(item.Role != "owner" && item.Role != "publisher" && item.Role != "viewer") {
			writeError(rsp, fmt.Errorf("publisher member requires user/group/role and owner/publisher/viewer"))
			return
		}
		item.PublisherID = publisher.ID
	}
	if err := h.service.Store().ReplaceSkillPublisherMembers(req.Request.Context(), publisher.ID, body.Items); err != nil {
		writeError(rsp, err)
		return
	}
	rsp.WriteHeader(http.StatusNoContent)
}
func (h *HTTPServer) revokePublisherKey(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, false)
	if err != nil {
		writeError(rsp, err)
		return
	}
	publisher, err := h.service.Store().GetSkillPublisher(req.Request.Context(), req.PathParameter("publisher"))
	if err != nil || (!actor.Admin && actor.ID != publisher.OwnerID) {
		writeError(rsp, market.ErrForbidden)
		return
	}
	version, _ := strconv.ParseUint(req.PathParameter("version"), 10, 32)
	if err := h.service.Store().RevokeSkillPublisherKey(req.Request.Context(), publisher.ID, uint32(version), time.Now()); err != nil {
		writeError(rsp, err)
		return
	}
	rsp.WriteHeader(http.StatusNoContent)
}

func (h *HTTPServer) listRegistrySources(req *restful.Request, rsp *restful.Response) {
	if _, err := h.principal(req, true); err != nil {
		writeError(rsp, err)
		return
	}
	items, err := h.service.Store().ListRegistrySources(req.Request.Context())
	if err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, map[string]any{"items": items}, restful.MIME_JSON)
}
func (h *HTTPServer) createRegistrySource(req *restful.Request, rsp *restful.Response) {
	actor, err := h.principal(req, true)
	if err != nil {
		writeError(rsp, err)
		return
	}
	var body registrySourceRequest
	if err := req.ReadEntity(&body); err != nil {
		writeError(rsp, err)
		return
	}
	source := body.toSource()
	if err := prepareRegistrySource(source, actor.ID, true); err != nil {
		writeError(rsp, err)
		return
	}
	if err := h.service.Store().CreateRegistrySource(req.Request.Context(), source); err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusCreated, source, restful.MIME_JSON)
}
func (h *HTTPServer) updateRegistrySource(req *restful.Request, rsp *restful.Response) {
	if _, err := h.principal(req, true); err != nil {
		writeError(rsp, err)
		return
	}
	var body registrySourceRequest
	if err := req.ReadEntity(&body); err != nil {
		writeError(rsp, err)
		return
	}
	source := body.toSource()
	source.ID = req.PathParameter("id")
	if err := prepareRegistrySource(source, "", false); err != nil {
		writeError(rsp, err)
		return
	}
	if err := h.service.Store().UpdateRegistrySource(req.Request.Context(), source); err != nil {
		writeError(rsp, err)
		return
	}
	_ = rsp.WriteHeaderAndJson(http.StatusOK, source, restful.MIME_JSON)
}
func (h *HTTPServer) deleteRegistrySource(req *restful.Request, rsp *restful.Response) {
	if _, err := h.principal(req, true); err != nil {
		writeError(rsp, err)
		return
	}
	if err := h.service.Store().DeleteRegistrySource(req.Request.Context(), req.PathParameter("id")); err != nil {
		writeError(rsp, err)
		return
	}
	rsp.WriteHeader(http.StatusNoContent)
}
func (h *HTTPServer) syncRegistrySource(req *restful.Request, rsp *restful.Response) {
	if _, err := h.principal(req, true); err != nil {
		writeError(rsp, err)
		return
	}
	if err := h.syncer.Sync(req.Request.Context(), req.PathParameter("id")); err != nil {
		writeError(rsp, err)
		return
	}
	rsp.WriteHeader(http.StatusNoContent)
}

func prepareRegistrySource(source *skilltypes.RegistrySource, actor string, create bool) error {
	if source.Name == "" {
		return fmt.Errorf("registry source name is required")
	}
	if _, err := url.ParseRequestURI(source.URL); err != nil {
		return fmt.Errorf("invalid registry URL")
	}
	if source.Type != skilltypes.RegistryHTTP && source.Type != skilltypes.RegistryPole && source.Type != skilltypes.RegistryGit {
		return fmt.Errorf("unsupported registry type")
	}
	if source.TrustLevel != skilltypes.TrustTrusted {
		source.TrustLevel = skilltypes.TrustUntrusted
	}
	if source.SyncIntervalSeconds == 0 {
		source.SyncIntervalSeconds = 3600
	}
	if create {
		source.ID, source.CreateBy, source.SyncStatus = utils.NewUUID(), actor, "idle"
	}
	source.NextSyncTime = time.Now().UTC()
	return nil
}

type registrySourceRequest struct {
	Name                     string                  `json:"name"`
	URL                      string                  `json:"url"`
	Type                     skilltypes.RegistryType `json:"type"`
	Enabled                  bool                    `json:"enabled"`
	TrustLevel               skilltypes.TrustLevel   `json:"trustLevel"`
	TrustLevelSnake          skilltypes.TrustLevel   `json:"trust_level"`
	SyncIntervalSeconds      uint32                  `json:"syncIntervalSeconds"`
	SyncIntervalSecondsSnake uint32                  `json:"sync_interval_seconds"`
}

func (r registrySourceRequest) toSource() *skilltypes.RegistrySource {
	trust, interval := r.TrustLevel, r.SyncIntervalSeconds
	if trust == "" {
		trust = r.TrustLevelSnake
	}
	if interval == 0 {
		interval = r.SyncIntervalSecondsSnake
	}
	return &skilltypes.RegistrySource{Name: r.Name, URL: r.URL, Type: r.Type, Enabled: r.Enabled,
		TrustLevel: trust, SyncIntervalSeconds: interval}
}

func (h *HTTPServer) optionalPrincipal(req *restful.Request) market.Principal {
	if req.HeaderParameter("Authorization") == "" && req.HeaderParameter("X-Polaris-Token") == "" {
		return market.Principal{}
	}
	principal, _ := h.principal(req, false)
	return principal
}
func (h *HTTPServer) principal(req *restful.Request, admin bool) (market.Principal, error) {
	if h.credentials == nil {
		return market.Principal{}, market.ErrForbidden
	}
	ctx := (&httpcommon.Handler{Request: req}).ParseHeaderContext()
	authCtx := authtypes.NewAcquireContext(authtypes.WithRequestContext(ctx), authtypes.WithModule(authtypes.MaintainModule), authtypes.WithOperation(authtypes.Read), authtypes.WithMethod(authtypes.DescribeSystemConfiguration))
	if admin {
		if h.permissions == nil {
			return market.Principal{}, market.ErrForbidden
		}
		pass, err := h.permissions.CheckConsolePermission(authCtx)
		if err != nil || !pass {
			return market.Principal{}, market.ErrForbidden
		}
	} else if err := h.credentials.CheckCredential(authCtx); err != nil {
		return market.Principal{}, market.ErrForbidden
	}
	principal := market.PrincipalFromAcquireContext(authCtx)
	principal.Admin = admin || authtypes.ParseUserRole(authCtx.GetRequestContext()) == authtypes.OwnerUserRole
	h.expandPrincipal(authCtx, &principal)
	return principal, nil
}
func (h *HTTPServer) expandPrincipal(authCtx *authtypes.AcquireContext, out *market.Principal) {
	value, ok := authCtx.GetAttachment(authtypes.PrincipalKey)
	if !ok || h.cache == nil {
		return
	}
	principal, ok := value.(authtypes.Principal)
	if !ok {
		return
	}
	for _, role := range h.cache.Role().GetPrincipalRoles(principal) {
		out.Expanded = append(out.Expanded, market.PrincipalRef{Type: "role", ID: role.ID})
	}
	if principal.PrincipalType == authtypes.PrincipalUser {
		for _, group := range h.cache.User().GetUserLinkGroupIds(principal.PrincipalID) {
			out.Expanded = append(out.Expanded, market.PrincipalRef{Type: "group", ID: group})
		}
	}
}

func requestBaseURL(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	if forwarded := req.Header.Get("X-Forwarded-Proto"); forwarded == "https" {
		scheme = "https"
	}
	return scheme + "://" + req.Host
}
func writeError(rsp *restful.Response, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, market.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, market.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, market.ErrVersionExists), errors.Is(err, market.ErrDigestConflict):
		status = http.StatusConflict
	}
	_ = rsp.WriteHeaderAndJson(status, map[string]any{"code": status, "info": err.Error()}, restful.MIME_JSON)
}
