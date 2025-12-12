package wrkb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

type BenchParam struct {
	ProcName    string
	ConnNum     int
	URL         string
	Method      string
	Duration    time.Duration
	Verbose     bool
	RPSLimit    float64
	Body        string
	Headers     []string
	HTTPVersion string
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
	total := r.Stat.GoodCnt + r.Stat.BadCnt
	if total == 0 {
		return r
	}

	r.RPS = int(float64(total) / r.Param.Duration.Seconds())
	r.Latency = time.Duration(r.Stat.Time.Nanoseconds() / int64(total))

	r.Min = time.Duration(r.Stat.Histogram.Min()) * time.Nanosecond
	r.P50 = time.Duration(r.Stat.Histogram.ValueAtQuantile(50.0)) * time.Nanosecond
	r.P90 = time.Duration(r.Stat.Histogram.ValueAtQuantile(90.0)) * time.Nanosecond
	r.P99 = time.Duration(r.Stat.Histogram.ValueAtQuantile(99.0)) * time.Nanosecond
	r.P999 = time.Duration(r.Stat.Histogram.ValueAtQuantile(99.9)) * time.Nanosecond
	r.Max = time.Duration(r.Stat.Histogram.Max()) * time.Nanosecond

	return r
}

func BenchHTTP(param BenchParam) BenchResult {
	client, cleanup, err := NewHTTPClient(param.HTTPVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize HTTP client: %v\n", err)
		return BenchResult{Param: param, Stat: BenchStat{}}
	}
	if cleanup != nil {
		defer cleanup()
	}

	if param.Verbose {
		fmt.Printf("* Using HTTP/%s transport\n", normalizeHTTPVersion(param.HTTPVersion))
	}
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

func runWorker(ctx context.Context, client *http.Client, param BenchParam) BenchStat {
	stat := newBenchStat()

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
			if limiter != nil {
				select {
				case <-limiter:
				case <-ctx.Done():
					return stat
				}
			}

			url := substitute(param.URL)

			var body string
			if param.Body != "" {
				body = substitute(param.Body)
			}

			req, err := http.NewRequestWithContext(ctx, param.Method, url, strings.NewReader(body))
			if err != nil {
				stat.ErrorCnt++
				continue
			}

			for _, h := range param.Headers {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}

			if len(body) > 0 {
				if req.Header.Get("Content-Type") == "" {
					req.Header.Set("Content-Type", "application/json")
				}
				stat.BodyReqSize += len(body)
			}

			if param.Verbose {
				fmt.Printf("*   Trying %s...\n", req.URL.Host)
				fmt.Printf("* Connected to %s (%s)\n", req.URL.Host, req.URL.Host)
				fmt.Printf("> %s %s HTTP/%s\n", req.Method, req.URL.Path, normalizeHTTPVersion(param.HTTPVersion))
				fmt.Printf("> Host: %s\n", req.URL.Host)
				for k, vs := range req.Header {
					for _, v := range vs {
						fmt.Printf("> %s: %s\n", k, v)
					}
				}
				if len(body) > 0 {
					fmt.Printf(">\n> %s\n", body)
				}
				fmt.Printf("> \n* Request completely sent off\n")
			}

			startReq := time.Now()
			resp, err := client.Do(req)
			elapsedReq := time.Since(startReq)

			var code int
			var bodyRespSize int

			if resp != nil {
				code = resp.StatusCode
				respBody, readErr := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if readErr == nil {
					bodyRespSize = len(respBody)
				}
			}

			if param.Verbose {
				statusLine := fmt.Sprintf("< HTTP/%s %d", normalizeHTTPVersion(respProto(resp)), code)
				fmt.Println(statusLine)
				if resp != nil {
					for k, vs := range resp.Header {
						for _, v := range vs {
							fmt.Printf("< %s: %s\n", k, v)
						}
					}
				}
				fmt.Printf("< \n")

				if resp != nil && bodyRespSize > 0 {
					fmt.Printf("* Response body length: %d bytes\n", bodyRespSize)
				}

				fmt.Printf("* Connection closed | time: %v | bodyRespSize: %d bytes\n\n", elapsedReq, bodyRespSize)
			}

			stat.Time += elapsedReq
			stat.Histogram.RecordValue(elapsedReq.Nanoseconds())

			switch {
			case err != nil:
				stat.ErrorCnt++
				if param.Verbose {
					fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
				}
			case code >= 200 && code < 400:
				stat.GoodCnt++
				stat.BodyRespSize += bodyRespSize
			default:
				stat.BadCnt++
			}
		}
	}
}

func respProto(resp *http.Response) string {
	if resp == nil {
		return "1.1"
	}
	if resp.ProtoMajor == 0 {
		return "1.1"
	}
	if resp.ProtoMinor == 0 {
		return fmt.Sprintf("%d", resp.ProtoMajor)
	}
	return fmt.Sprintf("%d.%d", resp.ProtoMajor, resp.ProtoMinor)
}
