package wrkb

import (
	"fmt"
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
		fmt.Printf("\nProcess %q starts with:\ncpu(s)  %f\nthreads %d\nmem(b)  %d\ndisk(b) %d\n\n",
			procName, ps.CPUTime, ps.CPUNumThreads, ps.MemRSS, ps.BinarySize)
	}

	fmt.Printf("┌────┬───────┬─────────┬─────┬────┬────────────┐\n")
	fmt.Printf("│%4s│%7s│%9s│%5s│%4s│%12s│\n", "conn", "rps", "latency", "cpu", "thr", "rss")
	fmt.Printf("├────┼───────┼─────────┼─────┼────┼────────────┤\n")

	var stats []BenchStat
	for _, c := range conns {

		var pss *PsStat
		if procName != "" {
			ps, err := Ps(procName)
			if err != nil {
				log.Fatal(err)
			}
			pss = ps
		}

		stat := Bench(c, url)
		stats = append(stats, stat)
		fmt.Printf("│%4d│%7d│%9s│", stat.ConnNum, stat.RPS, stat.Latency)

		if procName != "" {
			psf, err := Ps(procName)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%5.2f│%4d│%12d│\n", psf.CPUTime-pss.CPUTime, psf.CPUNumThreads, psf.MemRSS)
		} else {
			fmt.Printf("%5.2s│%4s│%12s│\n", "", "", "")
		}
	}
	fmt.Printf("└────┴───────┴─────────┴─────┴────┴────────────┘\n")

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

// "cpu", "thr", "rss"
// ps, err := Ps(procName)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Process %q starts with:\ncpu(s)  %f\nthreads %d\nmem(b)  %d\ndisk(b) %d\n\n",
//		procName, ps.CPUTime, ps.CPUNumThreads, ps.MemRSS, ps.BinarySize)
//	//CPUTime       float64
//	//CPUNumThreads int
//	//MemRSS        int
