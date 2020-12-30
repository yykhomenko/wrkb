package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	arg := strings.Split("wrk -t1 -c1 -d1s --latency http://127.0.0.1", " ")
	cmd := exec.Command(arg[0], arg[1:]...)
	b, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	p := split(string(b))
	fmt.Println()
	fmt.Println(p)
	fmt.Println(p[11][1])
	fmt.Println(p[3][1])
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
