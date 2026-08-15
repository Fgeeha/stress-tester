package stress

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"stress_tester/logger"

	"github.com/shirou/gopsutil/v3/mem"
)

type Config struct {
	MemoryMB    int
	NumThreads  int
	EnableCache bool
	UseFPU      bool
	StopOnError bool
	RNGCustom   bool
}

type Runner struct {
	Events  chan logger.Event
	OnAbort func()
}

func (r Runner) log(message string, eventType int) { logger.LogEvent(message, eventType, r.Events) }
func (r Runner) progress(coverage float64)         { logger.LogProgress(coverage, r.Events) }
func (r Runner) abortIfNeeded(config Config) {
	if config.StopOnError && r.OnAbort != nil {
		r.OnAbort()
	}
}

func (r Runner) RunMemory(ctx context.Context, config Config) {
	memPerThread := config.MemoryMB / config.NumThreads
	if memPerThread < 1 {
		memPerThread = 1
	}

	r.log("Starting memory test with configuration:", logger.EventLog)
	r.log(fmt.Sprintf("- Total memory to allocate: %d MB", config.MemoryMB), logger.EventLog)
	r.log(fmt.Sprintf("- Memory per thread: %d MB", memPerThread), logger.EventLog)
	r.log(fmt.Sprintf("- Number of threads: %d", config.NumThreads), logger.EventLog)
	r.log("- Algorithm: Random memory access with pattern writing", logger.EventLog)

	var memWg sync.WaitGroup
	memWg.Add(config.NumThreads)
	memoryBlocks := make([][]byte, config.NumThreads)
	allocDone := make(chan struct{}, config.NumThreads)

	for i := 0; i < config.NumThreads; i++ {
		go func(threadID int) {
			defer memWg.Done()
			memSize := memPerThread * 1024 * 1024
			r.log(fmt.Sprintf("Thread %d: Attempting to allocate %d MB", threadID, memPerThread), logger.EventLog)

			allocSucceeded := false
			defer func() {
				if !allocSucceeded {
					allocDone <- struct{}{}
				}
				if rec := recover(); rec != nil {
					r.log(fmt.Sprintf("Thread %d: Memory allocation failed: %v", threadID, rec), logger.EventError)
					r.abortIfNeeded(config)
				}
			}()

			memory := make([]byte, memSize)
			allocSucceeded = true
			memoryBlocks[threadID] = memory

			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(threadID)))
			pattern := make([]byte, 1024)
			for i := range pattern {
				pattern[i] = byte(rng.Intn(256))
			}
			for i := 0; i < len(memory); i += len(pattern) {
				copy(memory[i:], pattern)
			}

			r.log(fmt.Sprintf("Thread %d: Successfully allocated and filled %d MB with pattern data", threadID, memPerThread), logger.EventLog)
			allocDone <- struct{}{}

			for {
				select {
				case <-ctx.Done():
					r.log(fmt.Sprintf("Thread %d: Stopping and cleaning up memory", threadID), logger.EventLog)
					memory = nil
					memoryBlocks[threadID] = nil
					runtime.GC()
					return
				default:
					for j := 0; j < 1000; j++ {
						idx := rng.Intn(len(memory) - len(pattern))
						copy(memory[idx:], pattern)
						for k := 0; k < len(pattern); k++ {
							if memory[idx+k] != pattern[k] {
								r.log(fmt.Sprintf("Thread %d: Memory verification failed at offset %d", threadID, idx+k), logger.EventError)
							}
						}
					}

					vmStat, err := mem.VirtualMemory()
					if err != nil {
						r.log(fmt.Sprintf("Thread %d: Error checking memory: %v", threadID, err), logger.EventError)
						continue
					}
					usedPercent := 100.0 - (float64(vmStat.Available)/float64(vmStat.Total))*100.0
					r.progress(usedPercent)

					swapStat, err := mem.SwapMemory()
					if err != nil {
						r.log(fmt.Sprintf("Thread %d: Error checking swap memory: %v", threadID, err), logger.EventError)
					}
					if vmStat.Available < 10*1024*1024 && swapStat.Free < 10*1024*1024 {
						r.log(fmt.Sprintf("Thread %d: Memory critically low, stopping test", threadID), logger.EventError)
						if r.OnAbort != nil {
							r.OnAbort()
						}
						return
					}

					r.log(fmt.Sprintf("Memory coverage: %.2f%% (Used: %d MB, Available: %d MB, Total: %d MB) | Swap: Used: %d MB (%.2f%%), Free: %d MB",
						usedPercent, vmStat.Used/(1024*1024), vmStat.Available/(1024*1024), vmStat.Total/(1024*1024), swapStat.Used/(1024*1024), swapStat.UsedPercent, swapStat.Free/(1024*1024)), logger.EventLog)
					time.Sleep(50 * time.Millisecond)
				}
			}
		}(i)
	}

	for i := 0; i < config.NumThreads; i++ {
		<-allocDone
	}
	<-ctx.Done()
	memWg.Wait()
	for i := range memoryBlocks {
		memoryBlocks[i] = nil
	}
	for i := 0; i < 3; i++ {
		runtime.GC()
		debug.FreeOSMemory()
		time.Sleep(100 * time.Millisecond)
	}
	r.log("Memory test completed and cleaned up", logger.EventLog)
}

func (r Runner) RunCPU(ctx context.Context, config Config) {
	r.log("Starting CPU test with configuration:", logger.EventLog)
	r.log(fmt.Sprintf("- Number of threads: %d", config.NumThreads), logger.EventLog)
	r.log(fmt.Sprintf("- Cache stress: %v", config.EnableCache), logger.EventLog)
	r.log(fmt.Sprintf("- FPU stress: %v", config.UseFPU), logger.EventLog)
	r.log("- Algorithm: Mixed integer and floating-point calculations", logger.EventLog)

	var cpuWg sync.WaitGroup
	cpuWg.Add(config.NumThreads)
	for i := 0; i < config.NumThreads; i++ {
		go func(threadID int) {
			defer cpuWg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(threadID)))
			r.log(fmt.Sprintf("Thread %d: Starting CPU stress test", threadID), logger.EventLog)
			for {
				select {
				case <-ctx.Done():
					r.log(fmt.Sprintf("Thread %d: Stopping CPU test", threadID), logger.EventLog)
					return
				default:
					if config.EnableCache {
						var sum int64
						for j := 0; j < 1000000; j++ {
							sum += int64(rng.Intn(1000))
						}
						if sum == 0 {
							r.log(fmt.Sprintf("Thread %d: Cache test iteration completed", threadID), logger.EventLog)
						}
					}
					if config.UseFPU {
						var result float64
						for j := 0; j < 1000000; j++ {
							result += rng.Float64() * rng.Float64()
						}
						if result == 0 {
							r.log(fmt.Sprintf("Thread %d: FPU test iteration completed", threadID), logger.EventLog)
						}
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}
	cpuWg.Wait()
	r.log("CPU test completed", logger.EventLog)
}
