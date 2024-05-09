package wrkb

import (
	"log"
	"strings"
	"time"
)

type WrkStat struct {
	RPS     int
	Latency time.Duration
}

func Wrk(out []byte) WrkStat {
	p := split(string(out))
	rps, err := parseRPS(p[4][1])
	if err != nil {
		log.Fatal(err)
	}

	latency, err := time.ParseDuration(p[3][1])
	if err != nil {
		log.Fatal(err)
	}

	return WrkStat{
		RPS:     rps,
		Latency: latency,
	}
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
