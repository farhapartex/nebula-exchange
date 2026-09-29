package outbox

import "time"

const (
	firstRetryDelay   = 30 * time.Second
	maximumRetryDelay = time.Hour
)

func RetryDelay(failedAttemptCount int) time.Duration {
	if failedAttemptCount < 1 {
		return firstRetryDelay
	}
	retryDelay := firstRetryDelay
	for attemptIndex := 1; attemptIndex < failedAttemptCount; attemptIndex++ {
		retryDelay *= 2
		if retryDelay >= maximumRetryDelay {
			return maximumRetryDelay
		}
	}
	return retryDelay
}
