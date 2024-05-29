package wrkb

//
//import (
//	"fmt"
//	"log"
//	"os/exec"
//	"strconv"
//	"strings"
//	"time"
//)
//
//func BenchWRK(connNum int, url string) BenchStat {
//	args := strings.Split(fmt.Sprintf("wrk -t1 -c%d -d1s --latency %s", connNum, url), " ")
//	cmd := exec.Command(args[0], args[1:]...)
//
//	out, err := cmd.Output()
//	if err != nil {
//		log.Println("process wrk not response")
//		log.Fatal(string(out))
//	}
//
//	p := splitOut(string(out))
//	rps, err := parseRPS(p[4][1])
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	latency, err := time.ParseDuration(p[3][1])
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return BenchStat{
//		BenchParam: BenchParam{
//			ConnNum: connNum,
//		},
//		RPS:     rps,
//		Latency: latency,
//	}
//}
//
//func splitOut(in string) (out [][]string) {
//	rows := strings.Split(in, "\n")
//	for _, row := range rows {
//		var cs []string
//		for _, col := range strings.Split(row, " ") {
//			c := strings.TrimSpace(col)
//			if c != "" {
//				cs = append(cs, c)
//			}
//		}
//		if len(cs) > 0 {
//			out = append(out, cs)
//		}
//	}
//	return
//}
//
//func parseRPS(str string) (int, error) {
//	switch {
//	case strings.HasSuffix(str, "k"):
//		const kilo = 1000
//		tps, err := strconv.ParseFloat(strings.TrimSuffix(str, "k"), 64)
//		if err != nil {
//			return 0, err
//		}
//		return int(tps * kilo), nil
//	default:
//		tps, err := strconv.Atoi(str)
//		if err != nil {
//			return 0, err
//		}
//		return tps, nil
//	}
//}
