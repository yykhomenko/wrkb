package wrkb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// --- Основні типи ------------------------------------------------------------

type BenchParam struct {
	ProcName string
	ConnNum  int
	URL      string
	Method   string
	Duration time.Duration
	Verbose  bool
	RPSLimit float64
	Body     string
	Headers  []string
}

type BenchStat struct {
	GoodCnt, BadCnt, ErrorCnt int
	BodySize                  int
	Time                      time.Duration
}

type BenchResult struct {
	Param   BenchParam
	Stat    BenchStat
	RPS     int
	Latency time.Duration
}

// --- Агрегація ---------------------------------------------------------------

func (s BenchStat) Add(other BenchStat) BenchStat {
	s.GoodCnt += other.GoodCnt
	s.BadCnt += other.BadCnt
	s.ErrorCnt += other.ErrorCnt
	s.BodySize += other.BodySize
	s.Time += other.Time
	return s
}

func (r BenchResult) CalcStat() BenchResult {
	total := r.Stat.GoodCnt + r.Stat.BadCnt
	if total == 0 {
		return r
	}

	// Загальний RPS = кількість запитів / час тесту
	r.RPS = int(float64(total) / r.Param.Duration.Seconds())

	// Середня латентність = середній час одного запиту
	r.Latency = time.Duration(r.Stat.Time.Nanoseconds() / int64(total))

	return r
}

// --- Основна логіка ----------------------------------------------------------

func BenchHTTP(param BenchParam) BenchResult {
	client := newClient(param)
	ctx, cancel := context.WithTimeout(context.Background(), param.Duration)
	defer cancel()

	stats := make(chan BenchStat, param.ConnNum)
	wg := sync.WaitGroup{}

	for i := 0; i < param.ConnNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stat := runWorker(ctx, client, param)
			stats <- stat
		}()
	}

	go func() {
		wg.Wait()
		close(stats)
	}()

	final := BenchStat{}
	for s := range stats {
		final = final.Add(s)
	}

	return (BenchResult{
		Param: param,
		Stat:  final,
	}).CalcStat()
}

// --- Worker для одного потоку -----------------------------------------------

func runWorker(ctx context.Context, client *fasthttp.Client, param BenchParam) BenchStat {
	var stat BenchStat

	// Якщо встановлено RPS-ліміт — створюємо інтервал між запитами
	var limiter <-chan time.Time
	if param.RPSLimit > 0 {
		interval := time.Second / time.Duration(param.RPSLimit)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		limiter = ticker.C
	}

	for {
		select {
		case <-ctx.Done():
			return stat
		default:
			// чекаємо сигнал від обмежувача, якщо задано
			if limiter != nil {
				select {
				case <-limiter:
				case <-ctx.Done():
					return stat
				}
			}

			start := time.Now()
			code, size, err := makeRequest(client, param)
			stat.Time += time.Since(start)

			switch {
			case err != nil:
				stat.ErrorCnt++
				if param.Verbose {
					fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
				}
			case code >= 200 && code < 400:
				stat.GoodCnt++
				stat.BodySize += size
			default:
				stat.BadCnt++
			}
		}
	}
}

// --- HTTP логіка -------------------------------------------------------------

func makeRequest(client *fasthttp.Client, param BenchParam) (int, int, error) {
	url := substitute(param.URL)

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(param.Method)
	req.SetRequestURI(url)

	// --- встановлюємо заголовки ---
	for _, h := range param.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// --- тіло запиту ---
	var body string
	if param.Body != "" {
		body = substitute(param.Body)
		req.SetBodyString(body)

		// якщо користувач не вказав Content-Type — додаємо стандартний JSON
		if req.Header.Peek("Content-Type") == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	resp := fasthttp.AcquireResponse()
	err := client.Do(req, resp)

	code := resp.StatusCode()
	size := resp.Header.ContentLength()

	if param.Verbose {
		fmt.Printf("DEBUG %s → %d %s %s\n", url, code, param.Method, body)
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
	return code, size, err
}

func newClient(param BenchParam) *fasthttp.Client {
	return &fasthttp.Client{
		ReadTimeout:                   1 * time.Second,
		WriteTimeout:                  1 * time.Second,
		MaxIdleConnDuration:           1 * time.Minute,
		DisablePathNormalizing:        true,
		DisableHeaderNamesNormalizing: true,
		NoDefaultUserAgentHeader:      true,
		Dial: (&fasthttp.TCPDialer{
			DNSCacheDuration: 1 * time.Hour,
		}).Dial,
	}
}
