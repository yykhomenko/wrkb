package main

import (
	"os"
	"wrkb/pkg/wrkb"
)

func main() {
	conns := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 16, 32, 64}

	var procName string
	var url string

	if len(os.Args) == 2 {
		procName = ""
		url = os.Args[1]
	} else if len(os.Args) == 3 {
		procName = os.Args[1]
		url = os.Args[2]
	}

	wrkb.Start(conns, procName, url)
}
