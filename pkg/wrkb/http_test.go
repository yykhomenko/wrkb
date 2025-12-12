package wrkb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
	if res.Stat.GoodCnt == 0 {
		t.Fatalf("expected at least one successful response, got %d", res.Stat.GoodCnt)
	}
	if res.Latency <= 0 {
		t.Fatalf("expected latency > 0, got %v", res.Latency)
	}
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

			if tt.path == "/bad" {
				if res.Stat.BadCnt == 0 {
					t.Fatalf("expected server errors for %s, got good=%d bad=%d", tt.name, res.Stat.GoodCnt, res.Stat.BadCnt)
				}
				return
			}

			if res.Stat.BadCnt > 0 || res.Stat.ErrorCnt > 0 {
				t.Fatalf("expected only successful responses for %s, got good=%d bad=%d err=%d", tt.name, res.Stat.GoodCnt, res.Stat.BadCnt, res.Stat.ErrorCnt)
			}
			if res.Latency <= 0 {
				t.Fatalf("expected latency > 0 for %s, got %v", tt.name, res.Latency)
			}
		})
	}
}

func TestBenchStatAddMergesMetrics(t *testing.T) {
	first := newBenchStat()
	first.GoodCnt = 2
	first.BadCnt = 1
	first.ErrorCnt = 1
	first.BodyReqSize = 30
	first.BodyRespSize = 60
	first.Time = 10 * time.Millisecond
	first.Histogram.RecordValue(5_000_000)  // 5ms
	first.Histogram.RecordValue(10_000_000) // 10ms

	second := newBenchStat()
	second.GoodCnt = 1
	second.BodyReqSize = 10
	second.BodyRespSize = 20
	second.Time = 4 * time.Millisecond
	second.Histogram.RecordValue(2_000_000) // 2ms

	combined := first.Add(second)

	if combined.GoodCnt != 3 || combined.BadCnt != 1 || combined.ErrorCnt != 1 {
		t.Fatalf("unexpected counters: %+v", combined)
	}
	if combined.BodyReqSize != 40 || combined.BodyRespSize != 80 {
		t.Fatalf("unexpected body sizes: req=%d resp=%d", combined.BodyReqSize, combined.BodyRespSize)
	}
	if combined.Time != 14*time.Millisecond {
		t.Fatalf("unexpected total time: %v", combined.Time)
	}

	if combined.Histogram == nil {
		t.Fatalf("expected merged histogram")
	}
	if count := combined.Histogram.TotalCount(); count != 3 {
		t.Fatalf("expected 3 histogram records, got %d", count)
	}
	if min := time.Duration(combined.Histogram.Min()) * time.Nanosecond; min < 1900*time.Microsecond || min > 2100*time.Microsecond {
		t.Fatalf("expected min around 2ms, got %v", min)
	}
}

func TestBenchResultCalcStat(t *testing.T) {
	stat := newBenchStat()
	stat.GoodCnt = 3
	stat.BadCnt = 1
	stat.Time = 30 * time.Millisecond
	stat.Histogram.RecordValue(10_000_000) // 10ms
	stat.Histogram.RecordValue(15_000_000) // 15ms
	stat.Histogram.RecordValue(20_000_000) // 20ms
	stat.Histogram.RecordValue(5_000_000)  // 5ms (bad request still counted)

	res := BenchResult{
		Param: BenchParam{Duration: 1 * time.Second},
		Stat:  stat,
	}.CalcStat()

	if res.RPS != 4 {
		t.Fatalf("expected RPS 4, got %d", res.RPS)
	}
	if res.Latency != 7*time.Millisecond+500*time.Microsecond {
		t.Fatalf("expected average latency 7.5ms, got %v", res.Latency)
	}
	if res.Min < 4*time.Millisecond+800*time.Microsecond || res.Min > 5*time.Millisecond+200*time.Microsecond {
		t.Fatalf("expected min around 5ms, got %v", res.Min)
	}
	if res.P50 < 10*time.Millisecond || res.P50 > 15*time.Millisecond {
		t.Fatalf("expected p50 between 10ms and 15ms, got %v", res.P50)
	}
	if res.Max < 19*time.Millisecond+800*time.Microsecond || res.Max > 20*time.Millisecond+200*time.Microsecond {
		t.Fatalf("expected max around 20ms, got %v", res.Max)
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
