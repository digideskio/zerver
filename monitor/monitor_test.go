package pprof

import "testing"

func TestMonitor(t *testing.T) {
	s, e := NewMonitorServer("/")
	if e == nil {
		s.Start(nil)
	}
}
