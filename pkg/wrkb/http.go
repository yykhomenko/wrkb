package wrkb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
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
	Body     string
	Headers  []string
}

type BenchStat struct {
	GoodCnt, BadCnt, ErrorCnt int
	BodyReqSize               int
	BodyRespSize              int
	Time                      time.Duration
	Histogram                 *hdrhistogram.Histogram
}

type BenchResult struct {
	Param   BenchParam
	Stat    BenchStat
	RPS     int
	Latency time.Duration
	Min     time.Duration
	P50     time.Duration
	P90     time.Duration
	P99     time.Duration
	P999    time.Duration
	Max     time.Duration
}

func newBenchStat() BenchStat {
	h := hdrhistogram.New(1_000, 10_000_000_000, 3)
	return BenchStat{Histogram: h}
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

func (r BenchResult) CalcStat() BenchResult {
	totalRequests := r.Stat.GoodCnt + r.Stat.BadCnt + r.Stat.ErrorCnt
	if totalRequests == 0 {
		return r
	}

	r.RPS = int(float64(totalRequests) / r.Param.Duration.Seconds())

	measuredCount := int64(r.Stat.Histogram.TotalCount())
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
	client := newClient(param)
	ctx, cancel := context.WithTimeout(context.Background(), param.Duration)
	defer cancel()

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
	}

	if stopLimiter != nil {
		defer stopLimiter()
	}

	stats := make(chan BenchStat, param.ConnNum)
	wg := sync.WaitGroup{}

	for i := 0; i < param.ConnNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stat := runWorker(ctx, client, param, limiter)
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

func runWorker(ctx context.Context, client *fasthttp.Client, param BenchParam, limiter <-chan time.Time) BenchStat {
	stat := newBenchStat()

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

			url := substitute(param.URL)

			req := fasthttp.AcquireRequest()
			req.Header.SetMethod(param.Method)
			req.SetRequestURI(url)

			for _, h := range param.Headers {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}

			var body string
			if param.Body != "" {
				body = substitute(param.Body)
				req.SetBodyString(body)

				if req.Header.Peek("Content-Type") == nil {
					req.Header.Set("Content-Type", "application/json")
				}

				stat.BodyReqSize += len(body)
			}

			resp := fasthttp.AcquireResponse()

			if param.Verbose {
				fmt.Printf("*   Trying %s...\n", req.URI().Host())
				fmt.Printf("* Connected to %s (%s)\n", req.URI().Host(), req.URI().Host())
				fmt.Printf("> %s %s HTTP/1.1\n", req.Header.Method(), req.URI().PathOriginal())
				fmt.Printf("> Host: %s\n", req.URI().Host())
				req.Header.VisitAll(func(k, v []byte) {
					fmt.Printf("> %s: %s\n", k, v)
				})
				if len(body) > 0 {
					fmt.Printf(">\n> %s\n", body)
				}
				fmt.Printf("> \n* Request completely sent off\n")
			}

			startReq := time.Now()
			err := client.Do(req, resp)
			elapsedReq := time.Since(startReq)

			if err != nil {
				stat.ErrorCnt++
				if param.Verbose {
					fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
				}

				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(resp)
				continue
			}

			code := resp.StatusCode()
			bodyBytes := resp.Body()
			bodyRespSize := len(bodyBytes)

			if param.Verbose {
				statusLine := fmt.Sprintf("< HTTP/1.1 %d %s", code, fasthttp.StatusMessage(code))
				fmt.Println(statusLine)
				resp.Header.VisitAll(func(k, v []byte) {
					fmt.Printf("< %s: %s\n", k, v)
				})
				fmt.Printf("< \n")

				if len(bodyBytes) > 0 {
					fmt.Println(string(bodyBytes))
				}
			}

			if param.Verbose {
				fmt.Printf("* Connection closed | time: %v | bodyRespSize: %d bytes\n\n", elapsedReq, bodyRespSize)
			}

			stat.Time += elapsedReq
			stat.Histogram.RecordValue(elapsedReq.Nanoseconds())

			switch {
			case code >= 200 && code < 400:
				stat.GoodCnt++
				stat.BodyRespSize += bodyRespSize
			default:
				stat.BadCnt++
			}

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
		}
	}
}
