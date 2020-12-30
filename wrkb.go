package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	// appCmd := flag.String("run", "", "path to app for lunch")
	flag.Parse()
	link := flag.Arg(0)

	conns := []int{1, 2, 4, 8, 16, 32, 64, 128, 256}
	var stats []Stat
	for _, c := range conns {
		stat := *bench(c, link)
		stats = append(stats, stat)
		fmt.Println(stat)
	}
}

type Stat struct {
	c       int
	rps     int
	latency time.Duration
}

func bench(c int, link string) *Stat {
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

	return &Stat{
		c:       c,
		rps:     rps,
		latency: latency,
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
