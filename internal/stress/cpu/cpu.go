package cpu

import (
	"errors"
	"sync"
	"time"
)

// Tester represents a CPU stress tester
type Tester struct {
	duration time.Duration
	threads  int
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewTester creates a new CPU tester with defaults
func NewTester() *Tester {
	return &Tester{
		stopCh: make(chan struct{}),
	}
}

// Run starts the CPU stress test
func (t *Tester) Run(duration time.Duration, threads int, progress func(float64), reportError func(error)) error {
	if threads < 1 {
		return errors.New("threads must be >= 1")
	}
	t.duration = duration
	t.threads = threads
	t.wg.Add(threads)

	stop := time.After(duration)
	for i := 0; i < threads; i++ {
		go func(id int) {
			defer t.wg.Done()
			count := 0
			for {
				select {
				case <-stop:
					return
				case <-t.stopCh:
					return
				default:
					// Example workload: integer math
					x := id
					for i := 0; i < 1000000; i++ {
						x = x*i ^ (x >> 1)
					}
					count++
					progress(float64(count))
				}
			}
		}(i)
	}

	// Wait for completion or stop
	t.wg.Wait()
	return nil
}

// Stop signals the tester to stop early
func (t *Tester) Stop() {
	close(t.stopCh)
}
