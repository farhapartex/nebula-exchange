package outbox

import (
	"testing"
	"time"
)

func TestRetryDelayDoublesUpToOneHour(t *testing.T) {
	expectedDelays := map[int]time.Duration{
		1:  30 * time.Second,
		2:  time.Minute,
		3:  2 * time.Minute,
		4:  4 * time.Minute,
		7:  32 * time.Minute,
		8:  time.Hour,
		10: time.Hour,
	}
	for failedAttemptCount, expectedDelay := range expectedDelays {
		if actualDelay := RetryDelay(failedAttemptCount); actualDelay != expectedDelay {
			t.Fatalf("attempt %d: got %v, want %v", failedAttemptCount, actualDelay, expectedDelay)
		}
	}
}
