package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	p := strings.Split("wrk -t1 -c1 -d1s --latency http://127.0.0.1", " ")
	cmd := exec.Command(p[0], p[1:]...)
	b, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	lines := strings.Split(string(b), "\n")

	fmt.Println(lines[11])
}
