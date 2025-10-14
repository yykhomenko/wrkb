package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"wrkb/pkg/wrkb"

	"github.com/urfave/cli/v2"
)

func parseConnections(input string) []int {
	var conns []int
	for _, s := range strings.Split(input, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			log.Fatalf("invalid connection value: %s", s)
		}
		conns = append(conns, n)
	}
	return conns
}

func main() {
	app := &cli.App{
		Name:  "wrkb",
		Usage: "Flexible load testing CLI tool for benchmarking endpoints",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "p",
				Aliases:  []string{"proc"},
				Usage:    "Process name to benchmark (e.g. hashes, json, upload)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "c",
				Usage: "Comma-separated list of connections, e.g. 1,2,4,8,16",
				Value: "1,2,4,8,16,32,64,128",
			},
			&cli.IntFlag{
				Name:    "t",
				Usage:   "Duration of test in seconds",
				Value:   10,
				Aliases: []string{"time"},
			},
			&cli.BoolFlag{
				Name:    "v",
				Aliases: []string{"verbose"},
				Usage:   "Enable verbose output",
			},
			&cli.StringFlag{
				Name:    "m",
				Aliases: []string{"method"},
				Usage:   "HTTP method (GET, POST, etc.)",
				Value:   "GET",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Args().Len() < 1 {
				return cli.Exit("Usage: wrkb -p=<proc> [-c=<list>] [-t=<seconds>] [-v] [-m=<method>] <url>", 1)
			}

			url := c.Args().Get(0)
			procName := c.String("p")
			conns := parseConnections(c.String("c"))
			duration := time.Duration(c.Int("t")) * time.Second
			method := strings.ToUpper(c.String("m"))
			verbose := c.Bool("v")

			fmt.Printf("⚙️ Preparing benchmark: '%s' [%s] for %s\n", procName, method, url)
			fmt.Printf("   Connections: %v | Duration: %v | Verbose: %v\n", conns, duration, verbose)

			var params []wrkb.BenchParam
			for _, connNum := range conns {
				params = append(params, wrkb.BenchParam{
					ProcName: procName,
					ConnNum:  connNum,
					URL:      url,
					Method:   method,
					Duration: duration,
					Verbose:  verbose,
				})
			}

			wrkb.Start(params)
			//profileRun("cpu.prof", "mem.prof", func() {
			//	wrkb.Start(params)
			//})

			fmt.Println("✅ All benchmarks finished!")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
