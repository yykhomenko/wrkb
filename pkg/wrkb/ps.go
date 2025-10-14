package wrkb

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// ErrProcNotFound повертається, якщо процес з вказаним ім’ям не знайдено.
var ErrProcNotFound = errors.New("process not found")

// PsStat — знімок стану процесу.
type PsStat struct {
	CPUTime       float64 // total CPU time (user+system)
	CPUNumThreads int     // кількість активних потоків
	MemRSS        int64   // фізична пам’ять (Resident Set Size)
	BinarySize    int64   // розмір бінарника
	CPUPercent    float64 // відсоток завантаження CPU (опціонально)
	MemPercent    float32 // відсоток використання пам'яті
}

// Ps — шукає процес за частковою або точною назвою.
func Ps(procName string) (*PsStat, error) {
	if strings.TrimSpace(procName) == "" {
		return nil, fmt.Errorf("empty process name")
	}

	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("cannot list processes: %w", err)
	}

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		// гнучкіше порівняння — допускаємо частковий збіг
		if !strings.EqualFold(name, procName) && !strings.Contains(strings.ToLower(name), strings.ToLower(procName)) {
			continue
		}

		return collectProcStats(p)
	}

	return nil, ErrProcNotFound
}

// collectProcStats — витягує всі потрібні метрики про процес.
func collectProcStats(p *process.Process) (*PsStat, error) {
	cpuTimes, err := p.Times()
	if err != nil {
		return nil, fmt.Errorf("cannot get CPU times: %w", err)
	}

	numThreads, err := p.NumThreads()
	if err != nil {
		return nil, fmt.Errorf("cannot get thread count: %w", err)
	}

	memInfo, err := p.MemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("cannot get memory info: %w", err)
	}

	memPercent, _ := p.MemoryPercent()
	cpuPercent, _ := p.CPUPercent()

	exePath, err := p.Exe()
	if err != nil {
		return nil, fmt.Errorf("cannot get binary path: %w", err)
	}

	info, err := os.Stat(exePath)
	if err != nil {
		return nil, fmt.Errorf("cannot stat binary: %w", err)
	}

	return &PsStat{
		CPUTime:       cpuTimes.Total(),
		CPUNumThreads: int(numThreads),
		MemRSS:        int64(memInfo.RSS),
		BinarySize:    info.Size(),
		CPUPercent:    cpuPercent,
		MemPercent:    memPercent,
	}, nil
}
