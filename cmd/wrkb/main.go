package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"

	"wrkb/internal/wrkb"
)

func main() {
	procName := flag.String("name", "main", "process name")
	flag.Parse()
	link := flag.Arg(0)
	conns := []int{1, 2, 4, 8, 16, 32}

	ps, err := wrkb.Ps(*procName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Process %q starts with:\ncpu %f\nthreads %d\nmem %d\ndisk %d\n\n",
		*procName, ps.CPUTime, ps.CPUNumThreads, ps.MemRSS, ps.BinarySize)
	fmt.Printf("%3s|%7s|%8s|%4s|%3s|%s\n", "num", "rps", "latency", "cpu", "thr", "rss")
	fmt.Printf("----------------------------------------\n")
	results := wrkb.RunBench(conns, link, *procName)
	result := findBestBench(results)
	fmt.Printf("\nBest:\n%s\n", result.String())
}

func findBestBench(stats []wrkb.BenchStat) wrkb.BenchStat {
	sort.Slice(stats, func(i, j int) bool {
		w1 := float64(stats[i].RPS) / math.Log10(float64(stats[i].Latency.Nanoseconds()))
		w2 := float64(stats[j].RPS) / math.Log10(float64(stats[j].Latency.Nanoseconds()))
		return w1 > w2
	})
	return stats[0]
}
