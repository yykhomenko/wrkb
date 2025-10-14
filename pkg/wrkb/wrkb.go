package wrkb

import (
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/dustin/go-humanize"
)

func Start(params []BenchParam) {

	if params[0].ProcName != "" {
		ps, err := Ps(params[0].ProcName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\nProcess %q starts with:\ncpu: %f\nthreads: %d\nmem: %s\ndisk: %s\n\n",
			params[0].ProcName, ps.CPUTime, ps.CPUNumThreads, humanize.Bytes(uint64(ps.MemRSS)), humanize.Bytes(uint64(ps.BinarySize)))
	}

	fmt.Printf("┌────┬──────┬──────────┬────────┬────────┬────────┬───────┬─────┬────┬───────┐\n")
	fmt.Printf("│%4s│%6s│%10s│%8s|%8s|%8s|%7s|%5s│%4s│%7s│\n", "conn", "rps", "latency", "good", "bad", "err", "body", "cpu", "thr", "mem")
	fmt.Printf("├────┼──────┼──────────┼────────┼────────┼────────┼───────┼─────┼────┼───────┤\n")

	var results []BenchResult
	for _, p := range params {

		var pss *PsStat
		if p.ProcName != "" {
			ps, err := Ps(p.ProcName)
			if err != nil {
				log.Fatal(err)
			}
			pss = ps
		}

		//result := BenchWRK(connNum, url)

		result := BenchHTTP(p)
		results = append(results, result)

		if p.ProcName != "" {
			psf, err := Ps(p.ProcName)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("│%4d│%6d│%10s│%8d|%8d|%8d|%7.7s|%5.2f│%4d│%7.7s│\n",
				result.Param.ConnNum, result.RPS, result.Latency,
				result.Stat.GoodCnt, result.Stat.BadCnt, result.Stat.ErrorCnt, humanize.Bytes(uint64(result.Stat.BodySize)),
				(psf.CPUTime-pss.CPUTime)/result.Param.Duration.Seconds(), psf.CPUNumThreads, humanize.Bytes(uint64(psf.MemRSS)),
			)
		} else {
			fmt.Printf("│%4d│%6d│%10s│%8d|%8d|%8d|%7.7s|%5.2s│%4s│%7.7s│\n",
				result.Param.ConnNum, result.RPS, result.Latency,
				result.Stat.GoodCnt, result.Stat.BadCnt, result.Stat.ErrorCnt, humanize.Bytes(uint64(result.Stat.BodySize)),
				"", "", "",
			)
		}
	}
	fmt.Printf("└────┴──────┴──────────┴────────┴────────┴────────┴───────┴─────┴────┴───────┘\n")

	bestResult := findBestResult(results)
	fmt.Printf("\nBest: %d, rps: %d, latency: %s\n",
		bestResult.Param.ConnNum, bestResult.RPS, bestResult.Latency,
	)
}

func findBestResult(stats []BenchResult) BenchResult {
	sort.Slice(stats, func(i, j int) bool {
		w1 := float64(stats[i].RPS) / math.Log10(float64(stats[i].Latency.Nanoseconds()))
		w2 := float64(stats[j].RPS) / math.Log10(float64(stats[j].Latency.Nanoseconds()))
		return w1 > w2
	})
	return stats[0]
}
