package main

import (
	"errors"

	"github.com/shirou/gopsutil/v3/process"
)

type PsStat struct {
	CpuTime       float64
	CpuNumThreads int
	MemRSS        int
}

func psStat(procName string) (stat *PsStat, err error) {
	ps, err := process.Processes()
	if err != nil {
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
				return nil, err
			}

			cpuNumThreads, err := p.NumThreads()
			if err != nil {
				return nil, err
			}

			mem, err := p.MemoryInfo()
			if err != nil {
				return nil, err
			}

			return &PsStat{
				CpuTime:       cpuTime.Total(),
				CpuNumThreads: int(cpuNumThreads),
				MemRSS:        int(mem.RSS),
			}, nil
		}
	}

	return nil, errors.New("proc not found")
}
