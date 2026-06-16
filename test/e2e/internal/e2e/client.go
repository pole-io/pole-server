//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"
)

const ExecuteSuccess = 200000

type APIResponse struct {
	Code      int             `json:"code"`
	Info      string          `json:"info"`
	Data      json.RawMessage `json:"data"`
	Amount    int             `json:"amount"`
	Size      int             `json:"size"`
	Responses []APIResponse   `json:"responses"`
}

type HTTPClient struct {
	t         *testing.T
	env       *Env
	client    *http.Client
	userID    string
	authToken string
}

func NewHTTPClient(t *testing.T, env *Env) *HTTPClient {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}
	return &HTTPClient{
		t:   t,
		env: env,
		client: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HTTPClient) LoginAsMainUser() {
	c.t.Helper()
	createResp := c.Console("POST", "/admin/v1/mainuser/create", map[string]any{
		"name":     "admin",
		"password": "admin123",
		"comment":  "e2e admin",
		"source":   "pole-io",
	})
	if createResp.Code != ExecuteSuccess {
		c.t.Logf("main user create returned code=%d info=%s; login will still be attempted", createResp.Code, createResp.Info)
	}
	c.Login("admin", "admin123")
}

func (c *HTTPClient) Login(name, password string) {
	c.t.Helper()
	loginResp := c.Console("POST", "/auth/v1/user/login", map[string]any{
		"name":     name,
		"password": password,
	})
	RequireSuccess(c.t, loginResp, "login main user")
	data := loginResp.Map(c.t)
	if token, ok := data["token"].(string); ok && token != "" {
		c.authToken = token
	}
	if id, ok := data["user_id"].(string); ok && id != "" {
		c.userID = id
		return
	}
	if id, ok := data["id"].(string); ok && id != "" {
		c.userID = id
		return
	}
	c.t.Fatalf("login response does not contain user id: %s", string(loginResp.Data))
}

func (c *HTTPClient) AuthToken() string {
	return c.authToken
}

func (c *HTTPClient) Console(method, path string, body any) APIResponse {
	c.t.Helper()
	return c.do(method, c.env.ConsoleBaseURL, path, nil, body, true)
}

func (c *HTTPClient) ConsoleQuery(method, path string, query map[string]string, body any) APIResponse {
	c.t.Helper()
	return c.do(method, c.env.ConsoleBaseURL, path, query, body, true)
}

func (c *HTTPClient) Client(method, path string, body any) APIResponse {
	c.t.Helper()
	return c.do(method, c.env.ClientBaseURL, path, nil, body, false)
}

func (c *HTTPClient) ClientWithHeaders(method, path string, body any, headers map[string]string) APIResponse {
	c.t.Helper()
	return c.do(method, c.env.ClientBaseURL, path, nil, body, false, headers)
}

func (c *HTTPClient) ClientQuery(method, path string, query map[string]string, body any) APIResponse {
	c.t.Helper()
	return c.do(method, c.env.ClientBaseURL, path, query, body, false)
}

