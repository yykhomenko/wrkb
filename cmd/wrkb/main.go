package main

import (
	"os"
	"wrkb/pkg/wrkb"
)

func main() {
	conns := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 32, 64}
	procName := os.Args[1]
	url := os.Args[2]
	wrkb.Start(conns, procName, url)
}
