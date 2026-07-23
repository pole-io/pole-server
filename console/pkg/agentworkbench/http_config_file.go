package agentworkbench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

const executeSuccessCode uint32 = 200000
const downstreamFailureCode uint32 = 502001

type UpstreamError struct {
	Code      uint32
	Info      string
	RequestID string
	Status    int
}

func (e *UpstreamError) Error() string {
	if e.RequestID == "" {
		return fmt.Sprintf("pole-server request failed: code=%d info=%s", e.Code, e.Info)
	}
	return fmt.Sprintf("pole-server request failed: code=%d info=%s requestId=%s", e.Code, e.Info, e.RequestID)
}

func newDownstreamFailure(info, requestID string) *UpstreamError {
	return &UpstreamError{Code: downstreamFailureCode, Info: info, RequestID: requestID, Status: http.StatusBadGateway}
}

type HTTPConfigFilePort struct {
	baseURL string
	client  *http.Client
	timeout atomic.Int64
}

func NewHTTPConfigFilePort(address string, client *http.Client) *HTTPConfigFilePort {
	if client == nil {
		client = http.DefaultClient
	}
	baseURL := strings.TrimRight(address, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	port := &HTTPConfigFilePort{baseURL: baseURL, client: client}
	port.timeout.Store(int64(10 * time.Second))
	return port
}

func (p *HTTPConfigFilePort) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		p.timeout.Store(int64(timeout))
	}
}

func (p *HTTPConfigFilePort) requestContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Duration(p.timeout.Load()))
}

func (p *HTTPConfigFilePort) GetConfigFile(ctx context.Context, actor Actor, ref ResourceRef) (ConfigFile, string, error) {
	ctx, cancel := p.requestContext(ctx)
	defer cancel()
	query := url.Values{}
	query.Set("namespace", ref.Namespace)
	query.Set("group", ref.Group)
	query.Set("name", ref.Name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/config/v1/files/detail?"+query.Encode(), nil)
	if err != nil {
		return ConfigFile{}, "", err
	}
	setActorHeaders(req, actor)
	resp, err := p.client.Do(req)
	if err != nil {
		return ConfigFile{}, "", newDownstreamFailure("request pole-server config file: "+err.Error(), "")
	}
	defer resp.Body.Close()
	requestID := resp.Header.Get("X-Request-Id")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ConfigFile{}, requestID, newDownstreamFailure("read pole-server config response: "+err.Error(), requestID)
	}
	var envelope struct {
		Code uint32          `json:"code"`
		Info string          `json:"info"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return ConfigFile{}, requestID, newDownstreamFailure("decode pole-server config response: "+err.Error(), requestID)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices || envelope.Code != executeSuccessCode {
		return ConfigFile{}, requestID, &UpstreamError{Code: envelope.Code, Info: envelope.Info, RequestID: requestID, Status: resp.StatusCode}
	}
	var wire struct {
		ConfigFile
		EncryptAlgoSnake string `json:"encrypt_algo"`
	}
	if err := json.Unmarshal(envelope.Data, &wire); err != nil {
		return ConfigFile{}, requestID, newDownstreamFailure("decode pole-server config file: "+err.Error(), requestID)
	}
	if wire.EncryptAlgo == "" {
		wire.EncryptAlgo = wire.EncryptAlgoSnake
	}
	return cloneConfigFile(wire.ConfigFile), requestID, nil
}

func (p *HTTPConfigFilePort) UpdateConfigFile(ctx context.Context, actor Actor, file ConfigFile) (string, error) {
	ctx, cancel := p.requestContext(ctx)
	defer cancel()
	payload, err := json.Marshal([]ConfigFile{file})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, p.baseURL+"/config/v1/files", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	setActorHeaders(req, actor)
	resp, err := p.client.Do(req)
	if err != nil {
		return "", newDownstreamFailure("update pole-server config file: "+err.Error(), "")
	}
	defer resp.Body.Close()
	requestID := resp.Header.Get("X-Request-Id")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return requestID, newDownstreamFailure("read pole-server update response: "+err.Error(), requestID)
	}
	var envelope struct {
		Code      uint32 `json:"code"`
		Info      string `json:"info"`
		Responses []struct {
			Code uint32 `json:"code"`
			Info string `json:"info"`
		} `json:"responses"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return requestID, newDownstreamFailure("decode pole-server update response: "+err.Error(), requestID)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices || envelope.Code != executeSuccessCode {
		return requestID, &UpstreamError{Code: envelope.Code, Info: envelope.Info, RequestID: requestID, Status: resp.StatusCode}
	}
	if len(envelope.Responses) != 1 {
		return requestID, newDownstreamFailure("pole-server update response must contain exactly one result", requestID)
	}
	for _, item := range envelope.Responses {
		if item.Code != executeSuccessCode {
			return requestID, &UpstreamError{Code: item.Code, Info: item.Info, RequestID: requestID, Status: resp.StatusCode}
		}
	}
	return requestID, nil
}

func setActorHeaders(req *http.Request, actor Actor) {
	req.Header.Set("Authorization", actor.Token)
	req.Header.Set("X-Pole-User", actor.UserID)
	if actor.RequestID != "" {
		req.Header.Set("X-Request-Id", actor.RequestID)
	}
}
