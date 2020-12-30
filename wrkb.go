package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	procName := *flag.String("name", "main", "process name")
	flag.Parse()
	link := flag.Arg(0)

	conns := []int{6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	// conns := []int{1, 2, 4, 8, 16, 32, 64, 128, 256}

	var stats []Stat
	for _, c := range conns {
		stat := bench(c, link, procName)
		stats = append(stats, stat)
		fmt.Println(stat)
	}

	stat := findBestBench(stats)
	fmt.Println("\nBest:", stat)
}

func findBestBench(stats []Stat) Stat {
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Rate() > stats[j].Rate()
	})
	return stats[0]
}

type Stat struct {
	c       int
	rps     int
	latency time.Duration
	cpuTime float64
}

func (s *Stat) Rate() float64 {
	return float64(s.rps) / math.Log10(float64(s.latency.Nanoseconds()))
}

func bench(c int, link, procName string) Stat {

	cpuStart, err := cpuTime(procName)
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

	cpuFinish, err := cpuTime(procName)
	if err != nil {
		log.Fatal(err)
	}

	return Stat{
		c:       c,
		rps:     rps,
		latency: latency,
		cpuTime: cpuFinish - cpuStart,
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

func cpuTime(procName string) (float64, error) {
	ps, err := process.Processes()
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range ps {
		name, err := p.Name()
		if err != nil {
			return 0, err
		}

		if name == procName {
			t, err := p.Times()
			if err != nil {
				return 0, err
			}
			return t.Total(), nil
		}
	}

	return 0, errors.New("proc not found")
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
