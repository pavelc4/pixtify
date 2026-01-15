package handler

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pavelc4/pixtify/internal/storage"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type HealthHandler struct {
	StartTime time.Time
	DB        *sql.DB
	Storage   storage.Service
}

func NewHealthHandler(startTime time.Time, db *sql.DB, storage storage.Service) *HealthHandler {
	return &HealthHandler{
		StartTime: startTime,
		DB:        db,
		Storage:   storage,
	}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	uptime := time.Since(h.StartTime)
	uptimeStr := fmt.Sprintf("%dd %02dh %02dm %02ds",
		int(uptime.Hours())/24, int(uptime.Hours())%24,
		int(uptime.Minutes())%60, int(uptime.Seconds())%60)

	cpuUsage := 0.0
	cpuCores := int32(0)
	if cpuPerc, err := cpu.Percent(200*time.Millisecond, false); err == nil && len(cpuPerc) > 0 {
		cpuUsage = cpuPerc[0]
	}
	if cores, err := cpu.Counts(true); err == nil {
		cpuCores = int32(cores)
	}

	vmem := &mem.VirtualMemoryStat{}
	if v, err := mem.VirtualMemory(); err == nil {
		vmem = v
	}

	procCpu := 0.0
	procMemMB := 0.0
	if proc, err := process.NewProcess(int32(os.Getpid())); err == nil {
		if cpu, err := proc.CPUPercent(); err == nil {
			procCpu = cpu
		}
		if memInfo, err := proc.MemoryInfo(); err == nil {
			procMemMB = float64(memInfo.RSS) / 1024 / 1024
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	heapSysMB := float64(m.HeapSys) / 1024 / 1024
	heapAllocMB := float64(m.HeapAlloc) / 1024 / 1024
	gcPauseTotalNs := m.PauseTotalNs
	gcPauseMeanNs := float64(gcPauseTotalNs) / float64(m.NumGC)
	if m.NumGC == 0 {
		gcPauseMeanNs = 0
	}

	// Dependencies
	dbStatus := "healthy"
	if err := h.DB.PingContext(ctx); err != nil {
		dbStatus = "unhealthy"
	}
	storageStatus := "healthy"
	if h.Storage == nil {
		storageStatus = "unconfigured"
	}

	// Helpers
	toGB := func(b uint64) float64 { return float64(b) / 1024 / 1024 / 1024 }
	toMB := func(b uint64) float64 { return float64(b) / 1024 / 1024 }

	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    uptimeStr,

		"cpu": fiber.Map{
			"cores":         cpuCores,
			"usage_percent": cpuUsage,
			"usage":         fmt.Sprintf("%.2f%%", cpuUsage),
		},

		"memory": fiber.Map{
			"system": fiber.Map{
				"total_gb":     toGB(vmem.Total),
				"used_gb":      toGB(vmem.Used),
				"free_gb":      toGB(vmem.Free),
				"available_gb": toGB(vmem.Available),
				"used_percent": vmem.UsedPercent,
				"active_gb":    toGB(vmem.Active),
				"inactive_gb":  toGB(vmem.Inactive),
				"buffers_mb":   toMB(vmem.Buffers),
				"cached_mb":    toMB(vmem.Cached),
			},
			"process": fiber.Map{
				"rss_mb": procMemMB,
			},
		},

		"go_runtime": fiber.Map{
			"version":    runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"heap": fiber.Map{
				"sys_mb":        heapSysMB,
				"alloc_live_mb": heapAllocMB,
				"idle_mb":       float64(m.HeapIdle) / 1024 / 1024,
				"released_mb":   float64(m.HeapReleased) / 1024 / 1024,
				"objects":       int64(m.HeapObjects),
			},
			"gc": fiber.Map{
				"runs":           int64(m.NumGC),
				"pause_total_ns": int64(gcPauseTotalNs),
				"pause_mean_ns":  int64(gcPauseMeanNs),
			},
		},

		"process": fiber.Map{
			"pid":         os.Getpid(),
			"cpu_percent": procCpu,
			"cpu_usage":   fmt.Sprintf("%.2f%%", procCpu),
			"memory_mb":   procMemMB,
		},

		"dependencies": fiber.Map{
			"database": dbStatus,
			"storage":  storageStatus,
		},
	})
}
