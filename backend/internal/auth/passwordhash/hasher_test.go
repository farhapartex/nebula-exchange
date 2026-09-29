package passwordhash

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHasherNeverExceedsTheConcurrencyLimit(t *testing.T) {
	hasher := NewHasher(HasherOptions{Parameters: fastTestParameters, MaximumConcurrent: 2, MaximumQueueWait: 5 * time.Second})
	var activeHashes, peakActiveHashes atomic.Int32
	var waitGroup sync.WaitGroup

	for hashIndex := 0; hashIndex < 12; hashIndex++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if err := hasher.acquireSlot(context.Background()); err != nil {
				t.Errorf("acquire slot: %v", err)
				return
			}
			currentlyActive := activeHashes.Add(1)
			for {
				peak := peakActiveHashes.Load()
				if currentlyActive <= peak || peakActiveHashes.CompareAndSwap(peak, currentlyActive) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			activeHashes.Add(-1)
			hasher.releaseSlot()
		}()
	}
	waitGroup.Wait()

	if peakActiveHashes.Load() > 2 {
		t.Fatalf("peak concurrency was %d, limit is 2", peakActiveHashes.Load())
	}
}

func TestHasherReportsBusyWhenTheQueueWaitRunsOut(t *testing.T) {
	hasher := NewHasher(HasherOptions{Parameters: fastTestParameters, MaximumConcurrent: 1, MaximumQueueWait: 20 * time.Millisecond})
	if err := hasher.acquireSlot(context.Background()); err != nil {
		t.Fatalf("take the only slot: %v", err)
	}
	defer hasher.releaseSlot()

	_, err := hasher.Hash(context.Background(), "Mining4Crystal!Moon")
	if !errors.Is(err, ErrHasherBusy) {
		t.Fatalf("got %v, want ErrHasherBusy", err)
	}
}

func TestHasherHashesAndVerifies(t *testing.T) {
	hasher := NewHasher(HasherOptions{Parameters: fastTestParameters})
	encodedHash, err := hasher.Hash(context.Background(), "Mining4Crystal!Moon")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if isMatch, err := hasher.Verify(context.Background(), "Mining4Crystal!Moon", encodedHash); err != nil || !isMatch {
		t.Fatalf("verify: %v, %v", isMatch, err)
	}
}
