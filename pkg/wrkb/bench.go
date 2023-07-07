package wrkb

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
	CPUTime       float64
	CPUNumThreads int
	MemRSS        int
}

func (s *BenchStat) String() string {
	return fmt.Sprintf("│%4d│%7d│%9s│%5.2f│%4d│%12d│",
		s.ConnNum, s.RPS, s.Latency, s.CPUTime, s.CPUNumThreads, s.MemRSS)
}

func BenchAll(conns []int, procName, url string) (out []BenchStat) {
	for _, c := range conns {
		stat := Bench(c, procName, url)
		out = append(out, stat)
		fmt.Println(stat.String())
	}
	return
}

func Bench(c int, procName, url string) BenchStat {
	pss, err := Ps(procName)
	if err != nil {
		log.Fatal(err)
	}

	args := strings.Split(command(c, url), " ")
	cmd := exec.Command(args[0], args[1:]...)
	b, err := cmd.Output()
	if err != nil {
		log.Println("process wrk not response, probably wrong 'URL' parameter")
		log.Fatal(string(b))
	}

	wrk := Wrk(b)
	psf, err := Ps(procName)
	if err != nil {
		log.Fatal(err)
	}

	return BenchStat{
		ConnNum:       c,
		RPS:           wrk.RPS,
		Latency:       wrk.Latency,
		CPUTime:       psf.CPUTime - pss.CPUTime,
		CPUNumThreads: psf.CPUNumThreads,
		MemRSS:        psf.MemRSS,
	}
}

func command(c int, url string) string {
	return fmt.Sprintf("wrk -t1 -c%d -d1s --latency %s", c, url)
}

func parseRPS(s string) (int, error) {
	switch {
	case strings.HasSuffix(s, "k"):
		const kilo = 1000
		tps, err := strconv.ParseFloat(strings.TrimSuffix(s, "k"), 64)
		if err != nil {
			return 0, err
		}
		return int(tps * kilo), nil
	default:
		tps, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		return tps, nil
	}
}