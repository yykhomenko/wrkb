package wrkb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

var mockServerURL string

func TestMain(m *testing.M) {
	// простий сервер, який відповідає залежно від запиту
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/slow":
			time.Sleep(50 * time.Millisecond)
			w.WriteHeader(200)
			_, _ = w.Write([]byte("slow-ok"))
		case "/bad":
			w.WriteHeader(500)
			_, _ = w.Write([]byte("error"))
		default:
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ok"))
		}
	}))
	defer srv.Close()

	mockServerURL = srv.URL
	fmt.Printf("🧪 Mock server running at: %s\n", mockServerURL)

	code := m.Run()
	os.Exit(code)
}

func baseParams(conn int, path string) BenchParam {
	return BenchParam{
		URL:      mockServerURL + path,
		Method:   "GET",
		ConnNum:  conn,
		Duration: 1 * time.Second,
		Verbose:  false,
	}
}

func TestBenchHTTP_Basic(t *testing.T) {
	param := baseParams(4, "/")
	res := BenchHTTP(param)

	if res.RPS <= 0 {
		t.Fatalf("expected positive RPS, got %d", res.RPS)
	}
	if res.Latency <= 0 {
		t.Fatalf("expected latency > 0, got %v", res.Latency)
	}

	t.Logf("RPS: %d, latency: %v", res.RPS, res.Latency)
}

func TestBenchHTTP_Scenarios(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"fast", "/"},
		{"slow", "/slow"},
		{"error", "/bad"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := baseParams(2, tt.path)
			res := BenchHTTP(param)
			t.Logf("[%s] RPS=%d latency=%v good=%d bad=%d err=%d",
				tt.name, res.RPS, res.Latency,
				res.Stat.GoodCnt, res.Stat.BadCnt, res.Stat.ErrorCnt)
		})
	}
}

func TestBenchHTTP_AutoContentType(t *testing.T) {
	var total int64
	var missing int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&total, 1)
		if r.Header.Get("Content-Type") == "" {
			atomic.AddInt64(&missing, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	param := BenchParam{
		URL:      srv.URL,
		Method:   "PUT",
		ConnNum:  1,
		Duration: 2 * time.Second,
		MaxReqs:  2,
		Body:     `{"a": 1}`,
	}

	_ = BenchHTTP(param)

	if got := atomic.LoadInt64(&total); got != 2 {
		t.Fatalf("expected 2 requests, got %d", got)
	}
	if got := atomic.LoadInt64(&missing); got != 0 {
		t.Fatalf("expected Content-Type on all requests, missing on %d", got)
	}
}

func BenchmarkBenchHTTP(b *testing.B) {
	connLevels := []int{1, 2, 4, 8}

	for _, c := range connLevels {
		b.Run(fmt.Sprintf("conn_%02d", c), func(b *testing.B) {
			param := baseParams(c, "/slow")
			for i := 0; i < b.N; i++ {
				BenchHTTP(param)
			}
		})
	}
}
