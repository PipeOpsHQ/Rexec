package handlers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

var processStartTime = time.Now()

// GetRuntimeStats returns basic Go runtime memory/goroutine stats for debugging.
// This should only be exposed to admins.
func GetRuntimeStats(c *gin.Context) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	lastGC := ""
	if ms.LastGC != 0 {
		lastGC = time.Unix(0, int64(ms.LastGC)).UTC().Format(time.RFC3339Nano)
	}

	c.JSON(http.StatusOK, gin.H{
		"time":       time.Now().UTC().Format(time.RFC3339Nano),
		"uptime_sec": time.Since(processStartTime).Seconds(),
		"go": gin.H{
			"version":    runtime.Version(),
			"gomaxprocs": runtime.GOMAXPROCS(0),
			"goroutines": runtime.NumGoroutine(),
		},
		"mem": gin.H{
			"alloc_bytes":         ms.Alloc,
			"total_alloc_bytes":   ms.TotalAlloc,
			"sys_bytes":           ms.Sys,
			"heap_alloc_bytes":    ms.HeapAlloc,
			"heap_inuse_bytes":    ms.HeapInuse,
			"heap_idle_bytes":     ms.HeapIdle,
			"heap_released_bytes": ms.HeapReleased,
			"stack_inuse_bytes":   ms.StackInuse,
			"objects":             ms.Mallocs - ms.Frees,
			"gc": gin.H{
				"num":               ms.NumGC,
				"pause_total_ns":    ms.PauseTotalNs,
				"next_gc_bytes":     ms.NextGC,
				"gc_cpu_fraction":   ms.GCCPUFraction,
				"last_gc_rfc3339":   lastGC,
				"last_gc_unix_nano": ms.LastGC,
			},
		},
		"env": gin.H{
			"gogc":       os.Getenv("GOGC"),
			"gomemlimit": os.Getenv("GOMEMLIMIT"),
			"godebug":    os.Getenv("GODEBUG"),
		},
	})
}
