package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"
)

func main() {
	procName := *flag.String("name", "main", "process name")
	flag.Parse()
	link := flag.Arg(0)

	// conns := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	conns := []int{4, 8, 16}

	ps, err := psStat(procName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Process %q starts with cpu:%f, threads:%d, mem:%d, disk:%d\n",
		procName, ps.CpuTime, ps.CpuNumThreads, ps.MemRSS, ps.BinarySize)

	results := RunBench(conns, link, procName)
	result := findBestBench(results)
	fmt.Printf("\nBest:\n%s\n", result.String())
}

func findBestBench(stats []BenchStat) BenchStat {
	sort.Slice(stats, func(i, j int) bool {
		w1 := float64(stats[i].RPS) / math.Log10(float64(stats[i].Latency.Nanoseconds()))
		w2 := float64(stats[j].RPS) / math.Log10(float64(stats[j].Latency.Nanoseconds()))
		return w1 > w2
	})
	return stats[0]
}
