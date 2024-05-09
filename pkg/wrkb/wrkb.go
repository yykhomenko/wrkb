package wrkb

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"math"
	"sort"
)

func Start(conns []int, procName, url string) {

	if procName != "" {
		ps, err := Ps(procName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\nProcess %q starts with:\ncpu: %f\nthreads: %d\nmem: %s\ndisk: %s\n\n",
			procName, ps.CPUTime, ps.CPUNumThreads, humanize.Bytes(uint64(ps.MemRSS)), humanize.Bytes(uint64(ps.BinarySize)))
	}

	fmt.Printf("┌────┬───────┬─────────┬─────┬────┬───────┐\n")
	fmt.Printf("│%4s│%7s│%9s│%5s│%4s│%7s│\n", "conn", "rps", "latency", "cpu", "thr", "mem")
	fmt.Printf("├────┼───────┼─────────┼─────┼────┼───────┤\n")

	var stats []BenchStat
	for _, connNum := range conns {

		var pss *PsStat
		if procName != "" {
			ps, err := Ps(procName)
			if err != nil {
				log.Fatal(err)
			}
			pss = ps
		}

		stat := Bench(connNum, url)
		stats = append(stats, stat)
		fmt.Printf("│%4d│%7d│%9s│", stat.ConnNum, stat.RPS, stat.Latency)

		if procName != "" {
			psf, err := Ps(procName)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%5.2f│%4d│%7.7s│\n", psf.CPUTime-pss.CPUTime, psf.CPUNumThreads, humanize.Bytes(uint64(psf.MemRSS)))
		} else {
			fmt.Printf("%5.2s│%4s│%7.7s│\n", "", "", "")
		}
	}
	fmt.Printf("└────┴───────┴─────────┴─────┴────┴───────┘\n")

	stat := findBestBench(stats)
	fmt.Printf("\nBest: %d, rps: %d, latency: %s\n", stat.ConnNum, stat.RPS, stat.Latency)
}

func findBestBench(stats []BenchStat) BenchStat {
	sort.Slice(stats, func(i, j int) bool {
		w1 := float64(stats[i].RPS) / math.Log10(float64(stats[i].Latency.Nanoseconds()))
		w2 := float64(stats[j].RPS) / math.Log10(float64(stats[j].Latency.Nanoseconds()))
		return w1 > w2
	})
	return stats[0]
}
