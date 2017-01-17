package util

import (
	"time"
)

type Sleeper interface {
	Sleep(time.Duration)
}

type RealSleeper struct{}

func (s RealSleeper) Sleep(d time.Duration) {
	time.Sleep(d)
}

type RecordingNoopSleeper struct {
	sleptTimes []time.Duration
}

func NewRecordingNoopSleeper() *RecordingNoopSleeper {
	return &RecordingNoopSleeper{}
}

func (s *RecordingNoopSleeper) Sleep(d time.Duration) {
	s.sleptTimes = append(s.sleptTimes, d)
}

func (s *RecordingNoopSleeper) SleptTimes() []time.Duration {
	return s.sleptTimes
}
