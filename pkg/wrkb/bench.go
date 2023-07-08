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
	ConnNum int
	RPS     int
	Latency time.Duration
}

func Bench(c int, url string) BenchStat {
	args := strings.Split(command(c, url), " ")
	cmd := exec.Command(args[0], args[1:]...)
	b, err := cmd.Output()
	if err != nil {
		log.Println("process wrk not response")
		log.Fatal(string(b))
	}

	wrk := Wrk(b)
	return BenchStat{
		ConnNum: c,
		RPS:     wrk.RPS,
		Latency: wrk.Latency,
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
