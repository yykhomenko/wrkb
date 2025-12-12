package wrkb

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPClient_HTTP1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	client, cleanup, err := NewHTTPClient("1.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleanup != nil {
		cleanup()
	}

	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.ProtoMajor != 1 {
		t.Fatalf("expected HTTP/1.x, got %s", resp.Proto)
	}
}

func TestNewHTTPClient_HTTP2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := httptest.NewUnstartedServer(handler)
	if err := configureHTTP2Server(srv.Config); err != nil {
		t.Fatalf("configure h2: %v", err)
	}
	srv.StartTLS()
	defer srv.Close()

	client, cleanup, err := NewHTTPClient("2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.ProtoMajor != 2 {
		t.Fatalf("expected HTTP/2, got %s", resp.Proto)
	}
}

func TestNewHTTPClient_HTTP3Fallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fallback"))
	}))
	defer srv.Close()

	client, cleanup, err := NewHTTPClient("3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	done := make(chan struct{})
	go func() {
		resp, reqErr := client.Get(srv.URL)
		if reqErr != nil {
			t.Errorf("request failed: %v", reqErr)
			close(done)
			return
		}
		defer resp.Body.Close()

		if resp.ProtoMajor != 1 {
			t.Errorf("expected fallback to HTTP/1.x, got %s", resp.Proto)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("request timed out")
	}
}
