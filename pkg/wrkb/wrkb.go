package wrkb

import (
	"fmt"
	"log"
	"math"
	"sort"
)

func Start(conns []int, procName, url string) {
	ps, err := Ps(procName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Process %q starts with:\ncpu(s)  %f\nthreads %d\nmem(b)  %d\ndisk(b) %d\n\n",
		procName, ps.CPUTime, ps.CPUNumThreads, ps.MemRSS, ps.BinarySize)

	fmt.Printf("┌────┬───────┬─────────┬─────┬────┬────────────┐\n")
	fmt.Printf("│%4s│%7s│%9s│%5s│%4s│%12s│\n", "conn", "rps", "latency", "cpu", "thr", "rss")
	fmt.Printf("├────┼───────┼─────────┼─────┼────┼────────────┤\n")

	stats := BenchAll(conns, procName, url)
	stat := findBestBench(stats)
	fmt.Printf("└────┴───────┴─────────┴─────┴────┴────────────┘\n")
	fmt.Printf("\nBest:\n%s\n", stat.String())
}

func findBestBench(stats []BenchStat) BenchStat {
	sort.Slice(stats, func(i, j int) bool {
		w1 := float64(stats[i].RPS) / math.Log10(float64(stats[i].Latency.Nanoseconds()))
		w2 := float64(stats[j].RPS) / math.Log10(float64(stats[j].Latency.Nanoseconds()))
		return w1 > w2
	})
	return stats[0]
}
