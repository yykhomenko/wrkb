package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	load(1)
	load(10)
}

func load(c int) {
	args := strings.Split(command(c), " ")
	cmd := exec.Command(args[0], args[1:]...)
	b, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(string(b))

	p := split(string(b))
	// fmt.Println()
	// fmt.Println(p)
	fmt.Println(p[11][1], p[3][1])
}

func command(c int) string {
	return fmt.Sprintf("wrk -t1 -c%d -d1s --latency http://127.0.0.1", c)
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
