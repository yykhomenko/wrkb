package main

import (
	"os"

	wrkb "wrkb/internal"
)

func main() {
	conns := []int{1, 2, 4, 8, 16, 32, 64}
	procName := os.Args[1]
	link := os.Args[2]
	wrkb.Start(conns, link, procName)
}
