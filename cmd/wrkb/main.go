package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yykhomenko/wrkb/pkg/wrkb"

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
				Usage:    "Process name to benchmark (e.g. hashes, json, upload).",
				Required: false,
			},
			&cli.StringFlag{
				Name:    "c",
				Aliases: []string{"conns"},
				Usage:   "Comma-separated list of connections, e.g. 1,2,4,8,16",
				Value:   "1,2,4,8,16,32,64,128,256",
			},
			&cli.IntFlag{
				Name:    "t",
				Aliases: []string{"time"},
				Usage:   "Duration of test in seconds",
				Value:   1,
			},
			&cli.IntFlag{
				Name:    "n",
				Aliases: []string{"requests"},
				Usage:   "Total number of requests to send (0 = unlimited)",
				Value:   0,
			},
			&cli.Float64Flag{
				Name:    "rps",
				Aliases: []string{"rate"},
				Usage:   "Limit total requests per second across all connections (0 = unlimited)",
				Value:   0,
			},
			&cli.StringFlag{
				Name:    "X",
				Aliases: []string{"method"},
				Usage:   "HTTP method (GET, POST, etc.)",
				Value:   "GET",
			},
			&cli.StringSliceFlag{
				Name:    "H",
				Aliases: []string{"header"},
				Usage:   "Custom HTTP header(s), e.g. -H 'Authorization: Bearer XXX' -H 'Content-Type: application/json'",
			},
			&cli.StringFlag{
				Name:    "d",
				Aliases: []string{"data"},
				Usage:   "Request body to send with POST/PUT/PATCH requests (e.g. JSON string)",
			},
			&cli.BoolFlag{
				Name:    "v",
				Aliases: []string{"verbose"},
				Usage:   "Enable verbose output",
			},
			&cli.StringFlag{
				Name:  "best-json",
				Usage: "Write best benchmark result to JSON file (empty = stdout, use --best-json=)",
			},
			&cli.BoolFlag{
				Name:  "compare",
				Usage: "Compare best-json against existing file and write -2.json + -compare.csv",
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
			method := strings.ToUpper(c.String("X"))
			verbose := c.Bool("v")
			rpsLimit := c.Float64("rps")
			maxReqs := c.Int("n")
			body := c.String("d")
			headers := c.StringSlice("H")
			bestJSONPath := c.String("best-json")
			writeBestJSON := c.IsSet("best-json")
			compareBestJSON := c.Bool("compare")
			jsonOnly := writeBestJSON && bestJSONPath == ""

			if !jsonOnly {
				fmt.Printf("\n⚙️  Preparing benchmark: '%s' [%s] for %s\n", procName, method, url)
				fmt.Printf("   Connections: %v | Duration: %v | Requests: %d | Verbose: %v\n", conns, duration, maxReqs, verbose)
			}

			var params []wrkb.BenchParam
			for _, connNum := range conns {
				params = append(params, wrkb.BenchParam{
					ProcName:        procName,
					ConnNum:         connNum,
					URL:             url,
					Method:          method,
					Duration:        duration,
					Verbose:         verbose,
					RPSLimit:        rpsLimit,
					MaxReqs:         maxReqs,
					Body:            body,
					Headers:         headers,
					BestJSONPath:    bestJSONPath,
					WriteBestJSON:   writeBestJSON,
					CompareBestJSON: compareBestJSON,
				})
			}

			return wrkb.Start(params)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
