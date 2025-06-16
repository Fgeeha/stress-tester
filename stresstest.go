package main

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

type TestConfig struct {
	MemoryMB    int
	NumThreads  int
	EnableCache bool
	UseFPU      bool
	StopOnError bool
	RNGCustom   bool
}

func runMemoryTest(ctx context.Context, config TestConfig) {
	memPerThread := config.MemoryMB / config.NumThreads
	if memPerThread < 1 {
		memPerThread = 1
	}

	// Log test configuration
	logger.LogEvent("Starting memory test with configuration:", logger.EventLog, eventsChan)
	logger.LogEvent(fmt.Sprintf("- Total memory to allocate: %d MB", config.MemoryMB), logger.EventLog, eventsChan)
	logger.LogEvent(fmt.Sprintf("- Memory per thread: %d MB", memPerThread), logger.EventLog, eventsChan)
	logger.LogEvent(fmt.Sprintf("- Number of threads: %d", config.NumThreads), logger.EventLog, eventsChan)
	logger.LogEvent("- Algorithm: Random memory access with pattern writing", logger.EventLog, eventsChan)

	// Create a wait group for memory test threads
	var memWg sync.WaitGroup
	memWg.Add(config.NumThreads)

	// Create a slice to hold memory allocations for cleanup
	memoryBlocks := make([][]byte, config.NumThreads)

	// Create a channel to signal memory allocation completion
	allocDone := make(chan struct{})

	// Start memory test threads
	for i := 0; i < config.NumThreads; i++ {
		go func(threadID int) {
			defer memWg.Done()

			// Allocate memory for this thread
			memSize := memPerThread * 1024 * 1024 // Convert MB to bytes

			// Log memory allocation attempt
			logger.LogEvent(fmt.Sprintf("Thread %d: Attempting to allocate %d MB", threadID, memPerThread), logger.EventLog, eventsChan)

			allocSucceeded := false
			// Try to allocate memory with panic recovery
			defer func() {
				if !allocSucceeded {
					// signal allocation attempt even if failed
					allocDone <- struct{}{}
				}
				if r := recover(); r != nil {
					logger.LogEvent(fmt.Sprintf("Thread %d: Memory allocation failed: %v", threadID, r), logger.EventError, eventsChan)
					if config.StopOnError && currentCancel != nil {
						currentCancel()
					}
				}
			}()

			memory := make([]byte, memSize)
			allocSucceeded = true
			memoryBlocks[threadID] = memory // Store reference for cleanup

			// Fill memory with pattern data
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(threadID)))
			pattern := make([]byte, 1024) // 1KB pattern
			for i := range pattern {
				pattern[i] = byte(r.Intn(256))
			}

			// Fill memory with pattern
			for i := 0; i < len(memory); i += len(pattern) {
				copy(memory[i:], pattern)
			}

			logger.LogEvent(fmt.Sprintf("Thread %d: Successfully allocated and filled %d MB with pattern data", threadID, memPerThread), logger.EventLog, eventsChan)

			// Signal that allocation is complete
			allocDone <- struct{}{}

			// Continuously read and write to memory
			for {
				select {
				case <-ctx.Done():
					logger.LogEvent(fmt.Sprintf("Thread %d: Stopping and cleaning up memory", threadID), logger.EventLog, eventsChan)
					// Clear memory when stopped
					memory = nil
					memoryBlocks[threadID] = nil
					runtime.GC() // Force GC for this thread
					return
				default:
					// Read and write to random locations with pattern
					for j := 0; j < 1000; j++ {
						idx := r.Intn(len(memory) - len(pattern))
						// Write pattern
						copy(memory[idx:], pattern)
						// Read and verify pattern
						for k := 0; k < len(pattern); k++ {
							if memory[idx+k] != pattern[k] {
								logger.LogEvent(fmt.Sprintf("Thread %d: Memory verification failed at offset %d", threadID, idx+k), logger.EventError, eventsChan)
							}
						}
					}

					// Check memory usage
					vmStat, err := mem.VirtualMemory()
					if err != nil {
						logger.LogEvent(fmt.Sprintf("Thread %d: Error checking memory: %v", threadID, err), logger.EventError, eventsChan)
						continue
					}

					// Calculate and log memory coverage
					usedPercent := 100.0 - (float64(vmStat.Available) / float64(vmStat.Total) * 100.0)
					logger.LogProgress(usedPercent, eventsChan)

					// Get swap memory information
					swapStat, err := mem.SwapMemory()
					if err != nil {
						logger.LogEvent(fmt.Sprintf("Thread %d: Error checking swap memory: %v", threadID, err), logger.EventError, eventsChan)
					}

					// Abort if both RAM and swap are critically low
					if vmStat.Available < 10*1024*1024 && swapStat.Free < 10*1024*1024 {
						logger.LogEvent(fmt.Sprintf("Thread %d: Memory critically low, stopping test", threadID), logger.EventError, eventsChan)
						if currentCancel != nil {
							currentCancel()
						}
						return
					}

					// Log detailed memory coverage information with formula explanation
					logger.LogEvent(fmt.Sprintf("Memory coverage: %.2f%% (Used: %d MB, Available: %d MB, Total: %d MB) [Formula: 100 - (Available/Total * 100)] | Swap: Used: %d MB (%.2f%%), Free: %d MB",
						usedPercent,
						vmStat.Used/(1024*1024),
						vmStat.Available/(1024*1024),
						vmStat.Total/(1024*1024),
						swapStat.Used/(1024*1024),
						swapStat.UsedPercent,
						swapStat.Free/(1024*1024)),
						logger.EventLog, eventsChan)

					time.Sleep(50 * time.Millisecond)
				}
			}
		}(i)
	}

	// Wait for all threads to complete allocation
	for i := 0; i < config.NumThreads; i++ {
		<-allocDone
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Wait for all memory test threads to complete
	memWg.Wait()

	// Clear all memory blocks after test completion
	for i := range memoryBlocks {
		if memoryBlocks[i] != nil {
			memoryBlocks[i] = nil
		}
	}

	// Force garbage collection multiple times to ensure memory is freed
	for i := 0; i < 3; i++ {
		runtime.GC()
		debug.FreeOSMemory()
		time.Sleep(100 * time.Millisecond) // Give GC time to work
	}

	logger.LogEvent("Memory test completed and cleaned up", logger.EventLog, eventsChan)
}

