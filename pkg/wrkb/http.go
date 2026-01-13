package wrkb

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/valyala/fasthttp"
)

type BenchParam struct {
	ProcName string
	ConnNum  int
	URL      string
	Method   string
	Duration time.Duration
	Verbose  bool
	RPSLimit float64
	MaxReqs  int
	Body     string
	Headers  []string
}

type BenchStat struct {
	GoodCnt      int
	BadCnt       int
	ErrorCnt     int
	BodyReqSize  int
	BodyRespSize int
	Time         time.Duration
	Histogram    *hdrhistogram.Histogram
}

func (s BenchStat) Add(other BenchStat) BenchStat {
	s.GoodCnt += other.GoodCnt
	s.BadCnt += other.BadCnt
	s.ErrorCnt += other.ErrorCnt
	s.BodyRespSize += other.BodyRespSize
	s.BodyReqSize += other.BodyReqSize
	s.Time += other.Time

	if other.Histogram != nil {
		if s.Histogram == nil {
			s.Histogram = other.Histogram
		} else {
			s.Histogram.Merge(other.Histogram)
		}
	}
	return s
}

type BenchResult struct {
	Param          BenchParam
	Stat           BenchStat
	ActualDuration time.Duration
	RPS            int
	Latency        time.Duration
	Min            time.Duration
	P50            time.Duration
	P90            time.Duration
	P99            time.Duration
	P999           time.Duration
	Max            time.Duration
}

func (r BenchResult) CalcStat() BenchResult {

	totalRequests := r.Stat.GoodCnt + r.Stat.BadCnt + r.Stat.ErrorCnt
	if totalRequests == 0 {
		return r
	}

	duration := r.Param.Duration
	if r.ActualDuration > 0 {
		duration = r.ActualDuration
	}
	if duration > 0 {
		r.RPS = int(float64(totalRequests) / duration.Seconds())
	}

	measuredCount := r.Stat.Histogram.TotalCount()
	if measuredCount > 0 {
		r.Latency = time.Duration(r.Stat.Time.Nanoseconds() / measuredCount)

		r.Min = time.Duration(r.Stat.Histogram.Min()) * time.Nanosecond
		r.P50 = time.Duration(r.Stat.Histogram.ValueAtQuantile(50.0)) * time.Nanosecond
		r.P90 = time.Duration(r.Stat.Histogram.ValueAtQuantile(90.0)) * time.Nanosecond
		r.P99 = time.Duration(r.Stat.Histogram.ValueAtQuantile(99.0)) * time.Nanosecond
		r.P999 = time.Duration(r.Stat.Histogram.ValueAtQuantile(99.9)) * time.Nanosecond
		r.Max = time.Duration(r.Stat.Histogram.Max()) * time.Nanosecond
	}

	return r
}

func BenchHTTP(param BenchParam) BenchResult {

	start := time.Now()
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), param.Duration)
	defer cancelTimeout()

	ctx := ctxTimeout
	var cancelAll context.CancelFunc
	if param.MaxReqs > 0 {
		ctx, cancelAll = context.WithCancel(ctxTimeout)
		defer cancelAll()
	}

	client := &fasthttp.Client{
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

	var limiter <-chan time.Time
	var stopLimiter func()
	if param.RPSLimit > 0 {
		interval := time.Duration(float64(time.Second) / param.RPSLimit)
		if interval <= 0 {
			interval = time.Nanosecond
		}

		ticker := time.NewTicker(interval)
		limiter = ticker.C
		stopLimiter = ticker.Stop
		defer stopLimiter()
	}

	stats := make(chan BenchStat, param.ConnNum)
	wg := sync.WaitGroup{}
	var reqCount int64

	for i := 0; i < param.ConnNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats <- runWorker(ctx, param, client, limiter, &reqCount, cancelAll)
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
		Param:          param,
		Stat:           final,
		ActualDuration: time.Since(start),
	}).CalcStat()
}

