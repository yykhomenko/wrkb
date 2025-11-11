package wrkb

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
)

const (
	reset  = "\033[0m"
	green  = "\033[1;32m"
	red    = "\033[1;31m"
	yellow = "\033[1;33m"
	cyan   = "\033[1;36m"
	gray   = "\033[90m"
)

func Start(params []BenchParam) {
	if len(params) == 0 {
		log.Fatal("no benchmark parameters provided")
	}

	var results []BenchResult

	if params[0].ProcName != "" {
		ps, err := Ps(params[0].ProcName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\n%s⚙️  Process:%s %s\n", cyan, reset, params[0].ProcName)
		fmt.Printf("%s   CPU:%s %.2fs | %sThreads:%s %d | %sMem:%s %s | %sDisk:%s %s\n\n",
			gray, reset, ps.CPUTime,
			gray, reset, ps.CPUNumThreads,
			gray, reset, humanize.Bytes(uint64(ps.MemRSS)),
			gray, reset, humanize.Bytes(uint64(ps.BinarySize)))
	}

	printHeader()

	for _, p := range params {
		results = append(results, runSingleBenchmark(p))
	}

	printFooter()

	best := findBestResult(results)

	icon := randomStartIcon()

	fmt.Printf("\n%s %s Best result:%s %d connections | %s%d RPS%s | %s%s latency%s\n\n",
		icon,
		yellow, reset, best.Param.ConnNum,
		green, best.RPS, reset,
		red, best.Latency, reset,
	)
}

func runSingleBenchmark(p BenchParam) BenchResult {
	var psBefore *PsStat
	if p.ProcName != "" {
		ps, _ := Ps(p.ProcName)
		psBefore = ps
	}

	start := time.Now()
	result := BenchHTTP(p)
	elapsed := time.Since(start)

	var cpu float64
	var thr int
	var mem int64

	if p.ProcName != "" && psBefore != nil {
		psAfter, _ := Ps(p.ProcName)
		cpu = (psAfter.CPUTime - psBefore.CPUTime) / elapsed.Seconds()
		thr = psAfter.CPUNumThreads
		mem = int64(psAfter.MemRSS)
	}

	printRow(result, cpu, thr, mem)
	return result
}

func printHeader() {
	fmt.Printf("\n%s┌────┬────────┬────────────┬────────┬────────┬────────┬────────┬─────┬────┬────────┐%s\n", gray, reset)
	fmt.Printf("%s│%4s│%8s│%12s│%8s│%8s│%8s│%8s│%5s│%4s│%8s│%s\n",
		gray, "conn", "rps", "latency", "good", "bad", "err", "body", "cpu", "thr", "mem", reset)
	fmt.Printf("%s├────┼────────┼────────────┼────────┼────────┼────────┼────────┼─────┼────┼────────┤%s\n", gray, reset)
}

func printFooter() {
	fmt.Printf("%s└────┴────────┴────────────┴────────┴────────┴────────┴────────┴─────┴────┴────────┘%s\n", gray, reset)
}

func printRow(result BenchResult, cpu float64, threads int, memRSS int64) {
	bodySize := humanize.Bytes(uint64(result.Stat.BodyRespSize))
	fmt.Printf("│%4d│%s%8d%s│%s%12s%s│%8d│%8d│%8d│%8s│%s%5.2f%s│%4d│%8s│\n",
		result.Param.ConnNum,
		green, result.RPS, reset,
		red, result.Latency, reset,
		result.Stat.GoodCnt, result.Stat.BadCnt, result.Stat.ErrorCnt,
		bodySize,
		yellow, cpu, reset,
		threads,
		humanize.Bytes(uint64(memRSS)),
	)
}

// randomStartIcon — випадкова початкова іконка для "Best result"
func randomStartIcon() string {
	icons := []string{"✨", "🌟", "💫", "⚡️", "🚀", "🔥", "🏅", "💎", "🧠", "🎯"}
	rand.Seed(time.Now().UnixNano())
	return icons[rand.Intn(len(icons))]
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