func runCPUTest(ctx context.Context, config TestConfig) {
	logger.LogEvent("Starting CPU test with configuration:", logger.EventLog, eventsChan)
	logger.LogEvent(fmt.Sprintf("- Number of threads: %d", config.NumThreads), logger.EventLog, eventsChan)
	logger.LogEvent(fmt.Sprintf("- Cache stress: %v", config.EnableCache), logger.EventLog, eventsChan)
	logger.LogEvent(fmt.Sprintf("- FPU stress: %v", config.UseFPU), logger.EventLog, eventsChan)
	logger.LogEvent("- Algorithm: Mixed integer and floating-point calculations", logger.EventLog, eventsChan)

	// Create a wait group for CPU test threads
	var cpuWg sync.WaitGroup
	cpuWg.Add(config.NumThreads)

	// Start CPU test threads
	for i := 0; i < config.NumThreads; i++ {
		go func(threadID int) {
			defer cpuWg.Done()

			// Create a random number generator for this thread
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(threadID)))

			logger.LogEvent(fmt.Sprintf("Thread %d: Starting CPU stress test", threadID), logger.EventLog, eventsChan)

			// CPU stress test loop
			for {
				select {
				case <-ctx.Done():
					logger.LogEvent(fmt.Sprintf("Thread %d: Stopping CPU test", threadID), logger.EventLog, eventsChan)
					return
				default:
					// Perform CPU-intensive calculations
					if config.EnableCache {
						// Cache stress test
						var sum int64
						for j := 0; j < 1000000; j++ {
							sum += int64(r.Intn(1000))
						}
						if sum == 0 {
							logger.LogEvent(fmt.Sprintf("Thread %d: Cache test iteration completed", threadID), logger.EventLog, eventsChan)
						}
					}

					if config.UseFPU {
						var result float64
						for j := 0; j < 1000000; j++ {
							result += float64(r.Float64() * r.Float64())
						}
						if result == 0 {
							logger.LogEvent(fmt.Sprintf("Thread %d: FPU test iteration completed", threadID), logger.EventLog, eventsChan)
						}
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	cpuWg.Wait()
	logger.LogEvent("CPU test completed", logger.EventLog, eventsChan)
}
