package wrkb

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
)

// Start — основна функція запуску бенчмарку
func Start(params []BenchParam) {
	if len(params) == 0 {
		log.Fatal("no benchmark parameters provided")
	}

	var (
		results []BenchResult
		proc    *PsStat
	)

	// Якщо заданий процес — зафіксуємо його початковий стан
	if params[0].ProcName != "" {
		ps, err := Ps(params[0].ProcName)
		if err != nil {
			log.Fatal(err)
		}
		proc = ps
		fmt.Printf(
			"\nProcess %q starts with:\ncpu: %.2f\nthreads: %d\nmem: %s\ndisk: %s\n\n",
			params[0].ProcName, ps.CPUTime, ps.CPUNumThreads,
			humanize.Bytes(uint64(ps.MemRSS)), humanize.Bytes(uint64(ps.BinarySize)),
		)
	}

	printHeader()

	// Основний цикл — по кількості конекшнів
	for _, p := range params {
		results = append(results, runSingleBenchmark(p, proc))
	}

	printFooter()

	// Знаходимо найкращий результат
	best := findBestResult(results)
	fmt.Printf("\n✨ Best: %d connections | %d RPS | %s latency\n",
		best.Param.ConnNum, best.RPS, best.Latency)
}

// runSingleBenchmark — обробка одного параметра
func runSingleBenchmark(p BenchParam, procStart *PsStat) BenchResult {
	var psBefore *PsStat
	if p.ProcName != "" {
		ps, err := Ps(p.ProcName)
		if err != nil {
			log.Fatal(err)
		}
		psBefore = ps
	}

	start := time.Now()
	result := BenchHTTP(p)
	duration := time.Since(start)

	if p.ProcName != "" {
		psAfter, err := Ps(p.ProcName)
		if err != nil {
			log.Fatal(err)
		}
		cpuUsage := (psAfter.CPUTime - psBefore.CPUTime) / duration.Seconds()
		printRow(result, cpuUsage, psAfter.CPUNumThreads, int64(psAfter.MemRSS))
	} else {
		printRow(result, 0, 0, 0)
	}
	return result
}

func printHeader() {
	fmt.Println("┌────┬──────┬──────────┬────────┬────────┬────────┬───────┬─────┬────┬───────┐")
	fmt.Printf("│%4s│%6s│%10s│%8s|%8s|%8s|%7s|%5s│%4s│%7s│\n",
		"conn", "rps", "latency", "good", "bad", "err", "body", "cpu", "thr", "mem")
	fmt.Println("├────┼──────┼──────────┼────────┼────────┼────────┼───────┼─────┼────┼───────┤")
}

func printFooter() {
	fmt.Println("└────┴──────┴──────────┴────────┴────────┴────────┴───────┴─────┴────┴───────┘")
}

func printRow(result BenchResult, cpu float64, threads int, memRSS int64) {
	bodySize := humanize.Bytes(uint64(result.Stat.BodySize))
	if threads > 0 {
		fmt.Printf("│%4d│%6d│%10s│%8d|%8d|%8d|%7.7s|%5.2f│%4d│%7.7s│\n",
			result.Param.ConnNum, result.RPS, result.Latency,
			result.Stat.GoodCnt, result.Stat.BadCnt, result.Stat.ErrorCnt,
			bodySize, cpu, threads, humanize.Bytes(uint64(memRSS)))
	} else {
		fmt.Printf("│%4d│%6d│%10s│%8d|%8d|%8d|%7.7s|%5s│%4s│%7.7s│\n",
			result.Param.ConnNum, result.RPS, result.Latency,
			result.Stat.GoodCnt, result.Stat.BadCnt, result.Stat.ErrorCnt,
			bodySize, "", "", "")
	}
}

// findBestResult — знаходить оптимальний результат по співвідношенню RPS/latency
func findBestResult(stats []BenchResult) BenchResult {
	sort.Slice(stats, func(i, j int) bool {
		w1 := float64(stats[i].RPS) / math.Log10(float64(stats[i].Latency.Nanoseconds()))
		w2 := float64(stats[j].RPS) / math.Log10(float64(stats[j].Latency.Nanoseconds()))
		return w1 > w2
	})
	return stats[0]
}
