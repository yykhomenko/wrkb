package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	procName := *flag.String("name", "main", "process name")
	flag.Parse()
	link := flag.Arg(0)

	conns := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	// conns := []int{1, 2, 4, 8, 16, 32, 64, 128, 256}

	ps, err := stats(procName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Process %q starts with cpu:%f, threads:%d, mem:%d\n",
		procName, ps.CpuTime, ps.CpuThreadNum, ps.MemRSS)

	results := Run(conns, link, procName)
	result := findBestBench(results)
	fmt.Println("\nBest:", result)
}

func Run(conns []int, link string, procName string) (out []Result) {
	for _, c := range conns {
		stat := bench(c, link, procName)
		out = append(out, stat)
		fmt.Println(stat)
	}
	return
}

func findBestBench(stats []Result) Result {
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Rate() > stats[j].Rate()
	})
	return stats[0]
}

type Result struct {
	ConnNum      int
	RPS          int
	Latency      time.Duration
	CpuTime      float64
	CpuThreadNum int
	MemRSS       int
}

func (s *Result) Rate() float64 {
	return float64(s.RPS) / math.Log10(float64(s.Latency.Nanoseconds()))
}

func bench(c int, link, procName string) Result {
	pss, err := stats(procName)
	if err != nil {
		log.Fatal(err)
	}

	args := strings.Split(command(c, link), " ")
	cmd := exec.Command(args[0], args[1:]...)
	b, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	p := split(string(b))

	rps, err := parseRPS(p[4][1])
	if err != nil {
		log.Fatal(err)
	}

	latency, err := time.ParseDuration(p[3][1])
	if err != nil {
		log.Fatal(err)
	}

	psf, err := stats(procName)
	if err != nil {
		log.Fatal(err)
	}

	return Result{
		ConnNum:      c,
		RPS:          rps,
		Latency:      latency,
		CpuTime:      psf.CpuTime - pss.CpuTime,
		CpuThreadNum: psf.CpuThreadNum,
		MemRSS:       psf.MemRSS,
	}
}

func parseRPS(s string) (int, error) {
	switch {
	case strings.HasSuffix(s, "k"):
		tps, err := strconv.ParseFloat(strings.TrimSuffix(s, "k"), 64)
		if err != nil {
			return 0, err
		}
		return int(tps * 1000), nil
	default:
		tps, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		return tps, nil
	}
}

func command(c int, link string) string {
	return fmt.Sprintf("wrk -t1 -c%d -d1s --latency %s", c, link)
}

func split(in string) (out [][]string) {
	rows := strings.Split(in, "\n")
	for _, row := range rows {
		var cs []string
		for _, col := range strings.Split(row, " ") {
			c := strings.TrimSpace(col)
			if c != "" {
				cs = append(cs, c)
			}
		}
		if len(cs) > 0 {
			out = append(out, cs)
		}
	}
	return
}
