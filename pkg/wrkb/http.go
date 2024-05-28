package wrkb

import (
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"time"
)

func BenchHttp(connNum int, url string) BenchStat {

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

	resp, err := http.Get(url)
	if err != nil {
		log.Println("error: {}", err)
	}

	d, err := time.ParseDuration("1s")

	if resp.StatusCode == 200 {
		return BenchStat{
			ConnNum: connNum,
			RPS:     1,
			Latency: d,
		}
	}

	return BenchStat{}
}
