package wrkb

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"os"
	"sync"
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
	GoodCnt  int
	BadCnt   int
	ErrorCnt int
	BodySize int
	Time     time.Duration
}

func (s BenchStat) Add(other BenchStat) BenchStat {
	s.GoodCnt += other.GoodCnt
	s.BadCnt += other.BadCnt
	s.ErrorCnt += other.ErrorCnt
	s.BodySize += other.BodySize
	s.Time += other.Time
	return s
}

type BenchResult struct {
	Param   BenchParam
	Stat    BenchStat
	RPS     int
	Latency time.Duration
}

func (s BenchResult) CalcStat() BenchResult {
	s.RPS = int(float64(s.Stat.GoodCnt+s.Stat.BadCnt) / (s.Stat.Time.Seconds() / float64(s.Param.ConnNum)))
	s.Latency = time.Duration(s.Stat.Time.Nanoseconds() / int64(s.Stat.GoodCnt+s.Stat.BadCnt))
	return s
}

func BenchHTTP(param BenchParam) BenchResult {
	client := getClient(param)

	wg := &sync.WaitGroup{}
	stats := make(chan BenchStat, param.ConnNum)

	for i := 1; i <= param.ConnNum; i++ {
		wg.Add(1)
		go func(i int, results chan<- BenchStat, wg *sync.WaitGroup) {
			defer wg.Done()
			stats <- benchHTTP(client, param)
		}(i, stats, wg)
	}

	wg.Wait()
	close(stats)

	stat := BenchStat{}
	for s := range stats {
		//fmt.Printf("%+v\n", s)
		stat = stat.Add(s)
	}

	return (BenchResult{
		Param: param,
		Stat:  stat,
	}).CalcStat()
}

func getClient(param BenchParam) *fasthttp.Client {
	client := &fasthttp.Client{
		ReadTimeout:                   500 * time.Millisecond,
		WriteTimeout:                  500 * time.Millisecond,
		MaxIdleConnDuration:           1 * time.Hour,
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      param.ConnNum,
			DNSCacheDuration: 1 * time.Hour,
		}).Dial,
	}
	return client
}

func benchHTTP(client *fasthttp.Client, param BenchParam) BenchStat {
	stat := BenchStat{}
	startTime := time.Now()
	for {
		url := substitute(param.URL)
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod(param.Method)
		resp := fasthttp.AcquireResponse()

		startTimeReq := time.Now()
		err := client.Do(req, resp)
		stat.Time += time.Since(startTimeReq)

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
			//if param.Verbose {
			//fmt.Printf("DEBUG url: %s\tcode: %d\n", url, code)
			//}
		} else {
			stat.ErrorCnt++
			_, _ = fmt.Fprintf(os.Stderr, "ERR Connection error: %v\n", err)
		}

		if time.Since(startTime) >= param.Duration {
			break
		}
	}
	return stat
}
