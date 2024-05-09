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

func Bench(connNum int, url string) BenchStat {
	args := strings.Split(command(connNum, url), " ")
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.Output()
	if err != nil {
		log.Println("process wrk not response")
		log.Fatal(string(out))
	}

	wrk := Wrk(out)
	return BenchStat{
		ConnNum: connNum,
		RPS:     wrk.RPS,
		Latency: wrk.Latency,
	}
}

func command(connNum int, url string) string {
	return fmt.Sprintf("wrk -t1 -c%d -d1s --latency %s", connNum, url)
}

func parseRPS(str string) (int, error) {
	switch {
	case strings.HasSuffix(str, "k"):
		const kilo = 1000
		tps, err := strconv.ParseFloat(strings.TrimSuffix(str, "k"), 64)
		if err != nil {
			return 0, err
		}
		return int(tps * kilo), nil
	default:
		tps, err := strconv.Atoi(str)
		if err != nil {
			return 0, err
		}
		return tps, nil
	}
}
