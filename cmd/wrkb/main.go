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

			fmt.Println("✅ All benchmarks finished!")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

//func profileCPU(filename string, fn func()) {
//	f, err := os.Create(filename)
//	if err != nil {
//		log.Fatal("could not create CPU profile: ", err)
//	}
//	defer f.Close()
//
//	if err := pprof.StartCPUProfile(f); err != nil {
//		log.Fatal("could not start CPU profile: ", err)
//	}
//	defer pprof.StopCPUProfile()
//
//	fn()
//}
//
//func profileRun(cpuFile, memFile string, fn func()) {
//	// 🧠 CPU профайл
//	cpu, err := os.Create(cpuFile)
//	if err != nil {
//		log.Fatalf("could not create CPU profile: %v", err)
//	}
//	if err := pprof.StartCPUProfile(cpu); err != nil {
//		log.Fatalf("could not start CPU profile: %v", err)
//	}
//	log.Printf("⚙️ CPU profiling started -> %s", cpuFile)
//
//	// Виконуємо твою функцію
//	fn()
//
//	// 🧠 Зупиняємо CPU профілювання
//	pprof.StopCPUProfile()
//	cpu.Close()
//	log.Println("✅ CPU profiling stopped")
//
//	// 🧩 Збираємо профіль пам'яті
//	runtime.GC() // зібрати garbage перед заміром
//	mem, err := os.Create(memFile)
//	if err != nil {
//		log.Fatalf("could not create memory profile: %v", err)
//	}
//	if err := pprof.WriteHeapProfile(mem); err != nil {
//		log.Fatalf("could not write memory profile: %v", err)
//	}
//	mem.Close()
//	log.Printf("💾 Memory profile saved -> %s", memFile)
//}

//package main
//
//import (
//	"image/color"
//	"log"
//
//	"gonum.org/v1/plot"
//	"gonum.org/v1/plot/plotter"
//	"gonum.org/v1/plot/vg"
//)
//
//func main() {
//	// Дані з твоєї таблиці
//	conn := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}
//	rps := []float64{38085, 66863, 60701, 65650, 69237, 73739, 72860, 71825, 74307, 74740, 75898, 77554, 80459, 81651, 83862, 87634, 86925, 87747, 92097, 95458, 96785, 94965, 99730, 99951}
//	latency := []float64{26.256, 29.911, 49.422, 60.928, 72.214, 81.366, 96.073, 111.38, 121.118, 133.795, 144.93, 154.73, 161.571, 171.46, 178.865, 182.576, 195.568, 205.134, 206.304, 209.514, 216.974, 231.663, 230.621, 240.115}
//
//	p := plot.New()
//	p.Title.Text = "Benchmark scaling"
//	p.X.Label.Text = "Connections"
//	p.Y.Label.Text = "RPS / Latency (µs)"
//
//	// RPS
//	rpsPoints := make(plotter.XYs, len(conn))
//	for i := range conn {
//		rpsPoints[i].X = conn[i]
//		rpsPoints[i].Y = rps[i] / 1000 // масштабування, щоб не затьмарило latency
//	}
//
//	// Latency
//	latencyPoints := make(plotter.XYs, len(conn))
//	for i := range conn {
//		latencyPoints[i].X = conn[i]
//		latencyPoints[i].Y = latency[i]
//	}
//
//	// Створюємо лінії
//	rpsLine, err := plotter.NewLine(rpsPoints)
//	if err != nil {
//		log.Fatal(err)
//	}
//	rpsLine.Color = color.RGBA{R: 30, G: 144, B: 255, A: 255} // синій
//
//	latLine, err := plotter.NewLine(latencyPoints)
//	if err != nil {
//		log.Fatal(err)
//	}
//	latLine.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255} // червоний
//
//	p.Add(rpsLine, latLine)
//	p.Legend.Add("RPS (x1000)", rpsLine)
//	p.Legend.Add("Latency (µs)", latLine)
//	p.Legend.Top = true
//
//	if err := p.Save(8*vg.Inch, 4*vg.Inch, "benchmark.png"); err != nil {
//		log.Fatal(err)
//	}
//}
