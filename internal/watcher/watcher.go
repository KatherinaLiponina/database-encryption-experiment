package watcher

import (
	"runtime"
	"time"
)

// watcher checks time and memory spend on execution
type Report interface {
	Time() time.Duration
	Memory() uint64
}

type Watcher interface {
	Start()
	Stop() Report
}

type report struct {
	timeSpent        time.Duration
	allocationsCount uint64
	allocationsSum   uint64
}

func (r *report) Time() time.Duration {
	return r.timeSpent
}
func (r *report) Memory() uint64 {
	return r.allocationsSum
}

type watcher struct {
	startTime time.Time
	startMem  runtime.MemStats
}

func New() Watcher {
	return &watcher{}
}

func (w *watcher) Start() {
	runtime.ReadMemStats(&w.startMem)
	w.startTime = time.Now()
}
func (w *watcher) Stop() Report {
	endTime := time.Now()
	var endMem runtime.MemStats
	runtime.ReadMemStats(&endMem)
	return &report{timeSpent: endTime.Sub(w.startTime),
		allocationsCount: endMem.Mallocs - w.startMem.Mallocs,
		allocationsSum:   endMem.TotalAlloc - w.startMem.TotalAlloc}
}
