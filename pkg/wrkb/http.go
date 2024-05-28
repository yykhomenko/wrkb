package wrkb

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"math/rand"
	"os"
	"time"
)

type BenchParam struct {
	ConnNum  int
	URL      string
	Duration time.Duration
}

type BenchStat struct {
	ConnNum  int
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
		ConnNum: param.ConnNum,
	}

	low := int64(380670000001)
	high := int64(380679999999)

	startTime := time.Now()

	for {
		msisdn := low + rand.Int63n(high-low+1)
		req := fasthttp.AcquireRequest()

		//req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:8080/hashes/%d", msisdn))
		req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:8080/subscribers/%d", msisdn))

		req.Header.SetMethod(fasthttp.MethodGet)
		resp := fasthttp.AcquireResponse()

		startTimeReq := time.Now()
		err := client.Do(req, resp)
		stat.Latency += time.Since(startTimeReq)

		fasthttp.ReleaseRequest(req)
		code := resp.StatusCode()
		fasthttp.ReleaseResponse(resp)

		if err == nil {
			if code >= 200 && code < 399 {
				stat.GoodCnt++
			} else {
				stat.BadCnt++
			}
			//fmt.Printf("DEBUG Response: %s\n", resp.Body())
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
