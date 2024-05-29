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
}

type BenchStat struct {
	BenchParam
	GoodCnt  int
	BadCnt   int
	ErrorCnt int
	RPS      int
	Latency  time.Duration
}

func BenchHTTP(param BenchParam) BenchStat {

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

	stat := BenchStat{
		BenchParam: param,
	}

	startTime := time.Now()
	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(substitute(param.URL))
		req.Header.SetMethod(param.Method)
		resp := fasthttp.AcquireResponse()

		//println("=================================")
		//println(param.URL)
		//println(substitute(param.URL))
		//println("=================================")

		startTimeReq := time.Now()
		err := client.Do(req, resp)
		stat.Latency += time.Since(startTimeReq)

		fasthttp.ReleaseRequest(req)
		code := resp.StatusCode()
		fasthttp.ReleaseResponse(resp)

		if err == nil {
			if code >= 200 && code <= 399 {
				stat.GoodCnt++
			} else {
				stat.BadCnt++
			}
			//fmt.Printf("DEBUG Code: %d, Response: %s\n", code, resp.Body())
		} else {
			stat.ErrorCnt++
			_, _ = fmt.Fprintf(os.Stderr, "ERR Connection error: %v\n", err)
		}

		if time.Since(startTime) >= param.Duration {
			break
		}
	}

	stat.RPS = int(float64(stat.GoodCnt+stat.BadCnt) / param.Duration.Seconds())
	stat.Latency = time.Duration(stat.Latency.Nanoseconds() / int64(stat.GoodCnt+stat.BadCnt))
	return stat
}
