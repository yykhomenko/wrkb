package main

import (
	"errors"
	"log"

	"github.com/shirou/gopsutil/v3/process"
)

type PsStat struct {
	CpuTime      float64
	CpuThreadNum int
	MemRSS       int
}

func psStat(procName string) (stat *PsStat, err error) {
	ps, err := process.Processes()
	if err != nil {
		log.Println("1", err)
		return nil, err
	}

	for _, p := range ps {
		name, err := p.Name()
		if err != nil {
			continue
		}

		if name == procName {
			cpuTime, err := p.Times()
			if err != nil {
				log.Println("3", err)
				return nil, err
			}

			cpuThreadNum, err := p.NumThreads()
			if err != nil {
				log.Println("4", err)
				return nil, err
			}

			mem, err := p.MemoryInfo()
			if err != nil {
				log.Println("5", err)
				return nil, err
			}

			return &PsStat{
				CpuTime:      cpuTime.Total(),
				CpuThreadNum: int(cpuThreadNum),
				MemRSS:       int(mem.RSS),
			}, nil
		}
	}

	return nil, errors.New("proc not found")
}
