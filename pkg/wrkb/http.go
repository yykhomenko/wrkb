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
	BodyRespSize              int
	Time                      time.Duration
}

type BenchResult struct {
	Param   BenchParam
	Stat    BenchStat
	RPS     int
	Latency time.Duration
}

func (s BenchStat) Add(other BenchStat) BenchStat {
	s.GoodCnt += other.GoodCnt
	s.BadCnt += other.BadCnt
	s.ErrorCnt += other.ErrorCnt
	s.BodyRespSize += other.BodyRespSize
	s.Time += other.Time
	return s
}

func (r BenchResult) CalcStat() BenchResult {
	total := r.Stat.GoodCnt + r.Stat.BadCnt
	if total == 0 {
		return r
	}

	r.RPS = int(float64(total) / r.Param.Duration.Seconds())
	r.Latency = time.Duration(r.Stat.Time.Nanoseconds() / int64(total))

	return r
}

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

func runWorker(ctx context.Context, client *fasthttp.Client, param BenchParam) BenchStat {
	var stat BenchStat

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

			code := resp.StatusCode()
			size := resp.Header.ContentLength()

			if param.Verbose {
				statusLine := fmt.Sprintf("< HTTP/1.1 %d %s", code, fasthttp.StatusMessage(code))
				fmt.Println(statusLine)
				resp.Header.VisitAll(func(k, v []byte) {
					fmt.Printf("< %s: %s\n", k, v)
				})
				fmt.Printf("< \n")

				body := resp.Body()
				if len(body) > 0 {
					fmt.Println(string(body))
				}

				fmt.Printf("* Connection closed | time: %v | size: %d bytes\n\n", elapsedReq, size)
			}

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)

			stat.Time += elapsedReq

			switch {
			case err != nil:
				stat.ErrorCnt++
				if param.Verbose {
					fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
				}
			case code >= 200 && code < 400:
				stat.GoodCnt++
				stat.BodyRespSize += size
			default:
				stat.BadCnt++
			}
		}
	}
}
