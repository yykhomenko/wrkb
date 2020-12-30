package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type BenchStat struct {
	ConnNum       int
	RPS           int
	Latency       time.Duration
	CpuTime       float64
	CpuNumThreads int
	MemRSS        int
}

func (s *BenchStat) String() string {
	return fmt.Sprintf("%3d|%7d|%8s|%4.2f|%3d|%d",
		s.ConnNum, s.RPS, s.Latency, s.CpuTime, s.CpuNumThreads, s.MemRSS)
}

func RunBench(conns []int, link string, procName string) (out []BenchStat) {
	for _, c := range conns {
		stat := benchStat(c, link, procName)
		out = append(out, stat)
		fmt.Println(stat.String())
	}
	return
}

func benchStat(c int, link, procName string) BenchStat {
	pss, err := psStat(procName)
	if err != nil {
		log.Fatal(err)
	}

	args := strings.Split(command(c, link), " ")
	cmd := exec.Command(args[0], args[1:]...)
	b, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	wrk := wrkStat(b)
	psf, err := psStat(procName)
	if err != nil {
		log.Fatal(err)
	}

	return BenchStat{
		ConnNum:       c,
		RPS:           wrk.RPS,
		Latency:       wrk.Latency,
		CpuTime:       psf.CpuTime - pss.CpuTime,
		CpuNumThreads: psf.CpuNumThreads,
		MemRSS:        psf.MemRSS,
	}
}

func command(c int, link string) string {
	return fmt.Sprintf("wrk -t1 -c%d -d1s --latency %s", c, link)
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
