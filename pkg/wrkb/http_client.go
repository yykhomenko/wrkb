package wrkb

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go/http3"
	"golang.org/x/net/http2"
)

// NewHTTPClient builds an *http.Client configured for the requested HTTP version.
// It returns the client, an optional cleanup function, and an error.
func NewHTTPClient(version string) (*http.Client, func(), error) {
	normalized, err := ValidateHTTPVersion(version)
	if err != nil {
		return nil, nil, err
	}

	switch normalized {
	case "1.1":
		tr := defaultHTTPTransport()
		return &http.Client{Transport: tr}, tr.CloseIdleConnections, nil
	case "2":
		tr := &http2.Transport{}
		return &http.Client{Transport: tr}, tr.CloseIdleConnections, nil
	case "3":
		h1 := defaultHTTPTransport()
		h3 := &http3.RoundTripper{}
		rt := &http3FallbackTransport{primary: h3, fallback: h1}
		cleanup := func() {
			h3.Close()
			h1.CloseIdleConnections()
		}
		return &http.Client{Transport: rt}, cleanup, nil
	default:
		return nil, nil, fmt.Errorf("unsupported http version: %s", version)
	}
}

// defaultHTTPTransport mirrors the standard library defaults while allowing connection reuse and timeouts.
func defaultHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}
}

type http3FallbackTransport struct {
	primary  *http3.RoundTripper
	fallback http.RoundTripper

	mu          sync.RWMutex
	useFallback bool
}

func (t *http3FallbackTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	if t.shouldUseFallback() || t.primary == nil {
		return t.fallback.RoundTrip(req)
	}

	resp, err := t.primary.RoundTrip(req)
	if err == nil {
		return resp, nil
	}

	fmt.Fprintf(os.Stderr, "HTTP/3 failed, falling back to HTTP/1.1: %v\n", err)
	t.enableFallback()

	if t.fallback == nil {
		return nil, err
	}

	return t.fallback.RoundTrip(req)
}

func (t *http3FallbackTransport) shouldUseFallback() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.useFallback
}

func (t *http3FallbackTransport) enableFallback() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.useFallback = true
}

// ValidateHTTPVersion ensures the provided version string maps to a supported HTTP version.
// It returns the normalized value ("1.1", "2", or "3").
func ValidateHTTPVersion(version string) (string, error) {
	v := normalizeHTTPVersion(version)
	switch v {
	case "1.1", "2", "3":
		return v, nil
	default:
		return "", fmt.Errorf("unsupported http version: %s", version)
	}
}

func normalizeHTTPVersion(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		return "1.1"
	}

	switch v {
	case "1":
		return "1.1"
	case "2", "2.0":
		return "2"
	case "3", "3.0":
		return "3"
	default:
		return v
	}
}

// configureHTTP2Server is used only in tests to ensure TLS servers speak h2.
func configureHTTP2Server(srv *http.Server) error {
	srv.TLSConfig = &tls.Config{NextProtos: []string{"h2", "http/1.1"}}
	return http2.ConfigureServer(srv, &http2.Server{})
}
