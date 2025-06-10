package mem

import (
	"errors"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

// Pattern defines memory test pattern
type Pattern int

const (
	PatternAA Pattern = iota
	Pattern55
	PatternRandom
)

// Tester represents a memory stress tester
type Tester struct {
	sizeMB   int
	duration time.Duration
	pattern  Pattern
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewTester creates a new memory tester
func NewTester() *Tester {
	return &Tester{stopCh: make(chan struct{})}
}

// Run starts the memory stress test
func (t *Tester) Run(sizeMB int, pattern Pattern, duration time.Duration, progress func(float64), reportError func(error)) error {
	if sizeMB < 1 {
		return errors.New("sizeMB must be >= 1")
	}

	t.sizeMB = sizeMB
	t.pattern = pattern
	t.duration = duration
	t.wg.Add(1)

	// Allocate buffer
	buf := make([]byte, sizeMB*1024*1024)

	go func() {
		defer t.wg.Done()
		rand.Seed(time.Now().UnixNano())
		stop := time.After(duration)
		iterations := 0
		for {
			select {
			case <-stop:
				return
			case <-t.stopCh:
				return
			default:
				// Write pattern
				for i := range buf {
					switch t.pattern {
					case PatternAA:
						buf[i] = 0xAA
					case Pattern55:
						buf[i] = 0x55
					case PatternRandom:
						buf[i] = byte(rand.Intn(256))
					}
				}
				// Read-back check
				sum := 0
				ptr := unsafe.Pointer(&buf[0])
				for i := 0; i < len(buf); i++ {
					sum += int(*(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(i))))
				}
				if sum == 0 && t.pattern != PatternRandom {
					if reportError != nil {
						reportError(errors.New("memory read/write error"))
					}
				}
				iterations++
				if progress != nil {
					progress(float64(iterations))
				}
			}
		}
	}()

	// Wait completion
	t.wg.Wait()
	return nil
}

// Stop signals the tester to stop early
func (t *Tester) Stop() {
	close(t.stopCh)
}
