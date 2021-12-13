package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"

	wrkb "wrkb/internal"
)

func main() {
	procName := *flag.String("name", "main", "process name")
	flag.Parse()
	link := flag.Arg(0)
	conns := []int{1, 2, 4, 8, 16, 32, 64}

	ps, err := wrkb.Ps(procName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Process %q starts with:\ncpu(s)  %f\nthreads %d\nmem(b)  %d\ndisk(b) %d\n\n",
		procName, ps.CPUTime, ps.CPUNumThreads, ps.MemRSS, ps.BinarySize)
	fmt.Printf("%3s|%7s|%9s|%5s|%4s|%12s\n", "num", "rps", "latency", "cpu", "thr", "rss")
	fmt.Printf("---------------------------------------------\n")
	stats := wrkb.BenchAll(conns, link, procName)
	stat := findBestBench(stats)
	fmt.Printf("\nBest:\n%s\n", stat.String())
}

func findBestBench(stats []wrkb.BenchStat) wrkb.BenchStat {
	sort.Slice(stats, func(i, j int) bool {
		w1 := float64(stats[i].RPS) / math.Log10(float64(stats[i].Latency.Nanoseconds()))
		w2 := float64(stats[j].RPS) / math.Log10(float64(stats[j].Latency.Nanoseconds()))
		return w1 > w2
	})
	return stats[0]
}