func (c *HTTPClient) do(method, baseURL, path string, query map[string]string, body any, console bool, extraHeaders ...map[string]string) APIResponse {
	c.t.Helper()
	fullURL := baseURL + path
	if len(query) > 0 {
		values := url.Values{}
		for k, v := range query {
			if v != "" {
				values.Set(k, v)
			}
		}
		fullURL += "?" + values.Encode()
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("marshal %s %s: %v", method, fullURL, err)
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, fullURL, reader)
	if err != nil {
		c.t.Fatalf("new request %s %s: %v", method, fullURL, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Polaris-Request-Id", "e2e-"+strings.ReplaceAll(path, "/", "-"))
	if console && c.userID != "" {
		req.Header.Set("x-pole-user", c.userID)
	}
	for _, headers := range extraHeaders {
		for key, value := range headers {
			if key != "" && value != "" {
				req.Header.Set(key, value)
			}
		}
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatalf("request %s %s: %v\n%s", method, fullURL, err, c.env.LogTail())
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("read response %s %s: %v", method, fullURL, err)
	}
	var apiResp APIResponse
	if len(b) > 0 {
		if err := json.Unmarshal(b, &apiResp); err != nil {
			c.t.Fatalf("decode response %s %s http=%d body=%s err=%v", method, fullURL, resp.StatusCode, string(b), err)
		}
	}
	if resp.StatusCode >= 500 && apiResp.Code == 0 {
		c.t.Fatalf("request %s %s returned http=%d body=%s\n%s", method, fullURL, resp.StatusCode, string(b), c.env.LogTail())
	}
	return apiResp
}

func ExtractToken(t *testing.T, resp APIResponse) string {
	t.Helper()
	data := resp.Map(t)
	for _, key := range []string{"token", "auth_token", "access_token"} {
		if token, ok := data[key].(string); ok && token != "" {
			return token
		}
	}
	t.Fatalf("response does not contain token: %s", string(resp.Data))
	return ""
}

func RequireSuccess(t *testing.T, resp APIResponse, action string) {
	t.Helper()
	if resp.Code != ExecuteSuccess {
		t.Fatalf("%s: code=%d info=%s data=%s", action, resp.Code, resp.Info, string(resp.Data))
	}
	for i, item := range resp.Responses {
		if item.Code != 0 && item.Code != ExecuteSuccess {
			t.Fatalf("%s response[%d]: code=%d info=%s", action, i, item.Code, item.Info)
		}
	}
}

func RequireDenied(t *testing.T, resp APIResponse, action string) {
	t.Helper()
	if resp.Code == ExecuteSuccess {
		t.Fatalf("%s: expected denied response, got success: %+v", action, resp)
	}
}

func (r APIResponse) Map(t *testing.T) map[string]any {
	t.Helper()
	if len(r.Data) == 0 || string(r.Data) == "null" {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(r.Data, &out); err != nil {
		t.Fatalf("decode data map: %v data=%s", err, string(r.Data))
	}
	return out
}

func (r APIResponse) Slice(t *testing.T) []map[string]any {
	t.Helper()
	if len(r.Data) == 0 || string(r.Data) == "null" {
		return nil
	}
	var out []map[string]any
	if err := json.Unmarshal(r.Data, &out); err != nil {
		t.Fatalf("decode data slice: %v data=%s", err, string(r.Data))
	}
	return out
}

func FindByName(items []map[string]any, name string) (map[string]any, bool) {
	return FindByField(items, "name", name)
}

func FindByField(items []map[string]any, field, value string) (map[string]any, bool) {
	for _, item := range items {
		if item[field] == value {
			return item, true
		}
	}
	return nil, false
}

func RequireNotFoundByName(t *testing.T, resp APIResponse, name, action string) {
	t.Helper()
	RequireSuccess(t, resp, action)
	if item, ok := FindByName(resp.Slice(t), name); ok {
		t.Fatalf("%s: unexpected item %s still found: %+v", action, name, item)
	}
}

func RequireNotFoundByField(t *testing.T, resp APIResponse, field, value, action string) {
	t.Helper()
	RequireSuccess(t, resp, action)
	if item, ok := FindByField(resp.Slice(t), field, value); ok {
		t.Fatalf("%s: unexpected item %s=%s still found: %+v", action, field, value, item)
	}
}

func StringField(t *testing.T, item map[string]any, key string) string {
	t.Helper()
	val, ok := item[key].(string)
	if !ok || val == "" {
		t.Fatalf("missing string field %q in %+v", key, item)
	}
	return val
}

func Eventually(t *testing.T, timeout, interval time.Duration, check func() bool, desc string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if check() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("condition not met within %s: %s", timeout, desc)
		}
		time.Sleep(interval)
	}
}

func Query(offset, limit int, fields map[string]string) map[string]string {
	q := map[string]string{
		"offset": fmt.Sprintf("%d", offset),
		"limit":  fmt.Sprintf("%d", limit),
	}
	for k, v := range fields {
		q[k] = v
	}
	return q
}
