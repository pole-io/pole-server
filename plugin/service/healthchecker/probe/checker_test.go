package probe

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/pole-io/pole-server/apis/service/healthcheck"
)

func TestTCPChecker(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	host, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(portText)
	checker := newChecker(TCPPluginName, healthcheck.HealthCheckerDetectTCP)
	response, err := checker.Check(&healthcheck.CheckRequest{QueryRequest: healthcheck.QueryRequest{
		Host: host,
		Port: uint32(port),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !response.Healthy || response.StayUnchanged || !response.Regular {
		t.Fatalf("unexpected healthy response: %+v", response)
	}
	_ = listener.Close()
	response, err = checker.Check(&healthcheck.CheckRequest{QueryRequest: healthcheck.QueryRequest{
		Host:    host,
		Port:    uint32(port),
		Healthy: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if response.Healthy || response.StayUnchanged {
		t.Fatalf("unexpected unhealthy response: %+v", response)
	}
}

func TestHTTPChecker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/health" {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	host, portText, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(portText)
	checker := newChecker(HTTPPluginName, healthcheck.HealthCheckerDetectHTTP)
	healthy, err := checker.Check(&healthcheck.CheckRequest{
		QueryRequest: healthcheck.QueryRequest{Host: host, Port: uint32(port)},
		Path:         "health",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !healthy.Healthy || healthy.StayUnchanged {
		t.Fatalf("unexpected healthy response: %+v", healthy)
	}
	unhealthy, err := checker.Check(&healthcheck.CheckRequest{
		QueryRequest: healthcheck.QueryRequest{Host: host, Port: uint32(port), Healthy: true},
		Path:         "/failure",
	})
	if err != nil {
		t.Fatal(err)
	}
	if unhealthy.Healthy || unhealthy.StayUnchanged {
		t.Fatalf("unexpected unhealthy response: %+v", unhealthy)
	}
}
