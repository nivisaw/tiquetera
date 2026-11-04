package platform

import "time"

// Clock provides the current time; inject a fake one in tests.
type Clock interface {
	Now() time.Time
}

// SystemClock is the real-time Clock.
type SystemClock struct{}

// Now returns the current time.
func (SystemClock) Now() time.Time { return time.Now() }
