package main

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

const (
	// 性能信息收集间隔
	MachineStatsColInterval = 5 * time.Second
)

type MachineStats struct {
	Time   time.Time
	CPU    float64
	Load   float64
	Memory float64
	sync.RWMutex
}

func NewMachineStats() *MachineStats {
	return &MachineStats{
		Time: time.Now(),
	}
}

func (s *MachineStats) Run() {
	for {
		s.Lock()

		// CPU
		cpuPer, _ := cpu.Percent(0, false)
		s.CPU = cpuPer[0]
		// Load
		loadAvg, _ := load.Avg()
		s.Load = loadAvg.Load1
		// Memory
		memData, _ := mem.VirtualMemory()
		s.Memory = memData.UsedPercent

		s.Unlock()

		time.Sleep(MachineStatsColInterval)
	}
}
