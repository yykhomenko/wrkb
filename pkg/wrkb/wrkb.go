package wrkb

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
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

func Start(params []BenchParam) error {
	if len(params) == 0 {
		return fmt.Errorf("no benchmark parameters provided")
	}

	var results []BenchResult
	jsonOnly := params[0].WriteBestJSON && params[0].BestJSONPath == ""

	if !jsonOnly && params[0].ProcName != "" {
		ps, err := Ps(params[0].ProcName)
		if err != nil {
			return err
		}
		fmt.Printf("\n%s⚙️  Process:%s %s\n", cyan, reset, params[0].ProcName)
		fmt.Printf("%s   CPU:%s %.2fs | %sThreads:%s %d | %sMem:%s %s | %sDisk:%s %s\n\n",
			gray, reset, ps.CPUTime,
			gray, reset, ps.CPUNumThreads,
			gray, reset, humanize.Bytes(uint64(ps.MemRSS)),
			gray, reset, humanize.Bytes(uint64(ps.BinarySize)))
	}

	if !jsonOnly {
		printHeader()
	}

	for _, p := range params {
		results = append(results, runSingleBenchmark(p, !jsonOnly))
	}

	if !jsonOnly {
		printFooter()
	}

	best := findBestResult(results)

	if !jsonOnly {
		icon := randomStartIcon()

		fmt.Printf("\n%s %s Best result:%s %d connections | %s%d RPS%s | %v%s latency%s \nmin=%-8v \np50=%-8v \np90=%-8v \np99=%-8v \np999=%-8v \nmax=%-8v\n\n",
			icon,
			yellow, reset, best.Param.ConnNum,
			green, best.RPS, reset,
			red, best.Latency, reset,
			best.Min, best.P50, best.P90, best.P99, best.P999, best.Max,
		)
	}

	if params[0].WriteBestJSON {
		rows, err := writeBestResultJSON(best, params[0].BestJSONPath, params[0].CompareBestJSON)
		if err != nil {
			return err
		}
		if len(rows) > 0 && !jsonOnly {
			printCompareTable(rows)
		}
	}

	return nil
}

func runSingleBenchmark(p BenchParam, showOutput bool) BenchResult {
	var psBefore *PsStat
	if p.ProcName != "" {
		ps, err := Ps(p.ProcName)
		if err != nil {
			log.Printf("failed to read process stats before benchmark: %v", err)
		} else {
			psBefore = ps
		}
	}

	start := time.Now()
	result := BenchHTTP(p)
	elapsed := time.Since(start)

	var cpu float64
	var thr int
	var mem int64

	if p.ProcName != "" {
		psAfter, err := Ps(p.ProcName)
		if err != nil {
			log.Printf("failed to read process stats after benchmark: %v", err)
		} else {
			if psBefore != nil {
				cpu = (psAfter.CPUTime - psBefore.CPUTime) / elapsed.Seconds()
			}
			thr = psAfter.CPUNumThreads
			mem = int64(psAfter.MemRSS)
		}
	}

	if showOutput {
		printRow(result, cpu, thr, mem)
	}
	return result
}

func printHeader() {
	fmt.Printf("\n%s┌────┬────────┬────────────┬────────┬────────┬────────┬─────────┬─────────┬─────┬────┬────────┐%s\n", gray, reset)
	fmt.Printf("%s│%4s│%8s│%12s│%8s│%8s│%8s│%9s│%9s│%5s│%4s│%8s│%s\n",
		gray, "conn", "rps", "latency", "good", "bad", "err", "body req", "body resp", "cpu", "thr", "mem", reset)
	fmt.Printf("%s├────┼────────┼────────────┼────────┼────────┼────────┼─────────┼─────────┼─────┼────┼────────┤%s\n", gray, reset)
}

func printRow(result BenchResult, cpu float64, threads int, memRSS int64) {
	bodyReqSize := humanize.Bytes(uint64(result.Stat.BodyReqSize))
	bodyRespSize := humanize.Bytes(uint64(result.Stat.BodyRespSize))
	fmt.Printf("│%4d│%s%8d%s│%s%12s%s│%8d│%8d│%8d│%9s│%9s│%s%5.2f%s│%4d│%8s│\n",
		result.Param.ConnNum,
		green, result.RPS, reset,
		red, result.Latency, reset,
		result.Stat.GoodCnt, result.Stat.BadCnt, result.Stat.ErrorCnt,
		bodyReqSize,
		bodyRespSize,
		yellow, cpu, reset,
		threads,
		humanize.Bytes(uint64(memRSS)),
	)
}

