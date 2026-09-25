package ports

import "time"

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

func NewSystemClock() Clock {
	return SystemClock{}
}