func runWorker(
	ctx context.Context,
	param BenchParam,
	client *fasthttp.Client,
	limiter <-chan time.Time,
	reqCount *int64,
	cancelAll context.CancelFunc) BenchStat {

	stat := BenchStat{Histogram: hdrhistogram.New(1_000, 10_000_000_000, 3)}
	hasContentTypeHeader := hasHeader(param.Headers, "Content-Type")

	for {
		select {
		case <-ctx.Done():
			return stat
		default:

			if limiter != nil {
				select {
				case <-limiter:
				case <-ctx.Done():
					return stat
				}
			}

			if param.MaxReqs > 0 {
				next := int(atomic.AddInt64(reqCount, 1))
				if next > param.MaxReqs {
					if cancelAll != nil {
						cancelAll()
					}
					return stat
				}
			}

			req := fasthttp.AcquireRequest()

			req.Header.SetMethod(param.Method)
			req.SetRequestURI(substitute(param.URL))
			setHeaders(req, param.Headers)
			stat.BodyReqSize += setBody(req, param.Body, hasContentTypeHeader)
			logRequest(req, param.Verbose)

			resp := fasthttp.AcquireResponse()

			start := time.Now()
			err := client.Do(req, resp)
			elapsed := time.Since(start)
			fasthttp.ReleaseRequest(req)

			if err != nil {
				fasthttp.ReleaseResponse(resp)
				stat.ErrorCnt++
				if param.Verbose {
					fmt.Printf("ERR: %v\n", err)
				}
				continue
			}

			logResponse(resp, param.Verbose, elapsed)
			updateStatistic(&stat, resp, elapsed)

			fasthttp.ReleaseResponse(resp)
		}
	}
}

func updateStatistic(stat *BenchStat, resp *fasthttp.Response, elapsed time.Duration) {
	stat.Time += elapsed
	stat.Histogram.RecordValue(elapsed.Nanoseconds())

	code := resp.StatusCode()
	switch {
	case code >= 200 && code < 400:
		stat.GoodCnt++
		stat.BodyRespSize += len(resp.Body())
	default:
		stat.BadCnt++
	}
}

func logResponse(resp *fasthttp.Response, isVerbose bool, elapsedReq time.Duration) {
	if isVerbose {
		code := resp.StatusCode()
		statusLine := fmt.Sprintf("< HTTP/1.1 %d %s", code, fasthttp.StatusMessage(code))
		fmt.Println(statusLine)
		resp.Header.VisitAll(func(k, v []byte) {
			fmt.Printf("< %s: %s\n", k, v)
		})
		fmt.Printf("< \n")

		bodyBytes := resp.Body()
		if len(bodyBytes) > 0 {
			fmt.Println(string(bodyBytes))
		}
		fmt.Printf("* Connection closed | time: %v | bodyRespSize: %d bytes\n\n", elapsedReq, len(bodyBytes))
	}
}

func hasHeader(headers []string, key string) bool {

	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(parts[0]), key) {
			return true
		}
	}

	return false
}

func setHeaders(req *fasthttp.Request, headers []string) {
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
}

func setBody(req *fasthttp.Request, body string, hasContentTypeHeader bool) int {

	if body != "" {

		body = substitute(body)
		req.SetBodyString(body)

		if !hasContentTypeHeader {
			req.Header.Set("Content-Type", "application/json")
		}

		return len(body)
	}

	return 0
}

func logRequest(req *fasthttp.Request, isVerbose bool) {
	if isVerbose {
		fmt.Printf("*   Trying %s...\n", req.URI().Host())
		fmt.Printf("* Connected to %s (%s)\n", req.URI().Host(), req.URI().Host())
		fmt.Printf("> %s %s HTTP/1.1\n", req.Header.Method(), req.URI().PathOriginal())
		fmt.Printf("> Host: %s\n", req.URI().Host())
		req.Header.VisitAll(func(k, v []byte) {
			fmt.Printf("> %s: %s\n", k, v)
		})
		body := req.Body()
		if len(body) > 0 {
			fmt.Printf(">\n> %s\n", body)
		}
		fmt.Printf("> \n* Request completely sent off\n")
	}
}
