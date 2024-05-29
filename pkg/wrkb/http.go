package wrkb

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"os"
	"time"
)

type BenchParam struct {
	ConnNum  int
	URL      string
	Method   string
	Duration time.Duration
	Verbose  bool
}

type BenchStat struct {
	BenchParam
	GoodCnt  int
	BadCnt   int
	ErrorCnt int
	RPS      int
	Latency  time.Duration
	BodySize int
}

func BenchHTTP(param BenchParam) BenchStat {
	client := getClient()
	stat := BenchStat{
		BenchParam: param,
	}
	return benchHTTP(client, stat)
}

func getClient() *fasthttp.Client {
	client := &fasthttp.Client{
		ReadTimeout:                   500 * time.Millisecond,
		WriteTimeout:                  500 * time.Millisecond,
		MaxIdleConnDuration:           1 * time.Hour,
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      200,
			DNSCacheDuration: 1 * time.Hour,
		}).Dial,
	}
	return client
}

func benchHTTP(client *fasthttp.Client, stat BenchStat) BenchStat {
	startTime := time.Now()
	for {
		url := substitute(stat.URL)
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod(stat.Method)
		resp := fasthttp.AcquireResponse()

		startTimeReq := time.Now()
		err := client.Do(req, resp)
		stat.Latency += time.Since(startTimeReq)

		fasthttp.ReleaseRequest(req)
		code := resp.StatusCode()
		bodyLen := len(resp.Body())
		fasthttp.ReleaseResponse(resp)

		if err == nil {
			if code >= 200 && code <= 399 {
				stat.GoodCnt++
				stat.BodySize += bodyLen
			} else {
				stat.BadCnt++
			}
			if stat.Verbose {
				//fmt.Printf("DEBUG url: %s\tcode: %d\tbody: %s\n", url, code, body)
			}
		} else {
			stat.ErrorCnt++
			_, _ = fmt.Fprintf(os.Stderr, "ERR Connection error: %v\n", err)
		}

		if time.Since(startTime) >= stat.Duration {
			break
		}
	}

	stat.RPS = int(float64(stat.GoodCnt+stat.BadCnt) / stat.Duration.Seconds())
	stat.Latency = time.Duration(stat.Latency.Nanoseconds() / int64(stat.GoodCnt+stat.BadCnt))
	return stat
}