func printFooter() {
	fmt.Printf("%s└────┴────────┴────────────┴────────┴────────┴────────┴─────────┴─────────┴─────┴────┴────────┘%s\n", gray, reset)
}

func randomStartIcon() string {
	icons := []string{"✨", "🌟", "💫", "⚡️", "🚀", "🔥", "🏅", "💎"}
	rand.Seed(time.Now().UnixNano())
	return icons[rand.Intn(len(icons))]
}

func findBestResult(stats []BenchResult) BenchResult {
	sort.Slice(stats, func(i, j int) bool {
		w1 := float64(stats[i].RPS) / math.Log10(float64(stats[i].Latency.Nanoseconds()))
		w2 := float64(stats[j].RPS) / math.Log10(float64(stats[j].Latency.Nanoseconds()))
		return w1 > w2
	})
	return stats[0]
}

type bestResultJSON struct {
	ProcName      string  `json:"proc_name,omitempty" cmp:"proc_name"`
	URL           string  `json:"url" cmp:"url"`
	Method        string  `json:"method" cmp:"method"`
	Connections   int     `json:"connections" cmp:"connections"`
	Duration      string  `json:"duration" cmp:"duration" cmpKind:"duration"`
	RPSLimit      float64 `json:"rps_limit,omitempty" cmp:"rps_limit"`
	MaxRequests   int     `json:"max_requests,omitempty" cmp:"max_requests"`
	RPS           int     `json:"rps" cmp:"rps"`
	Latency       string  `json:"latency" cmp:"latency" cmpKind:"duration"`
	Min           string  `json:"min" cmp:"min" cmpKind:"duration"`
	P50           string  `json:"p50" cmp:"p50" cmpKind:"duration"`
	P90           string  `json:"p90" cmp:"p90" cmpKind:"duration"`
	P99           string  `json:"p99" cmp:"p99" cmpKind:"duration"`
	P999          string  `json:"p999" cmp:"p999" cmpKind:"duration"`
	Max           string  `json:"max" cmp:"max" cmpKind:"duration"`
	Good          int     `json:"good" cmp:"good"`
	Bad           int     `json:"bad" cmp:"bad"`
	Error         int     `json:"error" cmp:"error"`
	BodyReqBytes  int     `json:"body_req_bytes" cmp:"body_req_bytes"`
	BodyRespBytes int     `json:"body_resp_bytes" cmp:"body_resp_bytes"`
	Time          string  `json:"time" cmp:"time" cmpKind:"duration"`
}

func writeBestResultJSON(best BenchResult, path string, compare bool) ([]compareRow, error) {
	payload := bestResultJSON{
		ProcName:      best.Param.ProcName,
		URL:           best.Param.URL,
		Method:        best.Param.Method,
		Connections:   best.Param.ConnNum,
		Duration:      best.Param.Duration.String(),
		RPSLimit:      best.Param.RPSLimit,
		MaxRequests:   best.Param.MaxReqs,
		RPS:           best.RPS,
		Latency:       best.Latency.String(),
		Min:           best.Min.String(),
		P50:           best.P50.String(),
		P90:           best.P90.String(),
		P99:           best.P99.String(),
		P999:          best.P999.String(),
		Max:           best.Max.String(),
		Good:          best.Stat.GoodCnt,
		Bad:           best.Stat.BadCnt,
		Error:         best.Stat.ErrorCnt,
		BodyReqBytes:  best.Stat.BodyReqSize,
		BodyRespBytes: best.Stat.BodyRespSize,
		Time:          best.Stat.Time.String(),
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	if path == "" {
		_, err := os.Stdout.Write(data)
		return nil, err
	}

	if !compare {
		return nil, os.WriteFile(path, data, 0o644)
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, os.WriteFile(path, data, 0o644)
		}
		return nil, err
	}

	base, err := readBestResultJSON(path)
	if err != nil {
		return nil, err
	}

	nextPath := path + "-2.json"
	if err := os.WriteFile(nextPath, data, 0o644); err != nil {
		return nil, err
	}

	rows := buildCompareRows(base, payload)
	comparePath := path + "-compare.csv"
	if err := writeBestResultCompareCSV(comparePath, rows); err != nil {
		return nil, err
	}

	return rows, nil
}
